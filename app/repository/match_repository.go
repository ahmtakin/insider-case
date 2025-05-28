package repository

import (
	"fmt"
	"insider-case/app/database"
	"insider-case/app/models"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

type IMatchRepository interface {
	GenerateFixtures(league models.League) ([]models.Match, error)
	GetMatchesByLeagueIdAndWeek(leagueID uint, week int) ([]models.Match, error)
	GetMatchesByLeagueId(leagueID uint) ([]models.Match, error)
	SaveMatch(match models.Match) error
	GetMatchByID(matchID uint) (*models.Match, error)
}

type MatchRepository struct {
	db *gorm.DB
}

var _ IMatchRepository = &MatchRepository{}

func NewMatchRepository() *MatchRepository {
	return &MatchRepository{
		db: database.GetDB()}
}

func (r *MatchRepository) GenerateFixtures(league models.League) ([]models.Match, error) {
	if len(league.Teams)%2 != 0 {
		return nil, fmt.Errorf("number of teams must be even")
	}

	numTeams := len(league.Teams)
	maxWeeks := league.MaxWeeks
	halfSeason := maxWeeks / 2
	matchesPerWeek := numTeams / 2

	// Track played pairs using string keys
	played := make(map[string]bool)

	matches := make([]models.Match, 0)
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// First half of the season
	for week := 1; week <= halfSeason; week++ {
		usedTeams := make(map[string]bool)
		weekMatches := make([]models.Match, 0, matchesPerWeek)

		attempts := 0
		for len(weekMatches) < matchesPerWeek && attempts < 1000 {
			attempts++

			i := rand.Intn(numTeams)
			j := rand.Intn(numTeams)
			if i == j {
				continue
			}

			teamA := league.Teams[i]
			teamB := league.Teams[j]

			keyA := teamA.Name
			keyB := teamB.Name

			if usedTeams[keyA] || usedTeams[keyB] {
				continue
			}

			matchKey := getMatchKey(keyA, keyB)
			if played[matchKey] {
				continue
			}

			// Add match
			usedTeams[keyA] = true
			usedTeams[keyB] = true
			played[matchKey] = true

			match := models.Match{
				LeagueID:   league.ID,
				Week:       week,
				HomeTeamID: teamA.ID,
				AwayTeamID: teamB.ID,
			}
			weekMatches = append(weekMatches, match)
		}

		if len(weekMatches) < matchesPerWeek {
			return nil, fmt.Errorf("could not generate valid fixtures after many attempts")
		}

		matches = append(matches, weekMatches...)
	}

	// Second half of the season (reverse matches)
	totalFirstHalf := len(matches)
	for i := 0; i < totalFirstHalf; i++ {
		original := matches[i]
		reverse := models.Match{
			LeagueID:   original.LeagueID,
			Week:       original.Week + halfSeason,
			HomeTeamID: original.AwayTeamID,
			AwayTeamID: original.HomeTeamID,
		}
		matches = append(matches, reverse)
	}

	return matches, nil
}

// getMatchKey returns a consistent key for a match regardless of team order
func getMatchKey(teamA, teamB string) string {
	if teamA < teamB {
		return teamA + "-" + teamB
	}
	return teamB + "-" + teamA
}

func (r *MatchRepository) GetMatchesByLeagueIdAndWeek(leagueID uint, week int) ([]models.Match, error) {
	var matches []models.Match
	if err := r.db.Where("league_id = ? AND week = ?", leagueID, week).Find(&matches).Error; err != nil {
		return nil, fmt.Errorf("failed to get matches for league %d and week %d: %w", leagueID, week, err)
	}
	return matches, nil
}
func (r *MatchRepository) GetMatchesByLeagueId(leagueID uint) ([]models.Match, error) {
	var matches []models.Match
	if err := r.db.Where("league_id = ?", leagueID).Find(&matches).Error; err != nil {
		return nil, fmt.Errorf("failed to get matches for league %d: %w", leagueID, err)
	}
	return matches, nil
}

func (r *MatchRepository) SaveMatch(match models.Match) error {
	// Find the match in the database
	var existingMatch models.Match
	if err := r.db.First(&existingMatch, match.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("match with ID %d not found: %w", match.ID, err)
		}
		return fmt.Errorf("failed to find match with ID %d: %w", match.ID, err)
	}

	// Update the match state
	existingMatch.HomeScore = match.HomeScore
	existingMatch.AwayScore = match.AwayScore
	existingMatch.Played = true
	existingMatch.Result = match.Result

	if err := r.db.Save(&existingMatch).Error; err != nil {
		return fmt.Errorf("failed to update match with ID %d: %w", match.ID, err)
	}

	return nil
}

func (r *MatchRepository) GetMatchByID(matchID uint) (*models.Match, error) {
	var match models.Match
	if err := r.db.First(&match, matchID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("match with ID %d not found", matchID)
		}
		return nil, fmt.Errorf("failed to get match by ID %d: %w", matchID, err)
	}
	return &match, nil
}
