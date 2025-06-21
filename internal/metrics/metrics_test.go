package metrics

import (
	"testing"
)

func TestRecordMatchRequest(t *testing.T) {
	RecordMatchRequest("game1", "pending")
}

func TestRecordMatchCreated(t *testing.T) {
	RecordMatchCreated("game1", 2)
}

func TestSetQueueSize(t *testing.T) {
	SetQueueSize("game1", 5)
}

func TestRecordAllocationRequest(t *testing.T) {
	RecordAllocationRequest("game1", "success")
}

func TestRecordMatchmakingDuration(t *testing.T) {
	RecordMatchmakingDuration("game1", 0.123)
}

func TestRecordAllocationDuration(t *testing.T) {
	RecordAllocationDuration("game1", 0.456)
}

func TestRecordRuleEvaluation(t *testing.T) {
	RecordRuleEvaluation("game1", "level", "pass")
}

func TestSetActiveSessions(t *testing.T) {
	SetActiveSessions("game1", 3)
}

func TestRecordHTTPRequest(t *testing.T) {
	RecordHTTPRequest("GET", "/health", "200", 0.01)
} 