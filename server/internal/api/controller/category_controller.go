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
	logger.Info("Received request to create a new category for user: ", ctx.GetInt64("authUserId"))
	var input models.CreateCategoryInput
	if err := c.BindJSON(ctx, &input); err != nil {
		logger.Error("[CategoryController] Failed to bind JSON: ", err)
		return
	}
	input.CreatedBy = ctx.GetInt64("authUserId")
	category, err := c.categoryService.CreateCategory(ctx, input)
	if err != nil {
		logger.Error("[CategoryController] Error creating category: ", err)
		c.HandleError(ctx, err)
		return
	}
	logger.Info("Successfully created category with ID: ", category.Id, " for user: ", input.CreatedBy)
	c.SendSuccess(ctx, http.StatusCreated, "Category created successfully", category)
}

func (c *CategoryController) GetCategory(ctx *gin.Context) {
	logger.Info("Received request to get category details for user: ", ctx.GetInt64("authUserId"))
	categoryId, err := strconv.ParseInt(ctx.Param("categoryId"), 10, 64)
	if err != nil {
		c.SendError(ctx, http.StatusBadRequest, "invalid category id")
		return
	}
	userId := ctx.GetInt64("authUserId")
	category, err := c.categoryService.GetCategoryById(ctx, categoryId, userId)
	if err != nil {
		logger.Error("[CategoryController] Error getting category: ", err)
		c.HandleError(ctx, err)
		return
	}
	logger.Infof("Successfully retrieved category with ID: ", category.Id, " for user: ", userId)
	c.SendSuccess(ctx, http.StatusOK, "Category retrieved successfully", category)
}

func (c *CategoryController) ListCategories(ctx *gin.Context) {
	userId := ctx.GetInt64("authUserId")
	logger.Info("Received request to list categories for user: ", userId)
	categories, err := c.categoryService.ListCategories(ctx, userId)
	if err != nil {
		logger.Error("[CategoryController] Error listing categories: ", err)
		c.HandleError(ctx, err)
		return
	}
	logger.Infof("Successfully retrieved categories for user: ", userId)
	c.SendSuccess(ctx, http.StatusOK, "Categories retrieved successfully", categories)
}

func (c *CategoryController) UpdateCategory(ctx *gin.Context) {
	userId := ctx.GetInt64("authUserId")
	logger.Infof("Received request to update category for user: ", userId)
	categoryId, err := strconv.ParseInt(ctx.Param("categoryId"), 10, 64)
	if err != nil {
		c.SendError(ctx, http.StatusBadRequest, "invalid category id")
		return
	}
	var input models.UpdateCategoryInput
	if err := c.BindJSON(ctx, &input); err != nil {
		logger.Error("[CategoryController] Failed to bind JSON: ", err)
		return
	}
	category, err := c.categoryService.UpdateCategory(ctx, categoryId, userId, input)
	if err != nil {
		logger.Error("[CategoryController] Error updating category: ", err)
		c.HandleError(ctx, err)
		return
	}
	logger.Infof("Category updated successfully with ID: ", category.Id, " for user: ", userId)
	c.SendSuccess(ctx, http.StatusOK, "Category updated successfully", category)
}

func (c *CategoryController) DeleteCategory(ctx *gin.Context) {
	userId := ctx.GetInt64("authUserId")
	logger.Infof("Received request to delete category for user: ", userId)
	categoryId, err := strconv.ParseInt(ctx.Param("categoryId"), 10, 64)
	if err != nil {
		c.SendError(ctx, http.StatusBadRequest, "invalid category id")
		return
	}
	err = c.categoryService.DeleteCategory(ctx, categoryId, userId)
	if err != nil {
		logger.Error("[CategoryController] Error deleting category: ", err)
		c.HandleError(ctx, err)
		return
	}
	logger.Infof("Successfully deleted category with ID: ", categoryId, " for user: ", userId)
	c.SendSuccess(ctx, http.StatusNoContent, "", nil)
}
