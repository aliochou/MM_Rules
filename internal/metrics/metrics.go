package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// MatchRequestCounter counts total match requests
	MatchRequestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mm_rules_match_requests_total",
			Help: "Total number of match requests",
		},
		[]string{"game_id", "status"},
	)

	// MatchCreatedCounter counts total matches created
	MatchCreatedCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mm_rules_matches_created_total",
			Help: "Total number of matches created",
		},
		[]string{"game_id", "team_size"},
	)

	// QueueSizeGauge tracks current queue size
	QueueSizeGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mm_rules_queue_size",
			Help: "Current number of players in queue",
		},
		[]string{"game_id"},
	)

	// AllocationRequestCounter counts allocation requests
	AllocationRequestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mm_rules_allocation_requests_total",
			Help: "Total number of allocation requests",
		},
		[]string{"game_id", "status"},
	)

	// MatchmakingDurationHistogram tracks matchmaking processing time
	MatchmakingDurationHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mm_rules_matchmaking_duration_seconds",
			Help:    "Time spent processing matchmaking",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"game_id"},
	)

	// AllocationDurationHistogram tracks allocation processing time
	AllocationDurationHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mm_rules_allocation_duration_seconds",
			Help:    "Time spent processing allocation",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"game_id"},
	)

	// RuleEvaluationCounter counts rule evaluations
	RuleEvaluationCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mm_rules_rule_evaluations_total",
			Help: "Total number of rule evaluations",
		},
		[]string{"game_id", "rule_type", "result"},
	)

	// ActiveSessionsGauge tracks active game sessions
	ActiveSessionsGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mm_rules_active_sessions",
			Help: "Number of active game sessions",
		},
		[]string{"game_id"},
	)

	// HTTPRequestDurationHistogram tracks HTTP request processing time
	HTTPRequestDurationHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mm_rules_http_request_duration_seconds",
			Help:    "Time spent processing HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTPRequestCounter counts HTTP requests
	HTTPRequestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mm_rules_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)
)

// RecordMatchRequest records a new match request
func RecordMatchRequest(gameID, status string) {
	MatchRequestCounter.WithLabelValues(gameID, status).Inc()
}

// RecordMatchCreated records a new match creation
func RecordMatchCreated(gameID string, teamSize int) {
	MatchCreatedCounter.WithLabelValues(gameID, string(rune(teamSize+'0'))).Inc()
}

// SetQueueSize sets the current queue size
func SetQueueSize(gameID string, size int) {
	QueueSizeGauge.WithLabelValues(gameID).Set(float64(size))
}

// RecordAllocationRequest records an allocation request
func RecordAllocationRequest(gameID, status string) {
	AllocationRequestCounter.WithLabelValues(gameID, status).Inc()
}

// RecordMatchmakingDuration records matchmaking processing time
func RecordMatchmakingDuration(gameID string, duration float64) {
	MatchmakingDurationHistogram.WithLabelValues(gameID).Observe(duration)
}

// RecordAllocationDuration records allocation processing time
func RecordAllocationDuration(gameID string, duration float64) {
	AllocationDurationHistogram.WithLabelValues(gameID).Observe(duration)
}

// RecordRuleEvaluation records a rule evaluation
func RecordRuleEvaluation(gameID, ruleType, result string) {
	RuleEvaluationCounter.WithLabelValues(gameID, ruleType, result).Inc()
}

// SetActiveSessions sets the number of active sessions
func SetActiveSessions(gameID string, count int) {
	ActiveSessionsGauge.WithLabelValues(gameID).Set(float64(count))
}

// RecordHTTPRequest records an HTTP request
func RecordHTTPRequest(method, endpoint, status string, duration float64) {
	HTTPRequestCounter.WithLabelValues(method, endpoint, status).Inc()
	HTTPRequestDurationHistogram.WithLabelValues(method, endpoint, status).Observe(duration)
} 