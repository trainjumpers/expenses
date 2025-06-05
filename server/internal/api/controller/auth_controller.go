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
	*BaseController
	authService service.AuthServiceInterface
}

func NewAuthController(cfg *config.Config, authService service.AuthServiceInterface) *AuthController {
	return &AuthController{
		BaseController: NewBaseController(cfg),
		authService:    authService,
	}
}

// Signup controller handles creation of a new user, and returns the user data along with an access token
func (a *AuthController) Signup(ctx *gin.Context) {
	var newUser models.CreateUserInput
	if err := a.BindJSON(ctx, &newUser); err != nil {
		logger.Errorf("Failed to bind JSON: %v", err)
		return
	}
	logger.Infof("Creating new user with email %s", newUser.Email)
	authResponse, err := a.authService.Signup(ctx, newUser)
	if err != nil {
		logger.Errorf("Failed to sign up user: %v", err)
		a.HandleError(ctx, err)
		return
	}
	logger.Infof("User created successfully with ID %d", authResponse.User.Id)
	a.SendSuccess(ctx, http.StatusCreated, "User signed up successfully", authResponse)
}

// Login controller handles user login and sends back an access token
func (a *AuthController) Login(ctx *gin.Context) {
	var loginInput models.LoginInput
	if err := a.BindJSON(ctx, &loginInput); err != nil {
		logger.Errorf("Failed to bind JSON: %v", err)
		return
	}
	logger.Infof("User login attempt for email %s", loginInput.Email)

	authResponse, err := a.authService.Login(ctx, loginInput)
	if err != nil {
		logger.Errorf("Failed to log in user: %v", err)
		a.HandleError(ctx, err)
		return
	}

	logger.Infof("User logged in successfully with ID %d", authResponse.User.Id)
	a.SendSuccess(ctx, http.StatusOK, "User logged in successfully", gin.H{
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
	if err := a.BindJSON(ctx, &req); err != nil {
		logger.Errorf("Failed to bind JSON: %v", err)
		return
	}

	logger.Infof("Token refresh request received")
	authResponse, err := a.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		logger.Errorf("Failed to refresh token: %v", err)
		a.HandleError(ctx, err)
		return
	}

	logger.Infof("Token refreshed successfully for user ID %d", authResponse.User.Id)
	a.SendSuccess(ctx, http.StatusOK, "Token refreshed successfully", gin.H{
		"user":          authResponse.User,
		"access_token":  authResponse.AccessToken,
		"refresh_token": authResponse.RefreshToken,
	})
}
