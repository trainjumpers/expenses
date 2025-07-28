package controller

import (
	"expenses/internal/config"
	"expenses/internal/models"
	"expenses/internal/service"
	"expenses/pkg/logger"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type TransactionController struct {
	*BaseController
	transactionService service.TransactionServiceInterface
}

func NewTransactionController(cfg *config.Config, transactionService service.TransactionServiceInterface) *TransactionController {
	return &TransactionController{
		BaseController:     NewBaseController(cfg),
		transactionService: transactionService,
	}
}

// CreateTransaction creates a new transaction
// @Summary Create a new transaction
// @Description Create a new transaction for the authenticated user
// @Tags transactions
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param transaction body models.CreateTransactionInput true "Transaction data"
// @Success 201 {object} models.TransactionResponse "Transaction created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /transaction [post]
func (t *TransactionController) CreateTransaction(ctx *gin.Context) {
	var input models.CreateTransactionInput
	if err := t.BindJSON(ctx, &input); err != nil {
		logger.Errorf("Failed to bind JSON: %v", err)
		return
	}
	logger.Infof("Creating new transaction for user %d", input.CreatedBy)
	transaction, err := t.transactionService.CreateTransaction(ctx, input)
	if err != nil {
		logger.Errorf("Error creating transaction: %v", err)
		t.HandleError(ctx, err)
		return
	}

	logger.Infof("Transaction created successfully with ID %d for user %d", transaction.Id, transaction.CreatedBy)
	t.SendSuccess(ctx, http.StatusCreated, "Transaction created successfully", transaction)
}

// GetTransaction retrieves a specific transaction
// @Summary Get transaction by ID
// @Description Get transaction details by transaction ID for the authenticated user
// @Tags transactions
// @Produce json
// @Security BasicAuth
// @Param transactionId path int true "Transaction ID"
// @Success 200 {object} models.TransactionResponse "Transaction details"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Transaction not found"
// @Router /transaction/{transactionId} [get]
func (t *TransactionController) GetTransaction(ctx *gin.Context) {
	userId := t.GetAuthenticatedUserId(ctx)
	logger.Infof("Fetching transaction details for user %d", userId)

	transactionId, err := strconv.ParseInt(ctx.Param("transactionId"), 10, 64)
	if err != nil {
		t.SendError(ctx, http.StatusBadRequest, "invalid transaction id")
		return
	}

	transaction, err := t.transactionService.GetTransactionById(ctx, transactionId, userId)
	if err != nil {
		logger.Errorf("Error getting transaction: %v", err)
		t.HandleError(ctx, err)
		return
	}

	logger.Infof("Transaction retrieved successfully with Id %d for user %d", transaction.Id, userId)
	t.SendSuccess(ctx, http.StatusOK, "Transaction retrieved successfully", transaction)
}

// UpdateTransaction updates an existing transaction
// @Summary Update transaction
// @Description Update transaction details by transaction ID for the authenticated user
// @Tags transactions
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param transactionId path int true "Transaction ID"
// @Param transaction body models.UpdateTransactionInput true "Updated transaction data"
// @Success 200 {object} models.TransactionResponse "Transaction updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Transaction not found"
// @Router /transaction/{transactionId} [patch]
func (t *TransactionController) UpdateTransaction(ctx *gin.Context) {
	userId := t.GetAuthenticatedUserId(ctx)
	logger.Infof("Starting transaction update for user %d", userId)

	transactionId, err := strconv.ParseInt(ctx.Param("transactionId"), 10, 64)
	if err != nil {
		t.SendError(ctx, http.StatusBadRequest, "invalid transaction id")
		return
	}

	var input models.UpdateTransactionInput
	if err := t.BindJSON(ctx, &input); err != nil {
		logger.Errorf("Failed to bind JSON: %v", err)
		return
	}

	transaction, err := t.transactionService.UpdateTransaction(ctx, transactionId, userId, input)
	if err != nil {
		logger.Errorf("Error updating transaction: %v", err)
		t.HandleError(ctx, err)
		return
	}

	logger.Infof("Transaction updated successfully with Id %d for user %d", transaction.Id, userId)
	t.SendSuccess(ctx, http.StatusOK, "Transaction updated successfully", transaction)
}

// DeleteTransaction deletes a transaction
// @Summary Delete transaction
// @Description Delete transaction by transaction ID for the authenticated user
// @Tags transactions
// @Produce json
// @Security BasicAuth
// @Param transactionId path int true "Transaction ID"
// @Success 204 "Transaction deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Transaction not found"
// @Router /transaction/{transactionId} [delete]
func (t *TransactionController) DeleteTransaction(ctx *gin.Context) {
	userId := t.GetAuthenticatedUserId(ctx)
	logger.Infof("Starting transaction deletion for user %d", userId)

	transactionId, err := strconv.ParseInt(ctx.Param("transactionId"), 10, 64)
	if err != nil {
		t.SendError(ctx, http.StatusBadRequest, "invalid transaction id")
		return
	}

	err = t.transactionService.DeleteTransaction(ctx, transactionId, userId)
	if err != nil {
		logger.Errorf("Error deleting transaction: %v", err)
		t.HandleError(ctx, err)
		return
	}

	logger.Infof("Transaction deleted successfully with Id %d for user %d", transactionId, userId)
	t.SendSuccess(ctx, http.StatusNoContent, "", nil)
}

func parseInt64QueryParam(ctx *gin.Context, key string) *int64 {
	valStr := ctx.Query(key)
	if valStr == "" {
		return nil
	}
	if v, err := strconv.ParseInt(valStr, 10, 64); err == nil {
		return &v
	}
	return nil
}

func parseFloat64QueryParam(ctx *gin.Context, key string) *float64 {
	valStr := ctx.Query(key)
	if valStr == "" {
		return nil
	}
	if v, err := strconv.ParseFloat(valStr, 64); err == nil {
		return &v
	}
	return nil
}

func parseTimeQueryParam(ctx *gin.Context, key, layout string) *time.Time {
	valStr := ctx.Query(key)
	if valStr == "" {
		return nil
	}
	if v, err := time.Parse(layout, valStr); err == nil {
		return &v
	}
	return nil
}

func parseStringQueryParam(ctx *gin.Context, key string) *string {
	valStr := ctx.Query(key)
	if valStr == "" {
		return nil
	}
	return &valStr
}

func parseBoolQueryParam(ctx *gin.Context, key string) *bool {
	valStr := ctx.Query(key)
	if valStr == "" {
		return nil
	}
	if valStr == "true" || valStr == "1" {
		val := true
		return &val
	}
	if valStr == "false" || valStr == "0" {
		val := false
		return &val
	}
	return nil
}

func (t *TransactionController) bindTransactionListQuery(ctx *gin.Context) models.TransactionListQuery {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "15"))

	return models.TransactionListQuery{
		Page:          page,
		PageSize:      pageSize,
		SortBy:        ctx.DefaultQuery("sort_by", "date"),
		SortOrder:     ctx.DefaultQuery("sort_order", "desc"),
		AccountId:     parseInt64QueryParam(ctx, "account_id"),
		CategoryId:    parseInt64QueryParam(ctx, "category_id"),
		Uncategorized: parseBoolQueryParam(ctx, "uncategorized"),
		MinAmount:     parseFloat64QueryParam(ctx, "min_amount"),
		MaxAmount:     parseFloat64QueryParam(ctx, "max_amount"),
		DateFrom:      parseTimeQueryParam(ctx, "date_from", "2006-01-02"),
		DateTo:        parseTimeQueryParam(ctx, "date_to", "2006-01-02"),
		StatementId:   parseInt64QueryParam(ctx, "statement_id"),
		Search:        parseStringQueryParam(ctx, "search"),
	}
}

// ListTransactions retrieves all transactions for the user with filtering
// @Summary List transactions
// @Description Get all transactions for the authenticated user with optional filtering and pagination
// @Tags transactions
// @Produce json
// @Security BasicAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(15)
// @Param sort_by query string false "Sort by field" default(date)
// @Param sort_order query string false "Sort order (asc/desc)" default(desc)
// @Param account_id query int false "Filter by account ID"
// @Param category_id query int false "Filter by category ID"
// @Param uncategorized query bool false "Filter uncategorized transactions"
// @Param min_amount query number false "Minimum amount filter"
// @Param max_amount query number false "Maximum amount filter"
// @Param date_from query string false "Date from filter (YYYY-MM-DD)"
// @Param date_to query string false "Date to filter (YYYY-MM-DD)"
// @Param statement_id query int false "Filter by statement ID"
// @Param search query string false "Search in transaction descriptions"
// @Success 200 {object} models.PaginatedTransactionsResponse "List of transactions"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /transaction [get]
func (t *TransactionController) ListTransactions(ctx *gin.Context) {
	userId := t.GetAuthenticatedUserId(ctx)
	logger.Infof("Fetching transactions for user %d", userId)

	query := t.bindTransactionListQuery(ctx)

	transactions, err := t.transactionService.ListTransactions(ctx, userId, query)
	if err != nil {
		logger.Errorf("Error listing transactions: %v", err)
		t.HandleError(ctx, err)
		return
	}

	logger.Infof("Transactions retrieved successfully for user %d", userId)
	t.SendSuccess(ctx, http.StatusOK, "Transactions retrieved successfully", transactions)
}
