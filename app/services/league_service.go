package services

import (
	"fmt"
	"insider-case/app/dto"
	"insider-case/app/helpers"
	"insider-case/app/models"
	"insider-case/app/repository"
)

type ILeagueService interface {
	InitializeLeague(req dto.LeagueCreateRequest) (*dto.LeagueResponse, error)
}

type LeagueService struct {
	repo repository.ILeagueRepository
}

func NewLeagueService(repo repository.ILeagueRepository) *LeagueService {
	return &LeagueService{
		repo: repo,
	}
}

func (s *LeagueService) InitializeLeague(req dto.LeagueCreateRequest) (*dto.LeagueResponse, error) {
	if err := helpers.ValidateTeamCount(req.TeamCount); err != nil {
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
			Name:     team.Name,
			Strength: team.Strength,
			Stats:    team.Stats,
		}
	}
	for i, match := range league.Matches {
		response.Matches[i] = models.Match{
			ID:         match.ID,
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
