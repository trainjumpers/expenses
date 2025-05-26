package service

import (
	"expenses/internal/models"
	"expenses/internal/repository"

	"github.com/gin-gonic/gin"
)

// UserServiceInterface defines the contract for user service operations
type UserServiceInterface interface {
	CreateUser(c *gin.Context, newUser models.CreateUserInput) (models.UserResponse, error)
	GetUserByEmailWithPassword(c *gin.Context, email string) (models.UserWithPassword, error)
	GetUserByIdWithPassword(c *gin.Context, userId int64) (models.UserWithPassword, error)
	GetUserById(c *gin.Context, userId int64) (models.UserResponse, error)
	DeleteUser(c *gin.Context, userId int64) error
	UpdateUser(c *gin.Context, userId int64, updatedUser models.UpdateUserInput) (models.UserResponse, error)
	UpdateUserPassword(c *gin.Context, userId int64, password string) (models.UserResponse, error)
}

// UserService implements UserServiceInterface
type UserService struct {
	repo repository.UserRepositoryInterface
}

// NewUserService creates a new UserService instance that implements UserServiceInterface
func NewUserService(repo repository.UserRepositoryInterface) UserServiceInterface {
	return &UserService{repo: repo}
}

func (u *UserService) CreateUser(c *gin.Context, newUser models.CreateUserInput) (models.UserResponse, error) {
	return u.repo.CreateUser(c, newUser)
}

func (u *UserService) GetUserByEmailWithPassword(c *gin.Context, email string) (models.UserWithPassword, error) {
	return u.repo.GetUserByEmailWithPassword(c, email)
}

func (u *UserService) GetUserByIdWithPassword(c *gin.Context, userId int64) (models.UserWithPassword, error) {
	return u.repo.GetUserByIdWithPassword(c, userId)
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

func (u *UserService) UpdateUserPassword(c *gin.Context, userId int64, password string) (models.UserResponse, error) {
	return u.repo.UpdateUserPassword(c, userId, password)
}
