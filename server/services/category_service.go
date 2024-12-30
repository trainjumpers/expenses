package services

import (
	"expenses/entities"
	"expenses/logger"
	"expenses/models"
	"expenses/utils"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryService struct {
	db     *pgxpool.Pool
	schema string
}

func NewCategoryService(db *pgxpool.Pool) *CategoryService {
	return &CategoryService{
		db:     db,
		schema: utils.GetPGSchema(),
	}
}

func (c *CategoryService) CreateCategory(ctx *gin.Context, name string, color string, userID int64) (models.Category, error) {
	query := fmt.Sprintf(`
		INSERT INTO %s.categories (name, color, created_by)
		VALUES ($1, $2, $3)
		RETURNING id, name, color, created_by, created_at`, c.schema)

	logger.Info("Executing query to create category: ", query)
	var category models.Category
	err := c.db.QueryRow(ctx, query, name, color, userID).Scan(
		&category.ID,
		&category.Name,
		&category.Color,
		&category.CreatedBy,
		&category.CreatedAt,
	)
	if err != nil {
		logger.Error("Error creating category: ", err)
		return category, err
	}
	logger.Info("Successfully created category with ID: ", category.ID)
	return category, nil
}

func (c *CategoryService) CreateSubCategory(ctx *gin.Context, input models.SubCategory, categoryID int64, userID int64) (models.SubCategory, error) {
	query := fmt.Sprintf(`
		WITH inserted_subcategory AS (
			INSERT INTO %[1]s.subcategories (name, color, created_by)
			VALUES ($1, $2, $4)
			RETURNING id, name, color, created_by, created_at
		), new_category_subcategory_mapping AS (
			INSERT INTO %[1]s.category_subcategory_mapping (category_id, subcategory_id)
			VALUES ($3, (SELECT id FROM inserted_subcategory))
			RETURNING category_id, subcategory_id
		) SELECT * FROM inserted_subcategory;`, c.schema)
	logger.Info("Executing query to create sub-category: ", query)
	var subCategory models.SubCategory
	err := c.db.QueryRow(ctx, query, input.Name, input.Color, categoryID, userID).Scan(
		&subCategory.ID,
		&subCategory.Name,
		&subCategory.Color,
		&subCategory.CreatedBy,
		&subCategory.CreatedAt,
	)
	if err != nil {
		logger.Error("Error creating sub-category: ", err)
		return subCategory, err
	}
	logger.Info("Successfully created sub-category with ID: ", subCategory.ID)
	return subCategory, nil
}

func (c *CategoryService) GetCategory(ctx *gin.Context, categoryID string, userID int64) (models.CategoryWithSubs, error) {
	query := fmt.Sprintf(`
        SELECT 
            c.id, c.name, c.color,
            COALESCE(sc.id, 0), COALESCE(sc.name, ''), COALESCE(sc.color, '')
        FROM %[1]s.categories c
		LEFT JOIN %[1]s.category_subcategory_mapping csm ON c.id = csm.category_id
        LEFT JOIN %[1]s.subcategories sc ON csm.subcategory_id = sc.id
        WHERE c.id = $1 AND c.created_by = $2;
    `, c.schema)

	logger.Info("Executing query to get category: ", query)
	rows, err := c.db.Query(ctx, query, categoryID, userID)
	if err != nil {
		logger.Error("Error querying category: ", err)
		return models.CategoryWithSubs{}, err
	}
	defer rows.Close()

	var category models.CategoryWithSubs
	var subCategories []models.SubCategory

	for rows.Next() {
		var sc models.SubCategory
		err := rows.Scan(
			&category.ID, &category.Name, &category.Color,
			&sc.ID, &sc.Name, &sc.Color)
		if err != nil {
			logger.Error("Error scanning category row: ", err)
			return models.CategoryWithSubs{}, err
		}
		if sc.ID != 0 {
			subCategories = append(subCategories, sc)
		}
	}
	if subCategories == nil {
		category.SubCategories = []models.SubCategory{}
	} else {
		category.SubCategories = subCategories
	}
	logger.Info("Successfully retrieved category with ID: ", category.ID)
	return category, nil
}

func (c *CategoryService) GetAllCategories(ctx *gin.Context, userID int64) ([]models.CategoryWithSubs, error) {
	query := fmt.Sprintf(`
        SELECT 
            c.id, c.name, c.color,
            COALESCE(sc.id, 0), COALESCE(sc.name, ''), COALESCE(sc.color, '')
        FROM %[1]s.categories c
		LEFT JOIN %[1]s.category_subcategory_mapping csm ON c.id = csm.category_id
        LEFT JOIN %[1]s.subcategories sc ON csm.subcategory_id = sc.id
        WHERE c.created_by = $1
        ORDER BY c.id, sc.id
    `, c.schema)

	logger.Info("Executing query to get all categories: ", query)
	rows, err := c.db.Query(ctx, query, userID)
	if err != nil {
		logger.Error("Error querying all categories: ", err)
		return nil, err
	}
	defer rows.Close()

	categoriesMap := make(map[int64]*models.CategoryWithSubs)
	var categories []models.CategoryWithSubs

	for rows.Next() {
		var category models.CategoryWithSubs
		var sc models.SubCategory
		err := rows.Scan(
			&category.ID, &category.Name, &category.Color,
			&sc.ID, &sc.Name, &sc.Color)
		if err != nil {
			logger.Error("Error scanning categories row: ", err)
			return nil, err
		}

		if existingCategory, ok := categoriesMap[category.ID]; ok {
			if sc.ID != 0 {
				existingCategory.SubCategories = append(existingCategory.SubCategories, sc)
			}
		} else {
			if sc.ID != 0 {
				category.SubCategories = []models.SubCategory{sc}
			} else {
				category.SubCategories = []models.SubCategory{}
			}
			categoriesMap[category.ID] = &category
		}
	}

	for _, category := range categoriesMap {
		categories = append(categories, *category)
	}

	logger.Info("Successfully retrieved all categories")
	return categories, nil
}

func (c *CategoryService) UpdateCategory(ctx *gin.Context, categoryID string, input entities.UpdateCategoryInput, userID int64) (models.Category, error) {
	fields := map[string]interface{}{
		"name":  input.Name,
		"color": input.Color,
	}

	fieldsClause, argValues, argIndex, err := utils.CreateUpdateParamsQuery(fields)
	if err != nil {
		logger.Error("Error creating update params query: ", err)
		return models.Category{}, err
	}

	query := fmt.Sprintf("UPDATE %s.categories SET %s WHERE id = $%d AND created_by = $%d "+
		"RETURNING id, name, color, created_by, created_at;",
		c.schema, fieldsClause, argIndex, argIndex+1)

	logger.Info("Executing query to update category: ", query)
	result := c.db.QueryRow(ctx, query, append(argValues, categoryID, userID)...)

	var category models.Category
	err = result.Scan(&category.ID, &category.Name, &category.Color, &category.CreatedBy, &category.CreatedAt)
	if err != nil {
		logger.Error("Error updating category: ", err)
		return category, err
	}
	logger.Info("Successfully updated category with ID: ", category.ID)
	return category, err
}

func (c *CategoryService) UpdateSubCategory(ctx *gin.Context, subCategoryID string, input entities.UpdateSubCategoryInput, userID int64) (models.SubCategory, error) {
	fields := map[string]interface{}{
		"name":  input.Name,
		"color": input.Color,
	}

	fieldsClause, argValues, argIndex, err := utils.CreateUpdateParamsQuery(fields)
	if err != nil {
		logger.Error("Error creating update params query: ", err)
		return models.SubCategory{}, err
	}

	query := fmt.Sprintf("UPDATE %s.subcategories SET %s WHERE id = $%d AND created_by = $%d "+
		"RETURNING id, name, color, created_by, created_at;",
		c.schema, fieldsClause, argIndex, argIndex+1)

	logger.Info("Executing query to update sub-category: ", query)
	result := c.db.QueryRow(ctx, query, append(argValues, subCategoryID, userID)...)

	var subCategory models.SubCategory
	err = result.Scan(&subCategory.ID, &subCategory.Name, &subCategory.Color, &subCategory.CreatedBy, &subCategory.CreatedAt)
	if err != nil {
		logger.Error("Error updating sub-category: ", err)
		return subCategory, err
	}
	logger.Info("Successfully updated sub-category with ID: ", subCategory.ID)
	return subCategory, err
}

func (c *CategoryService) DeleteCategory(ctx *gin.Context, categoryID string, userID int64) error {
	subCategoryQuery := fmt.Sprintf(`
		WITH mapping_ids AS (
			DELETE FROM %[1]s.category_subcategory_mapping
			WHERE category_id = $1
			RETURNING subcategory_id
		)
		DELETE FROM %[1]s.subcategories 
		WHERE id IN (SELECT subcategory_id FROM mapping_ids) 
		AND created_by = $2;
	`, c.schema)
	categoryQuery := fmt.Sprintf(`
		DELETE FROM %s.categories 
		WHERE id = $1 AND created_by = $2
	`, c.schema)

	logger.Info("Acquiring transaction for deleting sub-categories...")
	txn, err := c.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer txn.Rollback(ctx)

	logger.Info("Executing query to delete sub-categories: ", subCategoryQuery)
	_, err = txn.Exec(ctx, subCategoryQuery, categoryID, userID)
	if err != nil {
		return err
	}
	logger.Info("Executing query to delete category: ", categoryQuery)
	_, err = txn.Exec(ctx, categoryQuery, categoryID, userID)
	if err != nil {
		return err
	}
	logger.Info("Successfully deleted category with ID: ", categoryID)
	return txn.Commit(ctx)
}

func (c *CategoryService) DeleteSubCategory(ctx *gin.Context, subCategoryID string, userID int64) error {
	query := fmt.Sprintf(`
        DELETE FROM %s.subcategories 
        WHERE id = $1 AND created_by = $2
    `, c.schema)

	logger.Info("Executing query to delete sub-category: ", query)
	_, err := c.db.Exec(ctx, query, subCategoryID, userID)
	if err != nil {
		logger.Error("Error deleting sub-category: ", err)
		return err
	}
	logger.Info("Successfully deleted sub-category with ID: ", subCategoryID)
	return nil
}
