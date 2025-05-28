package helpers

import (
	"fmt"
	"insider-case/app/dto"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

func ValidateTeamCount(count int) error {
	if count < 2 {
		return &ValidationError{
			Field:   "team_count",
			Message: "must have at least 2 teams",
		}
	}
	if count%2 != 0 {
		return &ValidationError{
			Field:   "team_count",
			Message: "must be an even number",
		}
	}
	return nil
}
func ValidateTeamStrength(teams []dto.TeamRequest) error {
	for _, team := range teams {
		if team.Strength < 1000 || team.Strength > 3000 {
			return &ValidationError{
				Field:   "team_strength",
				Message: fmt.Sprintf("strength for team %s must be between 1 and 100", team.Name),
			}
		}
	}
	return nil
}

func CalculateMaxWeeks(TeamCount int) int {
	return ((2 * TeamCount) - 2)
}
