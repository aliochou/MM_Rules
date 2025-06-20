package allocation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mm-rules/matchmaking/internal/models"
)

// Allocator handles game session allocation
type Allocator struct {
	webhookURL string
	client     *http.Client
}

// NewAllocator creates a new allocator instance
func NewAllocator(webhookURL string) *Allocator {
	return &Allocator{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// AllocateSession allocates a game session for a match
func (a *Allocator) AllocateSession(match *models.Match) (*models.GameSession, error) {
	// Create allocation request
	req := models.AllocationRequest{
		MatchID:  match.ID,
		GameID:   match.GameID,
		Players:  match.Players,
		TeamName: match.TeamName,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal allocation request: %w", err)
	}

	// Make HTTP request to allocation service
	httpReq, err := http.NewRequest("POST", a.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "MM-Rules-Allocator/1.0")

	// Send request
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send allocation request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var allocationResp models.AllocationResponse
	if err := json.NewDecoder(resp.Body).Decode(&allocationResp); err != nil {
		return nil, fmt.Errorf("failed to decode allocation response: %w", err)
	}

	// Check if allocation was successful
	if !allocationResp.Success {
		if allocationResp.Error != nil {
			return nil, fmt.Errorf("allocation failed: %s", *allocationResp.Error)
		}
		return nil, fmt.Errorf("allocation failed with unknown error")
	}

	if allocationResp.Session == nil {
		return nil, fmt.Errorf("allocation succeeded but no session returned")
	}

	return allocationResp.Session, nil
}

// AllocateSessionWithRetry allocates a session with retry logic
func (a *Allocator) AllocateSessionWithRetry(match *models.Match, maxRetries int, retryDelay time.Duration) (*models.GameSession, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		session, err := a.AllocateSession(match)
		if err == nil {
			return session, nil
		}

		lastErr = err

		// Don't sleep on the last attempt
		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("allocation failed after %d attempts: %w", maxRetries+1, lastErr)
}

// ValidateAllocationRequest validates an allocation request
func (a *Allocator) ValidateAllocationRequest(req *models.AllocationRequest) error {
	if req.MatchID == "" {
		return fmt.Errorf("match_id is required")
	}

	if req.GameID == "" {
		return fmt.Errorf("game_id is required")
	}

	if len(req.Players) == 0 {
		return fmt.Errorf("at least one player is required")
	}

	if req.TeamName == "" {
		return fmt.Errorf("team_name is required")
	}

	return nil
}

// MockAllocator is a mock allocator for testing
type MockAllocator struct {
	sessions map[string]*models.GameSession
	errors   map[string]error
}

// NewMockAllocator creates a new mock allocator
func NewMockAllocator() *MockAllocator {
	return &MockAllocator{
		sessions: make(map[string]*models.GameSession),
		errors:   make(map[string]error),
	}
}

// AllocateSession allocates a session using mock data
func (ma *MockAllocator) AllocateSession(match *models.Match) (*models.GameSession, error) {
	// Check if we should return an error for this match
	if err, exists := ma.errors[match.ID]; exists {
		return nil, err
	}

	// Check if we have a predefined session for this match
	if session, exists := ma.sessions[match.ID]; exists {
		return session, nil
	}

	// Generate a mock session
	session := &models.GameSession{
		IP:   "192.168.1.100",
		Port: 7777,
		ID:   fmt.Sprintf("session-%s", match.ID),
	}

	return session, nil
}

// SetMockSession sets a mock session for a specific match
func (ma *MockAllocator) SetMockSession(matchID string, session *models.GameSession) {
	ma.sessions[matchID] = session
}

// SetMockError sets a mock error for a specific match
func (ma *MockAllocator) SetMockError(matchID string, err error) {
	ma.errors[matchID] = err
} 