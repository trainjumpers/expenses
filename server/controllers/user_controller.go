package controllers

import (
	"expenses/entities"
	logger "expenses/logger"
	"expenses/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserController struct {
	userService *services.UserService
}

func NewUserController(db *pgxpool.Pool) *UserController {
	userService := services.NewUserService(db)
	return &UserController{userService: userService}
}

// GetUserById returns a user by ID
func (u *UserController) GetUserById(c *gin.Context) {
	userID := c.GetInt64("authUserID")
	logger.Info("Recieved request to get a user by ID: ", userID)
 
	user, err := u.userService.GetUserByID(c, userID)
	if err != nil {
		logger.Error("Error getting user by ID: ", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

// DeleteUser deletes a user by ID
func (u *UserController) DeleteUser(c *gin.Context) {
	userID := c.GetInt64("authUserID")
	logger.Info("Recieved request to delete a user by ID: ", userID)

	err := u.userService.DeleteUser(c, userID)
	if err != nil {
		logger.Error("Error deleting user: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting user"})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// UpdateUser updates a user by ID
func (u *UserController) UpdateUser(c *gin.Context) {
	userID := c.GetInt64("authUserID")
	logger.Info("Recieved request to update a user by ID: ", userID)

	var updatedUser entities.UpdateUserInput
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		logger.Error("Failed to bind JSON: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := u.userService.UpdateUser(c, userID, updatedUser)
	if err != nil {
		logger.Error("Error updating user: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user", "reason": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"data":    user,
	})
}

// Update User Password
func (u *UserController) UpdateUserPassword(c *gin.Context) {
	userID := c.GetInt64("authUserID")
	logger.Info("Recieved request to update a user password by ID: ", userID)
	var updatedUser entities.UpdateUserPasswordInput
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		logger.Error("Failed to bind JSON: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := u.userService.UpdateUserPassword(c, userID, updatedUser)
	if err != nil {
		logger.Error("Error updating user password: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user password", "reason": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User password updated successfully",
		"data":    user,
	})
}