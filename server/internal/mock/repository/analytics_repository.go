package mock_repository

import (
	"context"
	"expenses/internal/models"
	"fmt"
	"sync"
	"time"
)

type MockAnalyticsRepository struct {
	balances              map[string]map[int64]float64               // key: userId_startDate_endDate, value: accountId -> balance
	analytics             map[int64][]models.AccountBalanceAnalytics // key: userId, value: analytics
	networthData          map[string]networthMockData                // key: userId_startDate_endDate, value: networth data
	shouldErrorOnBalance  bool                                       // simulate GetBalance errors
	shouldErrorOnNetworth bool                                       // simulate GetNetworthTimeSeries errors
	mu                    sync.RWMutex
}

type networthMockData struct {
	initialBalance float64
	timeSeries     []map[string]interface{}
}

func NewMockAnalyticsRepository() *MockAnalyticsRepository {
	return &MockAnalyticsRepository{
		balances:              make(map[string]map[int64]float64),
		analytics:             make(map[int64][]models.AccountBalanceAnalytics),
		networthData:          make(map[string]networthMockData),
		shouldErrorOnBalance:  false,
		shouldErrorOnNetworth: false,
	}
}

func (m *MockAnalyticsRepository) GetBalance(ctx context.Context, userId int64, startDate *time.Time, endDate *time.Time) (map[int64]float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Simulate error if configured
	if m.shouldErrorOnBalance {
		return nil, fmt.Errorf("simulated GetBalance error")
	}

	// Create a key based on parameters
	key := m.createBalanceKey(userId, startDate, endDate)

	if balances, exists := m.balances[key]; exists {
		return balances, nil
	}

	// Return empty map if no data found
	return make(map[int64]float64), nil
}

func (m *MockAnalyticsRepository) GetNetworthTimeSeries(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (float64, []map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Simulate error if configured
	if m.shouldErrorOnNetworth {
		return 0, nil, fmt.Errorf("simulated GetNetworthTimeSeries error")
	}

	// Create a key based on parameters
	key := m.createNetworthKey(userId, startDate, endDate)

	if data, exists := m.networthData[key]; exists {
		return data.initialBalance, data.timeSeries, nil
	}

	// Return default sample data if no specific data set
	initialBalance := 1000.0
	timeSeries := []map[string]interface{}{
		{
			"date":         startDate.Format("2006-01-02"),
			"daily_change": 100.0,
		},
		{
			"date":         startDate.AddDate(0, 0, 1).Format("2006-01-02"),
			"daily_change": -50.0,
		},
	}

	return initialBalance, timeSeries, nil
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

func (m *MockAnalyticsRepository) SetNetworthTimeSeries(userId int64, startDate time.Time, endDate time.Time, initialBalance float64, timeSeries []map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.createNetworthKey(userId, startDate, endDate)
	m.networthData[key] = networthMockData{
		initialBalance: initialBalance,
		timeSeries:     timeSeries,
	}
}

func (m *MockAnalyticsRepository) SetShouldErrorOnBalance(shouldError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldErrorOnBalance = shouldError
}

func (m *MockAnalyticsRepository) SetShouldErrorOnNetworth(shouldError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldErrorOnNetworth = shouldError
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

func (m *MockAnalyticsRepository) createNetworthKey(userId int64, startDate time.Time, endDate time.Time) string {
	return fmt.Sprintf("%d_%s_%s", userId, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
}
