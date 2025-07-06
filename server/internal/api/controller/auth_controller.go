package controller

import (
	"errors"
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
	logger.Infof("User created successfully with Id %d", authResponse.User.Id)
	// Set HTTP-only cookies for tokens
	ctx.SetCookie("access_token", authResponse.AccessToken, int(a.cfg.AccessTokenDuration.Seconds()), "/", a.getCookieUrl(), true, true)
	ctx.SetCookie("refresh_token", authResponse.RefreshToken, int(a.cfg.RefreshTokenDuration.Seconds()), "/", a.getCookieUrl(), true, true)
	a.SendSuccess(ctx, http.StatusCreated, "User signed up successfully", gin.H{
		"user": authResponse.User,
	})
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

	logger.Infof("User logged in successfully with Id %d", authResponse.User.Id)
	// Set HTTP-only cookies for tokens
	ctx.SetCookie("access_token", authResponse.AccessToken, int(a.cfg.AccessTokenDuration.Seconds()), "/", a.getCookieUrl(), true, true)
	ctx.SetCookie("refresh_token", authResponse.RefreshToken, int(a.cfg.RefreshTokenDuration.Seconds()), "/", a.getCookieUrl(), true, true)
	a.SendSuccess(ctx, http.StatusOK, "User logged in successfully", gin.H{
		"user": authResponse.User,
	})
}

// RefreshToken endpoint issues a new access token if the refresh token is valid
func (a *AuthController) RefreshToken(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		logger.Errorf("No refresh_token cookie provided: %v", err)
		a.HandleError(ctx, errors.New("refresh token cookie missing"))
		return
	}

	logger.Infof("Token refresh request received")
	authResponse, err := a.authService.RefreshToken(ctx, refreshToken)
	if err != nil {
		logger.Errorf("Failed to refresh token: %v", err)
		a.HandleError(ctx, err)
		return
	}

	logger.Infof("Token refreshed successfully for user Id %d", authResponse.User.Id)
	// Set HTTP-only cookies for tokens
	ctx.SetCookie("access_token", authResponse.AccessToken, int(a.cfg.AccessTokenDuration.Seconds()), "/", a.getCookieUrl(), true, true)
	ctx.SetCookie("refresh_token", authResponse.RefreshToken, int(a.cfg.RefreshTokenDuration.Seconds()), "/", a.getCookieUrl(), true, true)
	a.SendSuccess(ctx, http.StatusOK, "Token refreshed successfully", gin.H{
		"user": authResponse.User,
	})
}

// Logout endpoint clears the auth cookies
func (a *AuthController) Logout(ctx *gin.Context) {
	// Set cookies to expire in the past
	ctx.SetCookie("access_token", "", -1, "/", a.getCookieUrl(), true, true)
	ctx.SetCookie("refresh_token", "", -1, "/", a.getCookieUrl(), true, true)
	ctx.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

func (a *AuthController) getCookieUrl() string {
	if a.cfg.Environment == "prod" {
		return "https://neurospend.vercel.app"
	}
	return ""
}
