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
	logger.Info("Creating a new account for the user: ", ctx.GetInt64("authUserId"))
	var input models.CreateAccountInput
	if err := a.BindJSON(ctx, &input); err != nil {
		logger.Error("[AccountController] Failed to bind JSON: ", err)
		return
	}
	input.CreatedBy = ctx.GetInt64("authUserId")
	account, err := a.accountService.CreateAccount(ctx, input)
	if err != nil {
		logger.Error("Error creating account: ", err)
		a.HandleError(ctx, err)
		return
	}
	logger.Infof("Account created successfully with ID: %d for user: %d", account.Id, input.CreatedBy)
	a.SendSuccess(ctx, http.StatusCreated, "Account created successfully", account)
}

func (a *AccountController) GetAccount(ctx *gin.Context) {
	accountId, err := strconv.ParseInt(ctx.Param("accountId"), 10, 64)
	if err != nil {
		a.SendError(ctx, http.StatusBadRequest, "invalid account id")
		return
	}
	logger.Info("Fetching account details for user: ", ctx.GetInt64("authUserId"), " and account ID: ", accountId)
	userId := ctx.GetInt64("authUserId")
	account, err := a.accountService.GetAccountById(ctx, accountId, userId)
	if err != nil {
		logger.Error("[AccountController] Error getting account: ", err)
		a.HandleError(ctx, err)
		return
	}
	logger.Infof("Account retrieved successfully with ID: ", account.Id, " for user: ", userId)
	a.SendSuccess(ctx, http.StatusOK, "Account retrieved successfully", account)
}

func (a *AccountController) UpdateAccount(ctx *gin.Context) {
	logger.Infof("Start UpdateAccount for user: ", ctx.GetInt64("authUserId"))
	accountId, err := strconv.ParseInt(ctx.Param("accountId"), 10, 64)
	if err != nil {
		a.SendError(ctx, http.StatusBadRequest, "invalid account id")
		return
	}
	var input models.UpdateAccountInput
	if err := a.BindJSON(ctx, &input); err != nil {
		logger.Error("[AccountController] Failed to bind JSON: ", err)
		return
	}
	userId := ctx.GetInt64("authUserId")
	account, err := a.accountService.UpdateAccount(ctx, accountId, userId, input)
	if err != nil {
		logger.Error("[AccountController] Error updating account: ", err)
		a.HandleError(ctx, err)
		return
	}
	logger.Infof("Account updated successfully with ID: ", account.Id, " for user: ", userId)
	a.SendSuccess(ctx, http.StatusOK, "Account updated successfully", account)
}

func (a *AccountController) DeleteAccount(ctx *gin.Context) {
	logger.Infof("Start DeleteAccount for user: ", ctx.GetInt64("authUserId"))
	accountId, err := strconv.ParseInt(ctx.Param("accountId"), 10, 64)
	if err != nil {
		a.SendError(ctx, http.StatusBadRequest, "invalid account id")
		return
	}
	userId := ctx.GetInt64("authUserId")
	err = a.accountService.DeleteAccount(ctx, accountId, userId)
	if err != nil {
		logger.Error("[AccountController] Error deleting account: ", err)
		a.HandleError(ctx, err)
		return
	}
	logger.Infof("Successfully deleted account with ID: ", accountId, " for user: ", userId)
	a.SendSuccess(ctx, http.StatusNoContent, "", nil)
}

func (a *AccountController) ListAccounts(ctx *gin.Context) {
	logger.Info("Fetching accounts for user: ", ctx.GetInt64("authUserId"))
	userId := ctx.GetInt64("authUserId")
	accounts, err := a.accountService.ListAccounts(ctx, userId)
	if err != nil {
		logger.Error("[AccountController] Error listing accounts: ", err)
		a.HandleError(ctx, err)
		return
	}
	logger.Infof("Accounts retrieved successfully for user: ", userId)
	a.SendSuccess(ctx, http.StatusOK, "Accounts retrieved successfully", accounts)
}
