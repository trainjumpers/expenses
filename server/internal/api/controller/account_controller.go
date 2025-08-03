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

// CreateAccount creates a new account
// @Summary Create a new account
// @Description Create a new bank account for the authenticated user
// @Tags accounts
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param account body models.CreateAccountInput true "Account data"
// @Success 201 {object} models.AccountResponse "Account created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /account [post]
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
	logger.Infof("Account created successfully with Id %d for user %d", account.Id, input.CreatedBy)
	a.SendSuccess(ctx, http.StatusCreated, "Account created successfully", account)
}

// GetAccount retrieves a specific account
// @Summary Get account by ID
// @Description Get account details by account ID for the authenticated user
// @Tags accounts
// @Produce json
// @Security BasicAuth
// @Param accountId path int true "Account ID"
// @Success 200 {object} models.AccountResponse "Account details"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Account not found"
// @Router /account/{accountId} [get]
func (a *AccountController) GetAccount(ctx *gin.Context) {
	accountId, err := strconv.ParseInt(ctx.Param("accountId"), 10, 64)
	if err != nil {
		a.SendError(ctx, http.StatusBadRequest, "invalid account id")
		return
	}
	logger.Infof("Fetching account details for user %d and account Id %d", a.GetAuthenticatedUserId(ctx), accountId)
	userId := a.GetAuthenticatedUserId(ctx)
	account, err := a.accountService.GetAccountById(ctx, accountId, userId)
	if err != nil {
		logger.Errorf("Error getting account: %v", err)
		a.HandleError(ctx, err)
		return
	}
	logger.Infof("Account retrieved successfully with Id %d for user %d", account.Id, userId)
	a.SendSuccess(ctx, http.StatusOK, "Account retrieved successfully", account)
}

// UpdateAccount updates an existing account
// @Summary Update account
// @Description Update account details by account ID for the authenticated user
// @Tags accounts
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param accountId path int true "Account ID"
// @Param account body models.UpdateAccountInput true "Updated account data"
// @Success 200 {object} models.AccountResponse "Account updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Account not found"
// @Router /account/{accountId} [patch]
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
	logger.Infof("Account updated successfully with Id %d for user %d", account.Id, userId)
	a.SendSuccess(ctx, http.StatusOK, "Account updated successfully", account)
}

// DeleteAccount deletes an account
// @Summary Delete account
// @Description Delete account by account ID for the authenticated user
// @Tags accounts
// @Produce json
// @Security BasicAuth
// @Param accountId path int true "Account ID"
// @Success 204 "Account deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Account not found"
// @Router /account/{accountId} [delete]
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
	logger.Infof("Successfully deleted account with Id %d for user %d", accountId, userId)
	a.SendSuccess(ctx, http.StatusNoContent, "", nil)
}

// ListAccounts retrieves all accounts for the user
// @Summary List all accounts
// @Description Get all accounts for the authenticated user
// @Tags accounts
// @Produce json
// @Security BasicAuth
// @Success 200 {array} models.AccountResponse "List of accounts"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /account [get]
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
