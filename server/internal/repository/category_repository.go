package repository

import (
	"errors"
	"expenses/internal/config"
	"expenses/internal/database/helper"
	database "expenses/internal/database/manager"
	customErrors "expenses/internal/errors"
	"expenses/internal/models"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type CategoryRepositoryInterface interface {
	CreateCategory(c *gin.Context, input models.CreateCategoryInput) (models.CategoryResponse, error)
	GetCategoryById(c *gin.Context, categoryId int64, userId int64) (models.CategoryResponse, error)
	ListCategories(c *gin.Context, userId int64) ([]models.CategoryResponse, error)
	UpdateCategory(c *gin.Context, categoryId int64, userId int64, input models.UpdateCategoryInput) (models.CategoryResponse, error)
	DeleteCategory(c *gin.Context, categoryId int64, userId int64) error
}

type CategoryRepository struct {
	db        database.DatabaseManager
	schema    string
	tableName string
}

func NewCategoryRepository(db database.DatabaseManager, cfg *config.Config) CategoryRepositoryInterface {
	return &CategoryRepository{
		db:        db,
		schema:    cfg.DBSchema,
		tableName: "categories",
	}
}

func (r *CategoryRepository) CreateCategory(c *gin.Context, input models.CreateCategoryInput) (models.CategoryResponse, error) {
	var category models.CategoryResponse
	query, values, ptrs, err := helper.CreateInsertQuery(&input, &category, r.tableName, r.schema)
	if err != nil {
		return models.CategoryResponse{}, err
	}

	err = r.db.FetchOne(c, query, values...).Scan(ptrs...)
	if err != nil {
		if customErrors.CheckForeignKey(err, "unique_category_name_created_by") {
			return models.CategoryResponse{}, customErrors.NewCategoryAlreadyExistsError(err)
		}
		return models.CategoryResponse{}, err
	}

	return category, nil
}

func (r *CategoryRepository) GetCategoryById(c *gin.Context, categoryId int64, userId int64) (models.CategoryResponse, error) {
	var category models.CategoryResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&category)
	if err != nil {
		return category, err
	}

	query := fmt.Sprintf(`SELECT %s FROM %s.%s WHERE id = $1 AND created_by = $2`, strings.Join(dbFields, ", "), r.schema, r.tableName)

	err = r.db.FetchOne(c, query, categoryId, userId).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return category, customErrors.NewCategoryNotFoundError(err)
		}
		return category, err
	}

	return category, nil
}

func (r *CategoryRepository) ListCategories(c *gin.Context, userId int64) ([]models.CategoryResponse, error) {
	categories := make([]models.CategoryResponse, 0)
	var category models.CategoryResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&category)
	if err != nil {
		return categories, err
	}

	query := fmt.Sprintf(`SELECT %s FROM %s.%s WHERE created_by = $1 ORDER BY id DESC;`, strings.Join(dbFields, ", "), r.schema, r.tableName)

	rows, err := r.db.FetchAll(c, query, userId)
	if err != nil {
		return categories, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(ptrs...)
		if err != nil {
			return categories, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *CategoryRepository) UpdateCategory(c *gin.Context, categoryId int64, userId int64, input models.UpdateCategoryInput) (models.CategoryResponse, error) {
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

	argValues = append(argValues, categoryId, userId)
	err = r.db.FetchOne(c, query, argValues...).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return category, customErrors.NewCategoryNotFoundError(err)
		}
		if customErrors.CheckForeignKey(err, "unique_category_name_created_by") {
			return category, customErrors.NewCategoryAlreadyExistsError(err)
		}
		return category, err
	}

	return category, nil
}

func (r *CategoryRepository) DeleteCategory(c *gin.Context, categoryId int64, userId int64) error {
	query := fmt.Sprintf(`DELETE FROM %s.%s WHERE id = $1 AND created_by = $2;`, r.schema, r.tableName)

	rowsAffected, err := r.db.ExecuteQuery(c, query, categoryId, userId)
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return customErrors.NewCategoryNotFoundError(fmt.Errorf("category with id %d not found", categoryId))
	}

	return nil
}
