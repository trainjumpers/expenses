package validators

import (
	"expenses/entities"
	"fmt"
)

func ValidateCreateExpense(expense entities.ExpenseInput) error {
	contributors := []int64{}
	contributions := []float64{}
	for k, v := range expense.Contributions {
		contributors = append(contributors, k)
		contributions = append(contributions, v)
	}
	if len(contributors) != len(contributions) {
		return fmt.Errorf("contributors and contributions must be of the same length")
	}
	// Sum of  contributions must be equal to the total expense
	totalContribution := 0.0
	for _, v := range contributions {
		totalContribution += v
	}
	if totalContribution != expense.Amount {
		return fmt.Errorf("total contribution must be equal to the total expense")
	}
	return nil
}
