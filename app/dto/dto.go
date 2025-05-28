package dto

import (
	"insider-case/app/models"
)

type LeagueCreateRequest struct {
	Name      string        `json:"name" binding:"required"`
	TeamCount int           `json:"team_count" binding:"required,min=2"`
	Teams     []TeamRequest `json:"teams" binding:"required,dive"`
}

type TeamRequest struct {
	Name     string `json:"name" binding:"required"`
	Strength int    `json:"strength" binding:"required,min=0,max=100"`
}

type LeagueResponse struct {
	ID        uint           `json:"id"`
	Name      string         `json:"name"`
	TeamCount int            `json:"team_count"`
	MaxWeeks  int            `json:"max_weeks"`
	CurrWeek  int            `json:"curr_week"`
	Teams     []models.Team  `json:"teams,omitempty"`
	Matches   []models.Match `json:"matches,omitempty"`
}

type Week struct {
	LeagueID  uint               `json:"league_id"`
	Week      int                `json:"week"`
	Matches   []models.Match     `json:"matches,omitempty"`
	TeamStats []models.TeamStats `json:"team_stats,omitempty"`
	Champion  models.Team        `json:"champion,omitempty"`
}

type ChampionshipEstimation struct {
	LeagueID   uint    `json:"league_id"`
	Week       int     `json:"week"`
	TeamID     uint    `json:"team_id"`
	Estimation float32 `json:"estimation"`
}
type LeagueState struct {
	LeagueID         uint               `json:"league_id"`
	Week             int                `json:"week"`
	Teams            []models.Team      `json:"teams"`
	RemainingMatches []models.Match     `json:"matches"`
	TeamStats        []models.TeamStats `json:"team_stats"`
}
type UserPlayedMatch struct {
	LeagueID   uint `json:"league_id"`
	Week       int  `json:"week"`
	MatchID    uint `json:"match_id"`
	HomeTeamID uint `json:"home_team_id"`
	AwayTeamID uint `json:"away_team_id"`
	HomeScore  int  `json:"home_score"`
	AwayScore  int  `json:"away_score"`
}
type Champion struct {
	TeamID   uint   `json:"team_id" gorm:"primaryKey"`
	LeagueID uint   `json:"league_id" gorm:"primaryKey"`
	TeamName string `json:"team_name"`
}
