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
}

type MatchRepository struct {
	db *gorm.DB
}

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
