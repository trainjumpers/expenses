package controllers

import (
	models "expenses/models"
	"expenses/services"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"expenses/entities"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	logger "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	userService *services.UserService
}

func NewAuthController(db *pgxpool.Pool) *AuthController {
	userService := services.NewUserService(db)
	return &AuthController{userService: userService}
}

func (a *AuthController) Signup(c *gin.Context) {
	var newUser entities.CreateUserInput
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := hashPassword(newUser.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	createdUser := a.userService.CreateUser(c, models.User{
		FirstName: newUser.FirstName,
		LastName:  newUser.LastName,
		Email:     newUser.Email,
		DOB:       newUser.DOB,
		Phone:     newUser.Phone,
		Password:  hashedPassword,
	})

	c.JSON(http.StatusOK, gin.H{"message": "User created successfully", "data": createdUser})

}

func (a *AuthController) Login(c *gin.Context) {
	var loginInput entities.LoginInput
	if err := c.ShouldBindJSON(&loginInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logger.Info("Recieved request to login a user for email: ", loginInput.Email)

	user := a.userService.GetUserByEmail(c, loginInput.Email)

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

func issueAuthToken(userId int64, email string) (string, error) {
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

	userId, ok := claims["user_id"].(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Malformed User ID"})
		c.Abort()
		return
	}
	c.Set("userID", int64(userId))

	c.Next()
}
