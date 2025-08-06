package mock_repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"expenses/internal/models"
)

type MockRuleRepository struct {
	mu              sync.Mutex
	rules           map[int64]models.RuleResponse
	actions         map[int64]models.RuleActionResponse
	conditions      map[int64]models.RuleConditionResponse
	mappings        map[string]bool // key: "ruleId:transactionId"
	nextRuleId      int64
	nextActionId    int64
	nextConditionId int64
}

func NewMockRuleRepository() *MockRuleRepository {
	return &MockRuleRepository{
		rules:           make(map[int64]models.RuleResponse),
		actions:         make(map[int64]models.RuleActionResponse),
		conditions:      make(map[int64]models.RuleConditionResponse),
		mappings:        make(map[string]bool),
		nextRuleId:      1,
		nextActionId:    1,
		nextConditionId: 1,
	}
}

func (m *MockRuleRepository) CreateRule(ctx context.Context, req models.CreateBaseRuleRequest) (models.RuleResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rule := models.RuleResponse{
		Id:            m.nextRuleId,
		Name:          req.Name,
		Description:   req.Description,
		EffectiveFrom: req.EffectiveFrom,
		CreatedBy:     req.CreatedBy,
	}
	m.rules[m.nextRuleId] = rule
	m.nextRuleId++
	return rule, nil
}

func (m *MockRuleRepository) CreateRuleActions(ctx context.Context, actions []models.CreateRuleActionRequest) ([]models.RuleActionResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []models.RuleActionResponse
	for _, a := range actions {
		action := models.RuleActionResponse{
			Id:          m.nextActionId,
			RuleId:      a.RuleId,
			ActionType:  a.ActionType,
			ActionValue: a.ActionValue,
		}
		m.actions[m.nextActionId] = action
		result = append(result, action)
		m.nextActionId++
	}
	return result, nil
}

func (m *MockRuleRepository) CreateRuleConditions(ctx context.Context, conditions []models.CreateRuleConditionRequest) ([]models.RuleConditionResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []models.RuleConditionResponse
	for _, cond := range conditions {
		condition := models.RuleConditionResponse{
			Id:                m.nextConditionId,
			RuleId:            cond.RuleId,
			ConditionType:     cond.ConditionType,
			ConditionValue:    cond.ConditionValue,
			ConditionOperator: cond.ConditionOperator,
		}
		m.conditions[m.nextConditionId] = condition
		result = append(result, condition)
		m.nextConditionId++
	}
	return result, nil
}

func (m *MockRuleRepository) GetRule(ctx context.Context, id int64, userId int64) (models.RuleResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rule, ok := m.rules[id]
	if !ok || rule.CreatedBy != userId {
		return models.RuleResponse{}, errors.New("rule not found")
	}
	return rule, nil
}

func (m *MockRuleRepository) ListRules(ctx context.Context, userId int64, query models.RuleListQuery) (models.PaginatedRulesResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var filteredRules []models.RuleResponse
	for _, rule := range m.rules {
		if rule.CreatedBy != userId {
			continue
		}

		// Apply search filter if provided
		if query.Search != nil && *query.Search != "" {
			searchTerm := strings.ToLower(*query.Search)
			nameMatch := strings.Contains(strings.ToLower(rule.Name), searchTerm)
			descMatch := rule.Description != nil && strings.Contains(strings.ToLower(*rule.Description), searchTerm)
			if !nameMatch && !descMatch {
				continue
			}
		}

		filteredRules = append(filteredRules, rule)
	}

	total := len(filteredRules)

	// If no pagination specified (PageSize <= 0), return all results
	if query.PageSize <= 0 {
		return models.PaginatedRulesResponse{
			Rules:    filteredRules,
			Total:    total,
			Page:     query.Page,
			PageSize: query.PageSize,
		}, nil
	}

	// Apply pagination
	start := (query.Page - 1) * query.PageSize
	end := start + query.PageSize

	if start >= total {
		return models.PaginatedRulesResponse{
			Rules:    []models.RuleResponse{},
			Total:    total,
			Page:     query.Page,
			PageSize: query.PageSize,
		}, nil
	}

	if end > total {
		end = total
	}

	return models.PaginatedRulesResponse{
		Rules:    filteredRules[start:end],
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func (m *MockRuleRepository) ListRuleActionsByRuleId(ctx context.Context, ruleId int64) ([]models.RuleActionResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []models.RuleActionResponse
	for _, action := range m.actions {
		if action.RuleId == ruleId {
			result = append(result, action)
		}
	}
	return result, nil
}

func (m *MockRuleRepository) ListRuleConditionsByRuleId(ctx context.Context, ruleId int64) ([]models.RuleConditionResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []models.RuleConditionResponse
	for _, cond := range m.conditions {
		if cond.RuleId == ruleId {
			result = append(result, cond)
		}
	}
	return result, nil
}

func (m *MockRuleRepository) UpdateRule(ctx context.Context, id int64, userId int64, req models.UpdateRuleRequest) (models.RuleResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rule, ok := m.rules[id]
	if !ok || rule.CreatedBy != userId {
		return models.RuleResponse{}, errors.New("rule not found")
	}
	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.Description != nil {
		rule.Description = req.Description
	}
	if req.EffectiveFrom != nil {
		rule.EffectiveFrom = *req.EffectiveFrom
	}
	m.rules[id] = rule
	return rule, nil
}

func (m *MockRuleRepository) UpdateRuleAction(ctx context.Context, id int64, ruleId int64, req models.UpdateRuleActionRequest) (models.RuleActionResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	action, ok := m.actions[id]
	if !ok || action.RuleId != ruleId {
		return models.RuleActionResponse{}, errors.New("action not found")
	}
	if req.ActionType != nil {
		action.ActionType = *req.ActionType
	}
	if req.ActionValue != nil {
		action.ActionValue = *req.ActionValue
	}
	m.actions[id] = action
	return action, nil
}

func (m *MockRuleRepository) UpdateRuleCondition(ctx context.Context, id int64, ruleId int64, req models.UpdateRuleConditionRequest) (models.RuleConditionResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cond, ok := m.conditions[id]
	if !ok || cond.RuleId != ruleId {
		return models.RuleConditionResponse{}, errors.New("condition not found")
	}
	if req.ConditionType != nil {
		cond.ConditionType = *req.ConditionType
	}
	if req.ConditionValue != nil {
		cond.ConditionValue = *req.ConditionValue
	}
	if req.ConditionOperator != nil {
		cond.ConditionOperator = *req.ConditionOperator
	}
	m.conditions[id] = cond
	return cond, nil
}

func (m *MockRuleRepository) DeleteRuleActionsByRuleId(ctx context.Context, ruleId int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, action := range m.actions {
		if action.RuleId == ruleId {
			delete(m.actions, id)
		}
	}
	return nil
}

func (m *MockRuleRepository) DeleteRuleConditionsByRuleId(ctx context.Context, ruleId int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, cond := range m.conditions {
		if cond.RuleId == ruleId {
			delete(m.conditions, id)
		}
	}
	return nil
}

func (m *MockRuleRepository) DeleteRule(ctx context.Context, id int64, userId int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	rule, ok := m.rules[id]
	if !ok || rule.CreatedBy != userId {
		return errors.New("rule not found")
	}
	delete(m.rules, id)
	// Also delete actions and conditions for this rule
	for aid, action := range m.actions {
		if action.RuleId == id {
			delete(m.actions, aid)
		}
	}
	for cid, cond := range m.conditions {
		if cond.RuleId == id {
			delete(m.conditions, cid)
		}
	}
	return nil
}

func (m *MockRuleRepository) CreateRuleTransactionMapping(ctx context.Context, ruleId int64, transactionId int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%d:%d", ruleId, transactionId)
	m.mappings[key] = true
	return nil
}

func (m *MockRuleRepository) PutRuleActions(ctx context.Context, ruleId int64, actions []models.CreateRuleActionRequest) ([]models.RuleActionResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Delete existing actions for this rule
	for id, action := range m.actions {
		if action.RuleId == ruleId {
			delete(m.actions, id)
		}
	}

	// Create new actions
	var result []models.RuleActionResponse
	for _, a := range actions {
		action := models.RuleActionResponse{
			Id:          m.nextActionId,
			RuleId:      ruleId,
			ActionType:  a.ActionType,
			ActionValue: a.ActionValue,
		}
		m.actions[m.nextActionId] = action
		result = append(result, action)
		m.nextActionId++
	}
	return result, nil
}

func (m *MockRuleRepository) PutRuleConditions(ctx context.Context, ruleId int64, conditions []models.CreateRuleConditionRequest) ([]models.RuleConditionResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Delete existing conditions for this rule
	for id, cond := range m.conditions {
		if cond.RuleId == ruleId {
			delete(m.conditions, id)
		}
	}

	// Create new conditions
	var result []models.RuleConditionResponse
	for _, c := range conditions {
		condition := models.RuleConditionResponse{
			Id:                m.nextConditionId,
			RuleId:            ruleId,
			ConditionType:     c.ConditionType,
			ConditionValue:    c.ConditionValue,
			ConditionOperator: c.ConditionOperator,
		}
		m.conditions[m.nextConditionId] = condition
		result = append(result, condition)
		m.nextConditionId++
	}
	return result, nil
}
