package services

import (
	"expenses/models"
	"expenses/utils"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatisticsService struct {
	db     *pgxpool.Pool
	schema string
}

func NewStatisticsService(db *pgxpool.Pool) *StatisticsService {
	return &StatisticsService{
		db:     db,
		schema: utils.GetPGSchema(), //unable to load as this is not inited anywhere in main, thus doesnt have access to env
	}
}

func (e *StatisticsService) GetExpensesBySubcategory(c *gin.Context, userID int64, startTime, endTime time.Time) ([]models.SubcategoryExpenseBreakdown, error) {
	query := fmt.Sprintf(`
        SELECT 
            c.name as category_name,
            c.color as category_color,
            s.name as subcategory_name,
            s.color as subcategory_color,
            SUM(e.amount) as total_amount,
            COUNT(e.id) as transaction_count
        FROM %[1]s.expense e
        JOIN %[1]s.subcategory_expense_mapping sem ON e.id = sem.expense_id
        JOIN %[1]s.subcategories s ON sem.subcategory_id = s.id
        JOIN %[1]s.category_subcategory_mapping csm ON s.id = csm.subcategory_id
        JOIN %[1]s.categories c ON csm.category_id = c.id
        JOIN %[1]s.expense_user_mapping eum ON e.id = eum.expense_id
        WHERE eum.user_id = $1
        AND e.created_at BETWEEN $2 AND $3
        GROUP BY c.name, c.color, s.name, s.color
        ORDER BY c.name, total_amount DESC;
    `, e.schema)

	rows, err := e.db.Query(c, query, userID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var breakdown []models.SubcategoryExpenseBreakdown
	for rows.Next() {
		var item models.SubcategoryExpenseBreakdown
		err := rows.Scan(
			&item.CategoryName,
			&item.CategoryColor,
			&item.SubcategoryName,
			&item.SubcategoryColor,
			&item.TotalAmount,
			&item.TransactionCount,
		)
		if err != nil {
			return nil, err
		}
		breakdown = append(breakdown, item)
	}
	return breakdown, nil
}
