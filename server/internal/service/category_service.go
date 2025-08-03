package service

import (
	"context"
	"expenses/internal/models"
	"expenses/internal/repository"
)

type CategoryServiceInterface interface {
	CreateCategory(ctx context.Context, input models.CreateCategoryInput) (models.CategoryResponse, error)
	GetCategoryById(ctx context.Context, categoryId int64, userId int64) (models.CategoryResponse, error)
	ListCategories(ctx context.Context, userId int64) ([]models.CategoryResponse, error)
	UpdateCategory(ctx context.Context, categoryId int64, userId int64, input models.UpdateCategoryInput) (models.CategoryResponse, error)
	DeleteCategory(ctx context.Context, categoryId int64, userId int64) error
}

type CategoryService struct {
	repo repository.CategoryRepositoryInterface
}

func NewCategoryService(repo repository.CategoryRepositoryInterface) CategoryServiceInterface {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) CreateCategory(ctx context.Context, input models.CreateCategoryInput) (models.CategoryResponse, error) {
	return s.repo.CreateCategory(ctx, input)
}

func (s *CategoryService) GetCategoryById(ctx context.Context, categoryId int64, userId int64) (models.CategoryResponse, error) {
	return s.repo.GetCategoryById(ctx, categoryId, userId)
}

func (s *CategoryService) ListCategories(ctx context.Context, userId int64) ([]models.CategoryResponse, error) {
	return s.repo.ListCategories(ctx, userId)
}

func (s *CategoryService) UpdateCategory(ctx context.Context, categoryId int64, userId int64, input models.UpdateCategoryInput) (models.CategoryResponse, error) {
	return s.repo.UpdateCategory(ctx, categoryId, userId, input)
}

func (s *CategoryService) DeleteCategory(ctx context.Context, categoryId int64, userId int64) error {
	return s.repo.DeleteCategory(ctx, categoryId, userId)
}
