package repository

import (
	"context"
	"expenses/internal/config"
	"expenses/internal/database/helper"
	database "expenses/internal/database/manager"
	errorsPkg "expenses/internal/errors"
	"expenses/internal/models"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type RuleRepositoryInterface interface {
	CreateRule(ctx context.Context, req *models.CreateRuleRequest) (*models.RuleResponse, error)
	GetRuleByID(ctx context.Context, id int64) (*models.RuleResponse, error)
	ListRules(ctx context.Context) ([]*models.RuleResponse, error)
	UpdateRule(ctx context.Context, id int64, req *models.UpdateRuleRequest) error
	DeleteRule(ctx context.Context, id int64) error
}

type RuleRepository struct {
	db                 database.DatabaseManager
	schema             string
	ruleTable          string
	ruleActionTable    string
	ruleConditionTable string
}

func NewRuleRepository(db database.DatabaseManager, cfg *config.Config) RuleRepositoryInterface {
	return &RuleRepository{
		db:                 db,
		schema:             cfg.DBSchema,
		ruleTable:          "rule",
		ruleActionTable:    "rule_action",
		ruleConditionTable: "rule_condition",
	}
}

func (r *RuleRepository) CreateRule(ctx context.Context, req *models.CreateRuleRequest) (*models.RuleResponse, error) {
	baseRule := models.Rule{
		BaseRule:  req.BaseRule,
		CreatedBy: req.CreatedBy,
	}

	var ruleResp *models.RuleResponse

	err := r.db.WithTxn(ctx, func(tx pgx.Tx) error {
		insertRule := baseRule
		insertRule.ID = 0
		query, values, ptrs, err := helper.CreateInsertQuery(&insertRule, &insertRule, r.ruleTable, r.schema)
		if err != nil {
			return err
		}
		err = tx.QueryRow(ctx, query, values...).Scan(ptrs...)
		if err != nil {
			return err
		}
		ruleID := insertRule.ID

		actions, err := r.insertRuleActions(ctx, tx, req.Actions, ruleID)
		if err != nil {
			return err
		}
		conditions, err := r.insertRuleConditions(ctx, tx, req.Conditions, ruleID)
		if err != nil {
			return err
		}

		// Prepare response
		ruleResp = &models.RuleResponse{
			ID:         ruleID,
			BaseRule:   req.BaseRule,
			CreatedBy:  req.CreatedBy,
			Actions:    actions,
			Conditions: conditions,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ruleResp, nil
}

// insertRuleActions inserts actions for a rule and returns the response objects
func (r *RuleRepository) insertRuleActions(ctx context.Context, tx pgx.Tx, actionsReq []models.CreateRuleActionRequest, ruleID int64) ([]models.RuleActionResponse, error) {
	var actions []models.RuleActionResponse
	for _, a := range actionsReq {
		action := models.RuleAction{
			RuleID:         ruleID,
			BaseRuleAction: a.BaseRuleAction,
		}
		query, values, ptrs, err := helper.CreateInsertQuery(&action, &action, r.ruleActionTable, r.schema)
		if err != nil {
			return nil, errorsPkg.NewRuleRepositoryError("failed to build insert query for rule_action", err)
		}
		err = tx.QueryRow(ctx, query, values...).Scan(ptrs...)
		if err != nil {
			return nil, errorsPkg.NewRuleActionInsertError(err)
		}
		actions = append(actions, models.RuleActionResponse{
			ID:             action.ID,
			RuleID:         action.RuleID,
			BaseRuleAction: action.BaseRuleAction,
		})
	}
	return actions, nil
}

// insertRuleConditions inserts conditions for a rule and returns the response objects
func (r *RuleRepository) insertRuleConditions(ctx context.Context, tx pgx.Tx, condsReq []models.CreateRuleConditionRequest, ruleID int64) ([]models.RuleConditionResponse, error) {
	var conditions []models.RuleConditionResponse
	for _, c := range condsReq {
		cond := models.RuleCondition{
			RuleID:            ruleID,
			BaseRuleCondition: c.BaseRuleCondition,
		}
		query, values, ptrs, err := helper.CreateInsertQuery(&cond, &cond, r.ruleConditionTable, r.schema)
		if err != nil {
			return nil, errorsPkg.NewRuleRepositoryError("failed to build insert query for rule_condition", err)
		}
		err = tx.QueryRow(ctx, query, values...).Scan(ptrs...)
		if err != nil {
			return nil, errorsPkg.NewRuleConditionInsertError(err)
		}
		conditions = append(conditions, models.RuleConditionResponse{
			ID:                cond.ID,
			RuleID:            cond.RuleID,
			BaseRuleCondition: cond.BaseRuleCondition,
		})
	}
	return conditions, nil
}

// Unified function to fetch all rules with actions and conditions, grouped by rule ID
func (r *RuleRepository) fetchRulesMap(ctx context.Context, filterID *int64) (map[int64]*models.RuleResponse, error) {
	var (
		whereClause string
		args        []interface{}
	)
	if filterID != nil {
		whereClause = "WHERE r.id = $1"
		args = append(args, *filterID)
	}
	query := fmt.Sprintf(`SELECT
		r.id, r.name, r.description, r.effective_from, r.created_by,
		a.id, a.rule_id, a.action_type, a.action_value,
		c.id, c.rule_id, c.condition_type, c.condition_value, c.condition_operator
	FROM %s.%s r
	LEFT JOIN %s.%s a ON r.id = a.rule_id
	LEFT JOIN %s.%s c ON r.id = c.rule_id
	%s
	ORDER BY r.id`,
		r.schema, r.ruleTable, r.schema, r.ruleActionTable, r.schema, r.ruleConditionTable, whereClause)

	rows, err := r.db.FetchAll(ctx, query, args...)
	if err != nil {
		return nil, errorsPkg.NewRuleRepositoryError("failed to fetch rules with actions and conditions", err)
	}
	defer rows.Close()

	ruleMap := make(map[int64]*models.RuleResponse)
	actionMap := make(map[int64]map[int64]models.RuleActionResponse)
	condMap := make(map[int64]map[int64]models.RuleConditionResponse)

	for rows.Next() {
		var (
			ruleID, createdBy           int64
			ruleName                    string
			ruleDesc                    *string
			effectiveFrom               time.Time
			actionID, actionRuleID      *int64
			actionType, actionValue     *string
			condID, condRuleID          *int64
			condType, condValue, condOp *string
		)
		err := rows.Scan(
			&ruleID, &ruleName, &ruleDesc, &effectiveFrom, &createdBy,
			&actionID, &actionRuleID, &actionType, &actionValue,
			&condID, &condRuleID, &condType, &condValue, &condOp,
		)
		if err != nil {
			return nil, errorsPkg.NewRuleRepositoryError("failed to scan rule row", err)
		}
		if _, exists := ruleMap[ruleID]; !exists {
			ruleMap[ruleID] = &models.RuleResponse{
				ID:         ruleID,
				BaseRule:   models.BaseRule{Name: ruleName, Description: ruleDesc, EffectiveFrom: effectiveFrom},
				CreatedBy:  createdBy,
				Actions:    []models.RuleActionResponse{},
				Conditions: []models.RuleConditionResponse{},
			}
			actionMap[ruleID] = make(map[int64]models.RuleActionResponse)
			condMap[ruleID] = make(map[int64]models.RuleConditionResponse)
		}
		if actionID != nil && *actionID != 0 {
			if _, exists := actionMap[ruleID][*actionID]; !exists {
				action := models.RuleActionResponse{
					ID:             *actionID,
					RuleID:         *actionRuleID,
					BaseRuleAction: models.BaseRuleAction{ActionType: models.RuleFieldType(derefStr(actionType)), ActionValue: derefStr(actionValue)},
				}
				actionMap[ruleID][*actionID] = action
			}
		}
		if condID != nil && *condID != 0 {
			if _, exists := condMap[ruleID][*condID]; !exists {
				cond := models.RuleConditionResponse{
					ID:                *condID,
					RuleID:            *condRuleID,
					BaseRuleCondition: models.BaseRuleCondition{ConditionType: models.RuleFieldType(derefStr(condType)), ConditionValue: derefStr(condValue), ConditionOperator: models.RuleOperator(derefStr(condOp))},
				}
				condMap[ruleID][*condID] = cond
			}
		}
	}

	for ruleID, rule := range ruleMap {
		for _, a := range actionMap[ruleID] {
			rule.Actions = append(rule.Actions, a)
		}
		for _, c := range condMap[ruleID] {
			rule.Conditions = append(rule.Conditions, c)
		}
	}
	return ruleMap, nil
}

func (r *RuleRepository) GetRuleByID(ctx context.Context, id int64) (*models.RuleResponse, error) {
	ruleMap, err := r.fetchRulesMap(ctx, &id)
	if err != nil {
		return nil, err
	}
	rule, exists := ruleMap[id]
	if !exists {
		return nil, errorsPkg.NewRuleNotFoundError(nil)
	}
	return rule, nil
}

func (r *RuleRepository) ListRules(ctx context.Context) ([]*models.RuleResponse, error) {
	ruleMap, err := r.fetchRulesMap(ctx, nil)
	if err != nil {
		return nil, err
	}
	rules := make([]*models.RuleResponse, 0, len(ruleMap))
	for _, rule := range ruleMap {
		rules = append(rules, rule)
	}
	return rules, nil
}

func (r *RuleRepository) UpdateRule(ctx context.Context, id int64, req *models.UpdateRuleRequest) error {
	return r.db.WithTxn(ctx, func(tx pgx.Tx) error {
		// Update base rule fields
		query := fmt.Sprintf(`UPDATE %s.%s SET name = $1, description = $2, effective_from = $3 WHERE id = $4`, r.schema, r.ruleTable)
		_, err := tx.Exec(ctx, query, req.Name, req.Description, req.EffectiveFrom, id)
		if err != nil {
			return errorsPkg.NewRuleRepositoryError("failed to update rule", err)
		}

		// Delete existing actions and conditions
		query = fmt.Sprintf(`DELETE FROM %s.%s WHERE rule_id = $1`, r.schema, r.ruleActionTable)
		_, err = tx.Exec(ctx, query, id)
		if err != nil {
			return errorsPkg.NewRuleRepositoryError("failed to delete old rule actions", err)
		}
		query = fmt.Sprintf(`DELETE FROM %s.%s WHERE rule_id = $1`, r.schema, r.ruleConditionTable)
		_, err = tx.Exec(ctx, query, id)
		if err != nil {
			return errorsPkg.NewRuleRepositoryError("failed to delete old rule conditions", err)
		}

		// Insert new actions and conditions
		_, err = r.insertRuleActions(ctx, tx, req.Actions, id)
		if err != nil {
			return err
		}
		_, err = r.insertRuleConditions(ctx, tx, req.Conditions, id)
		if err != nil {
			return err
		}
		return nil
	})
}

func (r *RuleRepository) DeleteRule(ctx context.Context, id int64) error {
	return r.db.WithTxn(ctx, func(tx pgx.Tx) error {
		// Delete actions and conditions first (to avoid FK constraint errors)
		query := fmt.Sprintf(`DELETE FROM %s.%s WHERE rule_id = $1`, r.schema, r.ruleActionTable)
		_, err := tx.Exec(ctx, query, id)
		if err != nil {
			return errorsPkg.NewRuleRepositoryError("failed to delete rule actions", err)
		}
		query = fmt.Sprintf(`DELETE FROM %s.%s WHERE rule_id = $1`, r.schema, r.ruleConditionTable)
		_, err = tx.Exec(ctx, query, id)
		if err != nil {
			return errorsPkg.NewRuleRepositoryError("failed to delete rule conditions", err)
		}
		// Delete the rule itself
		query = fmt.Sprintf(`DELETE FROM %s.%s WHERE id = $1`, r.schema, r.ruleTable)
		_, err = tx.Exec(ctx, query, id)
		if err != nil {
			return errorsPkg.NewRuleRepositoryError("failed to delete rule", err)
		}
		return nil
	})
}

// derefStr safely dereferences a *string, returning "" if nil
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
