package service

import (
	"expenses/internal/models"
	"expenses/internal/repository"

	"github.com/gin-gonic/gin"
)

type CategoryServiceInterface interface {
	CreateCategory(c *gin.Context, input models.CreateCategoryInput) (models.CategoryResponse, error)
	GetCategoryById(c *gin.Context, categoryId int64, userId int64) (models.CategoryResponse, error)
	ListCategories(c *gin.Context, userId int64) ([]models.CategoryResponse, error)
	UpdateCategory(c *gin.Context, categoryId int64, userId int64, input models.UpdateCategoryInput) (models.CategoryResponse, error)
	DeleteCategory(c *gin.Context, categoryId int64, userId int64) error
}

type CategoryService struct {
	repo repository.CategoryRepositoryInterface
}

func NewCategoryService(repo repository.CategoryRepositoryInterface) CategoryServiceInterface {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) CreateCategory(c *gin.Context, input models.CreateCategoryInput) (models.CategoryResponse, error) {
	return s.repo.CreateCategory(c, input)
}

func (s *CategoryService) GetCategoryById(c *gin.Context, categoryId int64, userId int64) (models.CategoryResponse, error) {
	return s.repo.GetCategoryById(c, categoryId, userId)
}

func (s *CategoryService) ListCategories(c *gin.Context, userId int64) ([]models.CategoryResponse, error) {
	return s.repo.ListCategories(c, userId)
}

func (s *CategoryService) UpdateCategory(c *gin.Context, categoryId int64, userId int64, input models.UpdateCategoryInput) (models.CategoryResponse, error) {
	return s.repo.UpdateCategory(c, categoryId, userId, input)
}

func (s *CategoryService) DeleteCategory(c *gin.Context, categoryId int64, userId int64) error {
	return s.repo.DeleteCategory(c, categoryId, userId)
}
