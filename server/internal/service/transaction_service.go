package service

import (
	"context"
	customErrors "expenses/internal/errors"
	"expenses/internal/models"
	"expenses/internal/repository"
	database "expenses/pkg/database/manager"
	"expenses/pkg/utils"
	"fmt"
	"time"
)

type TransactionServiceInterface interface {
	CreateTransaction(ctx context.Context, input models.CreateTransactionInput) (models.TransactionResponse, error)
	GetTransactionById(ctx context.Context, transactionId int64, userId int64) (models.TransactionResponse, error)
	UpdateTransaction(ctx context.Context, transactionId int64, userId int64, input models.UpdateTransactionInput) (models.TransactionResponse, error)
	DeleteTransaction(ctx context.Context, transactionId int64, userId int64) error
	ListTransactions(ctx context.Context, userId int64, query models.TransactionListQuery) (models.PaginatedTransactionsResponse, error)
}

type TransactionService struct {
	repo         repository.TransactionRepositoryInterface
	categoryRepo repository.CategoryRepositoryInterface
	accountRepo  repository.AccountRepositoryInterface
	db           database.DatabaseManager
}

func NewTransactionService(
	repo repository.TransactionRepositoryInterface,
	categoryRepo repository.CategoryRepositoryInterface,
	accountRepo repository.AccountRepositoryInterface,
	db database.DatabaseManager,
) TransactionServiceInterface {
	return &TransactionService{
		repo:         repo,
		categoryRepo: categoryRepo,
		accountRepo:  accountRepo,
		db:           db,
	}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, input models.CreateTransactionInput) (models.TransactionResponse, error) {
	if err := s.validateCreateTransaction(ctx, input); err != nil {
		return models.TransactionResponse{}, err
	}

	transactionInput := models.CreateBaseTransactionInput{}
	utils.ConvertStruct(&input, &transactionInput)
	return s.repo.CreateTransaction(ctx, transactionInput, input.CategoryIds)
}

func (s *TransactionService) GetTransactionById(ctx context.Context, transactionId int64, userId int64) (models.TransactionResponse, error) {
	return s.repo.GetTransactionById(ctx, transactionId, userId)
}

func (s *TransactionService) UpdateTransaction(ctx context.Context, transactionId int64, userId int64, input models.UpdateTransactionInput) (models.TransactionResponse, error) {
	if err := s.validateUpdateTransaction(ctx, input, userId); err != nil {
		return models.TransactionResponse{}, err
	}

	var transaction models.TransactionResponse
	err := s.db.WithTxn(ctx, func(txCtx context.Context) error {
		// Update base transaction if there are fields to update
		var baseInput models.UpdateBaseTransactionInput
		utils.ConvertStruct(&input, &baseInput)
		err := s.repo.UpdateTransaction(txCtx, transactionId, userId, baseInput)
		if err != nil && (err.Error() != customErrors.NoFieldsToUpdateError().Error() ||
			(input.CategoryIds == nil && input.AccountId == nil)) {
			return err
		}

		// Update category mapping if provided
		if input.CategoryIds != nil {
			err = s.repo.UpdateCategoryMapping(txCtx, transactionId, userId, *input.CategoryIds)
			if err != nil {
				return err
			}
		}

		// Get the updated transaction
		updatedTransaction, err := s.repo.GetTransactionById(txCtx, transactionId, userId)
		if err != nil {
			return err
		}
		transaction = updatedTransaction
		return nil
	})

	if err != nil {
		return models.TransactionResponse{}, err
	}

	return transaction, nil
}

func (s *TransactionService) DeleteTransaction(ctx context.Context, transactionId int64, userId int64) error {
	return s.repo.DeleteTransaction(ctx, transactionId, userId)
}

// ListTransactions returns paginated, sorted, and filtered transactions for a user
func (s *TransactionService) ListTransactions(ctx context.Context, userId int64, query models.TransactionListQuery) (models.PaginatedTransactionsResponse, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 15
	}

	return s.repo.ListTransactions(ctx, userId, query)
}

// validateCreateTransaction performs business rule validation for create operations
func (s *TransactionService) validateCreateTransaction(ctx context.Context, input models.CreateTransactionInput) error {
	if err := s.validateDateNotInFuture(input.Date); err != nil {
		return err
	}
	if err := s.validateAccountExists(ctx, input.AccountId, input.CreatedBy); err != nil {
		return err
	}
	if err := s.validateCategoryExists(ctx, input.CategoryIds, input.CreatedBy); err != nil {
		return err
	}
	return nil
}

func (s *TransactionService) validateUpdateTransaction(ctx context.Context, input models.UpdateTransactionInput, userId int64) error {
	if err := s.validateDateNotInFuture(input.Date); err != nil {
		return err
	}
	if id := input.AccountId; id != nil {
		if err := s.validateAccountExists(ctx, *id, userId); err != nil {
			return err
		}
	}

	if ids := input.CategoryIds; ids != nil {
		if err := s.validateCategoryExists(ctx, *ids, userId); err != nil {
			return err
		}
	}
	return nil
}

func (s *TransactionService) validateDateNotInFuture(date time.Time) error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	if date.After(today) {
		return customErrors.NewTransactionDateInFutureError(fmt.Errorf("transaction date cannot be in the future. Provided date: %s", date.Format("2006-01-02")))
	}

	return nil
}

func (s *TransactionService) validateAccountExists(ctx context.Context, accountId int64, userId int64) error {
	_, err := s.accountRepo.GetAccountById(ctx, accountId, userId)
	return err
}

func (s *TransactionService) validateCategoryExists(ctx context.Context, categoryIds []int64, userId int64) error {
	if len(categoryIds) == 0 {
		return nil
	}
	categories, err := s.categoryRepo.ListCategories(ctx, userId)
	if err != nil {
		return err
	}
	categoryMap := make(map[int64]bool)
	for _, category := range categories {
		categoryMap[category.Id] = true
	}
	for _, id := range categoryIds {
		if !categoryMap[id] {
			return customErrors.NewCategoryNotFoundError(fmt.Errorf("category with id %d not found for user %d", id, userId))
		}
	}
	return nil
}
