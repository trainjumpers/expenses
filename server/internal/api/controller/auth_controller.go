package controller

import (
	"expenses/internal/config"
	"expenses/internal/models"
	"expenses/internal/service"
	"expenses/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *service.AuthService
	cfg         *config.Config
}

func NewAuthController(cfg *config.Config, authService *service.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
		cfg:         cfg,
	}
}

// GetAuthService returns the auth service instance
func (a *AuthController) GetAuthService() *service.AuthService {
	return a.authService
}

// Signup controller handles creation of a new user, and returns the user data along with an access token
func (a *AuthController) Signup(ctx *gin.Context) {
	var newUser models.CreateUserInput
	if err := ctx.ShouldBindJSON(&newUser); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logger.Infof("Received request to create user with email: %s", newUser.Email)
	authResponse, err := a.authService.Signup(ctx, newUser)
	if err != nil {
		logger.Error("Failed to sign up user: ", err)
		handleError(ctx, a.cfg.IsDev(), err)
		return
	}
	logger.Infof("User created successfully with Id: %d", authResponse.User.Id)
	ctx.JSON(http.StatusCreated, authResponse)
}

// Login controller handles user login and sends back an access token
func (a *AuthController) Login(ctx *gin.Context) {
	var loginInput models.LoginInput
	if err := ctx.ShouldBindJSON(&loginInput); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logger.Info("Received request to log in a user for email: ", loginInput.Email)

	authResponse, err := a.authService.Login(ctx, loginInput)
	if err != nil {
		logger.Error("Failed to log in user: ", err)
		handleError(ctx, a.cfg.IsDev(), err)
		return
	}

	logger.Infof("User logged in successfully with Id: %d", authResponse.User.Id)
	ctx.JSON(http.StatusOK, gin.H{
		"message":       "User logged in successfully",
		"user":          authResponse.User,
		"access_token":  authResponse.AccessToken,
		"refresh_token": authResponse.RefreshToken,
	})
}

// RefreshToken endpoint issues a new access token if the refresh token is valid
func (a *AuthController) RefreshToken(ctx *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authResponse, err := a.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		logger.Error("Failed to refresh token: ", err)
		handleError(ctx, a.cfg.IsDev(), err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":       "Token refreshed successfully",
		"user":          authResponse.User,
		"access_token":  authResponse.AccessToken,
		"refresh_token": authResponse.RefreshToken,
	})
}
