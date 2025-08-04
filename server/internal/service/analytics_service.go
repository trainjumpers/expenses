package service

import (
	"context"
	"expenses/internal/models"
	"expenses/internal/repository"
	"time"
)

type AnalyticsServiceInterface interface {
	GetAccountAnalytics(ctx context.Context, userId int64) (models.AccountAnalyticsListResponse, error)
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
