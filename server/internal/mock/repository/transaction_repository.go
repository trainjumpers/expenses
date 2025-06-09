package mock_repository

import (
	customErrors "expenses/internal/errors"
	"expenses/internal/models"

	"github.com/gin-gonic/gin"
)

type MockTransactionRepository struct {
	transactions map[int64]models.TransactionResponse
	nextId       int64
}

func NewMockTransactionRepository() *MockTransactionRepository {
	return &MockTransactionRepository{
		transactions: make(map[int64]models.TransactionResponse),
		nextId:       1,
	}
}

func (m *MockTransactionRepository) CreateTransaction(c *gin.Context, input models.CreateTransactionInput) (models.TransactionResponse, error) {
	// Check for duplicate transaction based on composite uniqueness: created_by + date + name + description + amount
	for _, tx := range m.transactions {
		if tx.CreatedBy == input.CreatedBy &&
			tx.Date.Format("2006-01-02") == input.Date.Format("2006-01-02") &&
			tx.Name == input.Name &&
			tx.Amount == *input.Amount {

			// Handle description comparison (both NULL or both same value)
			existingDesc := ""
			if tx.Description != nil {
				existingDesc = *tx.Description
			}
			inputDesc := input.Description

			if existingDesc == inputDesc {
				return models.TransactionResponse{}, customErrors.NewTransactionAlreadyExistsError(nil)
			}
		}
	}

	tx := models.TransactionResponse{
		Id:        m.nextId,
		Name:      input.Name,
		Amount:    *input.Amount,
		Date:      input.Date,
		CreatedBy: input.CreatedBy,
	}
	if input.Description != "" {
		tx.Description = &input.Description
	}
	m.transactions[m.nextId] = tx
	m.nextId++
	return tx, nil
}

func (m *MockTransactionRepository) GetTransactionById(c *gin.Context, transactionId int64, userId int64) (models.TransactionResponse, error) {
	tx, ok := m.transactions[transactionId]
	if !ok || tx.CreatedBy != userId {
		return models.TransactionResponse{}, customErrors.NewTransactionNotFoundError(nil)
	}
	return tx, nil
}

func (m *MockTransactionRepository) UpdateTransaction(c *gin.Context, transactionId int64, userId int64, input models.UpdateTransactionInput) (models.TransactionResponse, error) {
	tx, ok := m.transactions[transactionId]
	if !ok || tx.CreatedBy != userId {
		return models.TransactionResponse{}, customErrors.NewTransactionNotFoundError(nil)
	}

	// Create updated transaction for duplicate checking
	updatedTx := tx
	if input.Name != "" {
		updatedTx.Name = input.Name
	}
	if input.Description != nil {
		updatedTx.Description = input.Description
	}
	if input.Amount != nil {
		updatedTx.Amount = *input.Amount
	}
	if !input.Date.IsZero() {
		updatedTx.Date = input.Date
	}

	// Check for conflicts with other transactions (excluding the current one)
	for id, existingTx := range m.transactions {
		if id != transactionId &&
			existingTx.CreatedBy == updatedTx.CreatedBy &&
			existingTx.Date.Format("2006-01-02") == updatedTx.Date.Format("2006-01-02") &&
			existingTx.Name == updatedTx.Name &&
			existingTx.Amount == updatedTx.Amount {

			// Handle description comparison
			existingDesc := ""
			if existingTx.Description != nil {
				existingDesc = *existingTx.Description
			}
			updatedDesc := ""
			if updatedTx.Description != nil {
				updatedDesc = *updatedTx.Description
			}

			if existingDesc == updatedDesc {
				return models.TransactionResponse{}, customErrors.NewTransactionAlreadyExistsError(nil)
			}
		}
	}

	// Apply the updates
	if input.Name != "" {
		tx.Name = input.Name
	}
	if input.Description != nil {
		tx.Description = input.Description
	}
	if input.Amount != nil {
		tx.Amount = *input.Amount
	}
	if !input.Date.IsZero() {
		tx.Date = input.Date
	}
	m.transactions[transactionId] = tx
	return tx, nil
}

func (m *MockTransactionRepository) DeleteTransaction(c *gin.Context, transactionId int64, userId int64) error {
	tx, ok := m.transactions[transactionId]
	if !ok || tx.CreatedBy != userId {
		return customErrors.NewTransactionNotFoundError(nil)
	}
	delete(m.transactions, transactionId)
	return nil
}

func (m *MockTransactionRepository) ListTransactions(c *gin.Context, userId int64) ([]models.TransactionResponse, error) {
	var result []models.TransactionResponse
	for _, tx := range m.transactions {
		if tx.CreatedBy == userId {
			result = append(result, tx)
		}
	}
	if len(result) == 0 {
		return nil, customErrors.NewTransactionNotFoundError(nil)
	}
	return result, nil
}
