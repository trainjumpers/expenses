package mock_repository

import (
	"context"
	"expenses/internal/models"
	"fmt"
	"sync"
	"time"
)

type MockAnalyticsRepository struct {
	balances  map[string]map[int64]float64               // key: userId_startDate_endDate, value: accountId -> balance
	analytics map[int64][]models.AccountBalanceAnalytics // key: userId, value: analytics
	mu        sync.RWMutex
}

func NewMockAnalyticsRepository() *MockAnalyticsRepository {
	return &MockAnalyticsRepository{
		balances:  make(map[string]map[int64]float64),
		analytics: make(map[int64][]models.AccountBalanceAnalytics),
	}
}

func (m *MockAnalyticsRepository) GetBalance(ctx context.Context, userId int64, startDate *time.Time, endDate *time.Time) (map[int64]float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a key based on parameters
	key := m.createBalanceKey(userId, startDate, endDate)

	if balances, exists := m.balances[key]; exists {
		return balances, nil
	}

	// Return empty map if no data found
	return make(map[int64]float64), nil
}

func (m *MockAnalyticsRepository) GetAccountAnalytics(ctx context.Context, userId int64) ([]models.AccountBalanceAnalytics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if analytics, exists := m.analytics[userId]; exists {
		return analytics, nil
	}

	// Return empty slice if no data found
	return []models.AccountBalanceAnalytics{}, nil
}

// Helper methods for testing
func (m *MockAnalyticsRepository) SetBalance(userId int64, startDate *time.Time, endDate *time.Time, balances map[int64]float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.createBalanceKey(userId, startDate, endDate)
	m.balances[key] = balances
}

func (m *MockAnalyticsRepository) SetAnalytics(userId int64, analytics []models.AccountBalanceAnalytics) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.analytics[userId] = analytics
}

func (m *MockAnalyticsRepository) createBalanceKey(userId int64, startDate *time.Time, endDate *time.Time) string {
	key := fmt.Sprintf("%d_", userId)
	if startDate != nil {
		key += startDate.Format("2006-01-02")
	} else {
		key += "nil"
	}
	key += "_"
	if endDate != nil {
		key += endDate.Format("2006-01-02")
	} else {
		key += "nil"
	}
	return key
}
