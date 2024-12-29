package mapper

import (
	"expenses/entities"
	"expenses/models"
	"fmt"
)

func ExpenseContributorToMapper(expenses []models.ExpenseWithAllContributions) (entities.ExpenseOutput, error) {
	var expenseOutput entities.ExpenseOutput
	if len(expenses) == 0 {
		return expenseOutput, fmt.Errorf("no expenses found")
	}
	expenseOutput.ID = expenses[0].ID
	expenseOutput.Name = expenses[0].Name
	expenseOutput.Description = expenses[0].Description
	expenseOutput.Amount = expenses[0].Amount
	expenseOutput.PayerID = expenses[0].PayerID
	expenseOutput.CreatedBy = expenses[0].CreatedBy
	expenseOutput.CreatedAt = expenses[0].CreatedAt
	expenseOutput.Contributions = make(map[int64]entities.ExpenseOutputContribution)
	for _, expense := range expenses {
		expenseOutput.Contributions[expense.ContributorId] = entities.ExpenseOutputContribution{
			Amount: expense.Contribution,
			Name:   expense.ContributorName,
		}
		if expense.ID != expenseOutput.ID {
			return expenseOutput, fmt.Errorf("expense id mismatch: %s != %s", expense.Name, expenseOutput.Name)
		}
	}
	return expenseOutput, nil
}

func StatementExpenseMapper(expenses []entities.Statement, userId int64) ([]models.Expense, error) {
	var expenseModels []models.Expense

	for _, expense := range expenses {
		expenseModels = append(expenseModels, models.Expense{
			Amount:      expense.Amount,
			PayerID:     userId,
			Name:        expense.TrasactionId,
			Description: expense.Description,
			CreatedBy:   userId,
			CreatedAt:   &expense.Date,
		})
		break
	}

	return expenseModels, nil
}
