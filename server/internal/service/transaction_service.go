package service

import (
	database "expenses/internal/database/manager"
	customErrors "expenses/internal/errors"
	"expenses/internal/models"
	"expenses/internal/repository"
	"expenses/pkg/utils"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type TransactionServiceInterface interface {
	CreateTransaction(c *gin.Context, input models.CreateTransactionInput) (models.TransactionResponse, error)
	GetTransactionById(c *gin.Context, transactionId int64, userId int64) (models.TransactionResponse, error)
	UpdateTransaction(c *gin.Context, transactionId int64, userId int64, input models.UpdateTransactionInput) (models.TransactionResponse, error)
	DeleteTransaction(c *gin.Context, transactionId int64, userId int64) error
	ListTransactions(c *gin.Context, userId int64, query models.TransactionListQuery) (models.PaginatedTransactionsResponse, error)
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

func (s *TransactionService) CreateTransaction(c *gin.Context, input models.CreateTransactionInput) (models.TransactionResponse, error) {
	if err := s.validateCreateTransaction(c, input); err != nil {
		return models.TransactionResponse{}, err
	}

	transactionInput := models.CreateBaseTransactionInput{}
	utils.ConvertStruct(&input, &transactionInput)
	return s.repo.CreateTransaction(c, transactionInput, input.CategoryIds)
}

func (s *TransactionService) GetTransactionById(c *gin.Context, transactionId int64, userId int64) (models.TransactionResponse, error) {
	return s.repo.GetTransactionById(c, transactionId, userId)
}

func (s *TransactionService) UpdateTransaction(c *gin.Context, transactionId int64, userId int64, input models.UpdateTransactionInput) (models.TransactionResponse, error) {
	if err := s.validateUpdateTransaction(c, input, userId); err != nil {
		return models.TransactionResponse{}, err
	}

	var transaction models.TransactionResponse
	err := s.db.WithTxn(c, func(tx pgx.Tx) error {
		// Update base transaction if there are fields to update
		var baseInput models.UpdateBaseTransactionInput
		utils.ConvertStruct(&input, &baseInput)
		err := s.repo.UpdateTransaction(c, transactionId, userId, baseInput)
		if err != nil && (err.Error() != customErrors.NoFieldsToUpdateError().Error() ||
			(input.CategoryIds == nil && input.AccountId == nil)) {
			return err
		}

		// Update category mapping if provided
		if input.CategoryIds != nil {
			err = s.repo.UpdateCategoryMapping(c, transactionId, userId, *input.CategoryIds)
			if err != nil {
				return err
			}
		}

		// Get the updated transaction
		updatedTransaction, err := s.repo.GetTransactionById(c, transactionId, userId)
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

func (s *TransactionService) DeleteTransaction(c *gin.Context, transactionId int64, userId int64) error {
	return s.repo.DeleteTransaction(c, transactionId, userId)
}

// ListTransactions returns paginated, sorted, and filtered transactions for a user
func (s *TransactionService) ListTransactions(c *gin.Context, userId int64, query models.TransactionListQuery) (models.PaginatedTransactionsResponse, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 15
	}

	return s.repo.ListTransactions(c, userId, query)
}

// validateCreateTransaction performs business rule validation for create operations
func (s *TransactionService) validateCreateTransaction(c *gin.Context, input models.CreateTransactionInput) error {
	if err := s.validateDateNotInFuture(input.Date); err != nil {
		return err
	}
	if err := s.validateAccountExists(c, input.AccountId, input.CreatedBy); err != nil {
		return err
	}
	if err := s.validateCategoryExists(c, input.CategoryIds, input.CreatedBy); err != nil {
		return err
	}
	return nil
}

func (s *TransactionService) validateUpdateTransaction(c *gin.Context, input models.UpdateTransactionInput, userId int64) error {
	if err := s.validateDateNotInFuture(input.Date); err != nil {
		return err
	}
	if id := input.AccountId; id != nil {
		if err := s.validateAccountExists(c, *id, userId); err != nil {
			return err
		}
	}

	if ids := input.CategoryIds; ids != nil {
		if err := s.validateCategoryExists(c, *ids, userId); err != nil {
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

func (s *TransactionService) validateAccountExists(c *gin.Context, accountId int64, userId int64) error {
	_, err := s.accountRepo.GetAccountById(c, accountId, userId)
	return err
}

func (s *TransactionService) validateCategoryExists(c *gin.Context, categoryIds []int64, userId int64) error {
	if len(categoryIds) == 0 {
		return nil
	}
	categories, err := s.categoryRepo.ListCategories(c, userId)
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
