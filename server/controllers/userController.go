package controllers

import (
	database "expenses/db"
	models "expenses/models"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"
)

type UserController struct{}

func (u *UserController) GetUsers(c *gin.Context) {
	var schema = os.Getenv("PGSCHEMA")

	logger.Info("Recieved request to get all users")
	var users []models.UserOutput

	query := fmt.Sprintf("SELECT id, first_name, last_name, email, dob, phone FROM %s.users;", schema)

	logger.Info("Executing query to get all users: ", query)
	result, err := database.DbPool.Query(c, query)
	if err != nil {
		logger.Fatal(fmt.Errorf("error querying the database: %v", err))
	}

	for result.Next() {
		var user models.UserOutput
		err := result.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.DOB, &user.Phone)
		if err != nil {
			panic(err.Error())
		}
		users = append(users, user)
	}

	logger.Info("Number of users found: ", len(users))
	c.JSON(http.StatusOK, gin.H{
		"data": users,
	})
}

func (u *UserController) GetUserById(c *gin.Context) {
	var schema = os.Getenv("PGSCHEMA")

	userID := c.Param("userID")
	logger.Info("Recieved request to get a user by ID: ", userID)

	var user models.User

	query := fmt.Sprintf("SELECT * FROM %s.users WHERE id = $1;", schema)

	logger.Info("Executing query to get a user by ID: ", query)
	result := database.DbPool.QueryRow(c, query, userID)

	err := result.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.DOB, &user.Phone)
	if err != nil {
		panic(err.Error())
	}

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

func (u *UserController) DeleteUser(c *gin.Context) {
	var schema = os.Getenv("PGSCHEMA")

	userID := c.Param("userID")
	logger.Info("Recieved request to delete a user by ID: ", userID)

	query := fmt.Sprintf("DELETE FROM %s.users WHERE id = $1;", schema)

	logger.Info("Executing query to delete a user by ID: ", query)
	_, err := database.DbPool.Exec(c, query, userID)
	if err != nil {
		panic(err.Error())
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}
