package controller

import (
	"expenses/internal/config"
	"expenses/internal/service"
	"expenses/internal/validator"
	"expenses/pkg/logger"
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

	err := ctx.Request.ParseMultipartForm(256 << 10) // 256KB max
	if err != nil {
		logger.Errorf("Failed to parse multipart form: %v", err)
		s.SendError(ctx, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	accountId, err := strconv.ParseInt(ctx.PostForm("account_id"), 10, 64)
	if err != nil {
		logger.Errorf("Failed to parse account_id: %v", err)
		s.SendError(ctx, http.StatusBadRequest, "Invalid account_id")
		return
	}

	statement, err := s.statementService.ParseStatement(ctx, accountId, userId)
	if err != nil {
		logger.Errorf("Error creating statement: %v", err)
		s.HandleError(ctx, err)
		return
	}

	logger.Infof("Statement created successfully with ID %d for user %d", statement.ID, userId)
	s.SendSuccess(ctx, http.StatusCreated, "Statement uploaded successfully and processing has begun", statement)
}

// GetStatements handles GET /statements
func (s *StatementController) GetStatements(ctx *gin.Context) {
	userID := s.GetAuthenticatedUserId(ctx)
	logger.Infof("Fetching statements for user %d", userID)

	resp, err := s.statementService.ListStatements(ctx, userID)
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
