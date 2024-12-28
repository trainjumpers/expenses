package validators

import (
	"fmt"
)

func ValidateContributions(contributions []float64, amount float64) error {
	totalContribution := 0.0
	for _, v := range contributions {
		totalContribution += v
	}
	if totalContribution != amount {
		return fmt.Errorf("total contribution must be equal to amount")
	}
	return nil
}
