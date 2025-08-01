package repository

import (
	"errors"
	"expenses/internal/config"
	"expenses/internal/database/helper"
	database "expenses/internal/database/manager"
	errorsPkg "expenses/internal/errors"
	"expenses/internal/models"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type RuleRepositoryInterface interface {
	CreateRule(c *gin.Context, rule models.CreateBaseRuleRequest) (models.RuleResponse, error)
	CreateRuleActions(c *gin.Context, actions []models.CreateRuleActionRequest) ([]models.RuleActionResponse, error)
	CreateRuleConditions(c *gin.Context, conditions []models.CreateRuleConditionRequest) ([]models.RuleConditionResponse, error)
	CreateRuleTransactionMapping(c *gin.Context, ruleId int64, transactionId int64) error
	GetRule(c *gin.Context, id int64, userId int64) (models.RuleResponse, error)
	ListRules(c *gin.Context, userId int64) ([]models.RuleResponse, error)
	ListRuleActionsByRuleId(c *gin.Context, ruleId int64) ([]models.RuleActionResponse, error)
	ListRuleConditionsByRuleId(c *gin.Context, ruleId int64) ([]models.RuleConditionResponse, error)
	UpdateRule(c *gin.Context, id int64, userId int64, rule models.UpdateRuleRequest) (models.RuleResponse, error)
	UpdateRuleAction(c *gin.Context, id int64, ruleId int64, action models.UpdateRuleActionRequest) (models.RuleActionResponse, error)
	UpdateRuleCondition(c *gin.Context, id int64, ruleId int64, condition models.UpdateRuleConditionRequest) (models.RuleConditionResponse, error)
	DeleteRuleActionsByRuleId(c *gin.Context, ruleId int64) error
	DeleteRuleConditionsByRuleId(c *gin.Context, ruleId int64) error
	DeleteRule(c *gin.Context, id int64, userId int64) error
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

func (r *RuleRepository) CreateRule(c *gin.Context, req models.CreateBaseRuleRequest) (models.RuleResponse, error) {
	var rule models.RuleResponse
	query, values, ptrs, err := helper.CreateInsertQuery(&req, &rule, r.ruleTable, r.schema)
	if err != nil {
		return rule, err
	}
	err = r.db.FetchOne(c, query, values...).Scan(ptrs...)
	if err != nil {
		return rule, errorsPkg.NewRuleRepositoryError("failed to create rule", err)
	}
	return rule, nil
}

func (r *RuleRepository) createRuleAction(c *gin.Context, req *models.CreateRuleActionRequest) (models.RuleActionResponse, error) {
	var ruleAction models.RuleActionResponse
	query, values, ptrs, err := helper.CreateInsertQuery(req, &ruleAction, r.ruleActionTable, r.schema)
	if err != nil {
		return ruleAction, err
	}
	err = r.db.FetchOne(c, query, values...).Scan(ptrs...)
	if err != nil {
		return ruleAction, errorsPkg.NewRuleActionInsertError(err)
	}
	return ruleAction, nil
}

func (r *RuleRepository) CreateRuleActions(c *gin.Context, actions []models.CreateRuleActionRequest) ([]models.RuleActionResponse, error) {
	ruleActions := make([]models.RuleActionResponse, 0)
	for _, action := range actions {
		newAction := action
		ruleAction, err := r.createRuleAction(c, &newAction)
		if err != nil {
			return ruleActions, err
		}
		ruleActions = append(ruleActions, ruleAction)
	}
	return ruleActions, nil
}

func (r *RuleRepository) createRuleCondition(c *gin.Context, req *models.CreateRuleConditionRequest) (models.RuleConditionResponse, error) {
	var ruleCondition models.RuleConditionResponse
	query, values, ptrs, err := helper.CreateInsertQuery(req, &ruleCondition, r.ruleConditionTable, r.schema)
	if err != nil {
		return ruleCondition, err
	}
	err = r.db.FetchOne(c, query, values...).Scan(ptrs...)
	if err != nil {
		return ruleCondition, errorsPkg.NewRuleConditionInsertError(err)
	}
	return ruleCondition, nil
}

func (r *RuleRepository) CreateRuleConditions(c *gin.Context, conditions []models.CreateRuleConditionRequest) ([]models.RuleConditionResponse, error) {
	ruleConditions := make([]models.RuleConditionResponse, 0)
	for _, cond := range conditions {
		newCond := cond
		ruleCondition, err := r.createRuleCondition(c, &newCond)
		if err != nil {
			return ruleConditions, err
		}
		ruleConditions = append(ruleConditions, ruleCondition)
	}
	return ruleConditions, nil
}

func (r *RuleRepository) GetRule(c *gin.Context, id int64, userId int64) (models.RuleResponse, error) {
	var rule models.RuleResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&rule)
	if err != nil {
		return rule, err
	}
	query := fmt.Sprintf(`SELECT %s FROM %s.%s WHERE id = $1 AND created_by = $2`, strings.Join(dbFields, ", "), r.schema, r.ruleTable)
	err = r.db.FetchOne(c, query, id, userId).Scan(ptrs...)
	if err != nil {
		if err == pgx.ErrNoRows {
			return rule, errorsPkg.NewRuleNotFoundError(err)
		}
		return rule, errorsPkg.NewRuleRepositoryError("failed to get rule", err)
	}
	return rule, nil
}

func (r *RuleRepository) ListRules(c *gin.Context, userId int64) ([]models.RuleResponse, error) {
	rules := make([]models.RuleResponse, 0)
	var rule models.RuleResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&rule)
	if err != nil {
		return rules, err
	}
	query := fmt.Sprintf(`SELECT %s FROM %s.%s WHERE created_by = $1`, strings.Join(dbFields, ", "), r.schema, r.ruleTable)
	rows, err := r.db.FetchAll(c, query, userId)
	if err != nil {
		return rules, errorsPkg.NewRuleRepositoryError("failed to list rules", err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(ptrs...)
		if err != nil {
			return rules, errorsPkg.NewRuleRepositoryError("failed to scan rule row", err)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (r *RuleRepository) ListRuleActionsByRuleId(c *gin.Context, ruleId int64) ([]models.RuleActionResponse, error) {
	var actions []models.RuleActionResponse
	var action models.RuleActionResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&action)
	if err != nil {
		return actions, err
	}
	query := fmt.Sprintf(`SELECT %s FROM %s.%s WHERE rule_id = $1`, strings.Join(dbFields, ", "), r.schema, r.ruleActionTable)
	rows, err := r.db.FetchAll(c, query, ruleId)
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

func (r *RuleRepository) ListRuleConditionsByRuleId(c *gin.Context, ruleId int64) ([]models.RuleConditionResponse, error) {
	var conditions []models.RuleConditionResponse
	var condition models.RuleConditionResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&condition)
	if err != nil {
		return conditions, err
	}
	query := fmt.Sprintf(`SELECT %s FROM %s.%s WHERE rule_id = $1`, strings.Join(dbFields, ", "), r.schema, r.ruleConditionTable)
	rows, err := r.db.FetchAll(c, query, ruleId)
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

func (r *RuleRepository) UpdateRule(c *gin.Context, id int64, userId int64, rule models.UpdateRuleRequest) (models.RuleResponse, error) {
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
	err = r.db.FetchOne(c, query, argValues...).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ruleResponse, errorsPkg.NewRuleNotFoundError(err)
		}
		return ruleResponse, errorsPkg.NewRuleRepositoryError("failed to update rule", err)
	}
	return ruleResponse, nil
}

func (r *RuleRepository) UpdateRuleAction(c *gin.Context, id int64, ruleId int64, action models.UpdateRuleActionRequest) (models.RuleActionResponse, error) {
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

	err = r.db.FetchOne(c, query, argValues...).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ruleAction, errorsPkg.NewRuleActionNotFoundError(err)
		}
		return ruleAction, errorsPkg.NewRuleRepositoryError("failed to update rule action", err)
	}
	return ruleAction, nil
}

func (r *RuleRepository) UpdateRuleCondition(c *gin.Context, id int64, ruleId int64, condition models.UpdateRuleConditionRequest) (models.RuleConditionResponse, error) {
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

	err = r.db.FetchOne(c, query, argValues...).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ruleCondition, errorsPkg.NewRuleConditionNotFoundError(err)
		}
		return ruleCondition, errorsPkg.NewRuleRepositoryError("failed to update rule condition", err)
	}
	return ruleCondition, nil
}

func (r *RuleRepository) DeleteRuleActionsByRuleId(c *gin.Context, ruleId int64) error {
	query := fmt.Sprintf(`DELETE FROM %s.%s WHERE rule_id = $1`, r.schema, r.ruleActionTable)
	rowsAffected, err := r.db.ExecuteQuery(c, query, ruleId)
	if err != nil {
		return errorsPkg.NewRuleRepositoryError("failed to delete rule actions", err)
	}
	if rowsAffected == 0 {
		return errorsPkg.NewRuleNotFoundError(fmt.Errorf("rule with id %d not found", ruleId))
	}
	return nil
}

func (r *RuleRepository) DeleteRuleConditionsByRuleId(c *gin.Context, ruleId int64) error {
	query := fmt.Sprintf(`DELETE FROM %s.%s WHERE rule_id = $1`, r.schema, r.ruleConditionTable)
	rowsAffected, err := r.db.ExecuteQuery(c, query, ruleId)
	if err != nil {
		return errorsPkg.NewRuleRepositoryError("failed to delete rule conditions", err)
	}
	if rowsAffected == 0 {
		return errorsPkg.NewRuleNotFoundError(fmt.Errorf("rule with id %d not found", ruleId))
	}
	return nil
}

func (r *RuleRepository) DeleteRule(c *gin.Context, id int64, userId int64) error {
	query := fmt.Sprintf(`DELETE FROM %s.%s WHERE id = $1 AND created_by = $2`, r.schema, r.ruleTable)
	rowsAffected, err := r.db.ExecuteQuery(c, query, id, userId)
	if err != nil {
		return errorsPkg.NewRuleRepositoryError("failed to delete rule", err)
	}
	if rowsAffected == 0 {
		return errorsPkg.NewRuleNotFoundError(fmt.Errorf("rule with id %d not found", id))
	}
	return nil
}

func (r *RuleRepository) CreateRuleTransactionMapping(c *gin.Context, ruleId int64, transactionId int64) error {
	query := fmt.Sprintf(`
		INSERT INTO %s.%s (rule_id, transaction_id) 
		VALUES ($1, $2) 
		ON CONFLICT (rule_id, transaction_id) DO NOTHING
	`, r.schema, r.ruleTransactionMappingTable)

	_, err := r.db.ExecuteQuery(c, query, ruleId, transactionId)
	if err != nil {
		return errorsPkg.NewRuleRepositoryError("failed to create rule transaction mapping", err)
	}
	return nil
}
