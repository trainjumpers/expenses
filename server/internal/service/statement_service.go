package service

import (
	"errors"
	"expenses/internal/models"
	"expenses/internal/parser"
	"expenses/internal/repository"
	"expenses/internal/validator"
	"expenses/pkg/logger"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

type StatementServiceInterface interface {
	ParseStatement(c *gin.Context, input models.ParseStatementInput, userId int64) (models.StatementResponse, error)
	GetStatementStatus(c *gin.Context, statementId int64, userId int64) (models.StatementResponse, error)
	ListStatements(c *gin.Context, userId int64, page int, pageSize int) (models.PaginatedStatementResponse, error)
	PreviewStatement(c *gin.Context, fileBytes []byte, fileName string, skipRows int, rowSize int) (*models.StatementPreview, error)
}

type StatementService struct {
	repo               repository.StatementRepositoryInterface
	accountService     AccountServiceInterface
	txService          TransactionServiceInterface
	statementValidator *validator.StatementValidator
}

func NewStatementService(
	repo repository.StatementRepositoryInterface,
	accountService AccountServiceInterface,
	statementValidator *validator.StatementValidator,
	txService TransactionServiceInterface,
) StatementServiceInterface {
	return &StatementService{
		repo:               repo,
		accountService:     accountService,
		txService:          txService,
		statementValidator: statementValidator,
	}
}

func (s *StatementService) ParseStatement(c *gin.Context, input models.ParseStatementInput, userId int64) (models.StatementResponse, error) {
	if err := s.statementValidator.ValidateStatementUpload(input.AccountId, input.FileBytes, input.OriginalFilename); err != nil {
		return models.StatementResponse{}, err
	}

	fileType := "csv"
	if strings.HasSuffix(input.OriginalFilename, ".xls") || strings.HasSuffix(input.OriginalFilename, ".xlsx") {
		fileType = "excel"
	}

	account, err := s.accountService.GetAccountById(c, input.AccountId, userId)
	if err != nil {
		return models.StatementResponse{}, err
	}

	// Create a statement record in the database.
	createStatement := models.CreateStatementInput{
		AccountId:        account.Id,
		CreatedBy:        userId,
		OriginalFilename: input.OriginalFilename,
		FileType:         fileType,
		Status:           models.StatementStatusPending,
	}

	statement, err := s.repo.CreateStatement(c, createStatement)
	if err != nil {
		return models.StatementResponse{}, err
	}

	// Process the statement asynchronously.
	go s.processStatementAsync(c, statement.Id, input, userId)
	return statement, nil
}

// processStatementAsync processes the statement in a separate goroutine.
func (s *StatementService) processStatementAsync(c *gin.Context, statementId int64, input models.ParseStatementInput, userId int64) {
	logger.Debugf("Processing statement ID %d for account ID %d by user ID %d", statementId, input.AccountId, userId)
	_, _ = s.repo.UpdateStatementStatus(c, statementId, models.UpdateStatementStatusInput{
		Status: models.StatementStatusProcessing,
	})

	parserType := input.BankType
	if parserType == "" {
		logger.Debugf("No bank type provided, fetching account details for account ID %d", input.AccountId)
		account, err := s.accountService.GetAccountById(c, input.AccountId, userId)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to fetch account: %v", err)
			_, _ = s.repo.UpdateStatementStatus(c, statementId, models.UpdateStatementStatusInput{
				Status:  models.StatementStatusError,
				Message: &errMsg,
			})
			return
		}
		parserType = string(account.BankType)
	}

	logger.Debugf("Fetching Parser for bank: %v", parserType)
	parserImpl, ok := parser.GetParser(models.BankType(parserType))
	if !ok {
		errMsg := fmt.Sprintf("No parser available for bank type: %s", parserType)
		_, _ = s.repo.UpdateStatementStatus(c, statementId, models.UpdateStatementStatusInput{
			Status:  models.StatementStatusError,
			Message: &errMsg,
		})
		return
	}

	logger.Debugf("Using parser: %T for bank type: %s", parserImpl, parserType)
	parsedTxs, err := parserImpl.Parse(input.FileBytes, input.Metadata, input.OriginalFilename)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to parse statement: %v", err)
		_, _ = s.repo.UpdateStatementStatus(c, statementId, models.UpdateStatementStatusInput{
			Status:  models.StatementStatusError,
			Message: &errMsg,
		})
		return
	}

	logger.Debugf("Parsed %d transactions from statement ID %d", len(parsedTxs), statementId)
	var successCount, failCount int
	for _, tx := range parsedTxs {
		tx.AccountId = input.AccountId
		tx.CreatedBy = userId
		transaction, err := s.txService.CreateTransaction(c, tx)
		if err != nil {
			failCount++
			continue
		}
		err = s.repo.CreateStatementTxn(c, statementId, transaction.Id)
		if err != nil {
			logger.Errorf("Failed to link transaction %d to statement %d: %v", transaction.Id, statementId, err)
			failCount++
			continue
		}
		successCount++
	}

	msg := fmt.Sprintf("Processed %d transactions, %d failed", successCount, failCount)
	status := models.StatementStatusDone
	if failCount == len(parsedTxs) {
		status = models.StatementStatusError
	}
	_, _ = s.repo.UpdateStatementStatus(c, statementId, models.UpdateStatementStatusInput{
		Status:  status,
		Message: &msg,
	})
}

func (s *StatementService) GetStatementStatus(c *gin.Context, statementId int64, userId int64) (models.StatementResponse, error) {
	if statementId <= 0 {
		return models.StatementResponse{}, errors.New("invalid statement id")
	}
	return s.repo.GetStatementByID(c, statementId, userId)
}

func (s *StatementService) ListStatements(c *gin.Context, userId int64, page int, pageSize int) (models.PaginatedStatementResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	statements, err := s.repo.ListStatementByUserId(c, userId, pageSize, (page-1)*pageSize)
	if err != nil {
		return models.PaginatedStatementResponse{}, err
	}
	total, err := s.repo.CountStatementsByUserId(c, userId)
	if err != nil {
		return models.PaginatedStatementResponse{}, err
	}

	return models.PaginatedStatementResponse{
		Statements: statements,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (s *StatementService) PreviewStatement(c *gin.Context, fileBytes []byte, fileName string, skipRows int, rowSize int) (*models.StatementPreview, error) {
	if rowSize == 0 {
		rowSize = 10
	}

	if err := s.statementValidator.ValidateStatementPreview(fileBytes, fileName, skipRows, rowSize); err != nil {
		return nil, err
	}

	p := parser.CustomParser{}
	return p.Preview(fileBytes, fileName, skipRows, rowSize)
}
