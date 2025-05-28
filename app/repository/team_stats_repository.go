package repository

import (
	"insider-case/app/database"
	"insider-case/app/dto"
	"insider-case/app/models"

	"gorm.io/gorm"
)

type ITeamStatsRepository interface {
	InitializeTeamStats(tx *gorm.DB, teams []models.Team) error
	GetTeamStatsByTeamID(TeamID uint) (models.TeamStats, error)
	UpdateTeamStats(teamStats models.TeamStats) error
	GetTeamStatsByLeagueID(leagueID uint) ([]models.TeamStats, error)
	UpdateChampionshipEstimation(estimations []dto.ChampionshipEstimation) error
	GetChampionshipEstimationByTeamID(teamID uint) (float32, error)
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
	if err := r.db.Model(&models.TeamStats{}).
		Where("team_id = ?", teamStats.TeamID).
		Omit("team_id", "estimation").
		Save(teamStats).Error; err != nil {
		return err
	}
	return nil
}
func (r *TeamStatsRepository) GetTeamStatsByLeagueID(leagueID uint) ([]models.TeamStats, error) {
	var teams []models.Team
	if err := r.db.Where("league_id = ?", leagueID).Find(&teams).Error; err != nil {
		return nil, err
	}

	teamIDs := make([]uint, len(teams))
	for i, team := range teams {
		teamIDs[i] = team.ID
	}

	var teamStats []models.TeamStats
	if err := r.db.Where("team_id IN ?", teamIDs).Find(&teamStats).Error; err != nil {
		return nil, err
	}

	return teamStats, nil
}

func (r *TeamStatsRepository) UpdateChampionshipEstimation(estimations []dto.ChampionshipEstimation) error {
	// Use a transaction to ensure all updates succeed or none do
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, est := range estimations {
			// Update each team's estimation individually
			if err := tx.Model(&models.TeamStats{}).
				Where("team_id = ?", est.TeamID).
				Update("estimation", est.Estimation).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (r *TeamStatsRepository) GetChampionshipEstimationByTeamID(teamID uint) (float32, error) {
	var estimation float32
	if err := r.db.Model(&models.TeamStats{}).
		Where("team_id = ?", teamID).
		Select("estimation").
		Scan(&estimation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0.0, nil // No estimation found for this team
		}
		return 0.0, err // Return any other error
	}
	return estimation, nil
}
