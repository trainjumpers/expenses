package controllers

import (
	"expenses/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	logger "github.com/sirupsen/logrus"
)

type UserController struct {
	userService *services.UserService
}

func NewUserController(db *pgxpool.Pool) *UserController {
	userService := services.NewUserService(db)
	return &UserController{userService: userService}
}

func (u *UserController) GetUsers(c *gin.Context) {
	logger.Info("Recieved request to get all users")

	users := u.userService.GetUsers(c)

	logger.Info("Number of users found: ", len(users))
	c.JSON(http.StatusOK, gin.H{
		"data": users,
	})
}

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

	user := u.userService.GetUserByID(c, userID)

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

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

	u.userService.DeleteUser(c, userID)

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}
