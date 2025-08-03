package mock_repository

import (
	"context"
	customErrors "expenses/internal/errors"
	"expenses/internal/models"
	"sort"
	"strings"
	"sync"
)

type statementTxnMapping struct {
	StatementId   int64
	TransactionId int64
}

type MockTransactionRepository struct {
	transactions                 map[int64]models.TransactionResponse
	nextId                       int64
	categoryMap                  map[int64][]int64
	mu                           sync.RWMutex
	statementTransactionMappings []statementTxnMapping // Use local struct for statement_id filtering
}

func NewMockTransactionRepository() *MockTransactionRepository {
	return &MockTransactionRepository{
		transactions:                 make(map[int64]models.TransactionResponse),
		nextId:                       1,
		categoryMap:                  make(map[int64][]int64),
		statementTransactionMappings: []statementTxnMapping{},
	}
}

func (m *MockTransactionRepository) CreateTransaction(ctx context.Context, input models.CreateBaseTransactionInput, categoryIds []int64) (models.TransactionResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
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

func (m *MockTransactionRepository) UpdateCategoryMapping(ctx context.Context, transactionId int64, userId int64, categoryIds []int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	tx, ok := m.transactions[transactionId]
	if !ok || tx.CreatedBy != userId {
		return customErrors.NewTransactionNotFoundError(nil)
	}
	m.categoryMap[transactionId] = categoryIds
	tx.CategoryIds = categoryIds
	m.transactions[transactionId] = tx
	return nil
}

func (m *MockTransactionRepository) GetTransactionById(ctx context.Context, transactionId int64, userId int64) (models.TransactionResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	tx, ok := m.transactions[transactionId]
	if !ok || tx.CreatedBy != userId {
		return models.TransactionResponse{}, customErrors.NewTransactionNotFoundError(nil)
	}
	return tx, nil
}

func (m *MockTransactionRepository) UpdateTransaction(ctx context.Context, transactionId int64, userId int64, input models.UpdateBaseTransactionInput) error {
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

func (m *MockTransactionRepository) DeleteTransaction(ctx context.Context, transactionId int64, userId int64) error {
	tx, ok := m.transactions[transactionId]
	if !ok || tx.CreatedBy != userId {
		return customErrors.NewTransactionNotFoundError(nil)
	}
	delete(m.transactions, transactionId)
	return nil
}

func (m *MockTransactionRepository) ListTransactions(ctx context.Context, userId int64, query models.TransactionListQuery) (models.PaginatedTransactionsResponse, error) {
	var result []models.TransactionResponse

	// Filter transactions by user Id and apply other filters
	for _, tx := range m.transactions {
		if tx.CreatedBy != userId {
			continue
		}

		// Apply filters
		if query.AccountId != nil && tx.AccountId != *query.AccountId {
			continue
		}
		if query.StatementId != nil {
			found := false
			for _, mapping := range m.statementTransactionMappings {
				if mapping.TransactionId == tx.Id && mapping.StatementId == *query.StatementId {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		if query.CategoryId != nil {
			found := false
			for _, catId := range tx.CategoryIds {
				if catId == *query.CategoryId {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		if query.Uncategorized != nil && *query.Uncategorized {
			if len(tx.CategoryIds) > 0 {
				continue
			}
		}

		if query.MinAmount != nil && tx.Amount < *query.MinAmount {
			continue
		}

		if query.MaxAmount != nil && tx.Amount > *query.MaxAmount {
			continue
		}

		if query.DateFrom != nil && tx.Date.Before(*query.DateFrom) {
			continue
		}

		if query.DateTo != nil && tx.Date.After(*query.DateTo) {
			continue
		}

		if query.Search != nil && *query.Search != "" {
			searchTerm := strings.ToLower(*query.Search)
			name := strings.ToLower(tx.Name)
			description := ""
			if tx.Description != nil {
				description = strings.ToLower(*tx.Description)
			}
			if !strings.Contains(name, searchTerm) && !strings.Contains(description, searchTerm) {
				continue
			}
		}

		result = append(result, tx)
	}

	// Sort transactions
	sortBy := query.SortBy
	if sortBy == "" {
		sortBy = "date"
	}
	sortOrder := strings.ToLower(query.SortOrder)
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// Sort the result slice
	switch sortBy {
	case "date":
		sort.Slice(result, func(i, j int) bool {
			if sortOrder == "asc" {
				return result[i].Date.Before(result[j].Date)
			}
			return result[i].Date.After(result[j].Date)
		})
	case "amount":
		sort.Slice(result, func(i, j int) bool {
			if sortOrder == "asc" {
				return result[i].Amount < result[j].Amount
			}
			return result[i].Amount > result[j].Amount
		})
	case "name":
		sort.Slice(result, func(i, j int) bool {
			if sortOrder == "asc" {
				return result[i].Name < result[j].Name
			}
			return result[i].Name > result[j].Name
		})
	}

	// Apply pagination
	page := query.Page
	if page < 1 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 15
	}

	total := len(result)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		return models.PaginatedTransactionsResponse{
			Transactions: []models.TransactionResponse{},
			Total:        total,
			Page:         page,
			PageSize:     pageSize,
		}, nil
	}
	if end > total {
		end = total
	}

	return models.PaginatedTransactionsResponse{
		Transactions: result[start:end],
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
	}, nil
}
