package engine

import (
	"testing"
	"time"

	"github.com/mm-rules/matchmaking/internal/models"
)

func TestRuleEngine_EvaluatePlayer(t *testing.T) {
	engine := NewRuleEngine()

	tests := []struct {
		name     string
		player   *models.MatchRequest
		rules    []models.Rule
		expected bool
	}{
		{
			name: "Level requirement met",
			player: &models.MatchRequest{
				Metadata: map[string]interface{}{
					"level": 25,
				},
			},
			rules: []models.Rule{
				{
					Field:  "level",
					Min:    &[]int{20}[0],
					Strict: true,
				},
			},
			expected: true,
		},
		{
			name: "Level requirement not met",
			player: &models.MatchRequest{
				Metadata: map[string]interface{}{
					"level": 15,
				},
			},
			rules: []models.Rule{
				{
					Field:  "level",
					Min:    &[]int{20}[0],
					Strict: true,
				},
			},
			expected: false,
		},
		{
			name: "Inventory contains item",
			player: &models.MatchRequest{
				Metadata: map[string]interface{}{
					"inventory": []string{"itemA", "itemB"},
				},
			},
			rules: []models.Rule{
				{
					Field:    "inventory",
					Contains: &[]string{"itemA"}[0],
					Strict:   false,
				},
			},
			expected: true,
		},
		{
			name: "Rule relaxation after time",
			player: &models.MatchRequest{
				Metadata: map[string]interface{}{
					"level": 15,
				},
				CreatedAt: time.Now().Add(-15 * time.Second),
			},
			rules: []models.Rule{
				{
					Field:      "level",
					Min:        &[]int{20}[0],
					RelaxAfter: &[]int{10}[0],
					Strict:     false,
				},
			},
			expected: true, // Should pass due to relaxation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := engine.EvaluatePlayer(tt.player, tt.rules, time.Since(tt.player.CreatedAt))
			if result != tt.expected {
				t.Errorf("EvaluatePlayer() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRuleEngine_ValidateGameConfig(t *testing.T) {
	engine := NewRuleEngine()

	tests := []struct {
		name    string
		config  *models.GameConfig
		wantErr bool
	}{
		{
			name: "Valid config",
			config: &models.GameConfig{
				GameID: "test-game",
				Teams: []models.Team{
					{Name: "Solo", Size: 1},
					{Name: "Duo", Size: 2},
				},
				Rules: []models.Rule{
					{
						Field:  "level",
						Min:    &[]int{20}[0],
						Strict: true,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Missing game ID",
			config: &models.GameConfig{
				Teams: []models.Team{
					{Name: "Solo", Size: 1},
				},
				Rules: []models.Rule{
					{
						Field:  "level",
						Min:    &[]int{20}[0],
						Strict: true,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "No teams",
			config: &models.GameConfig{
				GameID: "test-game",
				Rules: []models.Rule{
					{
						Field:  "level",
						Min:    &[]int{20}[0],
						Strict: true,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid team size",
			config: &models.GameConfig{
				GameID: "test-game",
				Teams: []models.Team{
					{Name: "Solo", Size: 0},
				},
				Rules: []models.Rule{
					{
						Field:  "level",
						Min:    &[]int{20}[0],
						Strict: true,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.ValidateGameConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGameConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRuleEngine_FindCompatiblePlayers(t *testing.T) {
	engine := NewRuleEngine()

	players := []*models.MatchRequest{
		{
			ID: "1",
			Metadata: map[string]interface{}{
				"level": 25,
			},
		},
		{
			ID: "2",
			Metadata: map[string]interface{}{
				"level": 15,
			},
		},
		{
			ID: "3",
			Metadata: map[string]interface{}{
				"level": 30,
			},
		},
	}

	rules := []models.Rule{
		{
			Field:  "level",
			Min:    &[]int{20}[0],
			Strict: true,
		},
	}

	compatible := engine.FindCompatiblePlayers(players, rules, 0)
	if len(compatible) != 2 {
		t.Errorf("Expected 2 compatible players, got %d", len(compatible))
	}

	// Check that only players with level >= 20 are included
	for _, player := range compatible {
		level := player.Metadata["level"].(int)
		if level < 20 {
			t.Errorf("Player with level %d should not be compatible", level)
		}
	}
} 