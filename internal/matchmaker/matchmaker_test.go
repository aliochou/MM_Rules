package matchmaker

import (
	"testing"
	"time"

	"github.com/mm-rules/matchmaking/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewMatchmaker(t *testing.T) {
	matchmaker := NewMatchmaker()
	assert.NotNil(t, matchmaker)
	assert.NotNil(t, matchmaker.ruleEngine)
}

func TestMatchmaker_ProcessMatchPool_Success(t *testing.T) {
	matchmaker := NewMatchmaker()
	
	config := &models.GameConfig{
		GameID: "test-game",
		Teams: []models.Team{
			{Name: "team1", Size: 2},
		},
		Rules: []models.Rule{{Field: "level", Min: &[]int{10}[0]}},
	}
	
	players := []*models.MatchRequest{
		{
			ID:       "req1",
			GameID:   "test-game",
			PlayerID: "player1",
			Metadata: map[string]interface{}{"level": 15},
			Status:   models.StatusPending,
			CreatedAt: time.Now().Add(-time.Minute),
		},
		{
			ID:       "req2",
			GameID:   "test-game",
			PlayerID: "player2",
			Metadata: map[string]interface{}{"level": 12},
			Status:   models.StatusPending,
			CreatedAt: time.Now(),
		},
	}
	
	matches := matchmaker.ProcessMatchPool(players, config)
	
	assert.Len(t, matches, 1)
	assert.Equal(t, "test-game", matches[0].GameID)
	assert.Equal(t, "team1", matches[0].TeamName)
	assert.Len(t, matches[0].Players, 2)
}

func TestMatchmaker_ProcessMatchPool_NoPlayers(t *testing.T) {
	matchmaker := NewMatchmaker()
	
	config := &models.GameConfig{
		GameID: "test-game",
		Teams: []models.Team{
			{Name: "team1", Size: 2},
		},
		Rules: []models.Rule{{Field: "level", Min: &[]int{10}[0]}},
	}
	
	players := []*models.MatchRequest{}
	
	matches := matchmaker.ProcessMatchPool(players, config)
	
	assert.Empty(t, matches)
}

func TestMatchmaker_ProcessMatchPool_NotEnoughPlayers(t *testing.T) {
	matchmaker := NewMatchmaker()
	
	config := &models.GameConfig{
		GameID: "test-game",
		Teams: []models.Team{
			{Name: "team1", Size: 3},
		},
		Rules: []models.Rule{{Field: "level", Min: &[]int{10}[0]}},
	}
	
	players := []*models.MatchRequest{
		{
			ID:       "req1",
			GameID:   "test-game",
			PlayerID: "player1",
			Metadata: map[string]interface{}{"level": 15},
			Status:   models.StatusPending,
			CreatedAt: time.Now(),
		},
		{
			ID:       "req2",
			GameID:   "test-game",
			PlayerID: "player2",
			Metadata: map[string]interface{}{"level": 12},
			Status:   models.StatusPending,
			CreatedAt: time.Now(),
		},
	}
	
	matches := matchmaker.ProcessMatchPool(players, config)
	
	assert.Empty(t, matches)
}

func TestMatchmaker_ValidateMatch_Success(t *testing.T) {
	matchmaker := NewMatchmaker()
	
	config := &models.GameConfig{
		GameID: "test-game",
		Teams: []models.Team{
			{Name: "team1", Size: 2},
		},
	}
	
	match := &models.Match{
		GameID:   "test-game",
		TeamName: "team1",
		Players:  []string{"player1", "player2"},
	}
	
	err := matchmaker.ValidateMatch(match, config)
	assert.NoError(t, err)
}

func TestMatchmaker_ValidateMatch_InvalidTeam(t *testing.T) {
	matchmaker := NewMatchmaker()
	
	config := &models.GameConfig{
		GameID: "test-game",
		Teams: []models.Team{
			{Name: "team1", Size: 2},
		},
	}
	
	match := &models.Match{
		GameID:   "test-game",
		TeamName: "invalid-team",
		Players:  []string{"player1", "player2"},
	}
	
	err := matchmaker.ValidateMatch(match, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "team 'invalid-team' not found")
}

func TestMatchmaker_ValidateMatch_WrongTeamSize(t *testing.T) {
	matchmaker := NewMatchmaker()
	
	config := &models.GameConfig{
		GameID: "test-game",
		Teams: []models.Team{
			{Name: "team1", Size: 3},
		},
	}
	
	match := &models.Match{
		GameID:   "test-game",
		TeamName: "team1",
		Players:  []string{"player1", "player2"},
	}
	
	err := matchmaker.ValidateMatch(match, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "team 'team1' requires 3 players, got 2")
}

func TestMatchmaker_GetMatchStats(t *testing.T) {
	matchmaker := NewMatchmaker()
	
	players := []*models.MatchRequest{
		{
			ID:       "req1",
			GameID:   "test-game",
			PlayerID: "player1",
			Metadata: map[string]interface{}{"level": 15},
			Status:   models.StatusPending,
			CreatedAt: time.Now().Add(-time.Minute),
		},
		{
			ID:       "req2",
			GameID:   "test-game",
			PlayerID: "player2",
			Metadata: map[string]interface{}{"level": 12},
			Status:   models.StatusPending,
			CreatedAt: time.Now(),
		},
	}
	
	matches := []*models.Match{
		{
			GameID:   "test-game",
			TeamName: "team1",
			Players:  []string{"player1", "player2"},
		},
	}
	
	stats := matchmaker.GetMatchStats(players, matches)
	
	assert.Equal(t, 2, stats["total_players"])
	assert.Equal(t, 1, stats["total_matches"])
	assert.Equal(t, 2, stats["matched_players"])
	assert.Equal(t, 0, stats["unmatched_players"])
	assert.NotNil(t, stats["avg_wait_time"])
	assert.NotNil(t, stats["max_wait_time"])
	assert.NotNil(t, stats["min_wait_time"])
	assert.NotNil(t, stats["median_wait_time"])
} 