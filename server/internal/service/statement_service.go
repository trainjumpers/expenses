package service

import (
	"expenses/internal/models"
	"expenses/internal/parser"
	"expenses/internal/repository"
	"expenses/internal/validator"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
)

type StatementServiceInterface interface {
	ParseStatement(c *gin.Context, accountId int64, userId int64) (models.StatementResponse, error)
	GetStatementStatus(c *gin.Context, statementId int64, userId int64) (models.StatementResponse, error)
	ListStatements(c *gin.Context, userId int64) (models.PaginatedStatementResponse, error)
}

type StatementService struct {
	repo               repository.StatementRepositoryInterface
	accountRepo        repository.AccountRepositoryInterface
	txService          TransactionServiceInterface
	statementValidator *validator.StatementValidator
}

func NewStatementService(
	repo repository.StatementRepositoryInterface,
	accountRepo repository.AccountRepositoryInterface,
	statementValidator *validator.StatementValidator,
	txService TransactionServiceInterface,
) StatementServiceInterface {
	return &StatementService{
		repo:               repo,
		accountRepo:        accountRepo,
		txService:          txService,
		statementValidator: statementValidator,
	}
}

func (s *StatementService) ParseStatement(c *gin.Context, accountId int64, userId int64) (models.StatementResponse, error) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		return models.StatementResponse{}, err
	}
	defer file.Close()

	err = s.statementValidator.ValidateStatementUpload(accountId, file, header)
	if err != nil {
		return models.StatementResponse{}, err
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return models.StatementResponse{}, err
	}

	fileType := "csv"
	filename := strings.ToLower(header.Filename)
	if strings.HasSuffix(filename, ".xls") || strings.HasSuffix(filename, ".xlsx") {
		fileType = "excel"
	}

	input := models.CreateStatementInput{
		AccountID:        accountId,
		CreatedBy:        userId,
		OriginalFilename: header.Filename,
		FileType:         fileType,
		Status:           models.StatementStatusPending,
	}

	statement, err := s.repo.CreateStatement(c, input)
	if err != nil {
		return models.StatementResponse{}, err
	}
	go s.processStatementAsync(statement.ID, input.AccountID, input.CreatedBy, fileBytes)
	return statement, nil
}

func (s *StatementService) processStatementAsync(statementId int64, accountId int64, userId int64, fileBytes []byte) {
	_, _ = s.repo.UpdateStatementStatus(&gin.Context{}, statementId, models.UpdateStatementStatusInput{
		Status: models.StatementStatusProcessing,
	})

	account, err := s.accountRepo.GetAccountById(&gin.Context{}, accountId, userId)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to fetch account: %v", err)
		_, _ = s.repo.UpdateStatementStatus(&gin.Context{}, statementId, models.UpdateStatementStatusInput{
			Status:  models.StatementStatusError,
			Message: &errMsg,
		})
		return
	}

	parserImpl, ok := parser.GetParser(account.BankType)
	if !ok {
		errMsg := fmt.Sprintf("No parser available for bank type: %s", account.BankType)
		_, _ = s.repo.UpdateStatementStatus(&gin.Context{}, statementId, models.UpdateStatementStatusInput{
			Status:  models.StatementStatusError,
			Message: &errMsg,
		})
		return
	}

	parsedTxs, err := parserImpl.Parse(fileBytes)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to parse statement: %v", err)
		_, _ = s.repo.UpdateStatementStatus(&gin.Context{}, statementId, models.UpdateStatementStatusInput{
			Status:  models.StatementStatusError,
			Message: &errMsg,
		})
		return
	}

	var successCount, failCount int
	for _, tx := range parsedTxs {
		tx.AccountId = accountId
		tx.CreatedBy = userId
		_, err := s.txService.CreateTransaction(&gin.Context{}, tx)
		if err != nil {
			failCount++
		} else {
			successCount++
		}
	}

	msg := fmt.Sprintf("Processed %d transactions, %d failed", successCount, failCount)
	status := models.StatementStatusDone
	if failCount > 0 {
		status = models.StatementStatusError
	}
	_, _ = s.repo.UpdateStatementStatus(&gin.Context{}, statementId, models.UpdateStatementStatusInput{
		Status:  status,
		Message: &msg,
	})
}

func (s *StatementService) GetStatementStatus(c *gin.Context, statementId int64, userId int64) (models.StatementResponse, error) {
	return s.repo.GetStatementByID(c, statementId, userId)
}

func (s *StatementService) ListStatements(c *gin.Context, userId int64) (models.PaginatedStatementResponse, error) {
	page := 1
	pageSize := 10
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
		if page < 1 {
			page = 1
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
		if pageSize < 1 || pageSize > 100 {
			pageSize = 10
		}
	}
	limit := pageSize
	offset := (page - 1) * pageSize

	_, _, err := s.statementValidator.ValidatePaginationParams(
		c.Query("limit"),
		c.Query("offset"),
	)
	if err != nil {
		return models.PaginatedStatementResponse{}, err
	}

	statements, err := s.repo.ListStatementByUserId(c, userId, limit, offset)
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
