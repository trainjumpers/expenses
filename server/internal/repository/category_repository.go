package repository

import (
	"errors"
	"expenses/internal/config"
	"expenses/internal/database/helper"
	database "expenses/internal/database/postgres"
	customErrors "expenses/internal/errors"
	"expenses/internal/models"
	"expenses/pkg/logger"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepositoryInterface interface {
	CreateCategory(c *gin.Context, input models.CreateCategoryInput) (models.CategoryResponse, error)
	GetCategoryById(c *gin.Context, categoryId int64, userId int64) (models.CategoryResponse, error)
	ListCategories(c *gin.Context, userId int64) ([]models.CategoryResponse, error)
	UpdateCategory(c *gin.Context, categoryId int64, userId int64, input models.UpdateCategoryInput) (models.CategoryResponse, error)
	DeleteCategory(c *gin.Context, categoryId int64, userId int64) error
}

type CategoryRepository struct {
	db        *pgxpool.Pool
	schema    string
	tableName string
}

func NewCategoryRepository(db *database.DatabaseManager, cfg *config.Config) *CategoryRepository {
	return &CategoryRepository{
		db:        db.GetPool(),
		schema:    cfg.DBSchema,
		tableName: "categories",
	}
}

func (r *CategoryRepository) CreateCategory(c *gin.Context, input models.CreateCategoryInput) (models.CategoryResponse, error) {
	logger.Debug("CategoryRepository: Creating category for user", input.CreatedBy)

	var category models.CategoryResponse
	query, values, ptrs, err := helper.CreateInsertQuery(&input, &category, r.tableName, r.schema)
	if err != nil {
		return models.CategoryResponse{}, err
	}

	logger.Info("Executing query to create category: ", query)
	err = r.db.QueryRow(c, query, values...).Scan(ptrs...)
	if err != nil {
		if customErrors.CheckForeignKey(err, "unique_category_name_created_by") {
			return models.CategoryResponse{}, customErrors.NewCategoryAlreadyExistsError(err)
		}
		return models.CategoryResponse{}, err
	}

	logger.Debug("CategoryRepository: Successfully created category with ID:", category.Id)
	return category, nil
}

func (r *CategoryRepository) GetCategoryById(c *gin.Context, categoryId int64, userId int64) (models.CategoryResponse, error) {
	logger.Debug("CategoryRepository: Getting category by id:", categoryId, "for user:", userId)

	var category models.CategoryResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&category)
	if err != nil {
		return category, err
	}

	query := fmt.Sprintf(`SELECT %s FROM %s.%s WHERE id = $1 AND created_by = $2`, strings.Join(dbFields, ", "), r.schema, r.tableName)
	logger.Info("Executing query to get category by id: ", query)

	err = r.db.QueryRow(c, query, categoryId, userId).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return category, customErrors.NewCategoryNotFoundError(err)
		}
		return category, err
	}

	logger.Debug("CategoryRepository: Successfully retrieved category:", category.Name)
	return category, nil
}

func (r *CategoryRepository) ListCategories(c *gin.Context, userId int64) ([]models.CategoryResponse, error) {
	logger.Debug("CategoryRepository: Listing categories for user:", userId)

	_, dbFields, err := helper.GetDbFieldsFromObject(&models.CategoryResponse{})
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`SELECT %s FROM %s.%s WHERE created_by = $1 ORDER BY id DESC;`, strings.Join(dbFields, ", "), r.schema, r.tableName)
	logger.Info("Executing query to list categories: ", query)

	rows, err := r.db.Query(c, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.CategoryResponse
	for rows.Next() {
		var category models.CategoryResponse
		ptrs, _, err := helper.GetDbFieldsFromObject(&category)
		if err != nil {
			return nil, err
		}
		err = rows.Scan(ptrs...)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	logger.Debug("CategoryRepository: Successfully listed", len(categories), "categories for user:", userId)
	return categories, nil
}

func (r *CategoryRepository) UpdateCategory(c *gin.Context, categoryId int64, userId int64, input models.UpdateCategoryInput) (models.CategoryResponse, error) {
	logger.Debug("CategoryRepository: Updating category id:", categoryId, "for user:", userId)

	fieldsClause, argValues, argIndex, err := helper.CreateUpdateParams(&input)
	if err != nil {
		return models.CategoryResponse{}, err
	}

	var category models.CategoryResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&category)
	if err != nil {
		return category, err
	}

	query := fmt.Sprintf(`UPDATE %s.%s SET %s WHERE id = $%d AND created_by = $%d RETURNING %s;`, r.schema, r.tableName, fieldsClause, argIndex, argIndex+1, strings.Join(dbFields, ", "))
	logger.Info("Executing query to update category: ", query)

	argValues = append(argValues, categoryId, userId)
	err = r.db.QueryRow(c, query, argValues...).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return category, customErrors.NewCategoryNotFoundError(err)
		}
		if customErrors.CheckForeignKey(err, "unique_category_name_created_by") {
			return category, customErrors.NewCategoryAlreadyExistsError(err)
		}
		return category, err
	}

	logger.Debug("CategoryRepository: Successfully updated category:", category.Name)
	return category, nil
}

func (r *CategoryRepository) DeleteCategory(c *gin.Context, categoryId int64, userId int64) error {
	logger.Debug("CategoryRepository: Deleting category id:", categoryId, "for user:", userId)

	query := fmt.Sprintf(`DELETE FROM %s.%s WHERE id = $1 AND created_by = $2;`, r.schema, r.tableName)
	logger.Info("Executing query to delete category: ", query)

	result, err := r.db.Exec(c, query, categoryId, userId)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return customErrors.NewCategoryNotFoundError(fmt.Errorf("category with id %d not found", categoryId))
	}

	logger.Debug("CategoryRepository: Successfully deleted category id:", categoryId)
	return nil
}
