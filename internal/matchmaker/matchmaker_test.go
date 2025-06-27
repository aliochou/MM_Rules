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
			ID:        "req1",
			GameID:    "test-game",
			PlayerID:  "player1",
			Metadata:  map[string]interface{}{"level": 15},
			Status:    models.StatusPending,
			CreatedAt: time.Now().Add(-time.Minute),
		},
		{
			ID:        "req2",
			GameID:    "test-game",
			PlayerID:  "player2",
			Metadata:  map[string]interface{}{"level": 12},
			Status:    models.StatusPending,
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
			ID:        "req1",
			GameID:    "test-game",
			PlayerID:  "player1",
			Metadata:  map[string]interface{}{"level": 15},
			Status:    models.StatusPending,
			CreatedAt: time.Now(),
		},
		{
			ID:        "req2",
			GameID:    "test-game",
			PlayerID:  "player2",
			Metadata:  map[string]interface{}{"level": 12},
			Status:    models.StatusPending,
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
			ID:        "req1",
			GameID:    "test-game",
			PlayerID:  "player1",
			Metadata:  map[string]interface{}{"level": 15},
			Status:    models.StatusPending,
			CreatedAt: time.Now().Add(-time.Minute),
		},
		{
			ID:        "req2",
			GameID:    "test-game",
			PlayerID:  "player2",
			Metadata:  map[string]interface{}{"level": 12},
			Status:    models.StatusPending,
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

func TestMatchmaker_ProcessFullTeamMatchPool_1v1_Success(t *testing.T) {
	matchmaker := NewMatchmaker()

	// 1v1 configuration with two teams
	config := &models.GameConfig{
		GameID: "game-1v1",
		Teams: []models.Team{
			{Name: "Player1", Size: 1},
			{Name: "Player2", Size: 1},
		},
		Rules: []models.Rule{
			{Field: "level", Min: &[]int{10}[0], Max: &[]int{50}[0], Strict: false},
		},
	}

	// Two players with compatible metadata
	players := []*models.MatchRequest{
		{
			ID:        "req1",
			GameID:    "game-1v1",
			PlayerID:  "player1",
			Metadata:  map[string]interface{}{"level": 25},
			Status:    models.StatusPending,
			CreatedAt: time.Now().Add(-time.Minute),
		},
		{
			ID:        "req2",
			GameID:    "game-1v1",
			PlayerID:  "player2",
			Metadata:  map[string]interface{}{"level": 30},
			Status:    models.StatusPending,
			CreatedAt: time.Now(),
		},
	}

	matches := matchmaker.ProcessFullTeamMatchPool(players, config)

	// Should create exactly one match with both teams filled
	assert.Len(t, matches, 1)
	match := matches[0]
	assert.Equal(t, "game-1v1", match.GameID)
	assert.Len(t, match.Teams, 2)

	// Check that both teams have exactly one player
	assert.Len(t, match.Teams["Player1"], 1)
	assert.Len(t, match.Teams["Player2"], 1)

	// Verify the players are assigned correctly
	assert.Contains(t, match.Teams["Player1"], "player1")
	assert.Contains(t, match.Teams["Player2"], "player2")
}

func TestMatchmaker_ProcessFullTeamMatchPool_1v1_NotEnoughPlayers(t *testing.T) {
	matchmaker := NewMatchmaker()

	// 1v1 configuration
	config := &models.GameConfig{
		GameID: "game-1v1",
		Teams: []models.Team{
			{Name: "Player1", Size: 1},
			{Name: "Player2", Size: 1},
		},
		Rules: []models.Rule{
			{Field: "level", Min: &[]int{10}[0], Max: &[]int{50}[0], Strict: false},
		},
	}

	// Only one player - should not create a match
	players := []*models.MatchRequest{
		{
			ID:        "req1",
			GameID:    "game-1v1",
			PlayerID:  "player1",
			Metadata:  map[string]interface{}{"level": 25},
			Status:    models.StatusPending,
			CreatedAt: time.Now(),
		},
	}

	matches := matchmaker.ProcessFullTeamMatchPool(players, config)

	// Should not create any matches since we need 2 players for 1v1
	assert.Empty(t, matches)
}

func TestMatchmaker_ProcessFullTeamMatchPool_1v3_Success(t *testing.T) {
	matchmaker := NewMatchmaker()

	// 1v3 configuration
	config := &models.GameConfig{
		GameID: "game-1v3",
		Teams: []models.Team{
			{Name: "Solo", Size: 1},
			{Name: "Trio", Size: 3},
		},
		Rules: []models.Rule{
			{Field: "level", Min: &[]int{15}[0], Max: &[]int{60}[0], Strict: false},
		},
	}

	// Four players with compatible metadata
	players := []*models.MatchRequest{
		{
			ID:        "req1",
			GameID:    "game-1v3",
			PlayerID:  "solo_player",
			Metadata:  map[string]interface{}{"level": 35},
			Status:    models.StatusPending,
			CreatedAt: time.Now().Add(-time.Minute),
		},
		{
			ID:        "req2",
			GameID:    "game-1v3",
			PlayerID:  "trio_player1",
			Metadata:  map[string]interface{}{"level": 28},
			Status:    models.StatusPending,
			CreatedAt: time.Now(),
		},
		{
			ID:        "req3",
			GameID:    "game-1v3",
			PlayerID:  "trio_player2",
			Metadata:  map[string]interface{}{"level": 32},
			Status:    models.StatusPending,
			CreatedAt: time.Now(),
		},
		{
			ID:        "req4",
			GameID:    "game-1v3",
			PlayerID:  "trio_player3",
			Metadata:  map[string]interface{}{"level": 29},
			Status:    models.StatusPending,
			CreatedAt: time.Now(),
		},
	}

	matches := matchmaker.ProcessFullTeamMatchPool(players, config)

	// Should create exactly one match with both teams filled
	assert.Len(t, matches, 1)
	match := matches[0]
	assert.Equal(t, "game-1v3", match.GameID)
	assert.Len(t, match.Teams, 2)

	// Check team sizes
	assert.Len(t, match.Teams["Solo"], 1)
	assert.Len(t, match.Teams["Trio"], 3)

	// Verify the players are assigned correctly
	assert.Contains(t, match.Teams["Solo"], "solo_player")
	assert.Contains(t, match.Teams["Trio"], "trio_player1")
	assert.Contains(t, match.Teams["Trio"], "trio_player2")
	assert.Contains(t, match.Teams["Trio"], "trio_player3")
}

func TestMatchmaker_ProcessFullTeamMatchPool_MultipleMatches(t *testing.T) {
	matchmaker := NewMatchmaker()

	// 1v1 configuration
	config := &models.GameConfig{
		GameID: "game-1v1",
		Teams: []models.Team{
			{Name: "Player1", Size: 1},
			{Name: "Player2", Size: 1},
		},
		Rules: []models.Rule{
			{Field: "level", Min: &[]int{10}[0], Max: &[]int{50}[0], Strict: false},
		},
	}

	// Four players - should create two 1v1 matches
	players := []*models.MatchRequest{
		{
			ID:        "req1",
			GameID:    "game-1v1",
			PlayerID:  "player1",
			Metadata:  map[string]interface{}{"level": 25},
			Status:    models.StatusPending,
			CreatedAt: time.Now().Add(-time.Minute * 2),
		},
		{
			ID:        "req2",
			GameID:    "game-1v1",
			PlayerID:  "player2",
			Metadata:  map[string]interface{}{"level": 30},
			Status:    models.StatusPending,
			CreatedAt: time.Now().Add(-time.Minute),
		},
		{
			ID:        "req3",
			GameID:    "game-1v1",
			PlayerID:  "player3",
			Metadata:  map[string]interface{}{"level": 20},
			Status:    models.StatusPending,
			CreatedAt: time.Now(),
		},
		{
			ID:        "req4",
			GameID:    "game-1v1",
			PlayerID:  "player4",
			Metadata:  map[string]interface{}{"level": 35},
			Status:    models.StatusPending,
			CreatedAt: time.Now(),
		},
	}

	matches := matchmaker.ProcessFullTeamMatchPool(players, config)

	// Should create exactly two matches
	assert.Len(t, matches, 2)

	// Verify each match has both teams filled
	for _, match := range matches {
		assert.Equal(t, "game-1v1", match.GameID)
		assert.Len(t, match.Teams, 2)
		assert.Len(t, match.Teams["Player1"], 1)
		assert.Len(t, match.Teams["Player2"], 1)
	}

	// Verify all players are used
	allPlayers := make(map[string]bool)
	for _, match := range matches {
		for _, teamPlayers := range match.Teams {
			for _, playerID := range teamPlayers {
				allPlayers[playerID] = true
			}
		}
	}
	assert.Len(t, allPlayers, 4)
	assert.True(t, allPlayers["player1"])
	assert.True(t, allPlayers["player2"])
	assert.True(t, allPlayers["player3"])
	assert.True(t, allPlayers["player4"])
}
