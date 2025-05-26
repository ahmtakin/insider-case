package repository

import (
	"insider-case/app/database"
	"insider-case/app/models"

	"gorm.io/gorm"
)

type ITeamStatsRepository interface {
	InitializeTeamStats(tx *gorm.DB, teams []models.Team) error
	GetTeamStatsByTeamID(TeamID uint) (models.TeamStats, error)
	UpdateTeamStats(teamStats models.TeamStats) error
}
type TeamStatsRepository struct {
	db *gorm.DB
}

var _ ITeamStatsRepository = &TeamStatsRepository{}

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

func (r *TeamStatsRepository) GetTeamStatsByTeamID(TeamID uint) (models.TeamStats, error) {
	var teamStats models.TeamStats
	if err := r.db.Where("team_id = ?", TeamID).First(&teamStats).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.TeamStats{}, nil // No stats found for this team
		}
		return models.TeamStats{}, err // Return any other error
	}
	return teamStats, nil
}

func (r *TeamStatsRepository) UpdateTeamStats(teamStats models.TeamStats) error {
	if err := r.db.Model(&models.TeamStats{}).Where("team_id = ?", teamStats.TeamID).Updates(teamStats).Error; err != nil {
		return err
	}
	return nil
}
