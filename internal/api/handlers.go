package api

import (
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
	start := time.Now()
	
	var req MatchRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		metrics.RecordHTTPRequest("POST", "/api/v1/match-request", "400", time.Since(start).Seconds())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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
		// This is a simplified approach - in a real implementation you'd need to
		// track which match a request belongs to
		h.logger.WithField("request_id", requestID).Warn("Match found but session info not available")
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

	// Get game configuration
	config, err := h.storage.GetGameConfig(c.Request.Context(), gameID)
	if err != nil {
		metrics.RecordHTTPRequest("POST", "/api/v1/process-matchmaking", "404", time.Since(start).Seconds())
		c.JSON(http.StatusNotFound, gin.H{"error": "Game configuration not found"})
		return
	}

	// Get pending match requests
	requests, err := h.storage.GetGameQueue(c.Request.Context(), gameID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get game queue")
		metrics.RecordHTTPRequest("POST", "/api/v1/process-matchmaking", "500", time.Since(start).Seconds())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get match requests"})
		return
	}

	// Update queue size metric
	metrics.SetQueueSize(gameID, len(requests))

	if len(requests) == 0 {
		metrics.RecordHTTPRequest("POST", "/api/v1/process-matchmaking", "200", time.Since(start).Seconds())
		c.JSON(http.StatusOK, gin.H{
			"message": "No pending match requests",
			"matches": []interface{}{},
		})
		return
	}

	// Process matchmaking
	matchResultsWithRequests := h.matchmaker.ProcessMatchPoolWithRequests(requests, config)

	// Record matchmaking duration
	metrics.RecordMatchmakingDuration(gameID, time.Since(start).Seconds())

	if len(matchResultsWithRequests) == 0 {
		metrics.RecordHTTPRequest("POST", "/api/v1/process-matchmaking", "200", time.Since(start).Seconds())
		c.JSON(http.StatusOK, gin.H{
			"message": "No matches could be formed",
			"matches": []interface{}{},
		})
		return
	}

	// Store matches and update request statuses
	var matchResults []gin.H
	for _, result := range matchResultsWithRequests {
		// Store the match
		if err := h.storage.StoreMatch(c.Request.Context(), result.Match); err != nil {
			h.logger.WithError(err).Error("Failed to store match")
			continue
		}

		// Record match creation metrics
		metrics.RecordMatchCreated(gameID, len(result.Match.Players))

		// Update request statuses using request IDs
		for _, requestID := range result.RequestIDs {
			if err := h.storage.UpdateMatchRequestStatus(c.Request.Context(), requestID, models.StatusMatched); err != nil {
				h.logger.WithError(err).WithField("request_id", requestID).Error("Failed to update request status")
			}
		}

		matchResults = append(matchResults, gin.H{
			"match_id":  result.Match.ID,
			"players":   result.Match.Players,
			"team_name": result.Match.TeamName,
			"created_at": result.Match.CreatedAt,
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
		session, err := h.allocator.AllocateSession(match)
		if err != nil {
			results = append(results, gin.H{
				"match_id": match.ID,
				"error":   err.Error(),
			})
			continue
		}
		results = append(results, gin.H{
			"match_id": match.ID,
			"session":  session,
		})
	}

	metrics.RecordAllocationRequest(gameID, "requested")
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
		"storage": storageStats,
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
		"status": "healthy",
		"timestamp": time.Now(),
	})
} 