package service

import (
	"expenses/internal/models"
	"expenses/internal/repository"

	"github.com/gin-gonic/gin"
)

// UserService implements UserServiceInterface
type UserService struct {
	repo repository.UserRepositoryInterface
}

// UserServiceInterface defines the contract for user service operations
type UserServiceInterface interface {
	CreateUser(c *gin.Context, newUser models.CreateUserInput) (models.UserResponse, error)
	GetUserByEmail(c *gin.Context, email string) (models.UserWithPassword, error)
	GetUserById(c *gin.Context, userId int64) (models.UserResponse, error)
	DeleteUser(c *gin.Context, userId int64) error
	UpdateUser(c *gin.Context, userId int64, updatedUser models.UpdateUserInput) (models.UserResponse, error)
}

// NewUserService creates a new UserService instance that implements UserServiceInterface
func NewUserService(repo repository.UserRepositoryInterface) UserServiceInterface {
	return &UserService{repo: repo}
}

func (u *UserService) CreateUser(c *gin.Context, newUser models.CreateUserInput) (models.UserResponse, error) {
	return u.repo.CreateUser(c, newUser)
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
