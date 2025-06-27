package engine

import (
	"fmt"
	"sort"
	"time"

	"github.com/mm-rules/matchmaking/internal/models"
)

// RuleEngine handles the evaluation of matchmaking rules
type RuleEngine struct{}

// NewRuleEngine creates a new rule engine instance
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{}
}

// EvaluatePlayer evaluates a single player against a set of rules
func (re *RuleEngine) EvaluatePlayer(player *models.MatchRequest, rules []models.Rule, elapsedTime time.Duration) (bool, []string) {
	var violations []string

	// Sort rules by priority (higher priority first)
	sortedRules := make([]models.Rule, len(rules))
	copy(sortedRules, rules)
	sort.Slice(sortedRules, func(i, j int) bool {
		return sortedRules[i].Priority > sortedRules[j].Priority
	})

	for _, rule := range sortedRules {
		if !re.evaluateRule(player, rule, elapsedTime) {
			violation := fmt.Sprintf("Rule '%s' failed", rule.Field)
			violations = append(violations, violation)
		}
	}

	return len(violations) == 0, violations
}

// evaluateRule evaluates a single rule against a player
func (re *RuleEngine) evaluateRule(player *models.MatchRequest, rule models.Rule, elapsedTime time.Duration) bool {
	// Check if rule should be relaxed
	if rule.RelaxAfter != nil && elapsedTime.Seconds() >= float64(*rule.RelaxAfter) {
		return true // Rule is relaxed, always pass
	}

	// Get the field value from player metadata
	fieldValue, exists := player.Metadata[rule.Field]
	if !exists {
		return !rule.Strict // If field doesn't exist and rule is not strict, pass
	}

	// Evaluate based on rule type
	switch {
	case rule.Min != nil:
		return re.evaluateMin(fieldValue, *rule.Min)
	case rule.Max != nil:
		return re.evaluateMax(fieldValue, *rule.Max)
	case rule.Contains != nil:
		return re.evaluateContains(fieldValue, *rule.Contains)
	case rule.Equals != nil:
		return re.evaluateEquals(fieldValue, *rule.Equals)
	default:
		return true // No specific evaluation criteria, pass
	}
}

// evaluateMin checks if a value is greater than or equal to min
func (re *RuleEngine) evaluateMin(value interface{}, min int) bool {
	switch v := value.(type) {
	case int:
		return v >= min
	case float64:
		return int(v) >= min
	case string:
		// Try to parse as number
		if num, ok := re.parseNumber(v); ok {
			return num >= min
		}
		return false
	default:
		return false
	}
}

// evaluateMax checks if a value is less than or equal to max
func (re *RuleEngine) evaluateMax(value interface{}, max int) bool {
	switch v := value.(type) {
	case int:
		return v <= max
	case float64:
		return int(v) <= max
	case string:
		// Try to parse as number
		if num, ok := re.parseNumber(v); ok {
			return num <= max
		}
		return false
	default:
		return false
	}
}

// evaluateContains checks if a value contains the specified string
func (re *RuleEngine) evaluateContains(value interface{}, contains string) bool {
	switch v := value.(type) {
	case string:
		// Check if string contains the substring
		return v == contains
	case []interface{}:
		for _, item := range v {
			if fmt.Sprintf("%v", item) == contains {
				return true
			}
		}
		return false
	case []string:
		for _, item := range v {
			if item == contains {
				return true
			}
		}
		return false
	default:
		// For other types, convert to string and check
		return fmt.Sprintf("%v", v) == contains
	}
}

// evaluateEquals checks if a value equals the specified string
func (re *RuleEngine) evaluateEquals(value interface{}, equals string) bool {
	return fmt.Sprintf("%v", value) == equals
}

// parseNumber attempts to parse a string as a number
func (re *RuleEngine) parseNumber(s string) (int, bool) {
	var num int
	_, err := fmt.Sscanf(s, "%d", &num)
	return num, err == nil
}

// FindCompatiblePlayers finds players that are compatible based on rules
func (re *RuleEngine) FindCompatiblePlayers(players []*models.MatchRequest, rules []models.Rule, elapsedTime time.Duration) []*models.MatchRequest {
	var compatible []*models.MatchRequest

	for _, player := range players {
		if valid, _ := re.EvaluatePlayer(player, rules, elapsedTime); valid {
			compatible = append(compatible, player)
		}
	}

	return compatible
}

// ValidateGameConfig validates a game configuration
func (re *RuleEngine) ValidateGameConfig(config *models.GameConfig) error {
	if config.GameID == "" {
		return fmt.Errorf("game_id is required")
	}

	if len(config.Teams) == 0 {
		return fmt.Errorf("at least one team must be defined")
	}

	for i, team := range config.Teams {
		if team.Name == "" {
			return fmt.Errorf("team %d: name is required", i)
		}
		if team.Size <= 0 {
			return fmt.Errorf("team %d: size must be greater than 0", i)
		}
	}

	for i, rule := range config.Rules {
		if rule.Field == "" {
			return fmt.Errorf("rule %d: field is required", i)
		}

		// Check that at least one evaluation criteria is set
		if rule.Min == nil && rule.Max == nil && rule.Contains == nil && rule.Equals == nil {
			return fmt.Errorf("rule %d: at least one evaluation criteria (min, max, contains, equals) must be set", i)
		}
	}

	return nil
}
