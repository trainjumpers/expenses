package mock_repository

import (
	"context"
	"expenses/internal/models"
	"fmt"
	"sync"
	"time"
)

type MockAnalyticsRepository struct {
	balances              map[string]map[int64]float64                 // key: userId_startDate_endDate, value: accountId -> balance
	analytics             map[int64][]models.AccountBalanceAnalytics   // key: userId, value: analytics
	networthData          map[string]networthMockData                  // key: userId_startDate_endDate, value: networth data
	categoryAnalytics     map[string]*models.CategoryAnalyticsResponse // key: userId_startDate_endDate, value: category analytics
	monthlyAnalytics      map[string]*models.MonthlyAnalyticsResponse  // key: userId_months, value: monthly analytics
	shouldErrorOnBalance  bool                                         // simulate GetBalance errors
	shouldErrorOnNetworth bool                                         // simulate GetNetworthTimeSeries errors
	shouldErrorOnCategory bool                                         // simulate GetCategoryAnalytics errors
	shouldErrorOnMonthly  bool                                         // simulate GetMonthlyAnalytics errors
	mu                    sync.RWMutex
}

type networthMockData struct {
	initialBalance float64
	timeSeries     []map[string]any
}

func NewMockAnalyticsRepository() *MockAnalyticsRepository {
	return &MockAnalyticsRepository{
		balances:              make(map[string]map[int64]float64),
		analytics:             make(map[int64][]models.AccountBalanceAnalytics),
		networthData:          make(map[string]networthMockData),
		categoryAnalytics:     make(map[string]*models.CategoryAnalyticsResponse),
		monthlyAnalytics:      make(map[string]*models.MonthlyAnalyticsResponse),
		shouldErrorOnBalance:  false,
		shouldErrorOnNetworth: false,
		shouldErrorOnCategory: false,
		shouldErrorOnMonthly:  false,
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
		// Negate values to mimic real repository's `* -1` on SUM(amount)
		negatedBalances := make(map[int64]float64)
		for accID, balance := range balances {
			negatedBalances[accID] = -balance
		}
		return negatedBalances, nil
	}

	// Return empty map if no data found
	return make(map[int64]float64), nil
}

func (m *MockAnalyticsRepository) GetNetworthTimeSeries(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (float64, float64, float64, []map[string]any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Simulate error if configured
	if m.shouldErrorOnNetworth {
		return 0, 0, 0, nil, fmt.Errorf("simulated GetNetworthTimeSeries error")
	}

	// Create a key based on parameters
	key := m.createNetworthKey(userId, startDate, endDate)

	if data, exists := m.networthData[key]; exists {
		// Negate values to mimic real repository's `* -1`
		negatedInitialBalance := -data.initialBalance
		var negatedTimeSeries []map[string]any
		for _, point := range data.timeSeries {
			dailyChange, ok := point["daily_change"].(float64)
			if !ok {
				return 0, 0, 0, nil, fmt.Errorf("invalid type for daily_change in daily data")
			}
			negatedPoint := map[string]any{
				"date":         point["date"],
				"daily_change": -dailyChange,
			}
			negatedTimeSeries = append(negatedTimeSeries, negatedPoint)
		}
		// Return sample total income and expenses
		return negatedInitialBalance, 500.0, 300.0, negatedTimeSeries, nil
	}

	// Return default sample data if no specific data set
	initialBalance := 1000.0
	timeSeries := []map[string]any{
		{
			"date":         startDate.Format("2006-01-02"),
			"daily_change": 100.0,
		},
		{
			"date":         startDate.AddDate(0, 0, 1).Format("2006-01-02"),
			"daily_change": -50.0,
		},
	}

	// Negate values to mimic real repository's `* -1`
	negatedInitialBalance := -initialBalance
	var negatedTimeSeries []map[string]interface{}
	for _, point := range timeSeries {
		negatedPoint := map[string]interface{}{
			"date":         point["date"],
			"daily_change": -point["daily_change"].(float64),
		}
		negatedTimeSeries = append(negatedTimeSeries, negatedPoint)
	}
	return negatedInitialBalance, 500.0, 300.0, negatedTimeSeries, nil
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

func (m *MockAnalyticsRepository) GetCategoryAnalytics(ctx context.Context, userId int64, startDate time.Time, endDate time.Time, categoryIds []int64) (*models.CategoryAnalyticsResponse, error) {
	_ = categoryIds
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Simulate error if configured
	if m.shouldErrorOnCategory {
		return nil, fmt.Errorf("simulated GetCategoryAnalytics error")
	}

	// Create a key based on parameters
	key := m.createCategoryKey(userId, startDate, endDate)

	if analytics, exists := m.categoryAnalytics[key]; exists {
		return analytics, nil
	}

	// Return default sample data if no specific data set
	defaultAnalytics := &models.CategoryAnalyticsResponse{
		CategoryTransactions: []models.CategoryTransaction{
			{
				CategoryID:   1,
				CategoryName: "Food",
				TotalAmount:  150.0,
			},
			{
				CategoryID:   2,
				CategoryName: "Transportation",
				TotalAmount:  200.0,
			},
		},
	}

	return defaultAnalytics, nil
}

func (m *MockAnalyticsRepository) GetMonthlyAnalytics(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (*models.MonthlyAnalyticsResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Simulate error if configured
	if m.shouldErrorOnMonthly {
		return nil, fmt.Errorf("simulated GetMonthlyAnalytics error")
	}

	// Create a key based on parameters
	key := m.createMonthlyKey(userId, startDate, endDate)

	if analytics, exists := m.monthlyAnalytics[key]; exists {
		return analytics, nil
	}

	// Return default sample data if no specific data set
	defaultAnalytics := &models.MonthlyAnalyticsResponse{
		TotalIncome:   1000.0,
		TotalExpenses: 600.0,
		TotalAmount:   400.0,
	}

	return defaultAnalytics, nil
}

func (m *MockAnalyticsRepository) GetAccountCashFlows(ctx context.Context, userId int64, accountIds []int64) ([]models.AccountCashFlow, error) {
	_ = userId
	_ = accountIds
	return []models.AccountCashFlow{}, nil
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

func (m *MockAnalyticsRepository) SetNetworthTimeSeries(userId int64, startDate time.Time, endDate time.Time, initialBalance float64, timeSeries []map[string]any) {
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

func (m *MockAnalyticsRepository) SetCategoryAnalytics(userId int64, startDate time.Time, endDate time.Time, analytics *models.CategoryAnalyticsResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.createCategoryKey(userId, startDate, endDate)
	m.categoryAnalytics[key] = analytics
}

func (m *MockAnalyticsRepository) SetShouldErrorOnCategory(shouldError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldErrorOnCategory = shouldError
}

func (m *MockAnalyticsRepository) SetMonthlyAnalytics(userId int64, startDate time.Time, endDate time.Time, analytics *models.MonthlyAnalyticsResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.createMonthlyKey(userId, startDate, endDate)
	m.monthlyAnalytics[key] = analytics
}

func (m *MockAnalyticsRepository) SetShouldErrorOnMonthly(shouldError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldErrorOnMonthly = shouldError
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

func (m *MockAnalyticsRepository) createCategoryKey(userId int64, startDate time.Time, endDate time.Time) string {
	return fmt.Sprintf("%d_%s_%s", userId, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
}

func (m *MockAnalyticsRepository) createMonthlyKey(userId int64, startDate time.Time, endDate time.Time) string {
	return fmt.Sprintf("%d_%s_%s", userId, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
}
