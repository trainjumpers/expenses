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

// CreateCategory creates a new category
// @Summary Create a new category
// @Description Create a new expense category for the authenticated user
// @Tags categories
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param category body models.CreateCategoryInput true "Category data"
// @Success 201 {object} models.CategoryResponse "Category created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /category [post]
func (c *CategoryController) CreateCategory(ctx *gin.Context) {
	var input models.CreateCategoryInput
	if err := c.BindJSON(ctx, &input); err != nil {
		logger.Errorf("Failed to bind JSON: %v", err)
		return
	}
	logger.Infof("Creating new category for user %d", input.CreatedBy)
	category, err := c.categoryService.CreateCategory(ctx, input)
	if err != nil {
		logger.Errorf("Error creating category: %v", err)
		c.HandleError(ctx, err)
		return
	}
	logger.Infof("Category created successfully with Id %d for user %d", category.Id, input.CreatedBy)
	c.SendSuccess(ctx, http.StatusCreated, "Category created successfully", category)
}

// GetCategory retrieves a specific category
// @Summary Get category by ID
// @Description Get category details by category ID for the authenticated user
// @Tags categories
// @Produce json
// @Security BasicAuth
// @Param categoryId path int true "Category ID"
// @Success 200 {object} models.CategoryResponse "Category details"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Category not found"
// @Router /category/{categoryId} [get]
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
	logger.Infof("Category retrieved successfully with Id %d for user %d", category.Id, userId)
	c.SendSuccess(ctx, http.StatusOK, "Category retrieved successfully", category)
}

// ListCategories retrieves all categories for the user
// @Summary List all categories
// @Description Get all categories for the authenticated user
// @Tags categories
// @Produce json
// @Security BasicAuth
// @Success 200 {array} models.CategoryResponse "List of categories"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /category [get]
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

// UpdateCategory updates an existing category
// @Summary Update category
// @Description Update category details by category ID for the authenticated user
// @Tags categories
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param categoryId path int true "Category ID"
// @Param category body models.UpdateCategoryInput true "Updated category data"
// @Success 200 {object} models.CategoryResponse "Category updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Category not found"
// @Router /category/{categoryId} [patch]
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
	logger.Infof("Category updated successfully with Id %d for user %d", category.Id, userId)
	c.SendSuccess(ctx, http.StatusOK, "Category updated successfully", category)
}

// DeleteCategory deletes a category
// @Summary Delete category
// @Description Delete category by category ID for the authenticated user
// @Tags categories
// @Produce json
// @Security BasicAuth
// @Param categoryId path int true "Category ID"
// @Success 204 "Category deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Category not found"
// @Router /category/{categoryId} [delete]
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
	logger.Infof("Category deleted successfully with Id %d for user %d", categoryId, userId)
	c.SendSuccess(ctx, http.StatusNoContent, "", nil)
}
