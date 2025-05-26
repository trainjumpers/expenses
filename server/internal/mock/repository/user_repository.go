package mock_repository

import (
	"expenses/internal/errors"
	"expenses/internal/models"

	"github.com/gin-gonic/gin"
)

type MockUserRepository struct {
	users  map[string]models.UserWithPassword
	nextId int64
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:  make(map[string]models.UserWithPassword),
		nextId: 1,
	}
}

func (m *MockUserRepository) CreateUser(c *gin.Context, newUser models.CreateUserInput) (models.UserResponse, error) {
	if _, exists := m.users[newUser.Email]; exists {
		return models.UserResponse{}, errors.NewUserAlreadyExistsError(nil)
	}

	user := models.UserWithPassword{
		Id:       m.nextId,
		Email:    newUser.Email,
		Name:     newUser.Name,
		Password: newUser.Password,
	}
	m.users[newUser.Email] = user
	m.nextId++

	return models.UserResponse{
		Id:    user.Id,
		Email: user.Email,
		Name:  user.Name,
	}, nil
}

func (m *MockUserRepository) GetUserByEmailWithPassword(c *gin.Context, email string) (models.UserWithPassword, error) {
	if user, exists := m.users[email]; exists {
		return user, nil
	}
	return models.UserWithPassword{}, errors.NewUserNotFoundError(nil)
}

func (m *MockUserRepository) GetUserByIdWithPassword(c *gin.Context, userId int64) (models.UserWithPassword, error) {
	for _, user := range m.users {
		if user.Id == userId {
			return user, nil
		}
	}
	return models.UserWithPassword{}, errors.NewUserNotFoundError(nil)
}

func (m *MockUserRepository) GetUserById(c *gin.Context, userId int64) (models.UserResponse, error) {
	for _, user := range m.users {
		if user.Id == userId {
			return models.UserResponse{
				Id:    user.Id,
				Email: user.Email,
				Name:  user.Name,
			}, nil
		}
	}
	return models.UserResponse{}, errors.NewUserNotFoundError(nil)
}

func (m *MockUserRepository) DeleteUser(c *gin.Context, userId int64) error {
	for email, user := range m.users {
		if user.Id == userId {
			delete(m.users, email)
			return nil
		}
	}
	return errors.NewUserNotFoundError(nil)
}

func (m *MockUserRepository) UpdateUser(c *gin.Context, userId int64, updatedUser models.UpdateUserInput) (models.UserResponse, error) {
	for email, user := range m.users {
		if user.Id == userId {
			if updatedUser.Name != "" {
				user.Name = updatedUser.Name
				m.users[email] = user
			}
			return models.UserResponse{
				Id:    user.Id,
				Email: user.Email,
				Name:  user.Name,
			}, nil
		}
	}
	return models.UserResponse{}, errors.NewUserNotFoundError(nil)
}

func (m *MockUserRepository) UpdateUserPassword(c *gin.Context, userId int64, password string) (models.UserResponse, error) {
	for email, user := range m.users {
		if user.Id == userId {
			user.Password = password
			m.users[email] = user
			return models.UserResponse{
				Id:    user.Id,
				Email: user.Email,
				Name:  user.Name,
			}, nil
		}
	}
	return models.UserResponse{}, errors.NewUserNotFoundError(nil)
}
