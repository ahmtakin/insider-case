package repository

import (
	"encoding/json"
	"fmt"
	"insider-case/app/database"
	"insider-case/app/models"

	"gorm.io/gorm"
)

type IWeeklyLogRepository interface {
	SaveWeeklyLog(leagueID uint, week int) error
	GetWeeklyLogByLeagueIDAndWeek(leagueID uint, week int) (*models.WeeklyLog, error)
}

type WeeklyLogRepository struct {
	db            *gorm.DB
	teamStatsRepo ITeamStatsRepository
}

var _ IWeeklyLogRepository = &WeeklyLogRepository{}

func NewWeeklyLogRepository(teamStatsRepo ITeamStatsRepository) *WeeklyLogRepository {
	return &WeeklyLogRepository{
		db:            database.GetDB(),
		teamStatsRepo: teamStatsRepo,
	}
}
func (r *WeeklyLogRepository) SaveWeeklyLog(leagueID uint, week int) error {

	log := models.WeeklyLog{
		LeagueID:      leagueID,
		Week:          week,
		TeamStatsJSON: "{}", // Placeholder for team stats JSON, should be populated with actual data
	}
	teamStats, err := r.teamStatsRepo.GetTeamStatsByLeagueID(leagueID)
	if err != nil {
		return fmt.Errorf("failed to get team stats for league %d: %w", leagueID, err)
	}
	teamStatsJSON, err := json.Marshal(&teamStats)
	if err != nil {
		return fmt.Errorf("failed to marshal team stats to JSON: %w", err)
	}
	log.TeamStatsJSON = string(teamStatsJSON)

	if err := r.db.Create(&log).Error; err != nil {
		return fmt.Errorf("failed to save weekly log: %w", err)
	}
	fmt.Println("Weekly log saved successfully for league:", log.LeagueID, "week:", log.Week)
	return nil
}
func (r *WeeklyLogRepository) GetWeeklyLogByLeagueIDAndWeek(leagueID uint, week int) (*models.WeeklyLog, error) {
	var log models.WeeklyLog
	if err := r.db.Where("league_id = ? AND week = ?", leagueID, week).First(&log).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no weekly log found for league %d and week %d", leagueID, week)
		}
		return nil, fmt.Errorf("failed to get weekly log: %w", err)
	}
	return &log, nil
}
