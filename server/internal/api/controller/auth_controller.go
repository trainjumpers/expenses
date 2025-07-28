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
// @Summary Create a new user account
// @Description Register a new user with email, name, and password
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.CreateUserInput true "User registration data"
// @Success 201 {object} map[string]interface{} "User created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 409 {object} map[string]interface{} "User already exists"
// @Router /signup [post]
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
	a.setCookies(ctx, authResponse.AccessToken, authResponse.RefreshToken)
	a.SendSuccess(ctx, http.StatusCreated, "User signed up successfully", gin.H{
		"user": authResponse.User,
	})
}

// Login controller handles user login and sends back an access token
// @Summary User login
// @Description Authenticate user with email and password. Returns user data and sets authentication cookies. For API testing, you can extract the access_token from cookies and use it in the Authorization header as "Bearer <token>".
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.LoginInput true "Login credentials"
// @Success 200 {object} map[string]interface{} "Login successful - check cookies for access_token"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Router /login [post]
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
	a.setCookies(ctx, authResponse.AccessToken, authResponse.RefreshToken)
	a.SendSuccess(ctx, http.StatusOK, "User logged in successfully", gin.H{
		"user": authResponse.User,
	})
}

// RefreshToken endpoint issues a new access token if the refresh token is valid
// @Summary Refresh access token
// @Description Get a new access token using refresh token from cookies
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{} "Token refreshed successfully"
// @Failure 401 {object} map[string]interface{} "Invalid refresh token"
// @Router /refresh [post]
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
	a.setCookies(ctx, authResponse.AccessToken, authResponse.RefreshToken)
	a.SendSuccess(ctx, http.StatusOK, "Token refreshed successfully", gin.H{
		"user": authResponse.User,
	})
}

// Logout endpoint clears the auth cookies
// @Summary User logout
// @Description Clear authentication cookies and log out user
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{} "Logged out successfully"
// @Router /logout [post]
func (a *AuthController) Logout(ctx *gin.Context) {
	// Set cookies to expire in the past
	a.setAuthCookie(ctx, "access_token", "", -1)
	a.setAuthCookie(ctx, "refresh_token", "", -1)
	ctx.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

func (a *AuthController) setCookies(ctx *gin.Context, accessToken string, refreshToken string) {
	a.setAuthCookie(ctx, "access_token", accessToken, int(a.cfg.AccessTokenDuration.Seconds()))
	a.setAuthCookie(ctx, "refresh_token", refreshToken, int(a.cfg.RefreshTokenDuration.Seconds()))
}
