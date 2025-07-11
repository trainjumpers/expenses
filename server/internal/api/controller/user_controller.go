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

// Update User Password
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
