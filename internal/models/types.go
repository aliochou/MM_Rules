package models

import (
	"time"

	"github.com/google/uuid"
)

// MatchRequest represents a player's request to join matchmaking
type MatchRequest struct {
	ID        string                 `json:"id"`
	PlayerID  string                 `json:"player_id"`
	GameID    string                 `json:"game_id"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	Status    MatchStatus            `json:"status"`
}

// MatchStatus represents the current status of a match request
type MatchStatus string

const (
	StatusPending  MatchStatus = "pending"
	StatusMatched  MatchStatus = "matched"
	StatusAllocated MatchStatus = "allocated"
	StatusFailed   MatchStatus = "failed"
)

// GameConfig represents the rules and team configuration for a game
type GameConfig struct {
	GameID string     `json:"game_id"`
	Teams  []Team     `json:"teams"`
	Rules  []Rule     `json:"rules"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Team represents a team configuration
type Team struct {
	Name string `json:"name"`
	Size int    `json:"size"`
}

// Rule represents a matchmaking rule
type Rule struct {
	Field       string      `json:"field"`
	Min         *int        `json:"min,omitempty"`
	Max         *int        `json:"max,omitempty"`
	Contains    *string     `json:"contains,omitempty"`
	Equals      *string     `json:"equals,omitempty"`
	Strict      bool        `json:"strict"`
	RelaxAfter  *int        `json:"relax_after,omitempty"` // seconds
	Priority    int         `json:"priority"`              // higher = more important
}

// Match represents a successful match of players
type Match struct {
	ID        string         `json:"id"`
	GameID    string         `json:"game_id"`
	TeamName  string         `json:"team_name"`
	Players   []string       `json:"players"`
	CreatedAt time.Time      `json:"created_at"`
	Session   *GameSession   `json:"session,omitempty"`
}

// GameSession represents the allocated game session
type GameSession struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
	ID   string `json:"id"`
}

// MatchStatusResponse represents the response for match status queries
type MatchStatusResponse struct {
	Status  MatchStatus   `json:"status"`
	Team    *string       `json:"team,omitempty"`
	Session *GameSession  `json:"session,omitempty"`
	Error   *string       `json:"error,omitempty"`
}

// AllocationRequest represents a request to allocate a game session
type AllocationRequest struct {
	MatchID string   `json:"match_id"`
	GameID  string   `json:"game_id"`
	Players []string `json:"players"`
	TeamName string  `json:"team_name"`
}

// AllocationResponse represents the response from the allocation service
type AllocationResponse struct {
	Success bool         `json:"success"`
	Session *GameSession `json:"session,omitempty"`
	Error   *string      `json:"error,omitempty"`
}

// NewMatchRequest creates a new match request with a generated ID
func NewMatchRequest(playerID, gameID string, metadata map[string]interface{}) *MatchRequest {
	return &MatchRequest{
		ID:        uuid.New().String(),
		PlayerID:  playerID,
		GameID:    gameID,
		Metadata:  metadata,
		CreatedAt: time.Now(),
		Status:    StatusPending,
	}
}

// NewMatch creates a new match with a generated ID
func NewMatch(gameID, teamName string, players []string) *Match {
	return &Match{
		ID:        uuid.New().String(),
		GameID:    gameID,
		TeamName:  teamName,
		Players:   players,
		CreatedAt: time.Now(),
	}
} 