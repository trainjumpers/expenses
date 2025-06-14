package mock_repository

import (
	customErrors "expenses/internal/errors"
	"expenses/internal/models"

	"github.com/gin-gonic/gin"
)

type MockTransactionRepository struct {
	transactions map[int64]models.TransactionResponse
	nextId       int64
	categoryMap  map[int64][]int64
}

func NewMockTransactionRepository() *MockTransactionRepository {
	return &MockTransactionRepository{
		transactions: make(map[int64]models.TransactionResponse),
		nextId:       1,
		categoryMap:  make(map[int64][]int64),
	}
}

func (m *MockTransactionRepository) CreateTransaction(c *gin.Context, input models.CreateBaseTransactionInput, categoryIds []int64) (models.TransactionResponse, error) {
	// Check for duplicate transaction based on composite uniqueness: created_by + date + name + description + amount
	for _, tx := range m.transactions {
		if tx.CreatedBy == input.CreatedBy &&
			tx.Date.Format("2006-01-02") == input.Date.Format("2006-01-02") &&
			tx.Name == input.Name &&
			tx.Amount == *input.Amount {
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

	// Create new transaction
	newId := m.nextId
	m.nextId++

	baseTx := models.TransactionBaseResponse{
		Id:          newId,
		Name:        input.Name,
		Description: &input.Description,
		Amount:      *input.Amount,
		Date:        input.Date,
		CreatedBy:   input.CreatedBy,
		AccountId:   input.AccountId,
	}

	tx := models.TransactionResponse{
		TransactionBaseResponse: baseTx,
		CategoryIds:             categoryIds,
	}

	m.transactions[newId] = tx
	m.categoryMap[newId] = categoryIds

	return tx, nil
}

func (m *MockTransactionRepository) UpdateCategoryMapping(c *gin.Context, transactionId int64, userId int64, categoryIds []int64) error {
	tx, ok := m.transactions[transactionId]
	if !ok || tx.CreatedBy != userId {
		return customErrors.NewTransactionNotFoundError(nil)
	}
	m.categoryMap[transactionId] = categoryIds
	tx.CategoryIds = categoryIds
	m.transactions[transactionId] = tx
	return nil
}

func (m *MockTransactionRepository) GetTransactionById(c *gin.Context, transactionId int64, userId int64) (models.TransactionResponse, error) {
	tx, ok := m.transactions[transactionId]
	if !ok || tx.CreatedBy != userId {
		return models.TransactionResponse{}, customErrors.NewTransactionNotFoundError(nil)
	}
	return tx, nil
}

func (m *MockTransactionRepository) UpdateTransaction(c *gin.Context, transactionId int64, userId int64, input models.UpdateBaseTransactionInput) error {
	tx, ok := m.transactions[transactionId]
	if !ok || tx.CreatedBy != userId {
		return customErrors.NewTransactionNotFoundError(nil)
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
	if input.AccountId != nil {
		updatedTx.AccountId = *input.AccountId
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
				return customErrors.NewTransactionAlreadyExistsError(nil)
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
	if input.AccountId != nil {
		tx.AccountId = *input.AccountId
	}
	m.transactions[transactionId] = tx
	return nil
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
	return result, nil
}
