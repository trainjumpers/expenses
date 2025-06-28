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

type AccountController struct {
	*BaseController
	accountService service.AccountServiceInterface
}

func NewAccountController(cfg *config.Config, accountService service.AccountServiceInterface) *AccountController {
	return &AccountController{
		BaseController: NewBaseController(cfg),
		accountService: accountService,
	}
}

func (a *AccountController) CreateAccount(ctx *gin.Context) {
	var input models.CreateAccountInput
	if err := a.BindJSON(ctx, &input); err != nil {
		logger.Errorf("Failed to bind JSON: %v", err)
		return
	}
	logger.Infof("Creating new account for user %d", input.CreatedBy)
	account, err := a.accountService.CreateAccount(ctx, input)
	if err != nil {
		logger.Errorf("Error creating account: %v", err)
		a.HandleError(ctx, err)
		return
	}
	logger.Infof("Account created successfully with ID %d for user %d", account.Id, input.CreatedBy)
	a.SendSuccess(ctx, http.StatusCreated, "Account created successfully", account)
}

func (a *AccountController) GetAccount(ctx *gin.Context) {
	accountId, err := strconv.ParseInt(ctx.Param("accountId"), 10, 64)
	if err != nil {
		a.SendError(ctx, http.StatusBadRequest, "invalid account id")
		return
	}
	logger.Infof("Fetching account details for user %d and account ID %d", a.GetAuthenticatedUserId(ctx), accountId)
	userId := a.GetAuthenticatedUserId(ctx)
	account, err := a.accountService.GetAccountById(ctx, accountId, userId)
	if err != nil {
		logger.Errorf("Error getting account: %v", err)
		a.HandleError(ctx, err)
		return
	}
	logger.Infof("Account retrieved successfully with ID %d for user %d", account.Id, userId)
	a.SendSuccess(ctx, http.StatusOK, "Account retrieved successfully", account)
}

func (a *AccountController) UpdateAccount(ctx *gin.Context) {
	logger.Infof("Starting update account for user %d", a.GetAuthenticatedUserId(ctx))
	accountId, err := strconv.ParseInt(ctx.Param("accountId"), 10, 64)
	if err != nil {
		a.SendError(ctx, http.StatusBadRequest, "invalid account id")
		return
	}
	var input models.UpdateAccountInput
	if err := a.BindJSON(ctx, &input); err != nil {
		logger.Errorf("Failed to bind JSON: %v", err)
		return
	}
	userId := a.GetAuthenticatedUserId(ctx)
	account, err := a.accountService.UpdateAccount(ctx, accountId, userId, input)
	if err != nil {
		logger.Errorf("Error updating account: %v", err)
		a.HandleError(ctx, err)
		return
	}
	logger.Infof("Account updated successfully with ID %d for user %d", account.Id, userId)
	a.SendSuccess(ctx, http.StatusOK, "Account updated successfully", account)
}

func (a *AccountController) DeleteAccount(ctx *gin.Context) {
	logger.Infof("Starting delete account for user %d", a.GetAuthenticatedUserId(ctx))
	accountId, err := strconv.ParseInt(ctx.Param("accountId"), 10, 64)
	if err != nil {
		a.SendError(ctx, http.StatusBadRequest, "invalid account id")
		return
	}
	userId := a.GetAuthenticatedUserId(ctx)
	err = a.accountService.DeleteAccount(ctx, accountId, userId)
	if err != nil {
		logger.Errorf("Error deleting account: %v", err)
		a.HandleError(ctx, err)
		return
	}
	logger.Infof("Successfully deleted account with ID %d for user %d", accountId, userId)
	a.SendSuccess(ctx, http.StatusNoContent, "", nil)
}

func (a *AccountController) ListAccounts(ctx *gin.Context) {
	logger.Infof("Fetching accounts for user %d", a.GetAuthenticatedUserId(ctx))
	userId := a.GetAuthenticatedUserId(ctx)
	accounts, err := a.accountService.ListAccounts(ctx, userId)
	if err != nil {
		logger.Errorf("Error listing accounts: %v", err)
		a.HandleError(ctx, err)
		return
	}
	logger.Infof("Accounts retrieved successfully for user %d", userId)
	a.SendSuccess(ctx, http.StatusOK, "Accounts retrieved successfully", accounts)
}
