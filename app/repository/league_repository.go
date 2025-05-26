package repository

import (
	"fmt"
	"insider-case/app/database"
	"insider-case/app/helpers"
	"insider-case/app/models"

	"gorm.io/gorm"
)

type ILeagueRepository interface {
	CreateLeague(league *models.League) (*models.League, error)
	GetLeagueByID(id uint) (*models.League, error)
	InitializeLeague(league *models.League) (*models.League, error)
	IncrementWeek(leagueID uint) (*models.League, error)
	GetMatchesByLeagueIdAndWeek(leagueID uint, week int) ([]models.Match, error)
}

type LeagueRepository struct {
	db                  *gorm.DB
	teamRepository      ITeamRepository
	matchRepository     IMatchRepository
	teamStatsRepository ITeamStatsRepository
}

var _ ILeagueRepository = &LeagueRepository{}

func NewLeagueRepository(
	TeamRepo ITeamRepository,
	MatchRepo IMatchRepository,
	TeamStatsRepo ITeamStatsRepository,
) *LeagueRepository {
	return &LeagueRepository{
		db:                  database.GetDB(),
		teamRepository:      TeamRepo,
		matchRepository:     MatchRepo,
		teamStatsRepository: TeamStatsRepo}
}

func (r *LeagueRepository) CreateLeague(league *models.League) (*models.League, error) {
	if err := r.db.Create(league).Error; err != nil {
		return nil, err
	}
	fmt.Println("League created successfully:", "id", league.ID, "name", league.Name, "teamcount", league.TeamCount, "maxweeks", league.MaxWeeks, "currentweek", league.CurrWeek)

	return league, nil
}

func (r *LeagueRepository) GetLeagueByID(id uint) (*models.League, error) {
	var league models.League
	if err := r.db.First(&league, id).Error; err != nil {
		return nil, err
	}
	return &league, nil
}

func (r *LeagueRepository) InitializeLeague(league *models.League) (*models.League, error) {
	// Validate league basic requirements
	if league.Name == "" {
		return nil, fmt.Errorf("league name is required")
	}
	if league.TeamCount <= 0 {
		return nil, fmt.Errorf("team count must be greater than 0")
	}

	// Validate teams before starting transaction
	if err := r.teamRepository.ValidateTeams(league.Teams, league.TeamCount); err != nil {
		return nil, fmt.Errorf("team validation failed: %w", err)
	}

	var createdLeague *models.League
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Create league first (without teams)
		leagueToCreate := &models.League{
			Name:      league.Name,
			TeamCount: league.TeamCount,
			MaxWeeks:  helpers.CalculateMaxWeeks(league.TeamCount),
			CurrWeek:  1}

		if err := tx.Create(leagueToCreate).Error; err != nil {
			return fmt.Errorf("failed to create league: %w", err)
		}
		leagueToCreate.Teams = league.Teams
		// Create teams with the new league ID
		if err := r.teamRepository.CreateTeams(tx, leagueToCreate.Teams, leagueToCreate.ID); err != nil {
			return err
		}
		// Initialize team stats
		if err := r.teamStatsRepository.InitializeTeamStats(tx, leagueToCreate.Teams); err != nil {
			return fmt.Errorf("failed to initialize team stats: %w", err)
		}

		fixtures, err := r.matchRepository.GenerateFixtures(*leagueToCreate)
		if err != nil {
			return fmt.Errorf("failed to generate fixtures: %w", err)
		}
		// Create all matches
		if err := tx.Create(&fixtures).Error; err != nil {
			return fmt.Errorf("failed to create fixtures: %w", err)
		}

		// Load the complete league with teams and matches
		if err := tx.Preload("Teams").Preload("Teams.Stats").Preload("Matches").First(&createdLeague, leagueToCreate.ID).Error; err != nil {
			return fmt.Errorf("failed to load created league: %w", err)
		}

		return nil

	})

	if err != nil {
		return nil, err
	}

	fmt.Printf("League created successfully: id=%d, name=%s, teams=%d\n",
		createdLeague.ID, createdLeague.Name, len(createdLeague.Teams))

	return createdLeague, nil
}
func (r *LeagueRepository) IncrementWeek(leagueID uint) (*models.League, error) {
	var league models.League
	if err := r.db.First(&league, leagueID).Error; err != nil {
		return nil, fmt.Errorf("failed to find league with ID %d: %w", leagueID, err)
	}

	// Increment the current week
	league.CurrWeek++
	if err := r.db.Save(&league).Error; err != nil {
		return nil, fmt.Errorf("failed to increment week for league %d: %w", leagueID, err)
	}

	fmt.Printf("League week incremented: id=%d, new week=%d\n", league.ID, league.CurrWeek)

	return &league, nil
}
func (r *LeagueRepository) GetMatchesByLeagueIdAndWeek(leagueID uint, week int) ([]models.Match, error) {
	var matches []models.Match
	if err := r.db.Where("league_id = ? AND week = ?", leagueID, week).Find(&matches).Error; err != nil {
		return nil, fmt.Errorf("failed to get matches for league %d and week %d: %w", leagueID, week, err)
	}
	return matches, nil
}
