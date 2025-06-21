package matchmaker

import (
	"fmt"
	"sort"
	"time"

	"github.com/mm-rules/matchmaking/internal/engine"
	"github.com/mm-rules/matchmaking/internal/models"
)

// Matchmaker handles the core matchmaking logic
type Matchmaker struct {
	ruleEngine *engine.RuleEngine
}

// NewMatchmaker creates a new matchmaker instance
func NewMatchmaker() *Matchmaker {
	return &Matchmaker{
		ruleEngine: engine.NewRuleEngine(),
	}
}

// MatchWithRequests represents a match and the corresponding match requests
// that were used to form it.
type MatchWithRequests struct {
	Match         *models.Match
	RequestIDs    []string
}

// ProcessMatchPool processes a pool of players and attempts to form matches
func (m *Matchmaker) ProcessMatchPool(players []*models.MatchRequest, config *models.GameConfig) []*models.Match {
	var matches []*models.Match
	usedPlayers := make(map[string]bool)

	// Sort teams by size (largest first) to prioritize larger matches
	sortedTeams := make([]models.Team, len(config.Teams))
	copy(sortedTeams, config.Teams)
	sort.Slice(sortedTeams, func(i, j int) bool {
		return sortedTeams[i].Size > sortedTeams[j].Size
	})

	for _, team := range sortedTeams {
		matches = append(matches, m.formMatchesForTeam(players, config, team, usedPlayers)...)
	}

	return matches
}

// ProcessMatchPoolWithRequests processes a pool of players and attempts to form matches,
// returning both the matches and the request IDs for each match.
func (m *Matchmaker) ProcessMatchPoolWithRequests(players []*models.MatchRequest, config *models.GameConfig) []MatchWithRequests {
	var results []MatchWithRequests
	usedPlayers := make(map[string]bool)

	// Sort teams by size (largest first) to prioritize larger matches
	sortedTeams := make([]models.Team, len(config.Teams))
	copy(sortedTeams, config.Teams)
	sort.Slice(sortedTeams, func(i, j int) bool {
		return sortedTeams[i].Size > sortedTeams[j].Size
	})

	for _, team := range sortedTeams {
		results = append(results, m.formMatchesForTeamWithRequests(players, config, team, usedPlayers)...)
	}

	return results
}

// formMatchesForTeam attempts to form matches for a specific team configuration
func (m *Matchmaker) formMatchesForTeam(
	players []*models.MatchRequest,
	config *models.GameConfig,
	team models.Team,
	usedPlayers map[string]bool,
) []*models.Match {
	var matches []*models.Match
	availablePlayers := m.getAvailablePlayers(players, usedPlayers)

	// Continue forming matches while we have enough players
	for len(availablePlayers) >= team.Size {
		// Calculate elapsed time for rule relaxation
		oldestPlayer := m.findOldestPlayer(availablePlayers)
		elapsedTime := time.Since(oldestPlayer.CreatedAt)

		// Find compatible players for this team
		compatiblePlayers := m.ruleEngine.FindCompatiblePlayers(availablePlayers, config.Rules, elapsedTime)

		if len(compatiblePlayers) < team.Size {
			break // Not enough compatible players
		}

		// Select the best players for this team
		selectedPlayers := m.selectBestPlayers(compatiblePlayers, team.Size, oldestPlayer.CreatedAt)

		// Create the match
		match := models.NewMatch(config.GameID, team.Name, m.getPlayerIDs(selectedPlayers))
		matches = append(matches, match)

		// Mark selected players as used
		for _, player := range selectedPlayers {
			usedPlayers[player.ID] = true
		}

		// Update available players list
		availablePlayers = m.getAvailablePlayers(players, usedPlayers)
	}

	return matches
}

// formMatchesForTeamWithRequests forms matches for a team and returns both the match and the request IDs.
func (m *Matchmaker) formMatchesForTeamWithRequests(
	players []*models.MatchRequest,
	config *models.GameConfig,
	team models.Team,
	usedPlayers map[string]bool,
) []MatchWithRequests {
	var results []MatchWithRequests
	availablePlayers := m.getAvailablePlayers(players, usedPlayers)

	for len(availablePlayers) >= team.Size {
		oldestPlayer := m.findOldestPlayer(availablePlayers)
		elapsedTime := time.Since(oldestPlayer.CreatedAt)
		compatiblePlayers := m.ruleEngine.FindCompatiblePlayers(availablePlayers, config.Rules, elapsedTime)
		if len(compatiblePlayers) < team.Size {
			break
		}
		selectedPlayers := m.selectBestPlayers(compatiblePlayers, team.Size, oldestPlayer.CreatedAt)
		match := models.NewMatch(config.GameID, team.Name, m.getPlayerIDs(selectedPlayers))
		requestIDs := make([]string, len(selectedPlayers))
		for i, player := range selectedPlayers {
			usedPlayers[player.ID] = true
			requestIDs[i] = player.ID
		}
		results = append(results, MatchWithRequests{
			Match:      match,
			RequestIDs: requestIDs,
		})
		availablePlayers = m.getAvailablePlayers(players, usedPlayers)
	}
	return results
}

// getAvailablePlayers returns players that haven't been used in matches yet
func (m *Matchmaker) getAvailablePlayers(players []*models.MatchRequest, usedPlayers map[string]bool) []*models.MatchRequest {
	var available []*models.MatchRequest
	for _, player := range players {
		if !usedPlayers[player.ID] {
			available = append(available, player)
		}
	}
	return available
}

// findOldestPlayer finds the player who has been waiting the longest
func (m *Matchmaker) findOldestPlayer(players []*models.MatchRequest) *models.MatchRequest {
	if len(players) == 0 {
		return nil
	}

	oldest := players[0]
	for _, player := range players {
		if player.CreatedAt.Before(oldest.CreatedAt) {
			oldest = player
		}
	}
	return oldest
}

// selectBestPlayers selects the best players for a team based on wait time
func (m *Matchmaker) selectBestPlayers(players []*models.MatchRequest, teamSize int, referenceTime time.Time) []*models.MatchRequest {
	// Sort players by wait time (longest waiting first)
	sortedPlayers := make([]*models.MatchRequest, len(players))
	copy(sortedPlayers, players)
	sort.Slice(sortedPlayers, func(i, j int) bool {
		return sortedPlayers[i].CreatedAt.Before(sortedPlayers[j].CreatedAt)
	})

	// Return the first N players (where N is team size)
	if len(sortedPlayers) > teamSize {
		return sortedPlayers[:teamSize]
	}
	return sortedPlayers
}

// getPlayerIDs extracts player IDs from a slice of match requests
func (m *Matchmaker) getPlayerIDs(players []*models.MatchRequest) []string {
	ids := make([]string, len(players))
	for i, player := range players {
		ids[i] = player.PlayerID
	}
	return ids
}

// ValidateMatch validates that a match meets the game configuration requirements
func (m *Matchmaker) ValidateMatch(match *models.Match, config *models.GameConfig) error {
	// Find the team configuration
	var teamConfig *models.Team
	for _, team := range config.Teams {
		if team.Name == match.TeamName {
			teamConfig = &team
			break
		}
	}

	if teamConfig == nil {
		return fmt.Errorf("team '%s' not found in game configuration", match.TeamName)
	}

	if len(match.Players) != teamConfig.Size {
		return fmt.Errorf("team '%s' requires %d players, got %d", match.TeamName, teamConfig.Size, len(match.Players))
	}

	return nil
}

// GetMatchStats returns statistics about the matchmaking process
func (m *Matchmaker) GetMatchStats(players []*models.MatchRequest, matches []*models.Match) map[string]interface{} {
	stats := make(map[string]interface{})
	
	// Calculate wait times
	var waitTimes []float64
	now := time.Now()
	for _, player := range players {
		waitTime := now.Sub(player.CreatedAt).Seconds()
		waitTimes = append(waitTimes, waitTime)
	}

	if len(waitTimes) > 0 {
		sort.Float64s(waitTimes)
		stats["avg_wait_time"] = m.calculateAverage(waitTimes)
		stats["max_wait_time"] = waitTimes[len(waitTimes)-1]
		stats["min_wait_time"] = waitTimes[0]
		stats["median_wait_time"] = m.calculateMedian(waitTimes)
	}

	stats["total_players"] = len(players)
	stats["total_matches"] = len(matches)
	stats["matched_players"] = m.countMatchedPlayers(matches)
	stats["unmatched_players"] = len(players) - m.countMatchedPlayers(matches)

	return stats
}

// calculateAverage calculates the average of a slice of float64 values
func (m *Matchmaker) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateMedian calculates the median of a slice of float64 values
func (m *Matchmaker) calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	if len(values)%2 == 0 {
		mid := len(values) / 2
		return (values[mid-1] + values[mid]) / 2
	}
	return values[len(values)/2]
}

// countMatchedPlayers counts the total number of players in all matches
func (m *Matchmaker) countMatchedPlayers(matches []*models.Match) int {
	count := 0
	for _, match := range matches {
		count += len(match.Players)
	}
	return count
} 