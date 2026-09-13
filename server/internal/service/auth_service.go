package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"expenses/internal/config"
	"expenses/internal/errors"
	"expenses/internal/models"
	"expenses/internal/repository"
	"expenses/pkg/logger"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthServiceInterface defines the contract for authentication service operations
type AuthServiceInterface interface {
	Signup(ctx context.Context, newUser models.CreateUserInput) (models.AuthResponse, error)
	Login(ctx context.Context, loginInput models.LoginInput) (models.AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (models.AuthResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAllSessions(ctx context.Context, userId int64) error
	UpdateUserPassword(ctx context.Context, userId int64, updatedUser models.UpdateUserPasswordInput) (models.UserResponse, error)
	// ExpireRefreshToken is a helper method for testing purposes only.
	// DO NOT USE IN PRODUCTION.
	ExpireRefreshToken(refreshToken string) error
}

// AuthService implements AuthServiceInterface
type AuthService struct {
	userService UserServiceInterface
	sessionRepo repository.SessionRepositoryInterface
	cfg         *config.Config
}

// NewAuthService creates a new AuthService instance that implements AuthServiceInterface
func NewAuthService(userService UserServiceInterface, sessionRepo repository.SessionRepositoryInterface, cfg *config.Config) AuthServiceInterface {
	return &AuthService{
		userService: userService,
		sessionRepo: sessionRepo,
		cfg:         cfg,
	}
}

// Signup handles user registration and returns auth tokens
func (a *AuthService) Signup(ctx context.Context, newUser models.CreateUserInput) (models.AuthResponse, error) {
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
	if err := a.createSession(ctx, createdUser.Id, refreshToken); err != nil {
		return models.AuthResponse{}, err
	}
	return models.AuthResponse{
		User:         createdUser,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Login handles user authentication and returns auth tokens
func (a *AuthService) Login(ctx context.Context, loginInput models.LoginInput) (models.AuthResponse, error) {
	user, err := a.userService.GetUserByEmailWithPassword(ctx, loginInput.Email)
	if err != nil {
		return models.AuthResponse{}, errors.NewInvalidCredentialsError(err)
	}

	if !a.checkPasswordHash(loginInput.Password, user.Password) {
		return models.AuthResponse{}, errors.NewInvalidCredentialsError(fmt.Errorf("invalid credentials"))
	}

	accessToken, err := a.issueAuthToken(user.Id, user.Email)
	if err != nil {
		return models.AuthResponse{}, errors.NewTokenGenerationError(err)
	}

	refreshToken, err := a.generateRefreshToken()
	if err != nil {
		return models.AuthResponse{}, errors.NewTokenGenerationError(err)
	}

	if err := a.createSession(ctx, user.Id, refreshToken); err != nil {
		return models.AuthResponse{}, errors.NewTokenGenerationError(err)
	}

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

// RefreshToken issues new auth tokens using a valid refresh token, rotating the session
func (a *AuthService) RefreshToken(ctx context.Context, refreshToken string) (models.AuthResponse, error) {
	session, err := a.sessionRepo.GetActiveByHash(ctx, hashRefreshToken(refreshToken))
	if err != nil {
		return models.AuthResponse{}, errors.NewInvalidTokenError(fmt.Errorf("refresh token not found or expired"))
	}

	user, err := a.userService.GetUserById(ctx, session.UserId)
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

	err = a.sessionRepo.Rotate(ctx, user.Id, session.TokenHash, hashRefreshToken(newRefreshToken), time.Now().Add(a.cfg.RefreshTokenDuration))
	if err != nil {
		return models.AuthResponse{}, errors.NewTokenGenerationError(err)
	}

	return models.AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// Logout revokes the session for the given refresh token. An empty token is a no-op.
func (a *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	_, err := a.sessionRepo.RevokeByHash(ctx, hashRefreshToken(refreshToken))
	return err
}

// LogoutAllSessions revokes every active session for a user.
func (a *AuthService) LogoutAllSessions(ctx context.Context, userId int64) error {
	return a.sessionRepo.RevokeAllForUser(ctx, userId)
}

func (a *AuthService) UpdateUserPassword(ctx context.Context, userId int64, updatedUser models.UpdateUserPasswordInput) (models.UserResponse, error) {
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

func (a *AuthService) createSession(ctx context.Context, userId int64, refreshToken string) error {
	if _, err := a.sessionRepo.DeleteExpired(ctx); err != nil {
		logger.Warnf("failed to prune expired sessions: %v", err)
	}
	return a.sessionRepo.Create(ctx, userId, hashRefreshToken(refreshToken), time.Now().Add(a.cfg.RefreshTokenDuration))
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
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
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
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

	affected, err := a.sessionRepo.ExpireByHash(context.Background(), hashRefreshToken(refreshToken))
	if err != nil {
		return errors.NewInvalidTokenError(err)
	}
	if affected == 0 {
		return errors.NewInvalidTokenError(fmt.Errorf("refresh token not found"))
	}
	return nil
}
