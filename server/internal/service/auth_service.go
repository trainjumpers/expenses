package service

import (
	"crypto/rand"
	"encoding/base64"
	"expenses/internal/errors"
	"expenses/internal/models"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type RefreshTokenData struct {
	UserId int64
	Email  string
	Expiry time.Time
}

type AuthService struct {
	refreshTokenStore struct {
		sync.RWMutex
		Tokens map[string]RefreshTokenData
	}
	userService *UserService
}

func NewAuthService(userService *UserService) *AuthService {
	return &AuthService{
		refreshTokenStore: struct {
			sync.RWMutex
			Tokens map[string]RefreshTokenData
		}{
			Tokens: make(map[string]RefreshTokenData),
		},
		userService: userService,
	}
}

// Signup handles user registration and returns auth tokens
func (a *AuthService) Signup(c *gin.Context, newUser models.CreateUserInput) (models.AuthResponse, error) {
	hashedPassword, err := a.hashPassword(newUser.Password)
	if err != nil {
		return models.AuthResponse{}, err
	}
	newUser.Password = hashedPassword
	createdUser, err := a.userService.CreateUser(c, newUser)
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
	a.saveRefreshToken(refreshToken, RefreshTokenData{
		UserId: createdUser.Id,
		Email:  createdUser.Email,
		Expiry: time.Now().Add(7 * 24 * time.Hour), // 7 days
	})
	return models.AuthResponse{
		User:         createdUser,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Login handles user authentication and returns auth tokens
func (a *AuthService) Login(c *gin.Context, loginInput models.LoginInput) (models.AuthResponse, error) {
	user, err := a.userService.GetUserByEmail(c, loginInput.Email)
	if err != nil {
		return models.AuthResponse{}, errors.NewUserNotFoundError(err)
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

	a.saveRefreshToken(refreshToken, RefreshTokenData{
		UserId: user.Id,
		Email:  user.Email,
		Expiry: time.Now().Add(7 * 24 * time.Hour), // 7 days
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
func (a *AuthService) RefreshToken(c *gin.Context, refreshToken string) (models.AuthResponse, error) {
	data, ok := a.getRefreshTokenData(refreshToken)
	if !ok {
		return models.AuthResponse{}, errors.NewInvalidTokenError(fmt.Errorf("refresh token not found or expired"))
	}

	user, err := a.userService.GetUserById(c, data.UserId)
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

	a.saveRefreshToken(newRefreshToken, RefreshTokenData{
		UserId: user.Id,
		Email:  user.Email,
		Expiry: time.Now().Add(7 * 24 * time.Hour), // 7 days
	})

	a.deleteRefreshToken(refreshToken)
	return models.AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (a *AuthService) saveRefreshToken(token string, data RefreshTokenData) {
	a.refreshTokenStore.Lock()
	defer a.refreshTokenStore.Unlock()
	a.refreshTokenStore.Tokens[token] = data
}

func (a *AuthService) getRefreshTokenData(token string) (RefreshTokenData, bool) {
	a.refreshTokenStore.RLock()
	defer a.refreshTokenStore.RUnlock()
	data, ok := a.refreshTokenStore.Tokens[token]
	if !ok || data.Expiry.Before(time.Now()) {
		return RefreshTokenData{}, false
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

func (a *AuthService) VerifyAuthToken(tokenString string) (jwt.MapClaims, error) {
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

func (a *AuthService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (a *AuthService) checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
