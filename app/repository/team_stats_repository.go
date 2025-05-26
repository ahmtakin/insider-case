package repository

import (
	"insider-case/app/database"
	"insider-case/app/models"

	"gorm.io/gorm"
)

type ITeamStatsRepository interface {
	InitializeTeamStats(tx *gorm.DB, teams []models.Team) error
}
type TeamStatsRepository struct {
	db *gorm.DB
}

func NewTeamStatsRepository() *TeamStatsRepository {
	return &TeamStatsRepository{
		db: database.GetDB(),
	}
}

func (r *TeamStatsRepository) InitializeTeamStats(tx *gorm.DB, teams []models.Team) error {
	teamStats := make([]models.TeamStats, len(teams))

	for i, team := range teams {
		teamStats[i] = models.TeamStats{
			TeamID:       team.ID,
			Points:       0,
			Played:       0,
			Won:          0,
			Lost:         0,
			Draw:         0,
			GoalsFor:     0,
			GoalsAgainst: 0,
			GoalDiff:     0,
			Estimation:   0.0,
		}
	}

	if err := tx.Create(&teamStats).Error; err != nil {
		return err
	}

	return nil
}
