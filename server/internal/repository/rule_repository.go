package repository

import (
	"context"
	"errors"
	"expenses/internal/config"
	"expenses/internal/database/helper"
	errorsPkg "expenses/internal/errors"
	"expenses/internal/models"
	database "expenses/pkg/database/manager"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type RuleRepositoryInterface interface {
	CreateRule(ctx context.Context, rule models.CreateBaseRuleRequest) (models.RuleResponse, error)
	CreateRuleActions(ctx context.Context, actions []models.CreateRuleActionRequest) ([]models.RuleActionResponse, error)
	CreateRuleConditions(ctx context.Context, conditions []models.CreateRuleConditionRequest) ([]models.RuleConditionResponse, error)
	CreateRuleTransactionMapping(ctx context.Context, ruleId int64, transactionId int64) error
	GetRule(ctx context.Context, id int64, userId int64) (models.RuleResponse, error)
	ListRules(ctx context.Context, userId int64, query models.RuleListQuery) (models.PaginatedRulesResponse, error)
	ListRuleActionsByRuleId(ctx context.Context, ruleId int64) ([]models.RuleActionResponse, error)
	ListRuleConditionsByRuleId(ctx context.Context, ruleId int64) ([]models.RuleConditionResponse, error)
	UpdateRule(ctx context.Context, id int64, userId int64, rule models.UpdateRuleRequest) (models.RuleResponse, error)
	UpdateRuleAction(ctx context.Context, id int64, ruleId int64, action models.UpdateRuleActionRequest) (models.RuleActionResponse, error)
	UpdateRuleCondition(ctx context.Context, id int64, ruleId int64, condition models.UpdateRuleConditionRequest) (models.RuleConditionResponse, error)
	PutRuleActions(ctx context.Context, ruleId int64, actions []models.CreateRuleActionRequest) ([]models.RuleActionResponse, error)
	PutRuleConditions(ctx context.Context, ruleId int64, conditions []models.CreateRuleConditionRequest) ([]models.RuleConditionResponse, error)
	DeleteRuleActionsByRuleId(ctx context.Context, ruleId int64) error
	DeleteRuleConditionsByRuleId(ctx context.Context, ruleId int64) error
	DeleteRule(ctx context.Context, id int64, userId int64) error
}

type RuleRepository struct {
	db                          database.DatabaseManager
	schema                      string
	ruleTable                   string
	ruleActionTable             string
	ruleConditionTable          string
	ruleTransactionMappingTable string
}

func NewRuleRepository(db database.DatabaseManager, cfg *config.Config) RuleRepositoryInterface {
	return &RuleRepository{
		db:                          db,
		schema:                      cfg.DBSchema,
		ruleTable:                   "rule",
		ruleActionTable:             "rule_action",
		ruleConditionTable:          "rule_condition",
		ruleTransactionMappingTable: "rule_transaction_mapping",
	}
}

func (r *RuleRepository) CreateRule(ctx context.Context, req models.CreateBaseRuleRequest) (models.RuleResponse, error) {
	var rule models.RuleResponse
	query, values, ptrs, err := helper.CreateInsertQuery(&req, &rule, r.ruleTable, r.schema)
	if err != nil {
		return rule, err
	}
	err = r.db.FetchOne(ctx, query, values...).Scan(ptrs...)
	if err != nil {
		return rule, errorsPkg.NewRuleRepositoryError("failed to create rule", err)
	}
	return rule, nil
}

func (r *RuleRepository) createRuleAction(ctx context.Context, req *models.CreateRuleActionRequest) (models.RuleActionResponse, error) {
	var ruleAction models.RuleActionResponse
	query, values, ptrs, err := helper.CreateInsertQuery(req, &ruleAction, r.ruleActionTable, r.schema)
	if err != nil {
		return ruleAction, err
	}
	err = r.db.FetchOne(ctx, query, values...).Scan(ptrs...)
	if err != nil {
		return ruleAction, errorsPkg.NewRuleActionInsertError(err)
	}
	return ruleAction, nil
}

func (r *RuleRepository) CreateRuleActions(ctx context.Context, actions []models.CreateRuleActionRequest) ([]models.RuleActionResponse, error) {
	ruleActions := make([]models.RuleActionResponse, 0)
	for _, action := range actions {
		newAction := action
		ruleAction, err := r.createRuleAction(ctx, &newAction)
		if err != nil {
			return ruleActions, err
		}
		ruleActions = append(ruleActions, ruleAction)
	}
	return ruleActions, nil
}

func (r *RuleRepository) createRuleCondition(ctx context.Context, req *models.CreateRuleConditionRequest) (models.RuleConditionResponse, error) {
	var ruleCondition models.RuleConditionResponse
	query, values, ptrs, err := helper.CreateInsertQuery(req, &ruleCondition, r.ruleConditionTable, r.schema)
	if err != nil {
		return ruleCondition, err
	}
	err = r.db.FetchOne(ctx, query, values...).Scan(ptrs...)
	if err != nil {
		return ruleCondition, errorsPkg.NewRuleConditionInsertError(err)
	}
	return ruleCondition, nil
}

func (r *RuleRepository) CreateRuleConditions(ctx context.Context, conditions []models.CreateRuleConditionRequest) ([]models.RuleConditionResponse, error) {
	ruleConditions := make([]models.RuleConditionResponse, 0)
	for _, cond := range conditions {
		newCond := cond
		ruleCondition, err := r.createRuleCondition(ctx, &newCond)
		if err != nil {
			return ruleConditions, err
		}
		ruleConditions = append(ruleConditions, ruleCondition)
	}
	return ruleConditions, nil
}

func (r *RuleRepository) GetRule(ctx context.Context, id int64, userId int64) (models.RuleResponse, error) {
	var rule models.RuleResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&rule)
	if err != nil {
		return rule, err
	}
	query := fmt.Sprintf(`SELECT %s FROM %s.%s WHERE id = $1 AND created_by = $2`, strings.Join(dbFields, ", "), r.schema, r.ruleTable)
	err = r.db.FetchOne(ctx, query, id, userId).Scan(ptrs...)
	if err != nil {
		if err == pgx.ErrNoRows {
			return rule, errorsPkg.NewRuleNotFoundError(err)
		}
		return rule, errorsPkg.NewRuleRepositoryError("failed to get rule", err)
	}
	return rule, nil
}

func (r *RuleRepository) ListRules(ctx context.Context, userId int64, query models.RuleListQuery) (models.PaginatedRulesResponse, error) {
	var response models.PaginatedRulesResponse
	response.Rules = make([]models.RuleResponse, 0)
	response.Page = query.Page
	response.PageSize = query.PageSize

	var rule models.RuleResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&rule)
	if err != nil {
		return response, err
	}

	// Build WHERE clause
	whereClause := "WHERE created_by = $1"
	args := []interface{}{userId}
	argIndex := 2

	// Add search filter if provided
	if query.Search != nil && *query.Search != "" {
		searchPattern := "%" + *query.Search + "%"
		whereClause += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex)
		args = append(args, searchPattern)
		argIndex++
	}

	// Count total records for pagination
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s.%s %s`, r.schema, r.ruleTable, whereClause)
	err = r.db.FetchOne(ctx, countQuery, args...).Scan(&response.Total)
	if err != nil {
		return response, errorsPkg.NewRuleRepositoryError("failed to count rules", err)
	}

	// Build the main query with pagination (if page_size > 0)
	var mainQuery string
	if query.PageSize > 0 {
		// Calculate offset for pagination
		offset := (query.Page - 1) * query.PageSize

		// Build query with pagination
		mainQuery = fmt.Sprintf(`
			SELECT %s 
			FROM %s.%s 
			%s 
			ORDER BY id DESC 
			LIMIT $%d OFFSET $%d`,
			strings.Join(dbFields, ", "), r.schema, r.ruleTable, whereClause, argIndex, argIndex+1)

		args = append(args, query.PageSize, offset)
	} else {
		// If no page_size specified, return all results
		mainQuery = fmt.Sprintf(`
			SELECT %s 
			FROM %s.%s 
			%s 
			ORDER BY id DESC`,
			strings.Join(dbFields, ", "), r.schema, r.ruleTable, whereClause)
	}

	rows, err := r.db.FetchAll(ctx, mainQuery, args...)
	if err != nil {
		return response, errorsPkg.NewRuleRepositoryError("failed to list rules", err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(ptrs...)
		if err != nil {
			return response, errorsPkg.NewRuleRepositoryError("failed to scan rule row", err)
		}
		response.Rules = append(response.Rules, rule)
	}

	return response, nil
}

func (r *RuleRepository) ListRuleActionsByRuleId(ctx context.Context, ruleId int64) ([]models.RuleActionResponse, error) {
	var actions []models.RuleActionResponse
	var action models.RuleActionResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&action)
	if err != nil {
		return actions, err
	}
	query := fmt.Sprintf(`SELECT %s FROM %s.%s WHERE rule_id = $1`, strings.Join(dbFields, ", "), r.schema, r.ruleActionTable)
	rows, err := r.db.FetchAll(ctx, query, ruleId)
	if err != nil {
		return actions, errorsPkg.NewRuleRepositoryError("failed to list rule actions", err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(ptrs...)
		if err != nil {
			return actions, errorsPkg.NewRuleRepositoryError("failed to scan rule action row", err)
		}
		actions = append(actions, action)
	}
	return actions, nil
}

func (r *RuleRepository) ListRuleConditionsByRuleId(ctx context.Context, ruleId int64) ([]models.RuleConditionResponse, error) {
	var conditions []models.RuleConditionResponse
	var condition models.RuleConditionResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&condition)
	if err != nil {
		return conditions, err
	}
	query := fmt.Sprintf(`SELECT %s FROM %s.%s WHERE rule_id = $1`, strings.Join(dbFields, ", "), r.schema, r.ruleConditionTable)
	rows, err := r.db.FetchAll(ctx, query, ruleId)
	if err != nil {
		return conditions, errorsPkg.NewRuleRepositoryError("failed to list rule conditions", err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(ptrs...)
		if err != nil {
			return conditions, errorsPkg.NewRuleRepositoryError("failed to scan rule condition row", err)
		}
		conditions = append(conditions, condition)
	}
	return conditions, nil
}

func (r *RuleRepository) UpdateRule(ctx context.Context, id int64, userId int64, rule models.UpdateRuleRequest) (models.RuleResponse, error) {
	var ruleResponse models.RuleResponse
	fieldsClause, argValues, argIndex, err := helper.CreateUpdateParams(&rule)
	if err != nil {
		return ruleResponse, err
	}
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&ruleResponse)
	if err != nil {
		return ruleResponse, err
	}
	argValues = append(argValues, id, userId)
	query := fmt.Sprintf(`UPDATE %s.%s SET %s WHERE id = $%d AND created_by = $%d RETURNING %s;`, r.schema, r.ruleTable, fieldsClause, argIndex, argIndex+1, strings.Join(dbFields, ", "))
	err = r.db.FetchOne(ctx, query, argValues...).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ruleResponse, errorsPkg.NewRuleNotFoundError(err)
		}
		return ruleResponse, errorsPkg.NewRuleRepositoryError("failed to update rule", err)
	}
	return ruleResponse, nil
}

func (r *RuleRepository) UpdateRuleAction(ctx context.Context, id int64, ruleId int64, action models.UpdateRuleActionRequest) (models.RuleActionResponse, error) {
	var ruleAction models.RuleActionResponse
	fieldsClause, argValues, argIndex, err := helper.CreateUpdateParams(&action)
	if err != nil {
		return ruleAction, err
	}
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&ruleAction)
	if err != nil {
		return ruleAction, err
	}
	argValues = append(argValues, id, ruleId)
	query := fmt.Sprintf(`
		UPDATE %s.%s SET %s WHERE id = $%d AND rule_id = $%d RETURNING %s;
    `, r.schema, r.ruleActionTable, fieldsClause, argIndex, argIndex+1, strings.Join(dbFields, ", "))

	err = r.db.FetchOne(ctx, query, argValues...).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ruleAction, errorsPkg.NewRuleActionNotFoundError(err)
		}
		return ruleAction, errorsPkg.NewRuleRepositoryError("failed to update rule action", err)
	}
	return ruleAction, nil
}

func (r *RuleRepository) UpdateRuleCondition(ctx context.Context, id int64, ruleId int64, condition models.UpdateRuleConditionRequest) (models.RuleConditionResponse, error) {
	var ruleCondition models.RuleConditionResponse
	fieldsClause, argValues, argIndex, err := helper.CreateUpdateParams(&condition)
	if err != nil {
		return ruleCondition, err
	}
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&ruleCondition)
	if err != nil {
		return ruleCondition, err
	}
	argValues = append(argValues, id, ruleId)
	query := fmt.Sprintf(`
		UPDATE %s.%s SET %s WHERE id = $%d AND rule_id = $%d RETURNING %s;
    `, r.schema, r.ruleConditionTable, fieldsClause, argIndex, argIndex+1, strings.Join(dbFields, ", "))

	err = r.db.FetchOne(ctx, query, argValues...).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ruleCondition, errorsPkg.NewRuleConditionNotFoundError(err)
		}
		return ruleCondition, errorsPkg.NewRuleRepositoryError("failed to update rule condition", err)
	}
	return ruleCondition, nil
}

func (r *RuleRepository) DeleteRuleActionsByRuleId(ctx context.Context, ruleId int64) error {
	query := fmt.Sprintf(`DELETE FROM %s.%s WHERE rule_id = $1`, r.schema, r.ruleActionTable)
	rowsAffected, err := r.db.ExecuteQuery(ctx, query, ruleId)
	if err != nil {
		return errorsPkg.NewRuleRepositoryError("failed to delete rule actions", err)
	}
	if rowsAffected == 0 {
		return errorsPkg.NewRuleNotFoundError(fmt.Errorf("rule with id %d not found", ruleId))
	}
	return nil
}

func (r *RuleRepository) DeleteRuleConditionsByRuleId(ctx context.Context, ruleId int64) error {
	query := fmt.Sprintf(`DELETE FROM %s.%s WHERE rule_id = $1`, r.schema, r.ruleConditionTable)
	rowsAffected, err := r.db.ExecuteQuery(ctx, query, ruleId)
	if err != nil {
		return errorsPkg.NewRuleRepositoryError("failed to delete rule conditions", err)
	}
	if rowsAffected == 0 {
		return errorsPkg.NewRuleNotFoundError(fmt.Errorf("rule with id %d not found", ruleId))
	}
	return nil
}

func (r *RuleRepository) DeleteRule(ctx context.Context, id int64, userId int64) error {
	query := fmt.Sprintf(`DELETE FROM %s.%s WHERE id = $1 AND created_by = $2`, r.schema, r.ruleTable)
	rowsAffected, err := r.db.ExecuteQuery(ctx, query, id, userId)
	if err != nil {
		return errorsPkg.NewRuleRepositoryError("failed to delete rule", err)
	}
	if rowsAffected == 0 {
		return errorsPkg.NewRuleNotFoundError(fmt.Errorf("rule with id %d not found", id))
	}
	return nil
}

func (r *RuleRepository) PutRuleActions(ctx context.Context, ruleId int64, actions []models.CreateRuleActionRequest) ([]models.RuleActionResponse, error) {
	var result []models.RuleActionResponse
	err := r.db.WithLock(ctx, ruleId, func(ctx context.Context) error {
		deleteQuery := fmt.Sprintf(`DELETE FROM %s.%s WHERE rule_id = $1`, r.schema, r.ruleActionTable)
		_, err := r.db.ExecuteQuery(ctx, deleteQuery, ruleId)
		if err != nil {
			return errorsPkg.NewRuleRepositoryError("failed to delete existing rule actions", err)
		}
		result = make([]models.RuleActionResponse, 0, len(actions))
		for _, action := range actions {
			action.RuleId = ruleId
			ruleAction, err := r.createRuleAction(ctx, &action)
			if err != nil {
				return err
			}
			result = append(result, ruleAction)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *RuleRepository) PutRuleConditions(ctx context.Context, ruleId int64, conditions []models.CreateRuleConditionRequest) ([]models.RuleConditionResponse, error) {
	var result []models.RuleConditionResponse
	err := r.db.WithLock(ctx, ruleId, func(ctx context.Context) error {
		deleteQuery := fmt.Sprintf(`DELETE FROM %s.%s WHERE rule_id = $1`, r.schema, r.ruleConditionTable)
		_, err := r.db.ExecuteQuery(ctx, deleteQuery, ruleId)
		if err != nil {
			return errorsPkg.NewRuleRepositoryError("failed to delete existing rule conditions", err)
		}

		result = make([]models.RuleConditionResponse, 0, len(conditions))
		for _, condition := range conditions {
			condition.RuleId = ruleId
			ruleCondition, err := r.createRuleCondition(ctx, &condition)
			if err != nil {
				return err
			}
			result = append(result, ruleCondition)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *RuleRepository) CreateRuleTransactionMapping(ctx context.Context, ruleId int64, transactionId int64) error {
	query := fmt.Sprintf(`
		INSERT INTO %s.%s (rule_id, transaction_id)
		VALUES ($1, $2)
		ON CONFLICT (rule_id, transaction_id) DO NOTHING
	`, r.schema, r.ruleTransactionMappingTable)

	_, err := r.db.ExecuteQuery(ctx, query, ruleId, transactionId)
	if err != nil {
		return errorsPkg.NewRuleRepositoryError("failed to create rule transaction mapping", err)
	}
	return nil
}
