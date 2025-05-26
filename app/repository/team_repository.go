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
}

type TeamRepository struct {
	db *gorm.DB
}

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
