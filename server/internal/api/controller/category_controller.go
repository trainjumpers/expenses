package controller

import (
	"expenses/internal/config"
	"expenses/internal/models"
	"expenses/internal/service"
	"expenses/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CategoryController struct {
	*BaseController
	categoryService service.CategoryServiceInterface
}

func NewCategoryController(cfg *config.Config, categoryService service.CategoryServiceInterface) *CategoryController {
	return &CategoryController{
		BaseController:  NewBaseController(cfg),
		categoryService: categoryService,
	}
}

func (c *CategoryController) CreateCategory(ctx *gin.Context) {
	logger.Infof("Creating new category for user %d", c.GetAuthenticatedUserId(ctx))
	var input models.CreateCategoryInput
	if err := c.BindJSON(ctx, &input); err != nil {
		logger.Errorf("Failed to bind JSON: %v", err)
		return
	}
	input.CreatedBy = c.GetAuthenticatedUserId(ctx)
	category, err := c.categoryService.CreateCategory(ctx, input)
	if err != nil {
		logger.Errorf("Error creating category: %v", err)
		c.HandleError(ctx, err)
		return
	}
	logger.Infof("Category created successfully with ID %d for user %d", category.Id, input.CreatedBy)
	c.SendSuccess(ctx, http.StatusCreated, "Category created successfully", category)
}

func (c *CategoryController) GetCategory(ctx *gin.Context) {
	logger.Infof("Fetching category details for user %d", c.GetAuthenticatedUserId(ctx))
	categoryId, err := strconv.ParseInt(ctx.Param("categoryId"), 10, 64)
	if err != nil {
		c.SendError(ctx, http.StatusBadRequest, "invalid category id")
		return
	}
	userId := c.GetAuthenticatedUserId(ctx)
	category, err := c.categoryService.GetCategoryById(ctx, categoryId, userId)
	if err != nil {
		logger.Errorf("Error getting category: %v", err)
		c.HandleError(ctx, err)
		return
	}
	logger.Infof("Category retrieved successfully with ID %d for user %d", category.Id, userId)
	c.SendSuccess(ctx, http.StatusOK, "Category retrieved successfully", category)
}

func (c *CategoryController) ListCategories(ctx *gin.Context) {
	userId := c.GetAuthenticatedUserId(ctx)
	logger.Infof("Fetching categories for user %d", userId)
	categories, err := c.categoryService.ListCategories(ctx, userId)
	if err != nil {
		logger.Errorf("Error listing categories: %v", err)
		c.HandleError(ctx, err)
		return
	}
	logger.Infof("Categories retrieved successfully for user %d", userId)
	c.SendSuccess(ctx, http.StatusOK, "Categories retrieved successfully", categories)
}

func (c *CategoryController) UpdateCategory(ctx *gin.Context) {
	userId := c.GetAuthenticatedUserId(ctx)
	logger.Infof("Starting category update for user %d", userId)
	categoryId, err := strconv.ParseInt(ctx.Param("categoryId"), 10, 64)
	if err != nil {
		c.SendError(ctx, http.StatusBadRequest, "invalid category id")
		return
	}
	var input models.UpdateCategoryInput
	if err := c.BindJSON(ctx, &input); err != nil {
		logger.Errorf("Failed to bind JSON: %v", err)
		return
	}
	category, err := c.categoryService.UpdateCategory(ctx, categoryId, userId, input)
	if err != nil {
		logger.Errorf("Error updating category: %v", err)
		c.HandleError(ctx, err)
		return
	}
	logger.Infof("Category updated successfully with ID %d for user %d", category.Id, userId)
	c.SendSuccess(ctx, http.StatusOK, "Category updated successfully", category)
}

func (c *CategoryController) DeleteCategory(ctx *gin.Context) {
	userId := c.GetAuthenticatedUserId(ctx)
	logger.Infof("Starting category deletion for user %d", userId)
	categoryId, err := strconv.ParseInt(ctx.Param("categoryId"), 10, 64)
	if err != nil {
		c.SendError(ctx, http.StatusBadRequest, "invalid category id")
		return
	}
	err = c.categoryService.DeleteCategory(ctx, categoryId, userId)
	if err != nil {
		logger.Errorf("Error deleting category: %v", err)
		c.HandleError(ctx, err)
		return
	}
	logger.Infof("Category deleted successfully with ID %d for user %d", categoryId, userId)
	c.SendSuccess(ctx, http.StatusNoContent, "", nil)
}
