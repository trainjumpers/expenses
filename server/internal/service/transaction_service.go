package service

import (
	customErrors "expenses/internal/errors"
	"expenses/internal/models"
	"expenses/internal/repository"
	"expenses/pkg/logger"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type TransactionServiceInterface interface {
	CreateTransaction(c *gin.Context, input models.CreateTransactionInput) (models.TransactionResponse, error)
	GetTransactionById(c *gin.Context, transactionId int64, userId int64) (models.TransactionResponse, error)
	UpdateTransaction(c *gin.Context, transactionId int64, userId int64, input models.UpdateTransactionInput) (models.TransactionResponse, error)
	DeleteTransaction(c *gin.Context, transactionId int64, userId int64) error
	ListTransactions(c *gin.Context, userId int64) ([]models.TransactionResponse, error)
}

type TransactionService struct {
	repo repository.TransactionRepositoryInterface
}

func NewTransactionService(repo repository.TransactionRepositoryInterface) TransactionServiceInterface {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) CreateTransaction(c *gin.Context, input models.CreateTransactionInput) (models.TransactionResponse, error) {
	logger.Debugf("Creating transaction for user %d", input.CreatedBy)

	if err := s.validateCreateTransaction(input); err != nil {
		return models.TransactionResponse{}, err
	}

	return s.repo.CreateTransaction(c, input)
}

func (s *TransactionService) GetTransactionById(c *gin.Context, transactionId int64, userId int64) (models.TransactionResponse, error) {
	logger.Debugf("Fetching transaction by ID %d for user %d", transactionId, userId)
	return s.repo.GetTransactionById(c, transactionId, userId)
}

func (s *TransactionService) UpdateTransaction(c *gin.Context, transactionId int64, userId int64, input models.UpdateTransactionInput) (models.TransactionResponse, error) {
	logger.Debugf("Updating transaction ID %d for user %d", transactionId, userId)

	if err := s.validateUpdateTransaction(input); err != nil {
		return models.TransactionResponse{}, err
	}

	return s.repo.UpdateTransaction(c, transactionId, userId, input)
}

func (s *TransactionService) DeleteTransaction(c *gin.Context, transactionId int64, userId int64) error {
	logger.Debugf("Deleting transaction ID %d for user %d", transactionId, userId)
	return s.repo.DeleteTransaction(c, transactionId, userId)
}

func (s *TransactionService) ListTransactions(c *gin.Context, userId int64) ([]models.TransactionResponse, error) {
	logger.Debugf("Listing transactions for user %d", userId)
	return s.repo.ListTransactions(c, userId)
}

// validateCreateTransaction performs business rule validation for create operations
func (s *TransactionService) validateCreateTransaction(input models.CreateTransactionInput) error {
	if err := s.validateDateNotInFuture(input.Date); err != nil {
		return err
	}
	return nil
}

func (s *TransactionService) validateUpdateTransaction(input models.UpdateTransactionInput) error {
	if err := s.validateDateNotInFuture(input.Date); err != nil {
		return err
	}
	return nil
}

// validateDateNotInFuture validates that a date is not in the future
func (s *TransactionService) validateDateNotInFuture(date time.Time) error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	if date.After(today) {
		return customErrors.NewTransactionDateInFutureError(fmt.Errorf("transaction date cannot be in the future. Provided date: %s", date.Format("2006-01-02")))
	}

	return nil
}
