package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mm-rules/matchmaking/internal/models"
	"github.com/mm-rules/matchmaking/internal/matchmaker"
	"github.com/mm-rules/matchmaking/internal/engine"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorage implements the storage interface for testing
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) StoreGameConfig(ctx context.Context, config *models.GameConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockStorage) GetGameConfig(ctx context.Context, gameID string) (*models.GameConfig, error) {
	args := m.Called(ctx, gameID)
	return args.Get(0).(*models.GameConfig), args.Error(1)
}

func (m *MockStorage) StoreMatchRequest(ctx context.Context, request *models.MatchRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockStorage) GetMatchRequest(ctx context.Context, requestID string) (*models.MatchRequest, error) {
	args := m.Called(ctx, requestID)
	return args.Get(0).(*models.MatchRequest), args.Error(1)
}

func (m *MockStorage) GetGameQueue(ctx context.Context, gameID string) ([]*models.MatchRequest, error) {
	args := m.Called(ctx, gameID)
	return args.Get(0).([]*models.MatchRequest), args.Error(1)
}

func (m *MockStorage) GetMatchStatus(ctx context.Context, requestID string) (*models.MatchStatusResponse, error) {
	args := m.Called(ctx, requestID)
	return args.Get(0).(*models.MatchStatusResponse), args.Error(1)
}

func (m *MockStorage) StoreMatch(ctx context.Context, match *models.Match) error {
	args := m.Called(ctx, match)
	return args.Error(0)
}

func (m *MockStorage) UpdateMatchRequestStatus(ctx context.Context, requestID string, status models.MatchStatus) error {
	args := m.Called(ctx, requestID, status)
	return args.Error(0)
}

func (m *MockStorage) GetStats(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockStorage) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockAllocator struct {
	mock.Mock
}

func (m *MockAllocator) AllocateSession(match *models.Match) (*models.GameSession, error) {
	args := m.Called(match)
	return args.Get(0).(*models.GameSession), args.Error(1)
}

func (m *MockAllocator) AllocateSessionWithRetry(match *models.Match, maxRetries int, retryDelay time.Duration) (*models.GameSession, error) {
	args := m.Called(match, maxRetries, retryDelay)
	return args.Get(0).(*models.GameSession), args.Error(1)
}

func (m *MockAllocator) ValidateAllocationRequest(req *models.AllocationRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func setupTestHandler() (*Handler, *MockStorage, *MockAllocator) {
	gin.SetMode(gin.TestMode)
	
	mockStorage := &MockStorage{}
	mockAllocator := &MockAllocator{}
	logger := logrus.New()
	
	// Create handler with mock storage
	handler := &Handler{
		storage:    mockStorage,
		matchmaker: matchmaker.NewMatchmaker(),
		ruleEngine: engine.NewRuleEngine(),
		allocator:  mockAllocator,
		logger:     logger,
	}
	
	return handler, mockStorage, mockAllocator
}

func TestHandler_HealthCheck(t *testing.T) {
	handler, mockStorage, _ := setupTestHandler()

	mockStorage.On("Ping", mock.Anything).Return(nil)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	
	handler.HealthCheck(ctx)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"healthy"`)
}

func TestHandler_CreateGameConfig_Success(t *testing.T) {
	handler, mockStorage, _ := setupTestHandler()
	
	config := &models.GameConfig{
		GameID: "test-game",
		Teams: []models.Team{
			{Name: "team1", Size: 2},
		},
		Rules: []models.Rule{{Field: "level", Min: &[]int{10}[0]}},
	}
	
	mockStorage.On("StoreGameConfig", mock.Anything, mock.AnythingOfType("*models.GameConfig")).Return(nil)
	
	body, _ := json.Marshal(config)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/rules/test-game", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "game_id", Value: "test-game"}}
	
	handler.CreateGameConfig(ctx)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	mockStorage.AssertExpectations(t)
}

func TestHandler_CreateGameConfig_InvalidJSON(t *testing.T) {
	handler, _, _ := setupTestHandler()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/rules/test-game", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "game_id", Value: "test-game"}}
	
	handler.CreateGameConfig(ctx)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_CreateGameConfig_StorageError(t *testing.T) {
	handler, mockStorage, _ := setupTestHandler()
	
	config := &models.GameConfig{
		GameID: "test-game",
		Teams: []models.Team{
			{Name: "team1", Size: 2},
		},
		Rules: []models.Rule{{Field: "level", Min: &[]int{10}[0]}},
	}
	
	mockStorage.On("StoreGameConfig", mock.Anything, mock.AnythingOfType("*models.GameConfig")).Return(assert.AnError)
	
	body, _ := json.Marshal(config)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/rules/test-game", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "game_id", Value: "test-game"}}
	
	handler.CreateGameConfig(ctx)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockStorage.AssertExpectations(t)
}

func TestHandler_CreateMatchRequest_Success(t *testing.T) {
	handler, mockStorage, _ := setupTestHandler()
	
	request := &MatchRequestRequest{
		PlayerID: "player1",
		GameID:   "test-game",
		Metadata: map[string]interface{}{"level": 10},
	}
	
	mockStorage.On("StoreMatchRequest", mock.Anything, mock.AnythingOfType("*models.MatchRequest")).Return(nil)
	
	body, _ := json.Marshal(request)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/match-request", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	
	handler.CreateMatchRequest(ctx)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response["request_id"])
	assert.Equal(t, "pending", response["status"])
	
	mockStorage.AssertExpectations(t)
}

func TestHandler_CreateMatchRequest_InvalidJSON(t *testing.T) {
	handler, _, _ := setupTestHandler()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/match-request", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	
	handler.CreateMatchRequest(ctx)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_CreateMatchRequest_StorageError(t *testing.T) {
	handler, mockStorage, _ := setupTestHandler()
	
	request := &MatchRequestRequest{
		PlayerID: "player1",
		GameID:   "test-game",
		Metadata: map[string]interface{}{"level": 10},
	}
	
	mockStorage.On("StoreMatchRequest", mock.Anything, mock.AnythingOfType("*models.MatchRequest")).Return(assert.AnError)
	
	body, _ := json.Marshal(request)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/match-request", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	
	handler.CreateMatchRequest(ctx)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockStorage.AssertExpectations(t)
}

func TestHandler_GetMatchStatus_Success(t *testing.T) {
	handler, mockStorage, _ := setupTestHandler()
	
	statusResponse := &models.MatchStatusResponse{
		Status: models.StatusMatched,
	}
	
	mockStorage.On("GetMatchStatus", mock.Anything, "req1").Return(statusResponse, nil)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/match-status/req1", nil)
	
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "request_id", Value: "req1"}}
	
	handler.GetMatchStatus(ctx)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.MatchStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, string(models.StatusMatched), string(response.Status))
	
	mockStorage.AssertExpectations(t)
}

func TestHandler_GetMatchStatus_NotFound(t *testing.T) {
	handler, mockStorage, _ := setupTestHandler()
	
	mockStorage.On("GetMatchStatus", mock.Anything, "req1").Return((*models.MatchStatusResponse)(nil), assert.AnError)
	mockStorage.On("GetMatchRequest", mock.Anything, "req1").Return((*models.MatchRequest)(nil), assert.AnError)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/match-status/req1", nil)
	
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "request_id", Value: "req1"}}
	
	handler.GetMatchStatus(ctx)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockStorage.AssertExpectations(t)
}

func TestHandler_ProcessMatchmaking_Success(t *testing.T) {
	handler, mockStorage, _ := setupTestHandler()
	
	config := &models.GameConfig{
		GameID: "test-game",
		Teams: []models.Team{
			{Name: "team1", Size: 2},
		},
		Rules: []models.Rule{{Field: "level", Min: &[]int{10}[0]}},
	}
	
	requests := []*models.MatchRequest{
		{
			ID:       "req1",
			PlayerID: "player1",
			GameID:   "test-game",
			Status:   models.StatusPending,
		},
		{
			ID:       "req2",
			PlayerID: "player2",
			GameID:   "test-game",
			Status:   models.StatusPending,
		},
	}
	
	mockStorage.On("GetGameConfig", mock.Anything, "test-game").Return(config, nil)
	mockStorage.On("GetGameQueue", mock.Anything, "test-game").Return(requests, nil)
	mockStorage.On("StoreMatch", mock.Anything, mock.AnythingOfType("*models.Match")).Return(nil)
	mockStorage.On("UpdateMatchRequestStatus", mock.Anything, "player1", models.StatusMatched).Return(nil)
	mockStorage.On("UpdateMatchRequestStatus", mock.Anything, "player2", models.StatusMatched).Return(nil)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/process-matchmaking/test-game", nil)
	
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "game_id", Value: "test-game"}}
	
	handler.ProcessMatchmaking(ctx)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockStorage.AssertExpectations(t)
}

func TestHandler_ProcessMatchmaking_GameConfigNotFound(t *testing.T) {
	handler, mockStorage, _ := setupTestHandler()
	
	mockStorage.On("GetGameConfig", mock.Anything, "test-game").Return((*models.GameConfig)(nil), assert.AnError)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/process-matchmaking/test-game", nil)
	
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "game_id", Value: "test-game"}}
	
	handler.ProcessMatchmaking(ctx)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockStorage.AssertExpectations(t)
}

func TestHandler_AllocateSessions_Success(t *testing.T) {
	handler, _, mockAllocator := setupTestHandler()
	
	matches := []*models.Match{
		{
			ID:       "match1",
			GameID:   "test-game",
			TeamName: "team1",
			Players:  []string{"player1", "player2"},
		},
	}
	
	session := &models.GameSession{
		ID:   "session1",
		IP:   "192.168.1.100",
		Port: 7777,
	}
	
	mockAllocator.On("AllocateSession", mock.MatchedBy(func(m *models.Match) bool {
		return m.ID == "match1" && m.GameID == "test-game" && m.TeamName == "team1"
	})).Return(session, nil)
	
	body, _ := json.Marshal(matches)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/allocate-sessions?game_id=test-game", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "game_id", Value: "test-game"}}
	
	handler.AllocateSessions(ctx)
	
	// Print response for debugging
	println("AllocateSessions response:", w.Body.String())

	assert.Equal(t, http.StatusOK, w.Code)
	mockAllocator.AssertExpectations(t)
}

func TestHandler_AllocateSessions_InvalidJSON(t *testing.T) {
	handler, _, _ := setupTestHandler()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/allocate-sessions", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	
	handler.AllocateSessions(ctx)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_GetStats_Success(t *testing.T) {
	handler, mockStorage, _ := setupTestHandler()

	mockStorage.On("GetStats", mock.Anything).Return(map[string]interface{}{"metrics": 1}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/stats", nil)

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	handler.GetStats(ctx)

	// Print response for debugging
	println("GetStats response:", w.Body.String())

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["storage"])
} 