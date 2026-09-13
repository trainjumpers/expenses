package service

import (
	"context"
	"expenses/internal/models"
	"expenses/internal/repository"
)

// UserServiceInterface defines the contract for user service operations
type UserServiceInterface interface {
	CreateUser(ctx context.Context, newUser models.CreateUserInput) (models.UserResponse, error)
	GetUserByEmailWithPassword(ctx context.Context, email string) (models.UserWithPassword, error)
	GetUserByIdWithPassword(ctx context.Context, userId int64) (models.UserWithPassword, error)
	GetUserById(ctx context.Context, userId int64) (models.UserResponse, error)
	DeleteUser(ctx context.Context, userId int64) error
	UpdateUser(ctx context.Context, userId int64, updatedUser models.UpdateUserInput) (models.UserResponse, error)
	UpdateUserPassword(ctx context.Context, userId int64, password string) (models.UserResponse, error)
}

// UserService implements UserServiceInterface
type UserService struct {
	repo        repository.UserRepositoryInterface
	sessionRepo repository.SessionRepositoryInterface
}

// NewUserService creates a new UserService instance that implements UserServiceInterface
func NewUserService(repo repository.UserRepositoryInterface, sessionRepo repository.SessionRepositoryInterface) UserServiceInterface {
	return &UserService{repo: repo, sessionRepo: sessionRepo}
}

func (u *UserService) CreateUser(ctx context.Context, newUser models.CreateUserInput) (models.UserResponse, error) {
	return u.repo.CreateUser(ctx, newUser)
}

func (u *UserService) GetUserByEmailWithPassword(ctx context.Context, email string) (models.UserWithPassword, error) {
	return u.repo.GetUserByEmailWithPassword(ctx, email)
}

func (u *UserService) GetUserByIdWithPassword(ctx context.Context, userId int64) (models.UserWithPassword, error) {
	return u.repo.GetUserByIdWithPassword(ctx, userId)
}

func (u *UserService) GetUserById(ctx context.Context, userId int64) (models.UserResponse, error) {
	return u.repo.GetUserById(ctx, userId)
}

func (u *UserService) DeleteUser(ctx context.Context, userId int64) error {
	if err := u.sessionRepo.RevokeAllForUser(ctx, userId); err != nil {
		return err
	}
	return u.repo.DeleteUser(ctx, userId)
}

func (u *UserService) UpdateUser(ctx context.Context, userId int64, updatedUser models.UpdateUserInput) (models.UserResponse, error) {
	return u.repo.UpdateUser(ctx, userId, updatedUser)
}

func (u *UserService) UpdateUserPassword(ctx context.Context, userId int64, password string) (models.UserResponse, error) {
	return u.repo.UpdateUserPassword(ctx, userId, password)
}
