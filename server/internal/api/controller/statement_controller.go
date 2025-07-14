package controller

import (
	"expenses/internal/config"
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

	fileBytes, fileName, accountId, err := s.extractFileFromContext(ctx)
	if err != nil {
		return
	}

	statement, err := s.statementService.ParseStatement(ctx, fileBytes, fileName, accountId, userId)
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

// Helper to extract file bytes, file name, and accountId from context
func (s *StatementController) extractFileFromContext(ctx *gin.Context) ([]byte, string, int64, error) {
	err := ctx.Request.ParseMultipartForm(256 << 10) // 256KB max
	if err != nil {
		logger.Errorf("Failed to parse multipart form: %v", err)
		s.SendError(ctx, http.StatusBadRequest, "Failed to parse form data")
		return nil, "", 0, err
	}

	accountId, err := strconv.ParseInt(ctx.PostForm("account_id"), 10, 64)
	if err != nil {
		logger.Errorf("Failed to parse account_id: %v", err)
		s.SendError(ctx, http.StatusBadRequest, "Invalid account_id")
		return nil, "", 0, err
	}

	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		logger.Errorf("Failed to get file from form: %v", err)
		s.SendError(ctx, http.StatusBadRequest, "File not found in form data")
		return nil, "", 0, err
	}
	defer file.Close()

	err = s.statementValidator.ValidateStatementUpload(accountId, file, header)
	if err != nil {
		logger.Errorf("Failed to validate statement upload: %v", err)
		s.SendError(ctx, http.StatusBadRequest, "Invalid statement upload")
		return nil, "", 0, err
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		logger.Errorf("Failed to read file bytes: %v", err)
		s.SendError(ctx, http.StatusBadRequest, "Failed to read file")
		return nil, "", 0, err
	}

	return fileBytes, header.Filename, accountId, nil
}
