package controller

import (
	"encoding/json"
	"expenses/internal/config"
	"expenses/internal/models"
	"expenses/internal/service"
	"expenses/internal/validator"
	"expenses/pkg/logger"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type StatementController struct {
	*BaseController
	statementService   service.StatementServiceInterface
	statementValidator *validator.StatementValidator
}

func NewStatementController(cfg *config.Config, statementService service.StatementServiceInterface) *StatementController {
	return &StatementController{
		BaseController:     NewBaseController(cfg),
		statementService:   statementService,
		statementValidator: validator.NewStatementValidator(),
	}
}

func (s *StatementController) CreateStatement(ctx *gin.Context) {
	userId := s.GetAuthenticatedUserId(ctx)
	logger.Infof("Creating statement for user %d", userId)

	fileBytes, fileName, accountId, metadata, err := s.extractFileFromContext(ctx)
	if err != nil {
		return
	}

	statement, err := s.statementService.ProcessStatement(ctx, fileBytes, fileName, accountId, userId, metadata)
	if err != nil {
		logger.Errorf("Error creating statement: %v", err)
		s.HandleError(ctx, err)
		return
	}

	logger.Infof("Statement created successfully with ID %d for user %d", statement.Id, userId)
	s.SendSuccess(ctx, http.StatusCreated, "Statement uploaded successfully and processing has begun", statement)
}

// GetStatements handles GET /statements
func (s *StatementController) GetStatements(ctx *gin.Context) {
	userID := s.GetAuthenticatedUserId(ctx)
	logger.Infof("Fetching statements for user %d", userID)

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "15"))

	resp, err := s.statementService.ListStatements(ctx, userID, page, pageSize)
	if err != nil {
		logger.Errorf("Error fetching statements: %v", err)
		s.HandleError(ctx, err)
		return
	}

	logger.Infof("Successfully fetched %d statements for user %d", len(resp.Statements), userID)
	s.SendSuccess(ctx, http.StatusOK, "Statements fetched successfully", resp)
}

// GetStatement handles GET /statements/:id
func (s *StatementController) GetStatementStatus(ctx *gin.Context) {
	userID := s.GetAuthenticatedUserId(ctx)
	statementId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		logger.Errorf("Failed to parse statement_id: %v", err)
		s.SendError(ctx, http.StatusBadRequest, "Invalid statement_id")
		return
	}

	logger.Infof("Fetching statement %d for user %d", statementId, userID)
	statement, err := s.statementService.GetStatementStatus(ctx, statementId, userID)
	if err != nil {
		logger.Errorf("Error fetching statement: %v", err)
		s.HandleError(ctx, err)
		return
	}
	logger.Infof("Successfully fetched statement %d for user %d", statementId, userID)
	s.SendSuccess(ctx, http.StatusOK, "Statement fetched successfully", statement)
}

// Helper to extract file bytes, file name, accountId, and metadata from context
func (s *StatementController) extractFileFromContext(ctx *gin.Context) ([]byte, string, int64, models.CreateStatementMetadata, error) {
	err := ctx.Request.ParseMultipartForm(256 << 10) // 256KB max
	if err != nil {
		logger.Errorf("Failed to parse multipart form: %v", err)
		s.SendError(ctx, http.StatusBadRequest, "Failed to parse form data")
		return nil, "", 0, models.CreateStatementMetadata{}, err
	}

	accountId, err := strconv.ParseInt(ctx.PostForm("account_id"), 10, 64)
	if err != nil {
		logger.Errorf("Failed to parse account_id: %v", err)
		s.SendError(ctx, http.StatusBadRequest, "Invalid account_id")
		return nil, "", 0, models.CreateStatementMetadata{}, err
	}

	// Extract optional metadata parameters
	metadata := models.NewCreateStatementMetadata()
	
	// Parse skip_rows parameter (optional, defaults to 0)
	if skipRowsStr := ctx.PostForm("skip_rows"); skipRowsStr != "" {
		if parsed, parseErr := strconv.Atoi(skipRowsStr); parseErr == nil && parsed >= 0 {
			metadata.SkipRows = parsed
		}
	}

	// Parse mappings JSON parameter (optional, defaults to empty)
	if mappingsJSON := ctx.PostForm("mappings"); mappingsJSON != "" {
		var mappings []models.ColumnMapping
		if parseErr := json.Unmarshal([]byte(mappingsJSON), &mappings); parseErr == nil {
			metadata.Mappings = mappings
		}
	}

	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		logger.Errorf("Failed to get file from form: %v", err)
		s.SendError(ctx, http.StatusBadRequest, "File not found in form data")
		return nil, "", 0, models.CreateStatementMetadata{}, err
	}
	defer file.Close()

	err = s.statementValidator.ValidateStatementUpload(accountId, file, header)
	if err != nil {
		logger.Errorf("Failed to validate statement upload: %v", err)
		s.SendError(ctx, http.StatusBadRequest, "Invalid statement upload")
		return nil, "", 0, models.CreateStatementMetadata{}, err
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		logger.Errorf("Failed to read file bytes: %v", err)
		s.SendError(ctx, http.StatusBadRequest, "Failed to read file")
		return nil, "", 0, models.CreateStatementMetadata{}, err
	}

	return fileBytes, header.Filename, accountId, metadata, nil
}

// PreviewCSV handles POST /statement/preview
func (s *StatementController) PreviewCSV(ctx *gin.Context) {
	userId := s.GetAuthenticatedUserId(ctx)
	logger.Infof("Previewing CSV for user %d", userId)

	fileBytes, fileName, accountId, metadata, err := s.extractFileFromContext(ctx)
	if err != nil {
		return
	}

	input := models.CSVPreviewInput{
		AccountId: accountId,
		FileBytes: fileBytes,
		Filename:  fileName,
		SkipRows:  metadata.SkipRows,
	}

	preview, err := s.statementService.PreviewCSV(ctx, input)
	if err != nil {
		logger.Errorf("Error previewing CSV: %v", err)
		s.HandleError(ctx, err)
		return
	}

	logger.Infof("Successfully generated CSV preview for user %d with %d skipped rows", userId, metadata.SkipRows)
	s.SendSuccess(ctx, http.StatusOK, "CSV preview generated successfully", preview)
}

// CustomImport handles POST /statement/custom
func (s *StatementController) CustomImport(ctx *gin.Context) {
	userId := s.GetAuthenticatedUserId(ctx)
	logger.Infof("Processing custom import for user %d", userId)

	// Parse multipart form
	err := ctx.Request.ParseMultipartForm(256 << 10) // 256KB max
	if err != nil {
		logger.Errorf("Failed to parse multipart form: %v", err)
		s.SendError(ctx, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// Extract account_id
	accountId, err := strconv.ParseInt(ctx.PostForm("account_id"), 10, 64)
	if err != nil {
		logger.Errorf("Failed to parse account_id: %v", err)
		s.SendError(ctx, http.StatusBadRequest, "Invalid account_id")
		return
	}

	// Extract skip_rows
	skipRows, _ := strconv.Atoi(ctx.PostForm("skip_rows"))

	// Extract mappings JSON
	mappingsJSON := ctx.PostForm("mappings")
	if mappingsJSON == "" {
		logger.Errorf("Missing mappings parameter")
		s.SendError(ctx, http.StatusBadRequest, "Missing mappings parameter")
		return
	}

	// Parse mappings
	var mappings []models.ColumnMapping
	if err := json.Unmarshal([]byte(mappingsJSON), &mappings); err != nil {
		logger.Errorf("Failed to parse mappings JSON: %v", err)
		s.SendError(ctx, http.StatusBadRequest, "Invalid mappings format")
		return
	}

	// Extract file
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		logger.Errorf("Failed to get file from form: %v", err)
		s.SendError(ctx, http.StatusBadRequest, "File not found in form data")
		return
	}
	defer file.Close()

	// Validate file using custom import validator
	customValidator := validator.NewCustomImportValidator()
	err = customValidator.ValidateCSVFile(file, header)
	if err != nil {
		logger.Errorf("Failed to validate CSV file: %v", err)
		s.HandleError(ctx, err)
		return
	}

	// Read file bytes
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		logger.Errorf("Failed to read file bytes: %v", err)
		s.SendError(ctx, http.StatusBadRequest, "Failed to read file")
		return
	}

	// Create custom import input
	input := models.CustomImportInput{
		AccountId: accountId,
		SkipRows:  skipRows,
		Mappings:  mappings,
	}

	// Process custom import
	result, err := s.statementService.ProcessCustomImport(ctx, input, fileBytes, header.Filename, userId)
	if err != nil {
		logger.Errorf("Error processing custom import: %v", err)
		s.HandleError(ctx, err)
		return
	}

	logger.Infof("Successfully started custom import for user %d", userId)
	s.SendSuccess(ctx, http.StatusCreated, "Custom import started successfully", result)
}