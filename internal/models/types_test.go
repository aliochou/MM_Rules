package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMatchRequest(t *testing.T) {
	metadata := map[string]interface{}{"foo": "bar"}
	req := NewMatchRequest("player1", "game1", metadata)
	assert.NotEmpty(t, req.ID)
	assert.Equal(t, "player1", req.PlayerID)
	assert.Equal(t, "game1", req.GameID)
	assert.Equal(t, metadata, req.Metadata)
	assert.WithinDuration(t, time.Now(), req.CreatedAt, time.Second)
	assert.Equal(t, StatusPending, req.Status)
}

func TestNewMatch(t *testing.T) {
	players := []string{"p1", "p2"}
	match := NewMatch("game1", "red", players)
	assert.NotEmpty(t, match.ID)
	assert.Equal(t, "game1", match.GameID)
	assert.Equal(t, "red", match.TeamName)
	assert.Equal(t, players, match.Players)
	assert.WithinDuration(t, time.Now(), match.CreatedAt, time.Second)
}

func TestMatchRequest_JSON(t *testing.T) {
	req := &MatchRequest{
		ID:        "id1",
		PlayerID:  "p1",
		GameID:    "g1",
		Metadata:  map[string]interface{}{"foo": "bar"},
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		Status:    StatusPending,
	}
	data, err := json.Marshal(req)
	assert.NoError(t, err)
	var out MatchRequest
	assert.NoError(t, json.Unmarshal(data, &out))
	assert.Equal(t, req.ID, out.ID)
	assert.Equal(t, req.PlayerID, out.PlayerID)
	assert.Equal(t, req.GameID, out.GameID)
	assert.Equal(t, req.Metadata["foo"], out.Metadata["foo"])
	assert.Equal(t, req.Status, out.Status)
}

func TestGameConfig_JSON(t *testing.T) {
	cfg := &GameConfig{
		GameID: "g1",
		Teams:  []Team{{Name: "red", Size: 2}},
		Rules:  []Rule{{Field: "level", Min: intPtr(10), Strict: true, Priority: 1}},
		UpdatedAt: time.Now().UTC().Truncate(time.Second),
	}
	data, err := json.Marshal(cfg)
	assert.NoError(t, err)
	var out GameConfig
	assert.NoError(t, json.Unmarshal(data, &out))
	assert.Equal(t, cfg.GameID, out.GameID)
	assert.Equal(t, cfg.Teams[0].Name, out.Teams[0].Name)
	assert.Equal(t, cfg.Rules[0].Field, out.Rules[0].Field)
	assert.Equal(t, *cfg.Rules[0].Min, *out.Rules[0].Min)
	assert.Equal(t, cfg.Rules[0].Strict, out.Rules[0].Strict)
}

func TestMatch_JSON(t *testing.T) {
	sess := &GameSession{IP: "1.2.3.4", Port: 1234, ID: "sess1"}
	match := &Match{
		ID:        "m1",
		GameID:    "g1",
		TeamName:  "red",
		Players:   []string{"p1", "p2"},
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		Session:   sess,
	}
	data, err := json.Marshal(match)
	assert.NoError(t, err)
	var out Match
	assert.NoError(t, json.Unmarshal(data, &out))
	assert.Equal(t, match.ID, out.ID)
	assert.Equal(t, match.GameID, out.GameID)
	assert.Equal(t, match.TeamName, out.TeamName)
	assert.Equal(t, match.Players, out.Players)
	assert.Equal(t, match.Session.IP, out.Session.IP)
}

func intPtr(i int) *int { return &i } 