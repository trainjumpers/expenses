package controllers

import (
	database "expenses/db"
	models "expenses/models"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	logger "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct{}

func (a *AuthController) Signup(c *gin.Context) {
	var newUser models.CreateUserInput
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := hashPassword(newUser.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	query := fmt.Sprintf("INSERT INTO %s.users (first_name, last_name, email, dob, phone, password) VALUES ($1, $2, $3, $4, $5, $6) "+
		"RETURNING id, first_name, last_name, email, dob, phone;", os.Getenv("PGSCHEMA"))
	insert := database.DbPool.QueryRow(c, query, newUser.FirstName, newUser.LastName, newUser.Email, newUser.DOB, newUser.Phone, hashedPassword)
	var createdUser models.UserOutput

	err = insert.Scan(&createdUser.ID, &createdUser.FirstName, &createdUser.LastName, &createdUser.Email, &createdUser.DOB, &createdUser.Phone)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("User with email: %s already exists", newUser.Email)})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created successfully", "data": createdUser})

}

func (a *AuthController) Login(c *gin.Context) {
	var loginInput models.LoginInput
	if err := c.ShouldBindJSON(&loginInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logger.Info("Recieved request to login a user for email: ", loginInput.Email)

	var user models.User
	query := fmt.Sprintf("SELECT id, email, password FROM %s.users WHERE email = $1;", os.Getenv("PGSCHEMA"))
	result := database.DbPool.QueryRow(c, query, loginInput.Email)

	err := result.Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	authenticated := checkPasswordHash(loginInput.Password, user.Password)
	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := issueAuthToken(user.ID, user.Email)
	if err != nil {
		logger.Error("Error generating token: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User logged in successfully", "access_token": token})
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func issueAuthToken(userId int, email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	key := []byte(os.Getenv("JWT_SECRET"))

	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func verifyAuthToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (a *AuthController) Protected(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
		c.Abort()
		return
	}

	tokenString = strings.Split(tokenString, " ")[1]
	claims, err := verifyAuthToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		c.Abort()
		return
	}
	logger.Info("Type of Claims['user_id']: ", fmt.Sprintf("%T", claims["user_id"]))
	userId, ok := claims["user_id"].(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Malformed User ID"})
		c.Abort()
		return
	}
	c.Set("userID", int64(userId))

	c.Next()
}
