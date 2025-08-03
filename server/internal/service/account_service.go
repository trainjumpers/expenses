package service

import (
	"context"
	"expenses/internal/models"
	"expenses/internal/repository"
)

type AccountServiceInterface interface {
	CreateAccount(ctx context.Context, input models.CreateAccountInput) (models.AccountResponse, error)
	GetAccountById(ctx context.Context, accountId int64, userId int64) (models.AccountResponse, error)
	UpdateAccount(ctx context.Context, accountId int64, userId int64, input models.UpdateAccountInput) (models.AccountResponse, error)
	DeleteAccount(ctx context.Context, accountId int64, userId int64) error
	ListAccounts(ctx context.Context, userId int64) ([]models.AccountResponse, error)
}

type AccountService struct {
	repo repository.AccountRepositoryInterface
}

func NewAccountService(repo repository.AccountRepositoryInterface) AccountServiceInterface {
	return &AccountService{repo: repo}
}

func (s *AccountService) CreateAccount(ctx context.Context, input models.CreateAccountInput) (models.AccountResponse, error) {
	if input.Balance == nil {
		zero := 0.0
		input.Balance = &zero
	}
	return s.repo.CreateAccount(ctx, input)
}

func (s *AccountService) GetAccountById(ctx context.Context, accountId int64, userId int64) (models.AccountResponse, error) {
	return s.repo.GetAccountById(ctx, accountId, userId)
}

func (s *AccountService) UpdateAccount(ctx context.Context, accountId int64, userId int64, input models.UpdateAccountInput) (models.AccountResponse, error) {
	return s.repo.UpdateAccount(ctx, accountId, userId, input)
}

func (s *AccountService) DeleteAccount(ctx context.Context, accountId int64, userId int64) error {
	return s.repo.DeleteAccount(ctx, accountId, userId)
}

func (s *AccountService) ListAccounts(ctx context.Context, userId int64) ([]models.AccountResponse, error) {
	return s.repo.ListAccounts(ctx, userId)
}
