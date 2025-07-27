package controller

import (
	"expenses/internal/config"
	"expenses/internal/models"
	"expenses/internal/service"
	"expenses/internal/validator"
	"expenses/pkg/logger"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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

func (s *StatementController) readFileFromRequest(fileHeader *multipart.FileHeader) ([]byte, string, error) {
	if fileHeader == nil {
		return nil, "", fmt.Errorf("file header is nil")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}
	fileName := strings.ToLower(fileHeader.Filename)
	return fileBytes, fileName, nil
}

func (s *StatementController) CreateStatement(ctx *gin.Context) {
	userId := s.GetAuthenticatedUserId(ctx)
	logger.Infof("Creating statement for user %d", userId)

	var form models.ParseStatementForm
	if err := ctx.ShouldBindWith(&form, binding.FormMultipart); err != nil {
		s.SendError(ctx, http.StatusBadRequest, fmt.Sprintf("Failed to parse form data: %v", err))
		return
	}

	fileBytes, fileName, err := s.readFileFromRequest(form.File)
	if err != nil {
		s.SendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	input := models.ParseStatementInput{
		AccountId:        form.AccountId,
		BankType:         form.BankType,
		Metadata:         form.Metadata,
		OriginalFilename: fileName,
		FileBytes:        fileBytes,
	}

	statement, err := s.statementService.ParseStatement(ctx, input, userId)
	if err != nil {
		logger.Errorf("Error creating statement: %v", err)
		s.HandleError(ctx, err)
		return
	}

	logger.Infof("Statement created successfully with ID %d for user %d", statement.Id, userId)
	s.SendSuccess(ctx, http.StatusCreated, "Statement uploaded successfully and processing has begun", statement)
}

func (s *StatementController) PreviewStatement(ctx *gin.Context) {
	logger.Info("Loading Preview for statement")

	var form models.PreviewStatementForm
	if err := ctx.ShouldBindWith(&form, binding.FormMultipart); err != nil {
		s.SendError(ctx, http.StatusBadRequest, fmt.Sprintf("Failed to parse form data: %v", err))
		return
	}

	fileBytes, fileName, err := s.readFileFromRequest(form.File)
	if err != nil {
		s.SendError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	preview, err := s.statementService.PreviewStatement(ctx, fileBytes, fileName, form.SkipRows, form.RowSize)
	if err != nil {
		logger.Errorf("Error previewing statement: %v", err)
		s.HandleError(ctx, err)
		return
	}

	logger.Info("Statement preview generated successfully")
	s.SendSuccess(ctx, http.StatusOK, "Statement preview generated successfully", preview)
}

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
