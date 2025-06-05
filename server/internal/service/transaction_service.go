package service

import (
	customErrors "expenses/internal/errors"
	"expenses/internal/models"
	"expenses/internal/repository"
	"expenses/pkg/logger"
	"fmt"
	"strings"
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
	logger.Info("Creating transaction for user: ", input.CreatedBy)
	if err := s.validateCreateTransaction(input); err != nil {
		return models.TransactionResponse{}, err
	}
	return s.repo.CreateTransaction(c, input)
}

func (s *TransactionService) GetTransactionById(c *gin.Context, transactionId int64, userId int64) (models.TransactionResponse, error) {
	logger.Info("Fetching transaction by ID: ", transactionId, " for user: ", userId)
	return s.repo.GetTransactionById(c, transactionId, userId)
}

func (s *TransactionService) UpdateTransaction(c *gin.Context, transactionId int64, userId int64, input models.UpdateTransactionInput) (models.TransactionResponse, error) {
	logger.Info("Updating transaction ID: ", transactionId, " for user: ", userId)

	if err := s.validateUpdateTransaction(input); err != nil {
		return models.TransactionResponse{}, err
	}

	return s.repo.UpdateTransaction(c, transactionId, userId, input)
}

func (s *TransactionService) DeleteTransaction(c *gin.Context, transactionId int64, userId int64) error {
	logger.Info("Deleting transaction ID: ", transactionId, " for user: ", userId)
	return s.repo.DeleteTransaction(c, transactionId, userId)
}

func (s *TransactionService) ListTransactions(c *gin.Context, userId int64) ([]models.TransactionResponse, error) {
	logger.Info("Listing transactions for user: ", userId)
	return s.repo.ListTransactions(c, userId)
}

// validateCreateTransaction performs business rule validation for create operations
func (s *TransactionService) validateCreateTransaction(input models.CreateTransactionInput) error {
	if err := s.validateDateNotInFuture(input.Date); err != nil {
		return err
	}
	if strings.TrimSpace(input.Name) == "" {
		return customErrors.New("transaction name cannot be empty")
	}

	return nil
}

// validateUpdateTransaction performs business rule validation for update operations
func (s *TransactionService) validateUpdateTransaction(input models.UpdateTransactionInput) error {
	// Validate date is not in the future (only if date is being updated)
	if !input.Date.IsZero() {
		if err := s.validateDateNotInFuture(input.Date); err != nil {
			return err
		}
	}

	// Validate name is not empty (only if name is being updated)
	if input.Name != "" && strings.TrimSpace(input.Name) == "" {
		return customErrors.New("transaction name cannot be empty")
	}

	return nil
}

// validateDateNotInFuture validates that a date is not in the future
func (s *TransactionService) validateDateNotInFuture(date time.Time) error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	if date.After(today) {
		return customErrors.New(fmt.Sprintf("transaction date cannot be in the future. Provided date: %s", date.Format("2006-01-02")))
	}

	return nil
}
