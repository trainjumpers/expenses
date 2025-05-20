package controllers

import (
	"expenses/internal/models"
	"expenses/internal/repository"
	"expenses/internal/service"
	"expenses/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthController struct {
	authService *service.AuthService
}

func NewAuthController(db *pgxpool.Pool) *AuthController {
	return &AuthController{
		authService: service.NewAuthService(
			service.NewUserService(repository.NewUserRepository(db)),
		),
	}
}

// GetAuthService returns the auth service instance
func (a *AuthController) GetAuthService() *service.AuthService {
	return a.authService
}

// Signup controller handles creation of a new user, and returns the user data along with an access token
func (a *AuthController) Signup(c *gin.Context) {
	var newUser models.CreateUserInput
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logger.Infof("Received request to create user with email: %s", newUser.Email)
	authResponse, err := a.authService.Signup(c, newUser)
	if err != nil {
		handleError(c, err)
		return
	}
	logger.Infof("User created successfully with Id: %d", authResponse.User.Id)
	c.JSON(http.StatusCreated, authResponse)
}

// Login controller handles user login and sends back an access token
func (a *AuthController) Login(c *gin.Context) {
	var loginInput models.LoginInput
	if err := c.ShouldBindJSON(&loginInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logger.Info("Received request to login a user for email: ", loginInput.Email)

	authResponse, err := a.authService.Login(c, loginInput)
	if err != nil {
		handleError(c, err)
		return
	}

	logger.Infof("User logged in successfully with Id: %d", authResponse.User.Id)
	c.JSON(http.StatusOK, gin.H{
		"message":       "User logged in successfully",
		"user":          authResponse.User,
		"access_token":  authResponse.AccessToken,
		"refresh_token": authResponse.RefreshToken,
	})
}

// RefreshToken endpoint issues a new access token if the refresh token is valid
func (a *AuthController) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authResponse, err := a.authService.RefreshToken(c, req.RefreshToken)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Token refreshed successfully",
		"user":          authResponse.User,
		"access_token":  authResponse.AccessToken,
		"refresh_token": authResponse.RefreshToken,
	})
}
