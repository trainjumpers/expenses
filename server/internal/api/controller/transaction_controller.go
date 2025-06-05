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
	logger.Info("Start CreateTransaction for user: ", ctx.GetInt64("authUserId"))
	
	var input models.CreateTransactionInput
	if err := t.BindJSON(ctx, &input); err != nil {
		logger.Error("[TransactionController] Failed to bind JSON: ", err)
		return
	}
	
	input.CreatedBy = ctx.GetInt64("authUserId")
	
	transaction, err := t.transactionService.CreateTransaction(ctx, input)
	if err != nil {
		logger.Error("[TransactionController] Error creating transaction: ", err)
		t.HandleError(ctx, err)
		return
	}
	
	logger.Infof("Transaction created successfully with ID: %d for user: %d", transaction.Id, input.CreatedBy)
	t.SendSuccess(ctx, http.StatusCreated, "Transaction created successfully", transaction)
}

func (t *TransactionController) GetTransaction(ctx *gin.Context) {
	transactionId, err := strconv.ParseInt(ctx.Param("transactionId"), 10, 64)
	if err != nil {
		t.SendError(ctx, http.StatusBadRequest, "invalid transaction id")
		return
	}
	
	logger.Info("Fetching transaction details for user: ", ctx.GetInt64("authUserId"), " and transaction ID: ", transactionId)
	userId := ctx.GetInt64("authUserId")
	
	transaction, err := t.transactionService.GetTransactionById(ctx, transactionId, userId)
	if err != nil {
		logger.Error("[TransactionController] Error getting transaction: ", err)
		t.HandleError(ctx, err)
		return
	}
	
	logger.Infof("Transaction retrieved successfully with ID: %d for user: %d", transaction.Id, userId)
	t.SendSuccess(ctx, http.StatusOK, "Transaction retrieved successfully", transaction)
}

func (t *TransactionController) UpdateTransaction(ctx *gin.Context) {
	logger.Infof("Start UpdateTransaction for user: ", ctx.GetInt64("authUserId"))
	
	transactionId, err := strconv.ParseInt(ctx.Param("transactionId"), 10, 64)
	if err != nil {
		t.SendError(ctx, http.StatusBadRequest, "invalid transaction id")
		return
	}
	
	var input models.UpdateTransactionInput
	if err := t.BindJSON(ctx, &input); err != nil {
		logger.Error("[TransactionController] Failed to bind JSON: ", err)
		return
	}
	
	userId := ctx.GetInt64("authUserId")
	transaction, err := t.transactionService.UpdateTransaction(ctx, transactionId, userId, input)
	if err != nil {
		logger.Error("[TransactionController] Error updating transaction: ", err)
		t.HandleError(ctx, err)
		return
	}
	
	logger.Infof("Transaction updated successfully with ID: %d for user: %d", transaction.Id, userId)
	t.SendSuccess(ctx, http.StatusOK, "Transaction updated successfully", transaction)
}

func (t *TransactionController) DeleteTransaction(ctx *gin.Context) {
	logger.Infof("Start DeleteTransaction for user: ", ctx.GetInt64("authUserId"))
	
	transactionId, err := strconv.ParseInt(ctx.Param("transactionId"), 10, 64)
	if err != nil {
		t.SendError(ctx, http.StatusBadRequest, "invalid transaction id")
		return
	}
	
	userId := ctx.GetInt64("authUserId")
	err = t.transactionService.DeleteTransaction(ctx, transactionId, userId)
	if err != nil {
		logger.Error("[TransactionController] Error deleting transaction: ", err)
		t.HandleError(ctx, err)
		return
	}
	
	logger.Infof("Successfully deleted transaction with ID: %d for user: %d", transactionId, userId)
	t.SendSuccess(ctx, http.StatusNoContent, "", nil)
}

func (t *TransactionController) ListTransactions(ctx *gin.Context) {
	logger.Info("Fetching transactions for user: ", ctx.GetInt64("authUserId"))
	
	userId := ctx.GetInt64("authUserId")
	transactions, err := t.transactionService.ListTransactions(ctx, userId)
	if err != nil {
		logger.Error("[TransactionController] Error listing transactions: ", err)
		t.HandleError(ctx, err)
		return
	}
	
	logger.Infof("Transactions retrieved successfully for user: %d", userId)
	t.SendSuccess(ctx, http.StatusOK, "Transactions retrieved successfully", transactions)
} 