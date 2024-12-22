package controllers

import (
	"expenses/entities"
	logger "expenses/logger"
	"expenses/services"
	"net/http"
	"strconv"

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

// GetUsers returns all users
func (u *UserController) GetUsers(c *gin.Context) {
	logger.Info("Recieved request to get all users")

	users, err := u.userService.GetUsers(c)
	if err != nil {
		logger.Error("Error getting users: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting users"})
		c.Abort()
		return
	}

	logger.Info("Number of users found: ", len(users))
	c.JSON(http.StatusOK, gin.H{
		"data": users,
	})
}

// GetUserById returns a user by ID
func (u *UserController) GetUserById(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("userID"), 10, 64)
	if err != nil {
		logger.Error("Failed to parse userID: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

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
	userID, err := strconv.ParseInt(c.Param("userID"), 10, 64)
	if err != nil {
		logger.Error("Failed to parse userID: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	logger.Info("Recieved request to delete a user by ID: ", userID)

	err = u.userService.DeleteUser(c, userID)
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
	userID, err := strconv.ParseInt(c.Param("userID"), 10, 64)
	if err != nil {
		logger.Error("Failed to parse userID: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user"})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"data":    user,
	})
}
