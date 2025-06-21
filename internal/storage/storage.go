package storage

import (
	"context"
	"github.com/mm-rules/matchmaking/internal/models"
)

type Storage interface {
	StoreGameConfig(ctx context.Context, config *models.GameConfig) error
	GetGameConfig(ctx context.Context, gameID string) (*models.GameConfig, error)
	StoreMatchRequest(ctx context.Context, request *models.MatchRequest) error
	GetMatchRequest(ctx context.Context, requestID string) (*models.MatchRequest, error)
	GetGameQueue(ctx context.Context, gameID string) ([]*models.MatchRequest, error)
	GetMatchStatus(ctx context.Context, requestID string) (*models.MatchStatusResponse, error)
	StoreMatch(ctx context.Context, match *models.Match) error
	UpdateMatchRequestStatus(ctx context.Context, requestID string, status models.MatchStatus) error
	GetStats(ctx context.Context) (map[string]interface{}, error)
	Ping(ctx context.Context) error
} 