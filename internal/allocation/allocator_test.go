package allocation

import (
	"testing"
	"time"

	"github.com/mm-rules/matchmaking/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewAllocator(t *testing.T) {
	allocator := NewAllocator("http://localhost:8080/webhook")
	assert.NotNil(t, allocator)
	assert.Equal(t, "http://localhost:8080/webhook", allocator.webhookURL)
	assert.NotNil(t, allocator.client)
}

func TestNewMockAllocator(t *testing.T) {
	mockAllocator := NewMockAllocator()
	assert.NotNil(t, mockAllocator)
	assert.NotNil(t, mockAllocator.sessions)
	assert.NotNil(t, mockAllocator.errors)
}

func TestMockAllocator_AllocateSession_Success(t *testing.T) {
	mockAllocator := NewMockAllocator()
	
	match := &models.Match{
		ID:       "match1",
		GameID:   "test-game",
		TeamName: "team1",
		Players:  []string{"player1", "player2"},
	}
	
	session, err := mockAllocator.AllocateSession(match)
	
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "192.168.1.100", session.IP)
	assert.Equal(t, 7777, session.Port)
	assert.Equal(t, "session-match1", session.ID)
}

func TestMockAllocator_AllocateSession_WithPredefinedSession(t *testing.T) {
	mockAllocator := NewMockAllocator()
	
	match := &models.Match{
		ID:       "match1",
		GameID:   "test-game",
		TeamName: "team1",
		Players:  []string{"player1", "player2"},
	}
	
	predefinedSession := &models.GameSession{
		IP:   "10.0.0.1",
		Port: 8080,
		ID:   "custom-session",
	}
	
	mockAllocator.SetMockSession("match1", predefinedSession)
	
	session, err := mockAllocator.AllocateSession(match)
	
	assert.NoError(t, err)
	assert.Equal(t, predefinedSession, session)
}

func TestMockAllocator_AllocateSession_WithError(t *testing.T) {
	mockAllocator := NewMockAllocator()
	
	match := &models.Match{
		ID:       "match1",
		GameID:   "test-game",
		TeamName: "team1",
		Players:  []string{"player1", "player2"},
	}
	
	expectedError := assert.AnError
	mockAllocator.SetMockError("match1", expectedError)
	
	session, err := mockAllocator.AllocateSession(match)
	
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, session)
}

func TestMockAllocator_AllocateSessionWithRetry_Success(t *testing.T) {
	mockAllocator := NewMockAllocator()
	
	match := &models.Match{
		ID:       "match1",
		GameID:   "test-game",
		TeamName: "team1",
		Players:  []string{"player1", "player2"},
	}
	
	session, err := mockAllocator.AllocateSessionWithRetry(match, 3, time.Millisecond)
	
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "session-match1", session.ID)
}

func TestMockAllocator_AllocateSessionWithRetry_WithError(t *testing.T) {
	mockAllocator := NewMockAllocator()
	
	match := &models.Match{
		ID:       "match1",
		GameID:   "test-game",
		TeamName: "team1",
		Players:  []string{"player1", "player2"},
	}
	
	expectedError := assert.AnError
	mockAllocator.SetMockError("match1", expectedError)
	
	session, err := mockAllocator.AllocateSessionWithRetry(match, 3, time.Millisecond)
	
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, session)
}

func TestMockAllocator_ValidateAllocationRequest_Success(t *testing.T) {
	mockAllocator := NewMockAllocator()
	
	req := &models.AllocationRequest{
		MatchID:  "match1",
		GameID:   "test-game",
		Players:  []string{"player1", "player2"},
		TeamName: "team1",
	}
	
	err := mockAllocator.ValidateAllocationRequest(req)
	assert.NoError(t, err)
}

func TestMockAllocator_ValidateAllocationRequest_MissingMatchID(t *testing.T) {
	mockAllocator := NewMockAllocator()
	
	req := &models.AllocationRequest{
		GameID:   "test-game",
		Players:  []string{"player1", "player2"},
		TeamName: "team1",
	}
	
	err := mockAllocator.ValidateAllocationRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "match_id is required")
}

func TestMockAllocator_ValidateAllocationRequest_MissingGameID(t *testing.T) {
	mockAllocator := NewMockAllocator()
	
	req := &models.AllocationRequest{
		MatchID:  "match1",
		Players:  []string{"player1", "player2"},
		TeamName: "team1",
	}
	
	err := mockAllocator.ValidateAllocationRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "game_id is required")
}

func TestMockAllocator_ValidateAllocationRequest_NoPlayers(t *testing.T) {
	mockAllocator := NewMockAllocator()
	
	req := &models.AllocationRequest{
		MatchID:  "match1",
		GameID:   "test-game",
		Players:  []string{},
		TeamName: "team1",
	}
	
	err := mockAllocator.ValidateAllocationRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one player is required")
}

func TestMockAllocator_ValidateAllocationRequest_MissingTeamName(t *testing.T) {
	mockAllocator := NewMockAllocator()
	
	req := &models.AllocationRequest{
		MatchID: "match1",
		GameID:  "test-game",
		Players: []string{"player1", "player2"},
	}
	
	err := mockAllocator.ValidateAllocationRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "team_name is required")
}

func TestRealAllocator_ValidateAllocationRequest_Success(t *testing.T) {
	allocator := NewAllocator("http://localhost:8080/webhook")
	
	req := &models.AllocationRequest{
		MatchID:  "match1",
		GameID:   "test-game",
		Players:  []string{"player1", "player2"},
		TeamName: "team1",
	}
	
	err := allocator.ValidateAllocationRequest(req)
	assert.NoError(t, err)
}

func TestRealAllocator_ValidateAllocationRequest_MissingMatchID(t *testing.T) {
	allocator := NewAllocator("http://localhost:8080/webhook")
	
	req := &models.AllocationRequest{
		GameID:   "test-game",
		Players:  []string{"player1", "player2"},
		TeamName: "team1",
	}
	
	err := allocator.ValidateAllocationRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "match_id is required")
} 