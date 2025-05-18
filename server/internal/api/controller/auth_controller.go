package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"expenses/internal/models"
	"expenses/internal/service"
	"expenses/pkg/logger"
	"expenses/pkg/utils"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthController struct {
	userService *service.UserService
}

func NewAuthController(db *pgxpool.Pool) *AuthController {
	userService := service.NewUserService(db)
	return &AuthController{userService: userService}
}

// Signup controller handles creation of a new user, and returns the user data along with an access token
func (a *AuthController) Signup(c *gin.Context) {
	var newUser models.CreateUserInput
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := utils.HashPassword(newUser.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	newUser.Password = hashedPassword
	createdUser, err := a.userService.CreateUser(c, newUser)
	if err != nil {
		logger.Error("Error creating user: ", err)
		if utils.CheckForeignKey(err, "user", "email") {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("User with email: %s already exists", newUser.Email)})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
		return
	}

	accessToken, err := issueAuthToken(createdUser.Id, createdUser.Email)
	if err != nil {
		logger.Error("Error generating token: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}
	refreshToken, err := generateRefreshToken()
	if err != nil {
		logger.Error("Error generating refresh token: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating refresh token"})
		return
	}
	utils.SaveRefreshToken(refreshToken, utils.RefreshTokenData{
		UserId: createdUser.Id,
		Email:  createdUser.Email,
		Expiry: time.Now().Add(7 * 24 * time.Hour), // 7 days
	})
	c.JSON(http.StatusOK, gin.H{
		"message":       "User created successfully",
		"data":          createdUser,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// Login controller handles user login and sends back an access token
func (a *AuthController) Login(c *gin.Context) {
	var loginInput models.LoginInput
	if err := c.ShouldBindJSON(&loginInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logger.Info("Recieved request to login a user for email: ", loginInput.Email)

	user, err := a.userService.GetUserByEmail(c, loginInput.Email)
	if err != nil {
		logger.Error("Error getting user by email: ", err)
		if strings.Contains(err.Error(), "no rows") {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			c.Abort()
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting user"})
		c.Abort()
		return
	}

	logger.Infof("User found with Id: %d", user.Id)
	authenticated := utils.CheckPasswordHash(loginInput.Password, user.Password)
	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	accessToken, err := issueAuthToken(user.Id, user.Email)
	if err != nil {
		logger.Error("Error generating token: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}
	refreshToken, err := generateRefreshToken()
	if err != nil {
		logger.Error("Error generating refresh token: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating refresh token"})
		return
	}
	utils.SaveRefreshToken(refreshToken, utils.RefreshTokenData{
		UserId: user.Id,
		Email:  user.Email,
		Expiry: time.Now().Add(7 * 24 * time.Hour), // 7 days
	})
	c.JSON(http.StatusOK, gin.H{
		"message":       "User logged in successfully",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// RefreshToken endpoint issues a new access token if the refresh token is valid
func (a *AuthController) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	data, ok := utils.GetRefreshTokenData(req.RefreshToken)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}
	accessToken, err := issueAuthToken(data.UserId, data.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not issue access token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}

// generateRefreshToken creates a random string for refresh token
func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// issueAuthToken issues a JWT token with the user Id and email
func issueAuthToken(userId int64, email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 12).Unix(),
	})

	key := []byte(os.Getenv("JWT_SECRET"))

	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// verifyAuthToken verifies the JWT token and returns the claims
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

// Protected is a middleware that checks if the request has a valid JWT token
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Malformed User Id"})
		c.Abort()
		return
	}
	c.Set("authUserId", int64(userId))
	logger.Info("Recieved request from user with Id: ", int64(userId))
	c.Next()
}
