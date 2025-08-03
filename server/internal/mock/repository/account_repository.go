package mock_repository

import (
	"context"
	customErrors "expenses/internal/errors"
	"expenses/internal/models"
	"sync"
)

type MockAccountRepository struct {
	accounts map[int64]models.AccountResponse
	nextId   int64
	mu       sync.RWMutex
}

func NewMockAccountRepository() *MockAccountRepository {
	return &MockAccountRepository{
		accounts: make(map[int64]models.AccountResponse),
		nextId:   1,
	}
}

func (m *MockAccountRepository) CreateAccount(ctx context.Context, input models.CreateAccountInput) (models.AccountResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	acc := models.AccountResponse{
		Id:        m.nextId,
		Name:      input.Name,
		BankType:  input.BankType,
		Currency:  input.Currency,
		Balance:   0,
		CreatedBy: input.CreatedBy,
	}
	if input.Balance != nil {
		acc.Balance = *input.Balance
	}
	m.accounts[m.nextId] = acc
	m.nextId++
	return acc, nil
}

func (m *MockAccountRepository) GetAccountById(ctx context.Context, accountId int64, userId int64) (models.AccountResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	acc, ok := m.accounts[accountId]
	if !ok || acc.CreatedBy != userId {
		return models.AccountResponse{}, customErrors.NewAccountNotFoundError(nil)
	}
	return acc, nil
}

func (m *MockAccountRepository) UpdateAccount(ctx context.Context, accountId int64, userId int64, input models.UpdateAccountInput) (models.AccountResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	acc, ok := m.accounts[accountId]
	if !ok || acc.CreatedBy != userId {
		return models.AccountResponse{}, customErrors.NewAccountNotFoundError(nil)
	}
	if input.Name != "" {
		acc.Name = input.Name
	}
	if input.BankType != "" {
		acc.BankType = input.BankType
	}
	if input.Currency != "" {
		acc.Currency = input.Currency
	}
	if input.Balance != nil {
		acc.Balance = *input.Balance
	}
	m.accounts[accountId] = acc
	return acc, nil
}

func (m *MockAccountRepository) DeleteAccount(ctx context.Context, accountId int64, userId int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	acc, ok := m.accounts[accountId]
	if !ok || acc.CreatedBy != userId {
		return customErrors.NewAccountNotFoundError(nil)
	}
	delete(m.accounts, accountId)
	return nil
}

func (m *MockAccountRepository) ListAccounts(ctx context.Context, userId int64) ([]models.AccountResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []models.AccountResponse
	for _, acc := range m.accounts {
		if acc.CreatedBy == userId {
			result = append(result, acc)
		}
	}
	if len(result) == 0 {
		return nil, customErrors.NewAccountNotFoundError(nil)
	}
	return result, nil
}
