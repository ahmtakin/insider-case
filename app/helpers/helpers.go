package helpers

import (
	"fmt"
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

func CalculateMaxWeeks(TeamCount int) int {
	return ((2 * TeamCount) - 2)
}
