package service

import (
	"crypto/rand"
	"encoding/base64"
	"expenses/internal/models"
	"expenses/internal/repository"
	"expenses/pkg/utils"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type UserService struct {
    repo     repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
    return &UserService{repo: repo}
}

func (u *UserService) CreateUser(c *gin.Context, newUser models.CreateUserInput) (models.AuthResponse, error) {
	hashedPassword, err := utils.HashPassword(newUser.Password)
	if err != nil {
		return models.AuthResponse{}, err
	}
	newUser.Password = hashedPassword
	createdUser, err := u.repo.CreateUser(c, newUser)
	if err != nil {
		return models.AuthResponse{}, err
	}

	accessToken, err := issueAuthToken(createdUser.Id, createdUser.Email)
	if err != nil {
		return models.AuthResponse{}, err
	}
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return models.AuthResponse{}, err
	}
	utils.SaveRefreshToken(refreshToken, utils.RefreshTokenData{
		UserId: createdUser.Id,
		Email:  createdUser.Email,
		Expiry: time.Now().Add(7 * 24 * time.Hour), // 7 days
	})
	return models.AuthResponse{
		User:          createdUser,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (u *UserService) GetUserByEmail(c *gin.Context, email string) (models.UserWithPassword, error) {
	return u.repo.GetUserByEmail(c, email)
}


func (u *UserService) GetUserById(c *gin.Context, userId int64) (models.UserResponse, error) {
	return u.repo.GetUserById(c, userId)
}

func (u *UserService) DeleteUser(c *gin.Context, userId int64) error {
	return u.repo.DeleteUser(c, userId)
}

func (u *UserService) UpdateUser(c *gin.Context, userId int64, updatedUser models.UpdateUserInput) (models.UserResponse, error) {
	return u.repo.UpdateUser(c, userId, updatedUser)
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
