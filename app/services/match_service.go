package services

import (
	"fmt"
	"insider-case/app/dto"
	"insider-case/app/models"
	"insider-case/app/repository"

	"insider-case/app/utils"
	"math/rand"
	"time"
)

const homeAdvantageMultiplier = 1.1 // 10% advantage to home team

type IMatchService interface {
	GetMatchesByLeagueIdAndWeek(leagueID uint, week int) ([]models.Match, error)
	GetMatchesByLeagueId(leagueID uint) ([]models.Match, error)
	// PlayMatch(match models.Match) error
	SimulateMatch(match models.Match) (models.Match, error)
	UserPlayMatch(week dto.UserPlayedMatch) (models.Match, error)
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
	homeTeamStats.GoalDiff += match.HomeScore - match.AwayScore
	awayTeamStats.GoalDiff += match.AwayScore - match.HomeScore

	s.teamStatsRepo.UpdateTeamStats(homeTeamStats)
	s.teamStatsRepo.UpdateTeamStats(awayTeamStats)

	return nil
}

func (s *MatchService) SimulateMatch(match models.Match) (models.Match, error) {
	// Seed the RNG
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fmt.Println("Simulating match:", match.ID, "between teams:", match.HomeTeamID, "and", match.AwayTeamID)

	homeTeam, err := s.teamRepo.GetTeamByID(match.HomeTeamID)
	if err != nil {
		return match, fmt.Errorf("failed to get home team %d: %w", match.HomeTeamID, err)
	}
	awayTeam, err := s.teamRepo.GetTeamByID(match.AwayTeamID)
	if err != nil {
		return match, fmt.Errorf("failed to get away team %d: %w", match.AwayTeamID, err)
	}
	teams := append([]models.Team{}, homeTeam, awayTeam)
	teamStats, err := s.teamStatsRepo.GetTeamStatsByLeagueID(match.LeagueID)
	if err != nil {
		return match, fmt.Errorf("failed to get team stats for league %d: %w", match.LeagueID, err)
	}
	formFactor := utils.CalculateFormFactor(teams, teamStats, match.HomeTeamID)
	if formFactor <= 0 {
		formFactor = 1.0 // Default form factor if calculation fails
	}

	// Total strength for probability calculations
	totalStrength := (float64(homeTeam.Strength) * homeAdvantageMultiplier * formFactor) + float64(awayTeam.Strength)

	// Compute win probabilities
	homeWinChance := float64(homeTeam.Strength) / totalStrength * 0.8
	drawChance := 0.2 // fixed draw chance

	// Simulate match result
	outcome := rand.Float64()

	var homeGoals, awayGoals int
	switch {
	case outcome < homeWinChance:
		homeGoals = r.Intn(3) + 1 // 1 to 3
		awayGoals = r.Intn(3)     // 0 to 1
	case outcome < homeWinChance+drawChance:
		goals := r.Intn(3) // 0 to 2
		homeGoals = goals
		awayGoals = goals
	default:
		awayGoals = r.Intn(3) + 1 // 1 to 3
		homeGoals = r.Intn(3)     // 0 to 1
	}
	s.setMatchWinner(&match)

	match.HomeScore = homeGoals
	match.AwayScore = awayGoals
	match.Played = true

	if err := s.matchRepo.SaveMatch(match); err != nil {
		return match, fmt.Errorf("failed to save simulated match %d: %w", match.ID, err)
	}
	if err := s.updateTeamStats(match); err != nil {
		return match, fmt.Errorf("failed to update team stats for match %d: %w", match.ID, err)
	}

	return match, nil

}
func (s *MatchService) UserPlayMatch(match dto.UserPlayedMatch) (models.Match, error) {
	existingMatch, err := s.matchRepo.GetMatchByID(match.MatchID)
	if err != nil {
		return models.Match{}, fmt.Errorf("failed to get match %d: %w", match.MatchID, err)
	}
	if existingMatch == nil {
		return models.Match{}, fmt.Errorf("match with ID %d not found", match.MatchID)
	}
	existingMatch.HomeScore = match.HomeScore
	existingMatch.AwayScore = match.AwayScore
	existingMatch.Played = true
	s.setMatchWinner(existingMatch)
	if err := s.matchRepo.SaveMatch(*existingMatch); err != nil {
		return models.Match{}, fmt.Errorf("failed to save match %d: %w", match.MatchID, err)
	}
	if err := s.updateTeamStats(*existingMatch); err != nil {
		return models.Match{}, fmt.Errorf("failed to update team stats for match %d: %w", match.MatchID, err)
	}
	return *existingMatch, nil

}
func (s *MatchService) setMatchWinner(match *models.Match) {
	if match.HomeScore > match.AwayScore {
		match.Result = &match.HomeTeamID
	} else if match.AwayScore > match.HomeScore {
		match.Result = &match.AwayTeamID
	} else {
		match.Result = nil // Draw
	}
}
