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

	logger.Infof("Transaction retrieved successfully with ID %d for user %d", transaction.Id, userId)
	t.SendSuccess(ctx, http.StatusOK, "Transaction retrieved successfully", transaction)
}

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

	logger.Infof("Transaction updated successfully with ID %d for user %d", transaction.Id, userId)
	t.SendSuccess(ctx, http.StatusOK, "Transaction updated successfully", transaction)
}

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

	logger.Infof("Transaction deleted successfully with ID %d for user %d", transactionId, userId)
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

func (t *TransactionController) bindTransactionListQuery(ctx *gin.Context) models.TransactionListQuery {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "15"))

	return models.TransactionListQuery{
		Page:       page,
		PageSize:   pageSize,
		SortBy:     ctx.DefaultQuery("sort_by", "date"),
		SortOrder:  ctx.DefaultQuery("sort_order", "desc"),
		AccountID:  parseInt64QueryParam(ctx, "account_id"),
		CategoryID: parseInt64QueryParam(ctx, "category_id"),
		MinAmount:  parseFloat64QueryParam(ctx, "min_amount"),
		MaxAmount:  parseFloat64QueryParam(ctx, "max_amount"),
		DateFrom:   parseTimeQueryParam(ctx, "date_from", "2006-01-02"),
		DateTo:     parseTimeQueryParam(ctx, "date_to", "2006-01-02"),
		Search:     parseStringQueryParam(ctx, "search"),
	}
}

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
