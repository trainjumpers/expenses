package mock_repository

import (
	"context"
	errorsPkg "expenses/internal/errors"
	"expenses/internal/models"
)

type MockRuleRepository struct {
	rules  map[int64]*models.RuleResponse
	nextId int64
}

func NewMockRuleRepository() *MockRuleRepository {
	return &MockRuleRepository{
		rules:  make(map[int64]*models.RuleResponse),
		nextId: 1,
	}
}

func (m *MockRuleRepository) CreateRule(ctx context.Context, req *models.CreateRuleRequest) (*models.RuleResponse, error) {
	id := m.nextId
	m.nextId++
	// Assign IDs to actions and conditions
	actions := make([]models.RuleActionResponse, len(req.Actions))
	for i, a := range req.Actions {
		actions[i] = models.RuleActionResponse{
			ID:             int64(i + 1),
			RuleID:         id,
			BaseRuleAction: a.BaseRuleAction,
		}
	}
	conditions := make([]models.RuleConditionResponse, len(req.Conditions))
	for i, c := range req.Conditions {
		conditions[i] = models.RuleConditionResponse{
			ID:                int64(i + 1),
			RuleID:            id,
			BaseRuleCondition: c.BaseRuleCondition,
		}
	}
	rule := &models.RuleResponse{
		ID:         id,
		BaseRule:   req.BaseRule,
		CreatedBy:  req.CreatedBy,
		Actions:    actions,
		Conditions: conditions,
	}
	m.rules[id] = rule
	return rule, nil
}

func (m *MockRuleRepository) GetRuleByID(ctx context.Context, id int64) (*models.RuleResponse, error) {
	rule, ok := m.rules[id]
	if !ok {
		return nil, errorsPkg.NewRuleNotFoundError(nil)
	}
	return rule, nil
}

func (m *MockRuleRepository) ListRules(ctx context.Context) ([]*models.RuleResponse, error) {
	var result []*models.RuleResponse
	for _, rule := range m.rules {
		result = append(result, rule)
	}
	return result, nil
}

func (m *MockRuleRepository) UpdateRule(ctx context.Context, id int64, req *models.UpdateRuleRequest) error {
	rule, ok := m.rules[id]
	if !ok {
		return errorsPkg.NewRuleNotFoundError(nil)
	}
	rule.BaseRule = req.BaseRule
	// Update actions and conditions
	actions := make([]models.RuleActionResponse, len(req.Actions))
	for i, a := range req.Actions {
		actions[i] = models.RuleActionResponse{
			ID:             int64(i + 1),
			RuleID:         id,
			BaseRuleAction: a.BaseRuleAction,
		}
	}
	conditions := make([]models.RuleConditionResponse, len(req.Conditions))
	for i, c := range req.Conditions {
		conditions[i] = models.RuleConditionResponse{
			ID:                int64(i + 1),
			RuleID:            id,
			BaseRuleCondition: c.BaseRuleCondition,
		}
	}
	rule.Actions = actions
	rule.Conditions = conditions
	return nil
}

func (m *MockRuleRepository) DeleteRule(ctx context.Context, id int64) error {
	_, ok := m.rules[id]
	if !ok {
		return errorsPkg.NewRuleNotFoundError(nil)
	}
	delete(m.rules, id)
	return nil
}
