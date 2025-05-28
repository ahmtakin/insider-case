package services

import (
	"fmt"
	"insider-case/app/dto"
	"insider-case/app/helpers"
	"insider-case/app/models"
	"insider-case/app/repository"
	"insider-case/app/utils"
)

type ILeagueService interface {
	InitializeLeague(req dto.LeagueCreateRequest) (*dto.LeagueResponse, error)
	SimulateWeek(leagueID uint) (*dto.Week, error)
	PlayRemainingMatches(leagueID uint) ([]*dto.Week, error)
	UserPlayWeek(matches []dto.UserPlayedMatch) (*dto.Week, error)
	GetChampionshipEstimationByLeagueID(leagueID uint) ([]dto.ChampionshipEstimation, error)
}

type LeagueService struct {
	repo          repository.ILeagueRepository
	matchService  IMatchService
	teamStatsRepo repository.ITeamStatsRepository
	weeklyLogRepo repository.IWeeklyLogRepository
	teamRepo      repository.ITeamRepository
}

var _ ILeagueService = &LeagueService{}

func NewLeagueService(repo repository.ILeagueRepository, matchService IMatchService, teamStatsRepo repository.ITeamStatsRepository, weeklyLogRepo repository.IWeeklyLogRepository, teamRepo repository.ITeamRepository) *LeagueService {
	return &LeagueService{
		repo:          repo,
		matchService:  matchService,
		teamStatsRepo: teamStatsRepo,
		weeklyLogRepo: weeklyLogRepo,
		teamRepo:      teamRepo,
	}
}

func (s *LeagueService) InitializeLeague(req dto.LeagueCreateRequest) (*dto.LeagueResponse, error) {
	if err := helpers.ValidateTeamCount(req.TeamCount); err != nil {
		return nil, err
	}
	if err := helpers.ValidateTeamStrength(req.Teams); err != nil {
		return nil, err
	}

	if len(req.Teams) != req.TeamCount {
		return nil, &helpers.ValidationError{
			Field:   "teams",
			Message: fmt.Sprintf("expected %d teams, got %d", req.TeamCount, len(req.Teams)),
		}
	}

	// Convert DTO to model
	league := &models.League{
		Name:      req.Name,
		TeamCount: req.TeamCount,
		MaxWeeks:  helpers.CalculateMaxWeeks(req.TeamCount),
		Teams:     make([]models.Team, len(req.Teams)),
	}

	for i, team := range req.Teams {
		league.Teams[i] = models.Team{
			Name:     team.Name,
			Strength: team.Strength,
		}
	}

	// Call repository
	createdLeague, err := s.repo.InitializeLeague(league)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize league: %w", err)
	}

	// Convert model to response DTO
	return convertToLeagueResponse(createdLeague), nil
}

func convertToLeagueResponse(league *models.League) *dto.LeagueResponse {
	response := &dto.LeagueResponse{
		ID:        league.ID,
		Name:      league.Name,
		TeamCount: league.TeamCount,
		MaxWeeks:  league.MaxWeeks,
		CurrWeek:  league.CurrWeek,
		Teams:     make([]models.Team, len(league.Teams)),
		Matches:   make([]models.Match, len(league.Matches)),
	}

	for i, team := range league.Teams {
		response.Teams[i] = models.Team{
			ID:       team.ID,
			LeagueID: team.LeagueID,
			Name:     team.Name,
			Strength: team.Strength,
			Stats:    team.Stats,
		}
	}
	for i, match := range league.Matches {
		response.Matches[i] = models.Match{
			ID:         match.ID,
			LeagueID:   match.LeagueID,
			Week:       match.Week,
			HomeTeamID: match.HomeTeamID,
			AwayTeamID: match.AwayTeamID,
			HomeScore:  match.HomeScore,
			AwayScore:  match.AwayScore,
			Played:     match.Played,
		}
	}

	return response
}

func (s *LeagueService) SimulateWeek(leagueID uint) (*dto.Week, error) {
	league, err := s.repo.GetLeagueByID(leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get league with ID %d: %w", leagueID, err)
	}

	// Get matches for the current week
	matches, err := s.repo.GetMatchesByLeagueIdAndWeek(leagueID, league.CurrWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches for league %d and week %d: %w", leagueID, league.CurrWeek, err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no matches found for league %d and week %d", leagueID, league.CurrWeek)
	}

	// Play all matches for the current week
	for i, match := range matches {
		if !match.Played {
			simulatedMatch, err := s.matchService.SimulateMatch(match)
			if err != nil {
				return nil, fmt.Errorf("failed to play match %d: %w", match.ID, err)
			}
			matches[i] = simulatedMatch // Update the match in the slice
		}
	}
	// Update championship probabilities if we're past week 3
	if league.CurrWeek > 3 {
		if err := s.updateChampionshipProbabilities(leagueID, league.CurrWeek); err != nil {
			return nil, fmt.Errorf("failed to update championship probabilities: %w", err)
		}
	}
	// Get Team stats after each week
	newStats, err := s.teamStatsRepo.GetTeamStatsByLeagueID(leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team stats for league %d: %w", leagueID, err)
	}
	// Log the weekly results
	if err := s.weeklyLogRepo.SaveWeeklyLog(leagueID, league.CurrWeek); err != nil {
		return nil, fmt.Errorf("failed to log weekly results for league %d and week %d: %w", leagueID, league.CurrWeek, err)
	}

	// Increment the league week
	updatedLeague, err := s.repo.IncrementWeek(leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to increment league week: %w", err)
	}
	if updatedLeague.CurrWeek > updatedLeague.MaxWeeks {
		champion, err := s.getChampionByLeagueID(leagueID)
		if err != nil {
			return nil, fmt.Errorf("failed to get champion for league %d: %w", leagueID, err)
		}
		return &dto.Week{
			LeagueID:  updatedLeague.ID,
			Week:      updatedLeague.CurrWeek,
			Matches:   matches,
			TeamStats: newStats,
			Champion:  champion,
		}, nil
	}

	return &dto.Week{
		LeagueID:  updatedLeague.ID,
		Week:      updatedLeague.CurrWeek,
		Matches:   matches,
		TeamStats: newStats,
	}, nil
}

func (s *LeagueService) PlayRemainingMatches(leagueID uint) ([]*dto.Week, error) {
	league, err := s.repo.GetLeagueByID(leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get league by ID %d: %w", leagueID, err)
	}

	var weeks []*dto.Week

	for week := league.CurrWeek; week <= league.MaxWeeks; week++ {
		matches, err := s.repo.GetMatchesByLeagueIdAndWeek(leagueID, week)
		if err != nil {
			return nil, fmt.Errorf("failed to get matches for league %d and week %d: %w", leagueID, week, err)
		}

		if len(matches) == 0 {
			continue // No matches for this week
		}

		for i, match := range matches {
			if !match.Played {
				simulatedMatch, err := s.matchService.SimulateMatch(match)
				if err != nil {
					return nil, fmt.Errorf("failed to play match %d: %w", match.ID, err)
				}
				matches[i] = simulatedMatch // Update the match in the slice
			}
		}
		// Update championship probabilities if we're past week 3
		if week > 3 {
			if err := s.updateChampionshipProbabilities(leagueID, week); err != nil {
				return nil, fmt.Errorf("failed to update championship probabilities: %w", err)
			}
		}
		// Get Team stats after each week
		newStats, err := s.teamStatsRepo.GetTeamStatsByLeagueID(leagueID)
		if err != nil {
			return nil, fmt.Errorf("failed to get team stats for league %d: %w", leagueID, err)
		}
		if err := s.weeklyLogRepo.SaveWeeklyLog(leagueID, week); err != nil {
			return nil, fmt.Errorf("failed to log weekly results for league %d and week %d: %w", leagueID, week, err)
		}
		// Increment the league week
		if _, err := s.repo.IncrementWeek(leagueID); err != nil {
			return nil, fmt.Errorf("failed to increment league week: %w", err)
		}
		if week == league.MaxWeeks {
			champion, err := s.getChampionByLeagueID(leagueID)
			if err != nil {
				return nil, fmt.Errorf("failed to get champion for league %d: %w", leagueID, err)
			}
			weeks = append(weeks, &dto.Week{
				LeagueID:  league.ID,
				Week:      week,
				Matches:   matches,
				TeamStats: newStats,
				Champion:  champion,
			})
			return weeks, nil

		}

		weeks = append(weeks, &dto.Week{
			LeagueID:  league.ID,
			Week:      week,
			Matches:   matches,
			TeamStats: newStats,
		})
	}

	return weeks, nil
}

// updateChampionshipProbabilities updates the championship probabilities for all teams
func (s *LeagueService) updateChampionshipProbabilities(leagueID uint, week int) error {
	currentLeagueState, err := s.populateLeagueState(leagueID, week)
	if err != nil {
		return fmt.Errorf("failed to populate league state: %w", err)
	}

	// Run Monte Carlo simulation
	results, err := utils.EstimateChampionshipProbabilities(*currentLeagueState)
	if err != nil {
		return fmt.Errorf("failed to estimate championship probabilities: %w", err)
	}

	// Update the estimations in the database
	if err := s.teamStatsRepo.UpdateChampionshipEstimation(results); err != nil {
		return fmt.Errorf("failed to update championship estimations: %w", err)
	}

	return nil
}
func (s *LeagueService) populateLeagueState(leagueID uint, week int) (*dto.LeagueState, error) {
	league, err := s.repo.GetLeagueByID(leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get league by ID %d: %w", leagueID, err)
	}

	matches, err := s.repo.GetRemainingMatches(leagueID, week)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches for league %d and week %d: %w", leagueID, week, err)
	}

	teams, err := s.repo.GetTeamsByLeagueID(leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams for league %d: %w", leagueID, err)
	}
	teamStats, err := s.teamStatsRepo.GetTeamStatsByLeagueID(leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team stats for league %d: %w", leagueID, err)
	}
	// Validate teams
	// Create a deep copy of the league state to prevent modifications to the original data
	copiedTeams := make([]models.Team, len(teams))
	copy(copiedTeams, teams)

	copiedMatches := make([]models.Match, len(matches))
	copy(copiedMatches, matches)

	copiedTeamStats := make([]models.TeamStats, len(teamStats))
	copy(copiedTeamStats, teamStats)

	return &dto.LeagueState{
		LeagueID:         league.ID,
		Week:             week,
		RemainingMatches: copiedMatches,
		Teams:            copiedTeams,
		TeamStats:        copiedTeamStats,
	}, nil
}
func (s *LeagueService) UserPlayWeek(matches []dto.UserPlayedMatch) (*dto.Week, error) {
	league, err := s.repo.GetLeagueByID(matches[0].LeagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get league by ID %d: %w", matches[0].LeagueID, err)
	}
	// Validate matches
	for _, match := range matches {
		if match.HomeScore < 0 || match.AwayScore < 0 {
			return nil, fmt.Errorf("scores cannot be negative")
		}
		if match.HomeTeamID == match.AwayTeamID {
			return nil, fmt.Errorf("home and away teams cannot be the same")
		}
	}

	// Play matches
	var playedMatches []models.Match
	for _, userMatch := range matches {
		playedMatch, err := s.matchService.UserPlayMatch(userMatch)
		if err != nil {
			return nil, fmt.Errorf("failed to play match %d: %w", userMatch.MatchID, err)
		}
		playedMatches = append(playedMatches, playedMatch)
	}

	if league.CurrWeek > 3 {
		if err := s.updateChampionshipProbabilities(league.ID, league.CurrWeek); err != nil {
			return nil, fmt.Errorf("failed to update championship probabilities: %w", err)
		}
	}
	// Get Team stats after each week
	newStats, err := s.teamStatsRepo.GetTeamStatsByLeagueID(matches[0].LeagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team stats for league %d: %w", matches[0].LeagueID, err)
	}

	// Log the weekly results
	if err := s.weeklyLogRepo.SaveWeeklyLog(matches[0].LeagueID, league.CurrWeek); err != nil {
		return nil, fmt.Errorf("failed to log weekly results for league %d and week %d: %w", matches[0].LeagueID, league.CurrWeek, err)
	}

	// Increment the league week
	updatedLeague, err := s.repo.IncrementWeek(matches[0].LeagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to increment league week: %w", err)
	}
	if updatedLeague.CurrWeek > updatedLeague.MaxWeeks {
		champion, err := s.getChampionByLeagueID(matches[0].LeagueID)
		if err != nil {
			return nil, fmt.Errorf("failed to get champion for league %d: %w", matches[0].LeagueID, err)
		}
		return &dto.Week{
			LeagueID:  updatedLeague.ID,
			Week:      updatedLeague.CurrWeek,
			Matches:   playedMatches,
			TeamStats: newStats,
			Champion:  champion,
		}, nil

	}

	return &dto.Week{
		LeagueID:  updatedLeague.ID,
		Week:      updatedLeague.CurrWeek,
		Matches:   playedMatches,
		TeamStats: newStats,
	}, nil
}

func (s *LeagueService) GetChampionshipEstimationByLeagueID(leagueID uint) ([]dto.ChampionshipEstimation, error) {
	league, err := s.repo.GetLeagueByID(leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get league by ID %d: %w", leagueID, err)
	}

	if league.CurrWeek <= 3 {
		return nil, fmt.Errorf("championship estimation is only available after week 3")
	}

	teams, err := s.repo.GetTeamsByLeagueID(leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams for league %d: %w", leagueID, err)
	}
	if len(teams) == 0 {
		return nil, fmt.Errorf("no teams found for league %d", leagueID)
	}

	var estimations []dto.ChampionshipEstimation
	for _, team := range teams {
		estimation, err := s.teamStatsRepo.GetChampionshipEstimationByTeamID(team.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get championship estimation for team %d: %w", team.ID, err)
		}
		estimations = append(estimations, dto.ChampionshipEstimation{
			TeamID:     team.ID,
			Week:       league.CurrWeek,
			LeagueID:   league.ID,
			Estimation: estimation,
		})

	}

	return estimations, nil
}
func (s *LeagueService) getChampionByLeagueID(leagueID uint) (team models.Team, err error) {
	teamStats, err := s.teamStatsRepo.GetTeamStatsByLeagueID(leagueID)
	if err != nil {
		return models.Team{}, fmt.Errorf("failed to get team Stats for league %d: %w", leagueID, err)
	}

	if len(teamStats) == 0 {
		return models.Team{}, fmt.Errorf("no teams found for league %d", leagueID)
	}

	champID := utils.DetermineChampion(teamStats)
	champion, err := s.teamRepo.GetTeamByID(champID)
	if err != nil {
		return models.Team{}, fmt.Errorf("failed to get champion team by ID %d: %w", champID, err)
	}

	return champion, nil
}
