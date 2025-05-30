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
	logger.Infof("[AccountController] Start CreateAccount")
	var input models.CreateAccountInput
	if err := a.BindJSON(ctx, &input); err != nil {
		logger.Error("[AccountController] Failed to bind JSON: ", err)
		return
	}
	input.CreatedBy = ctx.GetInt64("authUserId")
	account, err := a.accountService.CreateAccount(ctx, input)
	if err != nil {
		logger.Error("[AccountController] Error creating account: ", err)
		a.HandleError(ctx, err)
		return
	}
	logger.Infof("[AccountController] End CreateAccount")
	a.SendSuccess(ctx, http.StatusCreated, "Account created successfully", account)
}

func (a *AccountController) GetAccount(ctx *gin.Context) {
	logger.Infof("[AccountController] Start GetAccount")
	accountId, err := strconv.ParseInt(ctx.Param("accountId"), 10, 64)
	if err != nil {
		a.SendError(ctx, http.StatusBadRequest, "invalid account id")
		return
	}
	userId := ctx.GetInt64("authUserId")
	account, err := a.accountService.GetAccountById(ctx, accountId, userId)
	if err != nil {
		logger.Error("[AccountController] Error getting account: ", err)
		a.HandleError(ctx, err)
		return
	}
	logger.Infof("[AccountController] End GetAccount")
	a.SendSuccess(ctx, http.StatusOK, "Account retrieved successfully", account)
}

func (a *AccountController) UpdateAccount(ctx *gin.Context) {
	logger.Infof("[AccountController] Start UpdateAccount")
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
	logger.Infof("[AccountController] End UpdateAccount")
	a.SendSuccess(ctx, http.StatusOK, "Account updated successfully", account)
}

func (a *AccountController) DeleteAccount(ctx *gin.Context) {
	logger.Infof("[AccountController] Start DeleteAccount")
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
	logger.Infof("[AccountController] End DeleteAccount")
	a.SendSuccess(ctx, http.StatusNoContent, "", nil)
}

func (a *AccountController) ListAccounts(ctx *gin.Context) {
	logger.Infof("[AccountController] Start ListAccounts")
	userId := ctx.GetInt64("authUserId")
	accounts, err := a.accountService.ListAccounts(ctx, userId)
	if err != nil {
		logger.Error("[AccountController] Error listing accounts: ", err)
		a.HandleError(ctx, err)
		return
	}
	logger.Infof("[AccountController] End ListAccounts")
	a.SendSuccess(ctx, http.StatusOK, "Accounts retrieved successfully", accounts)
}
