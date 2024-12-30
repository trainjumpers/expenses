package controllers

import (
	"expenses/entities"
	"expenses/logger"
	models "expenses/models"
	"expenses/services"
	"expenses/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryController struct {
	categoryService *services.CategoryService
}

func NewCategoryController(db *pgxpool.Pool) *CategoryController {
	categoryService := services.NewCategoryService(db)
	return &CategoryController{categoryService: categoryService}
}

func (c *CategoryController) CreateCategory(ctx *gin.Context) {
	logger.Info("Starting category creation process")
	var input entities.CreateCategoryInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		logger.Error("Error binding JSON for category creation: ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetInt64("authUserID")
	logger.Info("Attempting to create category with name: ", input.Name)
	category, err := c.categoryService.CreateCategory(ctx, input.Name, input.Color, userID)
	if err != nil {
		logger.Error("Error creating category: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating category"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Category created successfully", "data": category})
}

func (c *CategoryController) CreateSubCategory(ctx *gin.Context) {
	logger.Info("Starting sub-category creation process")
	var input entities.CreateSubCategoryInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		logger.Error("Error binding JSON for sub-category creation: ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetInt64("authUserID")
	categoryID , err := strconv.ParseInt( ctx.Param("categoryID"), 10, 64)
	if err != nil {
		logger.Error("Error parsing category ID: ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}
	logger.Info("Attempting to create sub-category with name: ", input.Name, " for category ID: ", categoryID)
	createdSubCategory, err := c.categoryService.CreateSubCategory(ctx, models.SubCategory{
		Name:       input.Name,
		Color:      input.Color,
	}, categoryID, userID)
	if err != nil {
		if utils.CheckForeignKey(err, "category_subcategory_mapping", "category") {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
			return
		}
		logger.Error("Error creating sub-category: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating sub-category"})
		return
	}

	logger.Info("Successfully created sub-category with ID: ", createdSubCategory.ID)
	ctx.JSON(http.StatusOK, gin.H{"message": "Sub-category created successfully", "data": createdSubCategory})
}

func (c *CategoryController) GetCategory(ctx *gin.Context) {
	categoryID := ctx.Param("categoryID")
	userID := ctx.GetInt64("authUserID")
	logger.Info("Attempting to fetch category with ID: ", categoryID, " for user: ", userID)

	category, err := c.categoryService.GetCategory(ctx, categoryID, userID)
	if err != nil {
		logger.Error("Error fetching category: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching category"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": category})
}

func (c *CategoryController) GetAllCategories(ctx *gin.Context) {
	userID := ctx.GetInt64("authUserID")
	logger.Info("Attempting to fetch all categories for user: ", userID)

	categories, err := c.categoryService.GetAllCategories(ctx, userID)
	if err != nil {
		logger.Error("Error fetching categories: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching categories"})
		return
	}

	if len(categories) == 0 {
		logger.Info("No categories found for user: ", userID)
		ctx.JSON(http.StatusOK, gin.H{"message": "No categories found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": categories})
}

func (c *CategoryController) UpdateCategory(ctx *gin.Context) {
	var input entities.UpdateCategoryInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		logger.Error("Error binding JSON for category update: ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetInt64("authUserID")
	categoryID := ctx.Param("categoryID")
	logger.Info("Attempting to update category with ID: ", categoryID, " for user: ", userID)

	category, err := c.categoryService.UpdateCategory(ctx, categoryID, input, userID)
	if err != nil {
		logger.Error("Error updating category: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating category"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Category updated successfully", "data": category})
}

func (c *CategoryController) UpdateSubCategory(ctx *gin.Context) {
	var input entities.UpdateSubCategoryInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		logger.Error("Error binding JSON for sub-category update: ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetInt64("authUserID")
	subCategoryID := ctx.Param("subCategoryID")
	logger.Info("Attempting to update sub-category with ID: ", subCategoryID, " for user: ", userID)

	subCategory, err := c.categoryService.UpdateSubCategory(ctx, subCategoryID, input, userID)
	if err != nil {
		logger.Error("Error updating sub-category: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating sub-category"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Sub-category updated successfully", "data": subCategory})
}

func (c *CategoryController) DeleteCategory(ctx *gin.Context) {
	userID := ctx.GetInt64("authUserID")
	categoryID := ctx.Param("categoryID")
	logger.Info("Attempting to delete category with ID: ", categoryID, " for user: ", userID)

	err := c.categoryService.DeleteCategory(ctx, categoryID, userID)
	if err != nil {
		logger.Error("Error deleting category: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting category"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}

func (c *CategoryController) DeleteSubCategory(ctx *gin.Context) {
	userID := ctx.GetInt64("authUserID")
	subCategoryID := ctx.Param("subCategoryID")
	logger.Info("Attempting to delete sub-category with ID: ", subCategoryID, " for user: ", userID)

	err := c.categoryService.DeleteSubCategory(ctx, subCategoryID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "sub-category not found") {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Sub-category not found"})
			return
		}
		logger.Error("Error deleting sub-category: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting sub-category"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Sub-category deleted successfully"})
}
