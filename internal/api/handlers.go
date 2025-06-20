package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mm-rules/matchmaking/internal/allocation"
	"github.com/mm-rules/matchmaking/internal/engine"
	"github.com/mm-rules/matchmaking/internal/matchmaker"
	"github.com/mm-rules/matchmaking/internal/models"
	"github.com/mm-rules/matchmaking/internal/storage"
	"github.com/sirupsen/logrus"
)

// Handler handles HTTP requests for the matchmaking API
type Handler struct {
	storage    *storage.RedisStorage
	matchmaker *matchmaker.Matchmaker
	ruleEngine *engine.RuleEngine
	allocator  allocation.Allocator
	logger     *logrus.Logger
}

// NewHandler creates a new API handler
func NewHandler(storage *storage.RedisStorage, allocator allocation.Allocator, logger *logrus.Logger) *Handler {
	return &Handler{
		storage:    storage,
		matchmaker: matchmaker.NewMatchmaker(),
		ruleEngine: engine.NewRuleEngine(),
		allocator:  allocator,
		logger:     logger,
	}
}

// MatchRequestRequest represents the request body for creating a match request
type MatchRequestRequest struct {
	PlayerID string                 `json:"player_id" binding:"required"`
	Metadata map[string]interface{} `json:"metadata"`
	GameID   string                 `json:"game_id" binding:"required"`
}

// CreateMatchRequest handles POST /match-request
func (h *Handler) CreateMatchRequest(c *gin.Context) {
	var req MatchRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create match request
	matchRequest := models.NewMatchRequest(req.PlayerID, req.GameID, req.Metadata)

	// Store in Redis
	ctx := c.Request.Context()
	if err := h.storage.StoreMatchRequest(ctx, matchRequest); err != nil {
		h.logger.WithError(err).Error("Failed to store match request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create match request"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": matchRequest.ID,
		"player_id":  matchRequest.PlayerID,
		"game_id":    matchRequest.GameID,
	}).Info("Created match request")

	c.JSON(http.StatusCreated, gin.H{
		"request_id": matchRequest.ID,
		"status":     matchRequest.Status,
	})
}

// CreateGameConfig handles POST /rules/:game_id
func (h *Handler) CreateGameConfig(c *gin.Context) {
	gameID := c.Param("game_id")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game_id is required"})
		return
	}

	var config models.GameConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set the game ID from the URL parameter
	config.GameID = gameID

	// Validate the configuration
	if err := h.ruleEngine.ValidateGameConfig(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Store in Redis
	ctx := c.Request.Context()
	if err := h.storage.StoreGameConfig(ctx, &config); err != nil {
		h.logger.WithError(err).Error("Failed to store game config")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store game configuration"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"game_id": config.GameID,
		"teams":   len(config.Teams),
		"rules":   len(config.Rules),
	}).Info("Created game configuration")

	c.JSON(http.StatusCreated, gin.H{
		"game_id": config.GameID,
		"message": "Game configuration created successfully",
	})
}

// GetMatchStatus handles GET /match-status/:request_id
func (h *Handler) GetMatchStatus(c *gin.Context) {
	requestID := c.Param("request_id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request_id is required"})
		return
	}

	ctx := c.Request.Context()

	// Try to get cached status first
	status, err := h.storage.GetMatchStatus(ctx, requestID)
	if err == nil {
		c.JSON(http.StatusOK, status)
		return
	}

	// If no cached status, get the match request
	request, err := h.storage.GetMatchRequest(ctx, requestID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match request not found"})
		return
	}

	// Create status response
	statusResponse := &models.MatchStatusResponse{
		Status: request.Status,
	}

	// If matched, try to find the match and session info
	if request.Status == models.StatusMatched || request.Status == models.StatusAllocated {
		// This is a simplified approach - in a real implementation you'd need to
		// track which match a request belongs to
		h.logger.WithField("request_id", requestID).Warn("Match found but session info not available")
	}

	c.JSON(http.StatusOK, statusResponse)
}

// ProcessMatchmaking handles POST /process-matchmaking/:game_id
func (h *Handler) ProcessMatchmaking(c *gin.Context) {
	gameID := c.Param("game_id")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game_id is required"})
		return
	}

	ctx := c.Request.Context()

	// Get game configuration
	config, err := h.storage.GetGameConfig(ctx, gameID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game configuration not found"})
		return
	}

	// Get pending match requests
	requests, err := h.storage.GetGameQueue(ctx, gameID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get game queue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get match requests"})
		return
	}

	if len(requests) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No pending match requests",
			"matches": []interface{}{},
		})
		return
	}

	// Process matchmaking
	matches := h.matchmaker.ProcessMatchPool(requests, config)

	if len(matches) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No matches could be formed",
			"matches": []interface{}{},
		})
		return
	}

	// Store matches and update request statuses
	var matchResults []gin.H
	for _, match := range matches {
		// Store the match
		if err := h.storage.StoreMatch(ctx, match); err != nil {
			h.logger.WithError(err).Error("Failed to store match")
			continue
		}

		// Update request statuses
		for _, playerID := range match.Players {
			// Find the request for this player
			for _, request := range requests {
				if request.PlayerID == playerID {
					// Update status to matched
					if err := h.storage.UpdateMatchRequestStatus(ctx, request.ID, models.StatusMatched); err != nil {
						h.logger.WithError(err).Error("Failed to update request status")
					}

					// Store match status for this request
					status := &models.MatchStatusResponse{
						Status: models.StatusMatched,
						Team:   &match.TeamName,
					}
					if err := h.storage.StoreMatchStatus(ctx, request.ID, status); err != nil {
						h.logger.WithError(err).Error("Failed to store match status")
					}

					// Remove from queue
					if err := h.storage.RemoveFromQueue(ctx, gameID, request.ID); err != nil {
						h.logger.WithError(err).Error("Failed to remove from queue")
					}
					break
				}
			}
		}

		matchResults = append(matchResults, gin.H{
			"match_id":   match.ID,
			"team_name":  match.TeamName,
			"players":    match.Players,
			"created_at": match.CreatedAt,
		})
	}

	h.logger.WithFields(logrus.Fields{
		"game_id": gameID,
		"matches": len(matches),
	}).Info("Processed matchmaking")

	c.JSON(http.StatusOK, gin.H{
		"message": "Matchmaking processed successfully",
		"matches": matchResults,
	})
}

// AllocateSessions handles POST /allocate-sessions/:game_id
func (h *Handler) AllocateSessions(c *gin.Context) {
	gameID := c.Param("game_id")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game_id is required"})
		return
	}

	ctx := c.Request.Context()

	// Get all matches for this game (this is simplified - you'd need to track matches by game)
	// For now, we'll just return a message indicating this endpoint needs implementation
	c.JSON(http.StatusOK, gin.H{
		"message": "Session allocation endpoint - implementation needed",
		"game_id": gameID,
	})
}

// GetStats handles GET /stats
func (h *Handler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Get storage stats
	storageStats, err := h.storage.GetStats(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get storage stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"storage": storageStats,
		"timestamp": time.Now(),
	})
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	// Check Redis connection
	if err := h.storage.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "Redis connection failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"timestamp": time.Now(),
	})
} 