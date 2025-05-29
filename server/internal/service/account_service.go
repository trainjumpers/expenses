package service

import (
	"expenses/internal/models"
	"expenses/internal/repository"
	"expenses/pkg/logger"

	"github.com/gin-gonic/gin"
)

type AccountServiceInterface interface {
	CreateAccount(c *gin.Context, input models.CreateAccountInput) (models.AccountResponse, error)
	GetAccountById(c *gin.Context, accountId int64, userId int64) (models.AccountResponse, error)
	UpdateAccount(c *gin.Context, accountId int64, userId int64, input models.UpdateAccountInput) (models.AccountResponse, error)
	DeleteAccount(c *gin.Context, accountId int64, userId int64) error
	ListAccounts(c *gin.Context, userId int64) ([]models.AccountResponse, error)
}

type AccountService struct {
	repo repository.AccountRepositoryInterface
}

func NewAccountService(repo repository.AccountRepositoryInterface) AccountServiceInterface {
	return &AccountService{repo: repo}
}

func (s *AccountService) CreateAccount(c *gin.Context, input models.CreateAccountInput) (models.AccountResponse, error) {
	logger.Debug("AccountService: Creating account")
	if input.Balance == nil {
		zero := 0.0
		input.Balance = &zero
	}
	return s.repo.CreateAccount(c, input)
}

func (s *AccountService) GetAccountById(c *gin.Context, accountId int64, userId int64) (models.AccountResponse, error) {
	logger.Debug("AccountService: Getting account by id")
	return s.repo.GetAccountById(c, accountId, userId)
}

func (s *AccountService) UpdateAccount(c *gin.Context, accountId int64, userId int64, input models.UpdateAccountInput) (models.AccountResponse, error) {
	logger.Debug("AccountService: Updating account")
	return s.repo.UpdateAccount(c, accountId, userId, input)
}

func (s *AccountService) DeleteAccount(c *gin.Context, accountId int64, userId int64) error {
	logger.Debug("AccountService: Deleting account")
	return s.repo.DeleteAccount(c, accountId, userId)
}

func (s *AccountService) ListAccounts(c *gin.Context, userId int64) ([]models.AccountResponse, error) {
	logger.Debug("AccountService: Listing accounts")
	return s.repo.ListAccounts(c, userId)
}
