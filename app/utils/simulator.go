package utils

import (
	"fmt"
	"insider-case/app/dto"
	"insider-case/app/models"
	"math/rand"
	"sort"
	"sync"
	"time"
)

const (
	homeAdvantageMultiplier = 1.1
	simulationIterations    = 10000
)

// EstimateChampionshipProbabilities runs Monte Carlo simulations to estimate championship probabilities
func EstimateChampionshipProbabilities(leagueState dto.LeagueState) ([]dto.ChampionshipEstimation, error) {
	if len(leagueState.TeamStats) == 0 {
		return nil, fmt.Errorf("no team stats provided")
	}

	// Initialize championship counts
	championshipCounts := make(map[uint]int)
	for _, team := range leagueState.Teams {
		championshipCounts[team.ID] = 0
	}

	// Run simulations in parallel
	var wg sync.WaitGroup
	results := make(chan map[uint]int, simulationIterations)
	errors := make(chan error, simulationIterations)

	for i := 0; i < simulationIterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			finalStandings := simulateRemainingSeason(leagueState.TeamStats, leagueState.RemainingMatches, leagueState.Teams)
			championID := DetermineChampion(finalStandings)
			results <- map[uint]int{championID: 1}
		}()
	}

	// Collect results
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	// Aggregate results
	for result := range results {
		for teamID, count := range result {
			championshipCounts[teamID] += count
		}
	}
	for championshipCount := range championshipCounts {
		fmt.Printf("Team ID %d has %d championships\n", championshipCount, championshipCounts[championshipCount])
	}

	// Check if any team has zero championships
	// Check for errors
	for err := range errors {
		if err != nil {
			return nil, fmt.Errorf("simulation error: %w", err)
		}
	}

	// Calculate probabilities and create response
	estimations := make([]dto.ChampionshipEstimation, 0, len(leagueState.TeamStats))
	for _, stat := range leagueState.TeamStats {
		probability := float32(championshipCounts[stat.TeamID]) / float32(simulationIterations)
		estimations = append(estimations, dto.ChampionshipEstimation{
			LeagueID:   leagueState.LeagueID,
			Week:       leagueState.Week,
			TeamID:     stat.TeamID,
			Estimation: probability,
		})
	}

	return estimations, nil
}

// simulateRemainingSeason simulates all remaining matches and returns final standings
func simulateRemainingSeason(currentStats []models.TeamStats, remainingMatches []models.Match, teams []models.Team) []models.TeamStats {
	// Create a copy of current stats to avoid modifying the original
	simulatedStats := make([]models.TeamStats, len(currentStats))
	copy(simulatedStats, currentStats)

	// Create a map for quick lookup of team stats
	statsMap := make(map[uint]*models.TeamStats)
	for i := range simulatedStats {
		statsMap[simulatedStats[i].TeamID] = &simulatedStats[i]
	}

	// Simulate each remaining match
	for _, match := range remainingMatches {
		homeStats := statsMap[match.HomeTeamID]
		awayStats := statsMap[match.AwayTeamID]

		teamsPlaying := []models.Team{}
		for _, team := range teams {
			if team.ID == match.HomeTeamID || team.ID == match.AwayTeamID {
				teamsPlaying = append(teamsPlaying, team)
			}
		}

		if len(teamsPlaying) != 2 {
			fmt.Printf("Skipping match %d due to missing team stats\n", match.ID)
			continue
		}

		homeGoals, awayGoals := simulateMatch(match, teamsPlaying, currentStats)

		// Update stats based on match result
		homeStats.Played++
		awayStats.Played++
		homeStats.GoalsFor += homeGoals
		awayStats.GoalsFor += awayGoals
		homeStats.GoalsAgainst += awayGoals
		awayStats.GoalsAgainst += homeGoals
		homeStats.GoalDiff += homeGoals - awayGoals
		awayStats.GoalDiff += awayGoals - homeGoals

		if homeGoals > awayGoals {
			homeStats.Won++
			awayStats.Lost++
			homeStats.Points += 3
		} else if homeGoals < awayGoals {
			homeStats.Lost++
			awayStats.Won++
			awayStats.Points += 3
		} else {
			homeStats.Draw++
			awayStats.Draw++
			homeStats.Points++
			awayStats.Points++
		}
	}

	return simulatedStats
}

func CalculateFormFactor(teams []models.Team, stats []models.TeamStats, homeID uint) float64 {

	statsMap := make(map[uint]models.TeamStats)
	for _, stat := range stats {
		statsMap[stat.TeamID] = stat
	}

	var homeTeam, awayTeam models.TeamStats
	for _, team := range teams {
		if team.ID == homeID {
			homeTeam = statsMap[team.ID]
		} else {
			awayTeam = statsMap[team.ID]
		}
	}
	// Calculate form factor based on points and played matches
	if homeTeam.Played == 0 || awayTeam.Played == 0 {
		return 1.0 // Avoid division by zero
	}

	homeForm := float64(homeTeam.Won) / float64(homeTeam.Played)
	awayForm := float64(awayTeam.Won) / float64(awayTeam.Played)

	return homeForm / awayForm
}

// simulateMatch simulates a single match and returns the score
func simulateMatch(match models.Match, teams []models.Team, teamStats []models.TeamStats) (int, int) {
	var homeStrength int
	var awayStrength int
	// Find team strengths

	for _, team := range teams {
		if team.ID == match.HomeTeamID {
			homeStrength = team.Strength
			break
		} else if team.ID == match.AwayTeamID {
			awayStrength = team.Strength
			break
		}
	}
	// Calculate form factor
	formFactor := CalculateFormFactor(teams, teamStats, match.HomeTeamID)
	if formFactor <= 0 {
		formFactor = 1.0 // Default form factor if calculation fails
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Adjust win probability with form and home advantage
	totalStrength := (float64(homeStrength) * homeAdvantageMultiplier * formFactor) + float64(awayStrength)
	homeWinProb := float64(homeStrength) / totalStrength * 0.7 // 70% weight for home advantage
	drawProb := 0.2                                            // Fixed draw probability

	// Simulate match result
	outcome := r.Float64()
	var homeGoals, awayGoals int

	switch {
	case outcome < homeWinProb:
		homeGoals = r.Intn(3) + 1 // 1 to 3
		awayGoals = r.Intn(2)     // 0 to 1
	case outcome < homeWinProb+drawProb:
		goals := r.Intn(3) // 0 to 2
		homeGoals = goals
		awayGoals = goals
	default:
		awayGoals = r.Intn(3) + 1 // 1 to 3
		homeGoals = r.Intn(2)     // 0 to 1
	}

	return homeGoals, awayGoals
}

// determineChampion determines the champion based on final standings
func DetermineChampion(finalStandings []models.TeamStats) uint {
	// Sort teams by points, goal difference, and goals scored
	sort.Slice(finalStandings, func(i, j int) bool {
		if finalStandings[i].Points != finalStandings[j].Points {
			return finalStandings[i].Points > finalStandings[j].Points
		}
		if finalStandings[i].GoalDiff != finalStandings[j].GoalDiff {
			return finalStandings[i].GoalDiff > finalStandings[j].GoalDiff
		}
		return finalStandings[i].GoalsFor > finalStandings[j].GoalsFor
	})

	return finalStandings[0].TeamID
}
