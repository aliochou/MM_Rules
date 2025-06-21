package storage

import (
	"context"
	"testing"
	"time"

	"github.com/mm-rules/matchmaking/internal/models"
)

func TestStoreAndGetMatchRequest(t *testing.T) {
	storage := &RedisStorage{client: nil}
	ctx := context.Background()
	request := &models.MatchRequest{
		ID:        "req1",
		PlayerID:  "p1",
		GameID:    "g1",
		Metadata:  map[string]interface{}{"foo": "bar"},
		CreatedAt: time.Now(),
		Status:    models.StatusPending,
	}
	_ = storage
	_ = ctx
	_ = request
}

func TestStoreAndGetGameConfig(t *testing.T) {
	storage := &RedisStorage{client: nil}
	ctx := context.Background()
	cfg := &models.GameConfig{
		GameID: "g1",
		Teams:  []models.Team{{Name: "red", Size: 2}},
		Rules:  []models.Rule{{Field: "level", Min: intPtr(10), Strict: true, Priority: 1}},
		UpdatedAt: time.Now(),
	}
	_ = storage
	_ = ctx
	_ = cfg
}

func intPtr(i int) *int { return &i } 