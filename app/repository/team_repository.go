package repository

import (
	"fmt"
	"insider-case/app/database"
	"insider-case/app/models"

	"gorm.io/gorm"
)

type ITeamRepository interface {
	ValidateTeams(teams []models.Team, expectedCount int) error
	CreateTeams(tx *gorm.DB, teams []models.Team, leagueID uint) error
	GetTeamsByLeagueID(leagueID uint) ([]models.Team, error)
	GetTeamByID(TeamID uint) (models.Team, error)
	GetTeamStrengthByID(TeamID uint) (int, error)
}

type TeamRepository struct {
	db *gorm.DB
}

var _ ITeamRepository = &TeamRepository{}

func NewTeamRepository() *TeamRepository {
	return &TeamRepository{
		db: database.GetDB(),
	}
}

func (r *TeamRepository) ValidateTeams(teams []models.Team, expectedCount int) error {
	if len(teams) != expectedCount {
		return fmt.Errorf("expected %d teams, got %d", expectedCount, len(teams))
	}

	teamNames := make(map[string]bool)
	for _, team := range teams {
		if team.Name == "" {
			return fmt.Errorf("team name cannot be empty")
		}
		if teamNames[team.Name] {
			return fmt.Errorf("duplicate team name: %s", team.Name)
		}
		teamNames[team.Name] = true
	}

	return nil
}

func (r *TeamRepository) CreateTeams(tx *gorm.DB, teams []models.Team, leagueID uint) error {
	for i := range teams {
		teams[i].LeagueID = leagueID
	}

	if err := tx.Create(&teams).Error; err != nil {
		return fmt.Errorf("failed to create teams: %w", err)
	}

	return nil
}

func (r *TeamRepository) GetTeamsByLeagueID(leagueID uint) ([]models.Team, error) {
	var teams []models.Team
	if err := r.db.Where("league_id = ?", leagueID).Find(&teams).Error; err != nil {
		return nil, fmt.Errorf("failed to get teams for league %d: %w", leagueID, err)
	}
	return teams, nil
}

func (r *TeamRepository) GetTeamByID(TeamID uint) (models.Team, error) {
	var team models.Team
	if err := r.db.Where("id = ?", TeamID).First(&team).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.Team{}, fmt.Errorf("team with ID %d not found", TeamID)
		}
		return models.Team{}, fmt.Errorf("failed to get team by ID %d: %w", TeamID, err)
	}
	return team, nil
}

func (r *TeamRepository) GetTeamStrengthByID(TeamID uint) (int, error) {
	var strength int
	if err := r.db.Model(&models.Team{}).Select("strength").Where("id = ?", TeamID).Scan(&strength).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, fmt.Errorf("team with ID %d not found", TeamID)
		}
		return 0, fmt.Errorf("failed to get team strength by ID %d: %w", TeamID, err)
	}
	return strength, nil
}
