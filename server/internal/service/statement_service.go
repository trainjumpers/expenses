package service

import (
	"errors"
	"expenses/internal/models"
	"expenses/internal/parser"
	"expenses/internal/repository"
	"expenses/internal/validator"
	"expenses/pkg/logger"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/gin-gonic/gin"
	appErrors "expenses/internal/errors"
)

type StatementServiceInterface interface {
	ParseStatement(c *gin.Context, fileBytes []byte, fileName string, accountId int64, userId int64) (models.StatementResponse, error)
	GetStatementStatus(c *gin.Context, statementId int64, userId int64) (models.StatementResponse, error)
	ListStatements(c *gin.Context, userId int64, page int, pageSize int) (models.PaginatedStatementResponse, error)
	PreviewCSV(c *gin.Context, input models.CSVPreviewInput) (models.CSVPreviewResult, error)
	ProcessCustomImport(c *gin.Context, input models.CustomImportInput, fileBytes []byte, filename string, userId int64) (models.CustomImportResult, error)
	ValidateColumnMappings(mappings []models.ColumnMapping) error
	// Unified method for all statement imports
	ProcessStatement(c *gin.Context, fileBytes []byte, fileName string, accountId int64, userId int64, metadata models.CreateStatementMetadata) (models.StatementResponse, error)
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

func (s *StatementService) ParseStatement(c *gin.Context, fileBytes []byte, fileName string, accountId int64, userId int64) (models.StatementResponse, error) {

	fileType := "csv"
	fileName = strings.ToLower(fileName)
	if strings.HasSuffix(fileName, ".xls") || strings.HasSuffix(fileName, ".xlsx") {
		fileType = "excel"
	}

	account, err := s.accountService.GetAccountById(c, accountId, userId)
	if err != nil {
		return models.StatementResponse{}, err
	}

	input := models.CreateStatementInput{
		AccountId:        account.Id,
		CreatedBy:        userId,
		OriginalFilename: fileName,
		FileType:         fileType,
		Status:           models.StatementStatusPending,
	}

	statement, err := s.repo.CreateStatement(c, input)
	if err != nil {
		return models.StatementResponse{}, err
	}
	go s.processStatementAsync(c, statement.Id, input.AccountId, input.CreatedBy, fileBytes)
	return statement, nil
}

func (s *StatementService) processStatementAsync(c *gin.Context, statementId int64, accountId int64, userId int64, fileBytes []byte) {
	logger.Debugf("Processing statement ID %d for account ID %d by user ID %d", statementId, accountId, userId)
	_, _ = s.repo.UpdateStatementStatus(c, statementId, models.UpdateStatementStatusInput{
		Status: models.StatementStatusProcessing,
	})

	account, err := s.accountService.GetAccountById(c, accountId, userId)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to fetch account: %v", err)
		_, _ = s.repo.UpdateStatementStatus(c, statementId, models.UpdateStatementStatusInput{
			Status:  models.StatementStatusError,
			Message: &errMsg,
		})
		return
	}

	logger.Debugf("Fetching Parser for bank: %v", account.BankType)
	parserImpl, ok := parser.GetParser(account.BankType)
	if !ok {
		errMsg := fmt.Sprintf("No parser available for bank type: %s", account.BankType)
		_, _ = s.repo.UpdateStatementStatus(c, statementId, models.UpdateStatementStatusInput{
			Status:  models.StatementStatusError,
			Message: &errMsg,
		})
		return
	}

	logger.Debugf("Using parser: %T for bank type: %s", parserImpl, account.BankType)
	// Use empty metadata for existing bank statement parsing
	metadata := models.NewCreateStatementMetadata()
	parsedTxs, err := parserImpl.Parse(fileBytes, metadata)
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
		tx.AccountId = accountId
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
// PreviewCSV parses a CSV/XLS file and returns the first 10 rows for preview
func (s *StatementService) PreviewCSV(c *gin.Context, input models.CSVPreviewInput) (models.CSVPreviewResult, error) {
	// Validate file size (256KB limit)
	if len(input.FileBytes) > 256*1024 {
		return models.CSVPreviewResult{}, appErrors.NewCSVFileTooLargeError()
	}

	// Parse the file
	parseResult, err := parser.ParseCSVFile(input.FileBytes, input.Filename)
	if err != nil {
		return models.CSVPreviewResult{}, err
	}

	// Apply row skipping if specified
	if input.SkipRows > 0 {
		parseResult = parseResult.ApplySkipRows(input.SkipRows)
	}

	// Return preview with first 10 rows after skipping
	previewRows := parseResult.GetPreviewRows(10)
	
	return models.CSVPreviewResult{
		Columns: parseResult.Columns,
		Rows:    previewRows,
		Total:   parseResult.Total,
	}, nil
}

// ProcessCustomImport processes a custom CSV import with column mappings
func (s *StatementService) ProcessCustomImport(c *gin.Context, input models.CustomImportInput, fileBytes []byte, filename string, userId int64) (models.CustomImportResult, error) {
	logger.Infof("Starting custom import for user %d, file: %s, skip_rows: %d", userId, filename, input.SkipRows)
	logger.Debugf("Column mappings: %+v", input.Mappings)
	
	// Validate file size
	if len(fileBytes) > 256*1024 {
		logger.Errorf("File size too large: %d bytes", len(fileBytes))
		return models.CustomImportResult{}, appErrors.NewCSVFileTooLargeError()
	}
	logger.Debugf("File size validation passed: %d bytes", len(fileBytes))

	// Validate column mappings
	if err := s.ValidateColumnMappings(input.Mappings); err != nil {
		logger.Errorf("Column mapping validation failed: %v", err)
		return models.CustomImportResult{}, err
	}
	logger.Debugf("Column mapping validation passed")

	// Verify account belongs to user
	account, err := s.accountService.GetAccountById(c, input.AccountId, userId)
	if err != nil {
		logger.Errorf("Failed to get account %d for user %d: %v", input.AccountId, userId, err)
		return models.CustomImportResult{}, err
	}
	logger.Debugf("Account validation passed: %s (ID: %d)", account.Name, account.Id)

	// Parse the file
	parseResult, err := parser.ParseCSVFile(fileBytes, filename)
	if err != nil {
		logger.Errorf("Failed to parse CSV file: %v", err)
		return models.CustomImportResult{}, err
	}
	logger.Infof("CSV parsed successfully: %d columns, %d total rows", len(parseResult.Columns), parseResult.Total)
	logger.Debugf("CSV columns: %v", parseResult.Columns)

	// Apply row skipping
	if input.SkipRows > 0 {
		logger.Debugf("Applying row skipping: %d rows", input.SkipRows)
		parseResult = parseResult.ApplySkipRows(input.SkipRows)
		logger.Infof("After skipping %d rows: %d columns, %d remaining rows", input.SkipRows, len(parseResult.Columns), len(parseResult.Rows))
		logger.Debugf("New columns after skipping: %v", parseResult.Columns)
	}

	// Create statement record
	fileType := "csv"
	if strings.HasSuffix(strings.ToLower(filename), ".xls") || strings.HasSuffix(strings.ToLower(filename), ".xlsx") {
		fileType = "excel"
	}

	statementInput := models.CreateStatementInput{
		AccountId:        account.Id,
		CreatedBy:        userId,
		OriginalFilename: filename,
		FileType:         fileType,
		Status:           models.StatementStatusProcessing,
	}

	statement, err := s.repo.CreateStatement(c, statementInput)
	if err != nil {
		logger.Errorf("Failed to create statement record: %v", err)
		return models.CustomImportResult{}, err
	}
	logger.Infof("Statement record created with ID: %d", statement.Id)

	// Process transactions using custom parser
	logger.Debugf("Creating custom CSV parser with mappings: %+v", input.Mappings)
	customParser := parser.NewCustomCSVParser()
	metadata := models.CreateStatementMetadata{
		SkipRows: input.SkipRows,
		Mappings: input.Mappings,
	}
	transactions, err := customParser.Parse(fileBytes, metadata)
	if err != nil {
		logger.Errorf("Custom parser failed: %v", err)
		// Update statement status to error
		errMsg := fmt.Sprintf("Failed to parse custom CSV: %v", err)
		_, _ = s.repo.UpdateStatementStatus(c, statement.Id, models.UpdateStatementStatusInput{
			Status:  models.StatementStatusError,
			Message: &errMsg,
		})
		return models.CustomImportResult{}, err
	}
	logger.Infof("Custom parser processed %d transactions", len(transactions))

	// Log first few transactions for debugging
	for i, tx := range transactions {
		if i < 3 { // Log first 3 transactions
			logger.Debugf("Transaction %d: Name=%s, Amount=%v, Date=%v, Description=%s", 
				i+1, tx.Name, tx.Amount, tx.Date, tx.Description)
		}
	}

	// Create transactions
	var successCount, failCount int
	for i, tx := range transactions {
		tx.AccountId = account.Id
		tx.CreatedBy = userId
		
		logger.Debugf("Creating transaction %d: %+v", i+1, tx)
		transaction, err := s.txService.CreateTransaction(c, tx)
		if err != nil {
			logger.Warnf("Failed to create transaction %d: %v", i+1, err)
			failCount++
			continue
		}
		logger.Debugf("Transaction created with ID: %d", transaction.Id)
		
		err = s.repo.CreateStatementTxn(c, statement.Id, transaction.Id)
		if err != nil {
			logger.Errorf("Failed to link transaction %d to statement %d: %v", transaction.Id, statement.Id, err)
			failCount++
			continue
		}
		successCount++
	}

	logger.Infof("Transaction creation completed: %d successful, %d failed out of %d total", 
		successCount, failCount, len(transactions))

	// Update statement status
	msg := fmt.Sprintf("Successfully imported %d transactions", successCount)
	if failCount > 0 {
		msg = fmt.Sprintf("Imported %d transactions (%d failed)", successCount, failCount)
	}
	
	_, _ = s.repo.UpdateStatementStatus(c, statement.Id, models.UpdateStatementStatusInput{
		Status:  models.StatementStatusDone,
		Message: &msg,
	})

	logger.Infof("Custom import completed for user %d: %s", userId, msg)
	return models.CustomImportResult{
		Statement:          statement,
		TransactionsCreated: successCount,
		Message:            msg,
	}, nil
}

// ValidateColumnMappings validates that required fields are mapped correctly
// This method supports the unified validation approach through ParseOptions
func (s *StatementService) ValidateColumnMappings(mappings []models.ColumnMapping) error {
	validator := validator.NewCustomImportValidator()
	return validator.ValidateColumnMappings(mappings)
}

// ValidateStatementWithOptions provides unified validation for statement uploads with ParseOptions
func (s *StatementService) ValidateStatementWithOptions(accountId int64, file multipart.File, header *multipart.FileHeader, options models.ParseOptions) error {
	return s.statementValidator.ValidateStatementWithOptions(accountId, file, header, options)
}

// ProcessStatement is the unified method for all statement imports
func (s *StatementService) ProcessStatement(c *gin.Context, fileBytes []byte, fileName string, accountId int64, userId int64, metadata models.CreateStatementMetadata) (models.StatementResponse, error) {
	logger.Infof("Starting unified statement processing for user %d, file: %s", userId, fileName)
	logger.Debugf("Metadata: SkipRows=%d, Mappings=%+v", metadata.SkipRows, metadata.Mappings)

	// Validate file size (256KB limit)
	if len(fileBytes) > 256*1024 {
		logger.Errorf("File size too large: %d bytes", len(fileBytes))
		return models.StatementResponse{}, appErrors.NewCSVFileTooLargeError()
	}
	logger.Debugf("File size validation passed: %d bytes", len(fileBytes))

	// Validate column mappings if provided
	if metadata.HasCustomMappings() {
		if err := s.ValidateColumnMappings(metadata.Mappings); err != nil {
			logger.Errorf("Column mapping validation failed: %v", err)
			return models.StatementResponse{}, err
		}
		logger.Debugf("Column mapping validation passed")
	}

	// Verify account belongs to user
	account, err := s.accountService.GetAccountById(c, accountId, userId)
	if err != nil {
		logger.Errorf("Failed to get account %d for user %d: %v", accountId, userId, err)
		return models.StatementResponse{}, err
	}
	logger.Debugf("Account validation passed: %s (ID: %d)", account.Name, account.Id)

	// Determine file type
	fileType := "csv"
	if strings.HasSuffix(strings.ToLower(fileName), ".xls") || strings.HasSuffix(strings.ToLower(fileName), ".xlsx") {
		fileType = "excel"
	}

	// Create statement record
	statementInput := models.CreateStatementInput{
		AccountId:        account.Id,
		CreatedBy:        userId,
		OriginalFilename: fileName,
		FileType:         fileType,
		Status:           models.StatementStatusPending,
	}

	statement, err := s.repo.CreateStatement(c, statementInput)
	if err != nil {
		logger.Errorf("Failed to create statement record: %v", err)
		return models.StatementResponse{}, err
	}
	logger.Infof("Statement record created with ID: %d", statement.Id)

	// Process statement asynchronously with metadata
	go s.processStatementAsyncWithMetadata(c, statement.Id, account.Id, userId, fileBytes, metadata, account.BankType)
	
	logger.Infof("Unified statement processing initiated for user %d", userId)
	return statement, nil
}

// processStatementAsyncWithMetadata processes statements with metadata support
func (s *StatementService) processStatementAsyncWithMetadata(c *gin.Context, statementId int64, accountId int64, userId int64, fileBytes []byte, metadata models.CreateStatementMetadata, bankType models.BankType) {
	logger.Debugf("Processing statement ID %d for account ID %d by user ID %d with metadata", statementId, accountId, userId)
	_, _ = s.repo.UpdateStatementStatus(c, statementId, models.UpdateStatementStatusInput{
		Status: models.StatementStatusProcessing,
	})

	// Get the appropriate parser
	var parserImpl parser.BankStatementParser
	var ok bool

	if metadata.HasCustomMappings() {
		// Use custom CSV parser for custom mappings
		logger.Debugf("Using custom CSV parser due to custom mappings")
		parserImpl = parser.NewCustomCSVParser()
	} else {
		// Use bank-specific parser
		logger.Debugf("Fetching parser for bank: %v", bankType)
		parserImpl, ok = parser.GetParser(bankType)
		if !ok {
			errMsg := fmt.Sprintf("No parser available for bank type: %s", bankType)
			logger.Errorf(errMsg)
			_, _ = s.repo.UpdateStatementStatus(c, statementId, models.UpdateStatementStatusInput{
				Status:  models.StatementStatusError,
				Message: &errMsg,
			})
			return
		}
	}

	logger.Debugf("Using parser: %T for bank type: %s", parserImpl, bankType)
	parsedTxs, err := parserImpl.Parse(fileBytes, metadata)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to parse statement: %v", err)
		logger.Errorf(errMsg)
		_, _ = s.repo.UpdateStatementStatus(c, statementId, models.UpdateStatementStatusInput{
			Status:  models.StatementStatusError,
			Message: &errMsg,
		})
		return
	}

	logger.Infof("Parsed %d transactions from statement ID %d", len(parsedTxs), statementId)

	// Create transactions
	var successCount, failCount int
	for i, tx := range parsedTxs {
		tx.AccountId = accountId
		tx.CreatedBy = userId
		
		logger.Debugf("Creating transaction %d: %+v", i+1, tx)
		transaction, err := s.txService.CreateTransaction(c, tx)
		if err != nil {
			logger.Warnf("Failed to create transaction %d: %v", i+1, err)
			failCount++
			continue
		}
		logger.Debugf("Transaction created with ID: %d", transaction.Id)
		
		err = s.repo.CreateStatementTxn(c, statementId, transaction.Id)
		if err != nil {
			logger.Errorf("Failed to link transaction %d to statement %d: %v", transaction.Id, statementId, err)
			failCount++
			continue
		}
		successCount++
	}

	logger.Infof("Transaction creation completed: %d successful, %d failed out of %d total", 
		successCount, failCount, len(parsedTxs))

	// Update statement status
	msg := fmt.Sprintf("Processed %d transactions, %d failed", successCount, failCount)
	status := models.StatementStatusDone
	if failCount == len(parsedTxs) {
		status = models.StatementStatusError
	}
	
	_, _ = s.repo.UpdateStatementStatus(c, statementId, models.UpdateStatementStatusInput{
		Status:  status,
		Message: &msg,
	})

	logger.Infof("Unified statement processing completed for statement ID %d: %s", statementId, msg)
}