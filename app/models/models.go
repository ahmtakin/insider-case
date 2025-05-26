package models

type League struct {
	ID        uint    `json:"id" gorm:"primaryKey"`
	Name      string  `json:"name"`
	TeamCount int     `json:"team_count"`
	MaxWeeks  int     `json:"max_weeks"`
	CurrWeek  int     `json:"curr_week"`
	Teams     []Team  `json:"teams,omitempty" gorm:"foreignKey:LeagueID"`
	Matches   []Match `json:"matches,omitempty" gorm:"foreignKey:LeagueID"`
}

type Team struct {
	ID       uint      `json:"id" gorm:"primaryKey"`
	LeagueID uint      `json:"league_id"`
	Name     string    `json:"name"`
	Strength int       `json:"strength"`
	Stats    TeamStats `json:"stats,omitempty" gorm:"foreignKey:TeamID"`
}

type TeamStats struct {
	TeamID       uint    `json:"team_id" gorm:"primaryKey"`
	Points       int     `json:"points"`
	Played       int     `json:"played"`
	Won          int     `json:"won"`
	Lost         int     `json:"lost"`
	Draw         int     `json:"draw"`
	GoalsFor     int     `json:"goals_for"`
	GoalsAgainst int     `json:"goals_against"`
	GoalDiff     int     `json:"goal_diff"`
	Estimation   float32 `json:"estimation"`
}

type Match struct {
	ID         uint  `json:"id" gorm:"primaryKey"`
	LeagueID   uint  `json:"league_id"`
	Week       int   `json:"week"`
	Played     bool  `json:"played"`
	HomeTeamID uint  `json:"home_team"`
	AwayTeamID uint  `json:"away_team"`
	HomeScore  int   `json:"home_score"`
	AwayScore  int   `json:"away_score"`
	Result     *uint `json:"result,omitempty"` // ID of winning team or nil for draw
}

type WeeklyLog struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	LeagueID      uint   `json:"league_id"`
	Week          int    `json:"week"`
	TeamStatsJSON string `json:"team_stats_json" gorm:"type:jsonb"` // JSON snapshot of team stats
}
