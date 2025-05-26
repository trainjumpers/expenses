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
	userId := ctx.GetInt64("authUserId")
	logger.Infof("Received request to get a user by Id: %d", userId)

	user, err := u.userService.GetUserById(ctx, userId)
	if err != nil {
		logger.Error("Error getting user by Id: ", err)
		u.HandleError(ctx, err)
		return
	}

	logger.Infof("User retrieved successfully: %+v", user)
	u.SendSuccess(ctx, http.StatusOK, "User retrieved successfully", user)
}

// DeleteUser deletes a user by Id
func (u *UserController) DeleteUser(ctx *gin.Context) {
	userId := ctx.GetInt64("authUserId")
	logger.Infof("Received request to delete a user by Id: %d", userId)

	err := u.userService.DeleteUser(ctx, userId)
	if err != nil {
		logger.Error("Error deleting user: ", err)
		u.HandleError(ctx, err)
		return
	}

	logger.Infof("User deleted successfully: %d", userId)
	u.SendSuccess(ctx, http.StatusNoContent, "User deleted successfully", nil)
}

// UpdateUser updates a user by Id
func (u *UserController) UpdateUser(ctx *gin.Context) {
	userId := ctx.GetInt64("authUserId")
	var updatedUser models.UpdateUserInput
	if err := u.BindJSON(ctx, &updatedUser); err != nil {
		logger.Error("Failed to bind JSON for updating user: ", err)
		return
	}
	logger.Infof("Received request to update a user by Id: %d", userId)
	user, err := u.userService.UpdateUser(ctx, userId, updatedUser)
	if err != nil {
		logger.Error("Error updating user: ", err)
		u.HandleError(ctx, err)
		return
	}
	logger.Infof("User updated successfully: %+v", user)
	u.SendSuccess(ctx, http.StatusOK, "User updated successfully", user)
}

// Update User Password
func (u *UserController) UpdateUserPassword(ctx *gin.Context) {
	userId := ctx.GetInt64("authUserId")
	var updatedUser models.UpdateUserPasswordInput
	if err := u.BindJSON(ctx, &updatedUser); err != nil {
		logger.Error("Failed to bind JSON for updating user password: ", err)
		return
	}
	logger.Infof("Received request to update a user password by Id: %d", userId)
	user, err := u.authService.UpdateUserPassword(ctx, userId, updatedUser)
	if err != nil {
		logger.Error("Error updating user password: ", err)
		u.HandleError(ctx, err)
		return
	}

	logger.Infof("User password updated successfully: %+v", user)
	u.SendSuccess(ctx, http.StatusOK, "User password updated successfully", user)
}
