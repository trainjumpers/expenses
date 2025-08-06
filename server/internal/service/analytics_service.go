package service

import (
	"context"
	"expenses/internal/models"
	"expenses/internal/repository"
	"fmt"
	"time"
)

type AnalyticsServiceInterface interface {
	GetAccountAnalytics(ctx context.Context, userId int64) (models.AccountAnalyticsListResponse, error)
	GetNetworthTimeSeries(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (models.NetworthTimeSeriesResponse, error)
	GetCategoryAnalytics(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (*models.CategoryAnalyticsResponse, error)
}

type AnalyticsService struct {
	analyticsRepo repository.AnalyticsRepositoryInterface
	accountRepo   repository.AccountRepositoryInterface
}

func NewAnalyticsService(analyticsRepo repository.AnalyticsRepositoryInterface, accountRepo repository.AccountRepositoryInterface) AnalyticsServiceInterface {
	return &AnalyticsService{
		analyticsRepo: analyticsRepo,
		accountRepo:   accountRepo,
	}
}

func (s *AnalyticsService) GetAccountAnalytics(ctx context.Context, userId int64) (models.AccountAnalyticsListResponse, error) {
	// Get all user accounts to ensure we include accounts with no transactions
	accounts, err := s.accountRepo.ListAccounts(ctx, userId)
	if err != nil {
		// If no accounts found, return empty analytics (not an error)
		return models.AccountAnalyticsListResponse{
			AccountAnalytics: []models.AccountBalanceAnalytics{},
		}, nil
	}

	// Get current balances (all transactions)
	currentBalances, err := s.analyticsRepo.GetBalance(ctx, userId, nil, nil)
	if err != nil {
		return models.AccountAnalyticsListResponse{}, err
	}

	// Calculate one month ago date
	oneMonthAgo := time.Now().AddDate(0, -1, 0)

	// Get balances from one month ago
	historicalBalances, err := s.analyticsRepo.GetBalance(ctx, userId, nil, &oneMonthAgo)
	if err != nil {
		return models.AccountAnalyticsListResponse{}, err
	}

	// Build analytics response ensuring all accounts are included
	var accountAnalytics []models.AccountBalanceAnalytics
	for _, account := range accounts {
		currentBalance := currentBalances[account.Id]       // defaults to 0 if not found
		historicalBalance := historicalBalances[account.Id] // defaults to 0 if not found

		accountAnalytics = append(accountAnalytics, models.AccountBalanceAnalytics{
			AccountID:          account.Id,
			CurrentBalance:     currentBalance,
			BalanceOneMonthAgo: historicalBalance,
		})
	}

	return models.AccountAnalyticsListResponse{
		AccountAnalytics: accountAnalytics,
	}, nil
}

func (s *AnalyticsService) GetNetworthTimeSeries(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (models.NetworthTimeSeriesResponse, error) {
	accounts, err := s.accountRepo.ListAccounts(ctx, userId)

	if err != nil {
		return models.NetworthTimeSeriesResponse{}, err
	}

	// Get initial balance and daily changes from repository
	initialBalance, totalIncome, totalExpenses, dailyData, err := s.analyticsRepo.GetNetworthTimeSeries(ctx, userId, startDate, endDate)
	totalAccountBalance := 0.0
	if err != nil {
		return models.NetworthTimeSeriesResponse{}, err
	}

	for _, account := range accounts {
		initialBalance += account.Balance
		totalAccountBalance += account.Balance
	}

	var timeSeries []models.NetworthDataPoint
	runningBalance := initialBalance

	// Create a map of dates with daily changes for easy lookup
	dailyChanges := make(map[string]float64)
	for _, data := range dailyData {
		date, ok := data["date"].(string)
		if !ok {
			return models.NetworthTimeSeriesResponse{}, fmt.Errorf("invalid type for date in daily data")
		}
		dailyChange, ok := data["daily_change"].(float64)
		if !ok {
			return models.NetworthTimeSeriesResponse{}, fmt.Errorf("invalid type for daily_change in daily data")
		}
		dailyChanges[date] = dailyChange
	}

	// Generate time series for each day in the range
	currentDate := startDate
	for currentDate.Before(endDate.AddDate(0, 0, 1)) { // Include end date
		dateStr := currentDate.Format("2006-01-02")

		// Add daily change if it exists
		if dailyChange, exists := dailyChanges[dateStr]; exists {
			runningBalance += dailyChange
		}

		if runningBalance == totalAccountBalance {
			// Txn has not changed yet, so we can skip adding this point
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}

		timeSeries = append(timeSeries, models.NetworthDataPoint{
			Date:     dateStr,
			Networth: runningBalance,
		})

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	if len(timeSeries) == 0 {
		timeSeries = append(timeSeries, models.NetworthDataPoint{
			Date:     startDate.Format("2006-01-02"),
			Networth: initialBalance,
		})
	}

	return models.NetworthTimeSeriesResponse{
		InitialBalance: initialBalance, // Initial balance for frontend
		TotalIncome:    totalIncome,
		TotalExpenses:  totalExpenses,
		TimeSeries:     timeSeries,
	}, nil
}

func (s *AnalyticsService) GetCategoryAnalytics(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (*models.CategoryAnalyticsResponse, error) {
	return s.analyticsRepo.GetCategoryAnalytics(ctx, userId, startDate, endDate)
}
