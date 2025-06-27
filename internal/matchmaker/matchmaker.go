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
	Match      *models.Match
	RequestIDs []string
}

// MultiTeamMatch represents a match with multiple teams
// (new struct for multi-team matches)
type MultiTeamMatch struct {
	ID        string              `json:"id"`
	GameID    string              `json:"game_id"`
	Teams     map[string][]string `json:"teams"` // team name -> player IDs
	CreatedAt time.Time           `json:"created_at"`
	Session   *models.GameSession `json:"session,omitempty"`
}

// ProcessMatchPool processes a pool of players and attempts to form matches
// Now uses full-team matching by default - only creates matches when all teams can be filled
func (m *Matchmaker) ProcessMatchPool(players []*models.MatchRequest, config *models.GameConfig) []*models.Match {
	var matches []*models.Match
	usedPlayers := make(map[string]bool)
	teamCount := len(config.Teams)
	if teamCount == 0 {
		return matches
	}

	// Continue forming matches while all teams can be filled
	for {
		selected := make(map[string][]*models.MatchRequest) // team name -> players
		usedInThisRound := make(map[string]bool)

		// For each team, try to select enough compatible players
		for _, team := range config.Teams {
			available := m.getAvailablePlayers(players, usedPlayers)
			if len(available) < team.Size {
				goto done // Not enough players for this team
			}

			// Find compatible players for this team
			oldest := m.findOldestPlayer(available)
			elapsed := time.Since(oldest.CreatedAt)
			compatible := m.ruleEngine.FindCompatiblePlayers(available, config.Rules, elapsed)

			// Remove already selected in this round
			var filtered []*models.MatchRequest
			for _, p := range compatible {
				if !usedInThisRound[p.ID] {
					filtered = append(filtered, p)
				}
			}

			if len(filtered) < team.Size {
				goto done // Not enough compatible players for this team
			}

			// Select the best players for this team
			selectedPlayers := m.selectBestPlayers(filtered, team.Size, oldest.CreatedAt)
			selected[team.Name] = selectedPlayers
			for _, p := range selectedPlayers {
				usedInThisRound[p.ID] = true
			}
		}

		// If we get here, all teams are filled - create a single match with all players
		allPlayerIDs := make([]string, 0)
		for _, reqs := range selected {
			for _, req := range reqs {
				usedPlayers[req.ID] = true
				allPlayerIDs = append(allPlayerIDs, req.PlayerID)
			}
		}

		// Create a single match containing all players from all teams
		// Use the first team name as the match team name for backward compatibility
		firstTeamName := config.Teams[0].Name
		match := models.NewMatch(config.GameID, firstTeamName, allPlayerIDs)
		matches = append(matches, match)
	}
done:
	return matches
}

// ProcessMatchPoolWithRequests processes a pool of players and attempts to form matches,
// returning both the matches and the request IDs for each match.
// Now uses full-team matching by default
func (m *Matchmaker) ProcessMatchPoolWithRequests(players []*models.MatchRequest, config *models.GameConfig) []MatchWithRequests {
	var results []MatchWithRequests
	usedPlayers := make(map[string]bool)
	teamCount := len(config.Teams)
	if teamCount == 0 {
		return results
	}

	// Continue forming matches while all teams can be filled
	for {
		// Log the number of requests being processed
		fmt.Printf("[Matchmaker] Processing %d requests\n", len(players))

		selected := make(map[string][]*models.MatchRequest) // team name -> players
		usedInThisRound := make(map[string]bool)

		// For each team, try to select enough compatible players
		for _, team := range config.Teams {
			available := m.getAvailablePlayers(players, usedPlayers)
			if len(available) < team.Size {
				fmt.Printf("[Matchmaker] Not enough available players for team %s: have %d, need %d\n", team.Name, len(available), team.Size)
				goto done // Not enough players for this team
			}

			// Find compatible players for this team
			oldest := m.findOldestPlayer(available)
			elapsed := time.Since(oldest.CreatedAt)
			compatible := m.ruleEngine.FindCompatiblePlayers(available, config.Rules, elapsed)

			// Remove already selected in this round
			var filtered []*models.MatchRequest
			for _, p := range compatible {
				if !usedInThisRound[p.ID] {
					filtered = append(filtered, p)
				}
			}

			if len(filtered) < team.Size {
				fmt.Printf("[Matchmaker] Not enough compatible players for team %s after filtering: have %d, need %d\n", team.Name, len(filtered), team.Size)
				goto done // Not enough compatible players for this team
			}

			// Select the best players for this team
			selectedPlayers := m.selectBestPlayers(filtered, team.Size, oldest.CreatedAt)
			selected[team.Name] = selectedPlayers
			for _, p := range selectedPlayers {
				usedInThisRound[p.ID] = true
			}
		}

		// If we get here, all teams are filled - create a single match with all players and request IDs
		allPlayerIDs := make([]string, 0)
		allRequestIDs := make([]string, 0)
		playerIDSet := make(map[string]bool)
		for _, reqs := range selected {
			for _, req := range reqs {
				usedPlayers[req.ID] = true // Mark this player as used
				if !playerIDSet[req.PlayerID] {
					allPlayerIDs = append(allPlayerIDs, req.PlayerID)
					playerIDSet[req.PlayerID] = true
				}
				allRequestIDs = append(allRequestIDs, req.ID)
			}
		}
		fmt.Printf("[Matchmaker] Forming match with unique player IDs: %v\n", allPlayerIDs)

		// Create a single match containing all players from all teams
		firstTeamName := config.Teams[0].Name
		match := models.NewMatch(config.GameID, firstTeamName, allPlayerIDs)
		results = append(results, MatchWithRequests{
			Match:      match,
			RequestIDs: allRequestIDs,
		})
	}
done:
	return results
}

// ProcessFullTeamMatchPool processes a pool of players and forms matches only when all teams can be filled
// This is now the same as ProcessMatchPool - kept for backward compatibility
func (m *Matchmaker) ProcessFullTeamMatchPool(players []*models.MatchRequest, config *models.GameConfig) []*models.MultiTeamMatch {
	fmt.Printf("[MM] Starting ProcessFullTeamMatchPool: %d players, %d teams\n", len(players), len(config.Teams))
	var matches []*models.MultiTeamMatch
	usedPlayers := make(map[string]bool)
	teamCount := len(config.Teams)
	if teamCount == 0 {
		fmt.Println("[MM] No teams in config, aborting.")
		return matches
	}
	// Continue forming matches while all teams can be filled
	for {
		selected := make(map[string][]*models.MatchRequest) // team name -> players
		usedInThisRound := make(map[string]bool)
		// For each team, try to select enough compatible players
		for _, team := range config.Teams {
			available := m.getAvailablePlayers(players, usedPlayers)
			fmt.Printf("[MM] Team %s: need %d, available %d\n", team.Name, team.Size, len(available))
			if len(available) < team.Size {
				fmt.Printf("[MM] Not enough players for team %s: have %d, need %d\n", team.Name, len(available), team.Size)
				goto done // Not enough players for this team
			}
			// Find compatible players for this team
			oldest := m.findOldestPlayer(available)
			elapsed := time.Since(oldest.CreatedAt)
			compatible := m.ruleEngine.FindCompatiblePlayers(available, config.Rules, elapsed)
			fmt.Printf("[MM] Team %s: compatible %d\n", team.Name, len(compatible))
			// Remove already selected in this round
			var filtered []*models.MatchRequest
			for _, p := range compatible {
				if !usedInThisRound[p.ID] {
					filtered = append(filtered, p)
				}
			}
			fmt.Printf("[MM] Team %s: filtered %d\n", team.Name, len(filtered))
			if len(filtered) < team.Size {
				fmt.Printf("[MM] Not enough compatible players for team %s after filtering: have %d, need %d\n", team.Name, len(filtered), team.Size)
				goto done // Not enough compatible players for this team
			}
			// Select the best players for this team
			selectedPlayers := m.selectBestPlayers(filtered, team.Size, oldest.CreatedAt)
			selected[team.Name] = selectedPlayers
			for _, p := range selectedPlayers {
				usedInThisRound[p.ID] = true
			}
		}
		// If we get here, all teams are filled
		teamMap := make(map[string][]string)
		for teamName, reqs := range selected {
			for _, req := range reqs {
				usedPlayers[req.ID] = true
				teamMap[teamName] = append(teamMap[teamName], req.PlayerID)
			}
		}
		fmt.Printf("[MM] Forming match: %v\n", teamMap)
		match := &models.MultiTeamMatch{
			ID:        models.NewMatch(config.GameID, "multi", nil).ID, // reuse uuid
			GameID:    config.GameID,
			Teams:     teamMap,
			CreatedAt: time.Now(),
		}
		matches = append(matches, match)
	}
done:
	fmt.Printf("[MM] Done. Formed %d matches.\n", len(matches))
	return matches
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

// FlattenTeams flattens all player IDs for all teams
func (m *Matchmaker) FlattenTeams(teams map[string][]string) []string {
	var all []string
	for _, players := range teams {
		all = append(all, players...)
	}
	return all
}
