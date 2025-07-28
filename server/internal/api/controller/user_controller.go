package controller

import (
	"expenses/internal/config"
	"expenses/internal/models"
	"expenses/internal/service"
	"expenses/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	*BaseController
	userService service.UserServiceInterface
	authService service.AuthServiceInterface
}

func NewUserController(cfg *config.Config, userService service.UserServiceInterface, authService service.AuthServiceInterface) *UserController {
	return &UserController{
		BaseController: NewBaseController(cfg),
		userService:    userService,
		authService:    authService,
	}
}

// GetUserById returns a user by Id
// @Summary Get current user
// @Description Get the authenticated user's profile information
// @Tags users
// @Produce json
// @Security BasicAuth
// @Success 200 {object} models.UserResponse "User profile"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Router /user [get]
func (u *UserController) GetUserById(ctx *gin.Context) {
	userId := u.GetAuthenticatedUserId(ctx)
	logger.Infof("Fetching user details for Id %d", userId)

	user, err := u.userService.GetUserById(ctx, userId)
	if err != nil {
		logger.Errorf("Error getting user by Id: %v", err)
		u.HandleError(ctx, err)
		return
	}

	logger.Infof("User retrieved successfully for Id %d", userId)
	u.SendSuccess(ctx, http.StatusOK, "User retrieved successfully", user)
}

// DeleteUser deletes a user by Id
// @Summary Delete current user
// @Description Delete the authenticated user's account
// @Tags users
// @Produce json
// @Security BasicAuth
// @Success 204 "User deleted successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /user [delete]
func (u *UserController) DeleteUser(ctx *gin.Context) {
	userId := u.GetAuthenticatedUserId(ctx)
	logger.Infof("Starting user deletion for Id %d", userId)

	err := u.userService.DeleteUser(ctx, userId)
	if err != nil {
		logger.Errorf("Error deleting user: %v", err)
		u.HandleError(ctx, err)
		return
	}

	logger.Infof("User deleted successfully with Id %d", userId)
	u.SendSuccess(ctx, http.StatusNoContent, "User deleted successfully", nil)
}

// UpdateUser updates a user by Id
// @Summary Update current user
// @Description Update the authenticated user's profile information
// @Tags users
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param user body models.UpdateUserInput true "Updated user data"
// @Success 200 {object} models.UserResponse "User updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /user [patch]
func (u *UserController) UpdateUser(ctx *gin.Context) {
	userId := u.GetAuthenticatedUserId(ctx)
	var updatedUser models.UpdateUserInput
	if err := u.BindJSON(ctx, &updatedUser); err != nil {
		logger.Errorf("Failed to bind JSON for updating user: %v", err)
		return
	}
	logger.Infof("Starting user update for Id %d", userId)
	user, err := u.userService.UpdateUser(ctx, userId, updatedUser)
	if err != nil {
		logger.Errorf("Error updating user: %v", err)
		u.HandleError(ctx, err)
		return
	}
	logger.Infof("User updated successfully for Id %d", userId)
	u.SendSuccess(ctx, http.StatusOK, "User updated successfully", user)
}

// UpdateUserPassword updates user password
// @Summary Update user password
// @Description Update the authenticated user's password
// @Tags users
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param password body models.UpdateUserPasswordInput true "Password update data"
// @Success 200 {object} models.UserResponse "Password updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /user/password [post]
func (u *UserController) UpdateUserPassword(ctx *gin.Context) {
	userId := u.GetAuthenticatedUserId(ctx)
	var updatedUser models.UpdateUserPasswordInput
	if err := u.BindJSON(ctx, &updatedUser); err != nil {
		logger.Errorf("Failed to bind JSON for updating user password: %v", err)
		return
	}
	logger.Infof("Starting password update for user Id %d", userId)
	user, err := u.authService.UpdateUserPassword(ctx, userId, updatedUser)
	if err != nil {
		logger.Errorf("Error updating user password: %v", err)
		u.HandleError(ctx, err)
		return
	}

	logger.Infof("User password updated successfully for Id %d", userId)
	u.SendSuccess(ctx, http.StatusOK, "User password updated successfully", user)
}
