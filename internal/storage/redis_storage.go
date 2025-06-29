package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mm-rules/matchmaking/internal/models"
)

// RedisStorage handles data persistence using Redis
type RedisStorage struct {
	client *redis.Client
}

// NewRedisStorage creates a new Redis storage instance
func NewRedisStorage(addr, password string, db int) *RedisStorage {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisStorage{
		client: client,
	}
}

// Close closes the Redis connection
func (rs *RedisStorage) Close() error {
	return rs.client.Close()
}

// Ping tests the Redis connection
func (rs *RedisStorage) Ping(ctx context.Context) error {
	return rs.client.Ping(ctx).Err()
}

// StoreMatchRequest stores a match request in Redis
func (rs *RedisStorage) StoreMatchRequest(ctx context.Context, request *models.MatchRequest) error {
	// Store the request with its ID as key
	key := fmt.Sprintf("match_request:%s", request.ID)
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal match request: %w", err)
	}

	// Set with expiration (60 seconds)
	err = rs.client.Set(ctx, key, data, 60*time.Second).Err()
	if err != nil {
		return fmt.Errorf("failed to store match request: %w", err)
	}

	// Add to game-specific queue
	queueKey := fmt.Sprintf("game_queue:%s", request.GameID)
	err = rs.client.LPush(ctx, queueKey, request.ID).Err()
	if err != nil {
		return fmt.Errorf("failed to add to game queue: %w", err)
	}

	return nil
}

// GetMatchRequest retrieves a match request by ID
func (rs *RedisStorage) GetMatchRequest(ctx context.Context, requestID string) (*models.MatchRequest, error) {
	key := fmt.Sprintf("match_request:%s", requestID)
	data, err := rs.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("match request not found: %s", requestID)
		}
		return nil, fmt.Errorf("failed to get match request: %w", err)
	}

	var request models.MatchRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal match request: %w", err)
	}

	return &request, nil
}

// GetGameQueue retrieves all pending match requests for a game
func (rs *RedisStorage) GetGameQueue(ctx context.Context, gameID string) ([]*models.MatchRequest, error) {
	queueKey := fmt.Sprintf("game_queue:%s", gameID)

	// Get all request IDs in the queue
	requestIDs, err := rs.client.LRange(ctx, queueKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get game queue: %w", err)
	}

	var requests []*models.MatchRequest
	for _, requestID := range requestIDs {
		request, err := rs.GetMatchRequest(ctx, requestID)
		if err != nil {
			// Skip invalid requests
			continue
		}
		requests = append(requests, request)
	}

	return requests, nil
}

// RemoveFromQueue removes a match request from the game queue
func (rs *RedisStorage) RemoveFromQueue(ctx context.Context, gameID, requestID string) error {
	queueKey := fmt.Sprintf("game_queue:%s", gameID)

	result := rs.client.LRem(ctx, queueKey, 0, requestID)
	if result.Err() != nil {
		return result.Err()
	}

	return nil
}

// UpdateMatchRequestStatus updates the status of a match request
func (rs *RedisStorage) UpdateMatchRequestStatus(ctx context.Context, requestID string, status models.MatchStatus) error {
	request, err := rs.GetMatchRequest(ctx, requestID)
	if err != nil {
		return err
	}

	request.Status = status

	// Store the updated request without re-adding to queue
	key := fmt.Sprintf("match_request:%s", request.ID)
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal match request: %w", err)
	}

	// Set with expiration (60 seconds) but don't add to queue
	return rs.client.Set(ctx, key, data, 60*time.Second).Err()
}

// StoreGameConfig stores a game configuration
func (rs *RedisStorage) StoreGameConfig(ctx context.Context, config *models.GameConfig) error {
	key := fmt.Sprintf("game_config:%s", config.GameID)
	config.UpdatedAt = time.Now()

	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal game config: %w", err)
	}

	// Store without expiration (configs don't expire)
	return rs.client.Set(ctx, key, data, 0).Err()
}

// GetGameConfig retrieves a game configuration
func (rs *RedisStorage) GetGameConfig(ctx context.Context, gameID string) (*models.GameConfig, error) {
	key := fmt.Sprintf("game_config:%s", gameID)
	data, err := rs.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("game config not found: %s", gameID)
		}
		return nil, fmt.Errorf("failed to get game config: %w", err)
	}

	var config models.GameConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal game config: %w", err)
	}

	return &config, nil
}

// StoreMatch stores a completed match
func (rs *RedisStorage) StoreMatch(ctx context.Context, match *models.Match) error {
	key := fmt.Sprintf("match:%s", match.ID)
	data, err := json.Marshal(match)
	if err != nil {
		return fmt.Errorf("failed to marshal match: %w", err)
	}

	// Store with expiration (7 days)
	return rs.client.Set(ctx, key, data, 7*24*time.Hour).Err()
}

// GetMatch retrieves a match by ID
func (rs *RedisStorage) GetMatch(ctx context.Context, matchID string) (*models.Match, error) {
	key := fmt.Sprintf("match:%s", matchID)
	data, err := rs.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("match not found: %s", matchID)
		}
		return nil, fmt.Errorf("failed to get match: %w", err)
	}

	var match models.Match
	if err := json.Unmarshal(data, &match); err != nil {
		return nil, fmt.Errorf("failed to unmarshal match: %w", err)
	}

	return &match, nil
}

// StoreMatchStatus stores match status information
func (rs *RedisStorage) StoreMatchStatus(ctx context.Context, requestID string, status *models.MatchStatusResponse) error {
	key := fmt.Sprintf("match_status:%s", requestID)
	data, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal match status: %w", err)
	}

	// Store with expiration (1 hour)
	return rs.client.Set(ctx, key, data, time.Hour).Err()
}

// GetMatchStatus retrieves match status information
func (rs *RedisStorage) GetMatchStatus(ctx context.Context, requestID string) (*models.MatchStatusResponse, error) {
	key := fmt.Sprintf("match_status:%s", requestID)
	data, err := rs.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("match status not found: %s", requestID)
		}
		return nil, fmt.Errorf("failed to get match status: %w", err)
	}

	var status models.MatchStatusResponse
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal match status: %w", err)
	}

	return &status, nil
}

// CleanupExpiredRequests removes expired match requests
func (rs *RedisStorage) CleanupExpiredRequests(ctx context.Context) error {
	// Get all game configs to find active games
	pattern := "game_config:*"
	keys, err := rs.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get game config keys: %w", err)
	}

	totalRemoved := 0
	for _, key := range keys {
		gameID := key[len("game_config:"):]
		queueKey := fmt.Sprintf("game_queue:%s", gameID)

		// Get all request IDs in the queue
		requestIDs, err := rs.client.LRange(ctx, queueKey, 0, -1).Result()
		if err != nil {
			fmt.Printf("[CLEANUP] Failed to get queue for %s: %v\n", gameID, err)
			continue
		}

		if len(requestIDs) == 0 {
			continue
		}

		fmt.Printf("[CLEANUP] Checking %d requests in queue for %s\n", len(requestIDs), gameID)
		removedCount := 0

		// Check each request and remove if expired
		for _, requestID := range requestIDs {
			requestKey := fmt.Sprintf("match_request:%s", requestID)
			exists, err := rs.client.Exists(ctx, requestKey).Result()
			if err != nil {
				fmt.Printf("[CLEANUP] Error checking request %s: %v\n", requestID, err)
				continue
			}

			if exists == 0 {
				// Request doesn't exist, remove from queue
				result := rs.client.LRem(ctx, queueKey, 0, requestID)
				if result.Err() != nil {
					fmt.Printf("[CLEANUP] Failed to remove request %s from queue: %v\n", requestID, result.Err())
				} else {
					count, _ := result.Result()
					if count > 0 {
						removedCount += int(count)
						fmt.Printf("[CLEANUP] Removed expired request %s from queue\n", requestID)
					}
				}
			}
		}

		if removedCount > 0 {
			fmt.Printf("[CLEANUP] Removed %d expired requests from %s queue\n", removedCount, gameID)
			totalRemoved += removedCount
		}
	}

	if totalRemoved > 0 {
		fmt.Printf("[CLEANUP] Total cleanup: removed %d expired requests across all queues\n", totalRemoved)
	}

	return nil
}

// GetStats returns basic statistics about the storage
func (rs *RedisStorage) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count game configs
	configKeys, err := rs.client.Keys(ctx, "game_config:*").Result()
	if err == nil {
		stats["total_game_configs"] = len(configKeys)
	}

	// Count active queues
	queueKeys, err := rs.client.Keys(ctx, "game_queue:*").Result()
	if err == nil {
		stats["total_game_queues"] = len(queueKeys)
	}

	// Count total pending requests
	totalRequests := 0
	for _, queueKey := range queueKeys {
		count, err := rs.client.LLen(ctx, queueKey).Result()
		if err == nil {
			totalRequests += int(count)
		}
	}
	stats["total_pending_requests"] = totalRequests

	return stats, nil
}

// StoreRequestMatchMapping stores a mapping from requestID to matchID
func (rs *RedisStorage) StoreRequestMatchMapping(ctx context.Context, requestID, matchID string) error {
	key := fmt.Sprintf("request_match:%s", requestID)
	return rs.client.Set(ctx, key, matchID, 7*24*time.Hour).Err()
}

// GetMatchIDForRequest retrieves the matchID for a given requestID
func (rs *RedisStorage) GetMatchIDForRequest(ctx context.Context, requestID string) (string, error) {
	key := fmt.Sprintf("request_match:%s", requestID)
	matchID, err := rs.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return matchID, nil
}

// StoreMultiTeamMatch stores a MultiTeamMatch in Redis
func (rs *RedisStorage) StoreMultiTeamMatch(ctx context.Context, match *models.MultiTeamMatch) error {
	key := fmt.Sprintf("multi_team_match:%s", match.ID)
	data, err := json.Marshal(match)
	if err != nil {
		return fmt.Errorf("failed to marshal multi-team match: %w", err)
	}
	return rs.client.Set(ctx, key, data, 7*24*time.Hour).Err()
}

// GetMultiTeamMatch retrieves a MultiTeamMatch by ID
func (rs *RedisStorage) GetMultiTeamMatch(ctx context.Context, matchID string) (*models.MultiTeamMatch, error) {
	key := fmt.Sprintf("multi_team_match:%s", matchID)
	data, err := rs.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("multi-team match not found: %s", matchID)
		}
		return nil, fmt.Errorf("failed to get multi-team match: %w", err)
	}
	var match models.MultiTeamMatch
	if err := json.Unmarshal(data, &match); err != nil {
		return nil, fmt.Errorf("failed to unmarshal multi-team match: %w", err)
	}
	return &match, nil
}
