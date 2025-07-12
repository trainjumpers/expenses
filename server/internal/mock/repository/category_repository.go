package mock_repository

import (
	customErrors "expenses/internal/errors"
	"expenses/internal/models"
	"sync"

	"github.com/gin-gonic/gin"
)

type MockCategoryRepository struct {
	categories map[int64]models.CategoryResponse
	nextId     int64
	mu         sync.RWMutex
}

func NewMockCategoryRepository() *MockCategoryRepository {
	return &MockCategoryRepository{
		categories: make(map[int64]models.CategoryResponse),
		nextId:     1,
	}
}

func (m *MockCategoryRepository) CreateCategory(c *gin.Context, input models.CreateCategoryInput) (models.CategoryResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, cat := range m.categories {
		if cat.Name == input.Name && cat.CreatedBy == input.CreatedBy {
			return models.CategoryResponse{}, customErrors.NewCategoryAlreadyExistsError(nil)
		}
	}

	var icon *string
	if input.Icon != "" {
		icon = &input.Icon
	}

	category := models.CategoryResponse{
		Id:        m.nextId,
		Name:      input.Name,
		Icon:      icon,
		CreatedBy: input.CreatedBy,
	}

	m.categories[m.nextId] = category
	m.nextId++
	return category, nil
}

func (m *MockCategoryRepository) GetCategoryById(c *gin.Context, categoryId int64, userId int64) (models.CategoryResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cat, ok := m.categories[categoryId]
	if !ok || cat.CreatedBy != userId {
		return models.CategoryResponse{}, customErrors.NewCategoryNotFoundError(nil)
	}
	return cat, nil
}

func (m *MockCategoryRepository) ListCategories(c *gin.Context, userId int64) ([]models.CategoryResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []models.CategoryResponse
	for _, cat := range m.categories {
		if cat.CreatedBy == userId {
			result = append(result, cat)
		}
	}
	return result, nil
}

func (m *MockCategoryRepository) UpdateCategory(c *gin.Context, categoryId int64, userId int64, input models.UpdateCategoryInput) (models.CategoryResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cat, ok := m.categories[categoryId]
	if !ok || cat.CreatedBy != userId {
		return models.CategoryResponse{}, customErrors.NewCategoryNotFoundError(nil)
	}

	// Check for duplicate name if name is being updated
	if input.Name != "" && input.Name != cat.Name {
		for _, existingCat := range m.categories {
			if existingCat.Name == input.Name && existingCat.CreatedBy == userId && existingCat.Id != categoryId {
				return models.CategoryResponse{}, customErrors.NewCategoryAlreadyExistsError(nil)
			}
		}
		cat.Name = input.Name
	}

	if input.Icon != nil {
		cat.Icon = input.Icon
	}

	m.categories[categoryId] = cat
	return cat, nil
}

func (m *MockCategoryRepository) DeleteCategory(c *gin.Context, categoryId int64, userId int64) error {
	cat, ok := m.categories[categoryId]
	if !ok || cat.CreatedBy != userId {
		return customErrors.NewCategoryNotFoundError(nil)
	}
	delete(m.categories, categoryId)
	return nil
}
