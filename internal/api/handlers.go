package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mm-rules/matchmaking/internal/allocation"
	"github.com/mm-rules/matchmaking/internal/engine"
	"github.com/mm-rules/matchmaking/internal/matchmaker"
	"github.com/mm-rules/matchmaking/internal/metrics"
	"github.com/mm-rules/matchmaking/internal/models"
	"github.com/mm-rules/matchmaking/internal/storage"
	"github.com/sirupsen/logrus"
)

// Handler handles HTTP requests for the matchmaking API
type Handler struct {
	storage    storage.Storage
	matchmaker *matchmaker.Matchmaker
	ruleEngine *engine.RuleEngine
	allocator  allocation.Allocator
	logger     *logrus.Logger
}

// NewHandler creates a new API handler
func NewHandler(storage storage.Storage, allocator allocation.Allocator, logger *logrus.Logger) *Handler {
	handler := &Handler{
		storage:    storage,
		matchmaker: matchmaker.NewMatchmaker(),
		ruleEngine: engine.NewRuleEngine(),
		allocator:  allocator,
		logger:     logger,
	}

	// Start background cleanup routine
	go handler.startBackgroundCleanup()

	return handler
}

// startBackgroundCleanup runs a periodic cleanup of expired requests
func (h *Handler) startBackgroundCleanup() {
	ticker := time.NewTicker(30 * time.Second) // Run every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := h.storage.CleanupExpiredRequests(context.Background()); err != nil {
				h.logger.WithError(err).Error("Background cleanup failed")
			}
		}
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
	start := time.Now()

	var req MatchRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		metrics.RecordHTTPRequest("POST", "/api/v1/match-request", "400", time.Since(start).Seconds())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Remove any existing pending requests for this player and game
	requests, err := h.storage.GetGameQueue(c.Request.Context(), req.GameID)
	if err == nil {
		for _, r := range requests {
			if r.PlayerID == req.PlayerID && r.Status == models.StatusPending {
				_ = h.storage.RemoveFromQueue(c.Request.Context(), req.GameID, r.ID)
			}
		}
	}

	// Create match request
	matchRequest := models.NewMatchRequest(req.PlayerID, req.GameID, req.Metadata)

	// Store in Redis
	ctx := c.Request.Context()
	if err := h.storage.StoreMatchRequest(ctx, matchRequest); err != nil {
		h.logger.WithError(err).Error("Failed to store match request")
		metrics.RecordHTTPRequest("POST", "/api/v1/match-request", "500", time.Since(start).Seconds())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create match request"})
		return
	}

	// Record metrics
	metrics.RecordMatchRequest(req.GameID, "created")
	metrics.RecordHTTPRequest("POST", "/api/v1/match-request", "201", time.Since(start).Seconds())

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
	start := time.Now()
	gameID := c.Param("game_id")
	if gameID == "" {
		metrics.RecordHTTPRequest("POST", "/api/v1/rules", "400", time.Since(start).Seconds())
		c.JSON(http.StatusBadRequest, gin.H{"error": "game_id is required"})
		return
	}

	var config models.GameConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		metrics.RecordHTTPRequest("POST", "/api/v1/rules", "400", time.Since(start).Seconds())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set the game ID from the URL parameter
	config.GameID = gameID

	// Validate the configuration
	if err := h.ruleEngine.ValidateGameConfig(&config); err != nil {
		metrics.RecordHTTPRequest("POST", "/api/v1/rules", "400", time.Since(start).Seconds())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Store in Redis
	ctx := c.Request.Context()
	if err := h.storage.StoreGameConfig(ctx, &config); err != nil {
		h.logger.WithError(err).Error("Failed to store game config")
		metrics.RecordHTTPRequest("POST", "/api/v1/rules", "500", time.Since(start).Seconds())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store game configuration"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"game_id": config.GameID,
		"teams":   len(config.Teams),
		"rules":   len(config.Rules),
	}).Info("Created game configuration")

	metrics.RecordHTTPRequest("POST", "/api/v1/rules", "201", time.Since(start).Seconds())
	c.JSON(http.StatusCreated, gin.H{
		"game_id": config.GameID,
		"message": "Game configuration created successfully",
	})
}

// GetMatchStatus handles GET /match-status/:request_id
func (h *Handler) GetMatchStatus(c *gin.Context) {
	start := time.Now()
	requestID := c.Param("request_id")
	if requestID == "" {
		metrics.RecordHTTPRequest("GET", "/api/v1/match-status", "400", time.Since(start).Seconds())
		c.JSON(http.StatusBadRequest, gin.H{"error": "request_id is required"})
		return
	}

	// Try to get cached status first
	status, err := h.storage.GetMatchStatus(c.Request.Context(), requestID)
	if err == nil {
		metrics.RecordHTTPRequest("GET", "/api/v1/match-status", "200", time.Since(start).Seconds())
		c.JSON(http.StatusOK, status)
		return
	}

	// If no cached status, get the match request
	request, err := h.storage.GetMatchRequest(c.Request.Context(), requestID)
	if err != nil {
		metrics.RecordHTTPRequest("GET", "/api/v1/match-status", "404", time.Since(start).Seconds())
		c.JSON(http.StatusNotFound, gin.H{"error": "Match request not found"})
		return
	}

	// Create status response
	statusResponse := &models.MatchStatusResponse{
		Status: request.Status,
	}

	// If matched, try to find the match and session info
	if request.Status == models.StatusMatched || request.Status == models.StatusAllocated {
		matchID, err := h.storage.GetMatchIDForRequest(c.Request.Context(), requestID)
		if err == nil {
			match, err := h.storage.GetMatch(c.Request.Context(), matchID)
			if err == nil {
				statusResponse.Session = match.Session
				statusResponse.MatchID = match.ID
				statusResponse.Players = match.Players
				statusResponse.TeamName = match.TeamName
				statusResponse.CreatedAt = match.CreatedAt.Format(time.RFC3339)
			}
		}
		// If not found, fallback to warning
		if statusResponse.Session == nil {
			h.logger.WithField("request_id", requestID).Warn("Match found but session info not available")
		}
	}

	metrics.RecordHTTPRequest("GET", "/api/v1/match-status", "200", time.Since(start).Seconds())
	c.JSON(http.StatusOK, statusResponse)
}

// ProcessMatchmaking handles POST /process-matchmaking/:game_id
func (h *Handler) ProcessMatchmaking(c *gin.Context) {
	start := time.Now()
	gameID := c.Param("game_id")
	if gameID == "" {
		metrics.RecordHTTPRequest("POST", "/api/v1/process-matchmaking", "400", time.Since(start).Seconds())
		c.JSON(http.StatusBadRequest, gin.H{"error": "game_id is required"})
		return
	}

	_ = h.storage.CleanupExpiredRequests(c.Request.Context())

	config, err := h.storage.GetGameConfig(c.Request.Context(), gameID)
	if err != nil {
		metrics.RecordHTTPRequest("POST", "/api/v1/process-matchmaking", "404", time.Since(start).Seconds())
		c.JSON(http.StatusNotFound, gin.H{"error": "Game configuration not found"})
		return
	}

	requests, err := h.storage.GetGameQueue(c.Request.Context(), gameID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get game queue")
		metrics.RecordHTTPRequest("POST", "/api/v1/process-matchmaking", "500", time.Since(start).Seconds())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get match requests"})
		return
	}

	metrics.SetQueueSize(gameID, len(requests))

	if len(requests) == 0 {
		metrics.RecordHTTPRequest("POST", "/api/v1/process-matchmaking", "200", time.Since(start).Seconds())
		c.JSON(http.StatusOK, gin.H{
			"message": "No pending match requests",
			"matches": []interface{}{},
		})
		return
	}

	// Use ProcessFullTeamMatchPool for multi-team support
	multiTeamMatches := h.matchmaker.ProcessFullTeamMatchPool(requests, config)
	metrics.RecordMatchmakingDuration(gameID, time.Since(start).Seconds())

	if len(multiTeamMatches) == 0 {
		metrics.RecordHTTPRequest("POST", "/api/v1/process-matchmaking", "200", time.Since(start).Seconds())
		c.JSON(http.StatusOK, gin.H{
			"message": "No matches could be formed",
			"matches": []interface{}{},
		})
		return
	}

	var matchResults []gin.H
	for _, match := range multiTeamMatches {
		h.logger.WithFields(logrus.Fields{
			"match_id": match.ID,
			"teams":    match.Teams,
		}).Info("Storing multi-team match and updating all request statuses")

		if err := h.storage.StoreMultiTeamMatch(c.Request.Context(), match); err != nil {
			h.logger.WithError(err).Error("Failed to store multi-team match")
			continue
		}

		metrics.RecordMatchCreated(gameID, len(h.matchmaker.FlattenTeams(match.Teams)))

		// For each team, for each player, update status and mapping
		for teamName, playerIDs := range match.Teams {
			for _, playerID := range playerIDs {
				// Find the request ID for this player
				var requestID string
				for _, req := range requests {
					if req.PlayerID == playerID {
						requestID = req.ID
						break
					}
				}
				if requestID == "" {
					h.logger.WithField("player_id", playerID).Warn("No request ID found for player")
					continue
				}

				if err := h.storage.UpdateMatchRequestStatus(c.Request.Context(), requestID, models.StatusMatched); err != nil {
					h.logger.WithError(err).WithField("request_id", requestID).Error("Failed to update request status")
				}
				if err := h.storage.RemoveFromQueue(c.Request.Context(), gameID, requestID); err != nil {
					h.logger.WithError(err).WithField("request_id", requestID).Error("Failed to remove request from queue")
				}
				if err := h.storage.StoreRequestMatchMapping(c.Request.Context(), requestID, match.ID); err != nil {
					h.logger.WithError(err).WithField("request_id", requestID).Error("Failed to store request-match mapping")
				}

				// Build teammates (all players on the same team)
				teammates := make([]string, 0, len(playerIDs))
				for _, pid := range playerIDs {
					teammates = append(teammates, pid)
				}
				// Build all players in match
				allPlayers := h.matchmaker.FlattenTeams(match.Teams)

				statusResp := &models.MatchStatusResponse{
					Status:     models.StatusMatched,
					MatchID:    match.ID,
					Players:    teammates,
					TeamName:   teamName,
					CreatedAt:  match.CreatedAt.Format(time.RFC3339),
					AllPlayers: allPlayers,
				}
				if err := h.storage.StoreMatchStatus(c.Request.Context(), requestID, statusResp); err != nil {
					h.logger.WithError(err).WithField("request_id", requestID).Error("Failed to store match status response")
				}
			}
		}

		matchResults = append(matchResults, gin.H{
			"match_id":   match.ID,
			"teams":      match.Teams,
			"created_at": match.CreatedAt.Format(time.RFC3339),
		})
	}

	h.logger.WithFields(logrus.Fields{
		"game_id": gameID,
		"matches": len(matchResults),
	}).Info("Processed matchmaking")

	metrics.RecordHTTPRequest("POST", "/api/v1/process-matchmaking", "200", time.Since(start).Seconds())
	c.JSON(http.StatusOK, gin.H{
		"message": "Matchmaking processed successfully",
		"matches": matchResults,
	})
}

// AllocateSessions handles POST /allocate-sessions/:game_id
func (h *Handler) AllocateSessions(c *gin.Context) {
	start := time.Now()
	gameID := c.Param("game_id")
	if gameID == "" {
		metrics.RecordHTTPRequest("POST", "/api/v1/allocate-sessions", "400", time.Since(start).Seconds())
		c.JSON(http.StatusBadRequest, gin.H{"error": "game_id is required"})
		return
	}

	var matches []*models.Match
	if err := c.ShouldBindJSON(&matches); err != nil {
		metrics.RecordHTTPRequest("POST", "/api/v1/allocate-sessions", "400", time.Since(start).Seconds())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	results := make([]gin.H, 0, len(matches))
	for _, match := range matches {
		metrics.RecordAllocationRequest(gameID, "requested")
		session, err := h.allocator.AllocateSession(match)
		if err != nil {
			metrics.RecordAllocationError(gameID)
			results = append(results, gin.H{
				"match_id": match.ID,
				"error":    err.Error(),
			})
			continue
		}
		results = append(results, gin.H{
			"match_id": match.ID,
			"session":  session,
		})
	}

	metrics.RecordAllocationDuration(gameID, time.Since(start).Seconds())
	metrics.RecordHTTPRequest("POST", "/api/v1/allocate-sessions", "200", time.Since(start).Seconds())

	c.JSON(http.StatusOK, gin.H{
		"game_id":     gameID,
		"allocations": results,
	})
}

// GetStats handles GET /stats
func (h *Handler) GetStats(c *gin.Context) {
	start := time.Now()

	//ctx := c.Request.Context()

	// Get storage stats
	storageStats, err := h.storage.GetStats(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get storage stats")
		metrics.RecordHTTPRequest("GET", "/api/v1/stats", "500", time.Since(start).Seconds())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get statistics"})
		return
	}

	metrics.RecordHTTPRequest("GET", "/api/v1/stats", "200", time.Since(start).Seconds())
	c.JSON(http.StatusOK, gin.H{
		"storage":   storageStats,
		"timestamp": time.Now(),
	})
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(c *gin.Context) {
	start := time.Now()

	// Check Redis connection
	if err := h.storage.Ping(c.Request.Context()); err != nil {
		metrics.RecordHTTPRequest("GET", "/health", "503", time.Since(start).Seconds())
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "Redis connection failed",
		})
		return
	}

	metrics.RecordHTTPRequest("GET", "/health", "200", time.Since(start).Seconds())
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
	})
}
