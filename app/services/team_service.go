package services

import (
	"insider-case/app/models"
	"insider-case/app/repository"
)

type ITeamService interface {
	GetTeamsByLeagueID(leagueID uint) ([]models.Team, error)
	GetTeamByID(teamID uint) (models.Team, error)
}

type TeamService struct {
	teamRepo  repository.ITeamRepository
	statsRepo repository.ITeamStatsRepository
}

var _ ITeamService = &TeamService{}

func NewTeamService(teamRepo repository.ITeamRepository, statsRepo repository.ITeamStatsRepository) *TeamService {
	return &TeamService{
		teamRepo:  teamRepo,
		statsRepo: statsRepo,
	}
}
func (s *TeamService) GetTeamsByLeagueID(leagueID uint) ([]models.Team, error) {
	teams, err := s.teamRepo.GetTeamsByLeagueID(leagueID)
	if err != nil {
		return nil, err
	}
	for _, team := range teams {
		stats, err := s.statsRepo.GetTeamStatsByTeamID(team.ID)
		if err != nil {
			return nil, err
		}
		team.Stats = stats
	}

	return teams, nil
}

func (s *TeamService) GetTeamByID(teamID uint) (models.Team, error) {
	team, err := s.teamRepo.GetTeamByID(teamID)
	if err != nil {
		return models.Team{}, err
	}

	stats, err := s.statsRepo.GetTeamStatsByTeamID(team.ID)
	if err != nil {
		return models.Team{}, err
	}
	team.Stats = stats

	return team, nil
}
