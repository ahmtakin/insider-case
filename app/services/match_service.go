package services

import (
	"fmt"
	"insider-case/app/models"
	"insider-case/app/repository"
	"math/rand"
	"time"
)

const homeAdvantageMultiplier = 1.1 // 10% advantage to home team

type IMatchService interface {
	GetMatchesByLeagueIdAndWeek(leagueID uint, week int) ([]models.Match, error)
	GetMatchesByLeagueId(leagueID uint) ([]models.Match, error)
	PlayMatch(match models.Match) error
	SimulateMatch(match models.Match) error
}

type MatchService struct {
	matchRepo     repository.IMatchRepository
	teamRepo      repository.ITeamRepository
	teamStatsRepo repository.ITeamStatsRepository
}

var _ IMatchService = &MatchService{}

func NewMatchService(matchRepo repository.IMatchRepository, teamRepo repository.ITeamRepository, teamStatsRepo repository.ITeamStatsRepository) *MatchService {
	if matchRepo == nil || teamRepo == nil || teamStatsRepo == nil {
		fmt.Println("repositories not initialized")
		return nil
	}
	return &MatchService{
		matchRepo:     matchRepo,
		teamRepo:      teamRepo,
		teamStatsRepo: teamStatsRepo,
	}
}
func (s *MatchService) GetMatchesByLeagueIdAndWeek(leagueID uint, week int) ([]models.Match, error) {

	fmt.Println("Fetching matches for league:", leagueID, "week:", week)
	matches, err := s.matchRepo.GetMatchesByLeagueIdAndWeek(leagueID, week)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches for league %d and week %d: %w", leagueID, week, err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no matches found for league %d and week %d", leagueID, week)
	}
	return matches, nil
}

func (s *MatchService) GetMatchesByLeagueId(leagueID uint) ([]models.Match, error) {
	matches, err := s.matchRepo.GetMatchesByLeagueId(leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches for league %d: %w", leagueID, err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no matches found for league %d", leagueID)
	}
	return matches, nil
}

func (s *MatchService) PlayMatch(match models.Match) error {
	fmt.Println("Playing match:", match.ID, "between teams:", match.HomeTeamID, "and", match.AwayTeamID)

	if err := s.matchRepo.SaveMatch(match); err != nil {
		return fmt.Errorf("failed to play match %d: %w", match.ID, err)
	}

	// Update team stats after match
	if err := s.updateTeamStats(match); err != nil {
		return fmt.Errorf("failed to update team stats for match %d: %w", match.ID, err)
	}

	return nil
}

func (s *MatchService) updateTeamStats(match models.Match) error {
	homeTeamStats, err := s.teamStatsRepo.GetTeamStatsByTeamID(match.HomeTeamID)
	if err != nil {
		return fmt.Errorf("failed to get home team %d: %w", match.HomeTeamID, err)
	}

	awayTeamStats, err := s.teamStatsRepo.GetTeamStatsByTeamID(match.AwayTeamID)
	if err != nil {
		return fmt.Errorf("failed to get away team %d: %w", match.AwayTeamID, err)
	}

	// Update stats logic here
	// ...
	homeTeamStats.Played++
	awayTeamStats.Played++

	if match.HomeScore > match.AwayScore {
		homeTeamStats.Won++
		awayTeamStats.Lost++
		homeTeamStats.Points += 3

	} else if match.HomeScore < match.AwayScore {
		homeTeamStats.Lost++
		awayTeamStats.Won++
		awayTeamStats.Points += 3
	} else {
		homeTeamStats.Draw++
		awayTeamStats.Draw++
		homeTeamStats.Points += 1
		awayTeamStats.Points += 1
	}

	homeTeamStats.GoalsFor += match.HomeScore
	awayTeamStats.GoalsFor += match.AwayScore
	homeTeamStats.GoalsAgainst += match.AwayScore
	awayTeamStats.GoalsAgainst += match.HomeScore
	homeTeamStats.GoalDiff += homeTeamStats.GoalsFor - homeTeamStats.GoalsAgainst
	awayTeamStats.GoalDiff += awayTeamStats.GoalsFor - awayTeamStats.GoalsAgainst

	//if match.Week >3 estimation logic here
	s.teamStatsRepo.UpdateTeamStats(homeTeamStats)
	s.teamStatsRepo.UpdateTeamStats(awayTeamStats)

	return nil
}

func (s *MatchService) SimulateMatch(match models.Match) error {
	// Seed the RNG
	rand.New(rand.NewSource(time.Now().UnixNano()))
	fmt.Println("Simulating match:", match.ID, "between teams:", match.HomeTeamID, "and", match.AwayTeamID)

	intHomeStrength, err := s.teamRepo.GetTeamStrengthByID(match.HomeTeamID)
	if err != nil {
		return fmt.Errorf("failed to get home team strength: %w", err)
	}
	// If away team is nil, return an error
	intAwayStrength, err := s.teamRepo.GetTeamStrengthByID(match.AwayTeamID)
	if err != nil {
		return fmt.Errorf("failed to get away team strength: %w", err)
	}
	homeStrength := float64(intHomeStrength) * homeAdvantageMultiplier
	// Adjust strength with home advantage
	awayStrength := float64(intAwayStrength)

	// Total strength for probability calculations
	totalStrength := homeStrength + awayStrength

	// Compute win probabilities
	homeWinChance := homeStrength / totalStrength
	drawChance := 0.2 // fixed draw chance

	// Simulate match result
	outcome := rand.Float64()

	var homeGoals, awayGoals int
	switch {
	case outcome < homeWinChance:
		homeGoals = rand.Intn(3) + 1// 1 to 3
		awayGoals = rand.Intn(3)     // 0 to 1
		match.Result = &match.HomeTeamID
	case outcome < homeWinChance+drawChance:
		goals := rand.Intn(3) // 0 to 2
		homeGoals = goals
		awayGoals = goals
		match.Result = nil // 0 indicates a draw
	default:
		awayGoals = rand.Intn(3) + 1  // 1 to 3
		homeGoals = rand.Intn(3)     // 0 to 1
		match.Result = &match.AwayTeamID
	}

	match.HomeScore = homeGoals
	match.AwayScore = awayGoals
	match.Played = true

	if err := s.matchRepo.SaveMatch(match); err != nil {
		return fmt.Errorf("failed to play match %d: %w", match.ID, err)
	}
	if err := s.updateTeamStats(match); err != nil {
		return fmt.Errorf("failed to update team stats for match %d: %w", match.ID, err)
	}

	return nil

}
