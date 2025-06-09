package controller

import (
	"expenses/internal/config"
	"expenses/internal/models"
	"expenses/internal/service"
	"expenses/pkg/logger"
	"net/http"
	"strconv"

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
	userId := t.GetAuthenticatedUserId(ctx)
	logger.Infof("Creating new transaction for user %d", userId)

	var input models.CreateTransactionInput
	if err := t.BindJSON(ctx, &input); err != nil {
		logger.Errorf("Failed to bind JSON: %v", err)
		return
	}
	input.CreatedBy = userId
	transaction, err := t.transactionService.CreateTransaction(ctx, input)
	if err != nil {
		logger.Errorf("Error creating transaction: %v", err)
		t.HandleError(ctx, err)
		return
	}

	logger.Infof("Transaction created successfully with ID %d for user %d", transaction.Id, userId)
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

func (t *TransactionController) ListTransactions(ctx *gin.Context) {
	userId := t.GetAuthenticatedUserId(ctx)
	logger.Infof("Fetching transactions for user %d", userId)

	transactions, err := t.transactionService.ListTransactions(ctx, userId)
	if err != nil {
		logger.Errorf("Error listing transactions: %v", err)
		t.HandleError(ctx, err)
		return
	}

	logger.Infof("Transactions retrieved successfully for user %d", userId)
	t.SendSuccess(ctx, http.StatusOK, "Transactions retrieved successfully", transactions)
}
