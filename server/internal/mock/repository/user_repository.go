package mock_repository

import (
	"context"
	"expenses/internal/errors"
	"expenses/internal/models"
	"sync"
)

type MockUserRepository struct {
	users    map[string]models.UserWithPassword
	nextId   int64
	failures map[string]error
	mu       sync.RWMutex
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:  make(map[string]models.UserWithPassword),
		nextId: 1,
	}
}

// FailOn makes the named method return err, so service tests can exercise
// failure paths.
func (m *MockUserRepository) FailOn(method string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failures == nil {
		m.failures = make(map[string]error)
	}
	m.failures[method] = err
}

func (m *MockUserRepository) CreateUser(ctx context.Context, newUser models.CreateUserInput) (models.UserResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.failures["CreateUser"]; err != nil {
		return models.UserResponse{}, err
	}
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

func (m *MockUserRepository) GetUserByEmailWithPassword(ctx context.Context, email string) (models.UserWithPassword, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if user, exists := m.users[email]; exists {
		return user, nil
	}
	return models.UserWithPassword{}, errors.NewUserNotFoundError(nil)
}

func (m *MockUserRepository) GetUserByIdWithPassword(ctx context.Context, userId int64) (models.UserWithPassword, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, user := range m.users {
		if user.Id == userId {
			return user, nil
		}
	}
	return models.UserWithPassword{}, errors.NewUserNotFoundError(nil)
}

func (m *MockUserRepository) GetUserById(ctx context.Context, userId int64) (models.UserResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
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

func (m *MockUserRepository) DeleteUser(ctx context.Context, userId int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for email, user := range m.users {
		if user.Id == userId {
			delete(m.users, email)
			return nil
		}
	}
	return errors.NewUserNotFoundError(nil)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, userId int64, updatedUser models.UpdateUserInput) (models.UserResponse, error) {
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

func (m *MockUserRepository) UpdateUserPassword(ctx context.Context, userId int64, password string) (models.UserResponse, error) {
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
