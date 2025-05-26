package service

import (
	"crypto/rand"
	"encoding/base64"
	"expenses/internal/config"
	"expenses/internal/errors"
	"expenses/internal/models"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthServiceInterface defines the contract for authentication service operations
type AuthServiceInterface interface {
	Signup(ctx *gin.Context, newUser models.CreateUserInput) (models.AuthResponse, error)
	Login(ctx *gin.Context, loginInput models.LoginInput) (models.AuthResponse, error)
	RefreshToken(ctx *gin.Context, refreshToken string) (models.AuthResponse, error)
	UpdateUserPassword(ctx *gin.Context, userId int64, updatedUser models.UpdateUserPasswordInput) (models.UserResponse, error)
	// ExpireRefreshToken is a helper method for testing purposes only.
	// DO NOT USE IN PRODUCTION.
	ExpireRefreshToken(refreshToken string) error
}

// AuthService implements AuthServiceInterface
type AuthService struct {
	refreshTokenStore struct {
		sync.RWMutex
		Tokens map[string]models.RefreshTokenData
	}
	userService UserServiceInterface
	cfg         *config.Config
}

// NewAuthService creates a new AuthService instance that implements AuthServiceInterface
func NewAuthService(userService UserServiceInterface, cfg *config.Config) AuthServiceInterface {
	return &AuthService{
		refreshTokenStore: struct {
			sync.RWMutex
			Tokens map[string]models.RefreshTokenData
		}{
			Tokens: make(map[string]models.RefreshTokenData),
		},
		userService: userService,
		cfg:         cfg,
	}
}

// Signup handles user registration and returns auth tokens
func (a *AuthService) Signup(ctx *gin.Context, newUser models.CreateUserInput) (models.AuthResponse, error) {
	hashedPassword, err := a.hashPassword(newUser.Password)
	if err != nil {
		return models.AuthResponse{}, err
	}
	newUser.Password = hashedPassword
	createdUser, err := a.userService.CreateUser(ctx, newUser)
	if err != nil {
		if errors.CheckForeignKey(err, "unique_active_email") {
			return models.AuthResponse{}, errors.NewUserAlreadyExistsError(err)
		}
		return models.AuthResponse{}, err
	}

	accessToken, err := a.issueAuthToken(createdUser.Id, createdUser.Email)
	if err != nil {
		return models.AuthResponse{}, err
	}
	refreshToken, err := a.generateRefreshToken()
	if err != nil {
		return models.AuthResponse{}, err
	}
	a.saveRefreshToken(refreshToken, models.RefreshTokenData{
		UserId: createdUser.Id,
		Email:  createdUser.Email,
		Expiry: time.Now().Add(a.cfg.RefreshTokenDuration),
	})
	return models.AuthResponse{
		User:         createdUser,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Login handles user authentication and returns auth tokens
func (a *AuthService) Login(ctx *gin.Context, loginInput models.LoginInput) (models.AuthResponse, error) {
	user, err := a.userService.GetUserByEmailWithPassword(ctx, loginInput.Email)
	if err != nil {
		return models.AuthResponse{}, errors.NewInvalidCredentialsError(err)
	}

	if !a.checkPasswordHash(loginInput.Password, user.Password) {
		return models.AuthResponse{}, errors.NewInvalidCredentialsError(fmt.Errorf("password mismatch for user %s", user.Email))
	}

	accessToken, err := a.issueAuthToken(user.Id, user.Email)
	if err != nil {
		return models.AuthResponse{}, errors.NewTokenGenerationError(err)
	}

	refreshToken, err := a.generateRefreshToken()
	if err != nil {
		return models.AuthResponse{}, errors.NewTokenGenerationError(err)
	}

	a.saveRefreshToken(refreshToken, models.RefreshTokenData{
		UserId: user.Id,
		Email:  user.Email,
		Expiry: time.Now().Add(a.cfg.RefreshTokenDuration),
	})

	return models.AuthResponse{
		User: models.UserResponse{
			Id:    user.Id,
			Email: user.Email,
			Name:  user.Name,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken issues new auth tokens using a valid refresh token
func (a *AuthService) RefreshToken(ctx *gin.Context, refreshToken string) (models.AuthResponse, error) {
	data, ok := a.getRefreshTokenData(refreshToken)
	if !ok {
		return models.AuthResponse{}, errors.NewInvalidTokenError(fmt.Errorf("refresh token not found or expired"))
	}

	user, err := a.userService.GetUserById(ctx, data.UserId)
	if err != nil {
		return models.AuthResponse{}, errors.NewUserNotFoundError(err)
	}

	accessToken, err := a.issueAuthToken(user.Id, user.Email)
	if err != nil {
		return models.AuthResponse{}, errors.NewTokenGenerationError(err)
	}

	newRefreshToken, err := a.generateRefreshToken()
	if err != nil {
		return models.AuthResponse{}, errors.NewTokenGenerationError(err)
	}

	a.saveRefreshToken(newRefreshToken, models.RefreshTokenData{
		UserId: user.Id,
		Email:  user.Email,
		Expiry: time.Now().Add(a.cfg.RefreshTokenDuration),
	})

	a.deleteRefreshToken(refreshToken)
	return models.AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (a *AuthService) UpdateUserPassword(ctx *gin.Context, userId int64, updatedUser models.UpdateUserPasswordInput) (models.UserResponse, error) {
	userWithPassword, err := a.userService.GetUserByIdWithPassword(ctx, userId)
	if err != nil {
		return models.UserResponse{}, err
	}
	if !a.checkPasswordHash(updatedUser.OldPassword, userWithPassword.Password) {
		return models.UserResponse{}, errors.NewInvalidCredentialsError(fmt.Errorf("old password is incorrect"))
	}
	hashedPassword, err := a.hashPassword(updatedUser.NewPassword)
	if err != nil {
		return models.UserResponse{}, err
	}
	return a.userService.UpdateUserPassword(ctx, userId, hashedPassword)
}

func (a *AuthService) saveRefreshToken(token string, data models.RefreshTokenData) {
	a.refreshTokenStore.Lock()
	defer a.refreshTokenStore.Unlock()
	a.refreshTokenStore.Tokens[token] = data
}

func (a *AuthService) getRefreshTokenData(token string) (models.RefreshTokenData, bool) {
	a.refreshTokenStore.RLock()
	defer a.refreshTokenStore.RUnlock()
	data, ok := a.refreshTokenStore.Tokens[token]
	if !ok || data.Expiry.Before(time.Now()) {
		return models.RefreshTokenData{}, false
	}
	return data, true
}

func (a *AuthService) deleteRefreshToken(token string) {
	a.refreshTokenStore.Lock()
	defer a.refreshTokenStore.Unlock()
	delete(a.refreshTokenStore.Tokens, token)
}

func (a *AuthService) generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (a *AuthService) issueAuthToken(userId int64, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userId,
		"email":   email,
		"exp":     time.Now().Add(a.cfg.AccessTokenDuration).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.cfg.JWTSecret)
}

func (a *AuthService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (a *AuthService) checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ExpireRefreshToken manually expires a refresh token for testing purposes
func (a *AuthService) ExpireRefreshToken(refreshToken string) error {
	if !a.cfg.IsTest() {
		return errors.New("expiring refresh token is allowed only in test environment")
	}

	a.refreshTokenStore.Lock()
	defer a.refreshTokenStore.Unlock()

	data, exists := a.refreshTokenStore.Tokens[refreshToken]
	if !exists {
		return errors.NewInvalidTokenError(fmt.Errorf("refresh token not found"))
	}

	// Set expiry to past time
	data.Expiry = time.Now().Add(-time.Hour)
	a.refreshTokenStore.Tokens[refreshToken] = data
	return nil
}
