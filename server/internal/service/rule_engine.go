package service

import (
	"expenses/internal/models"
	"expenses/pkg/logger"
	"strconv"
	"strings"
)

type RuleChangeset struct {
	TransactionId int64
	NameUpdate    *string
	DescUpdate    *string
	CategoryAdds  []int64
	AppliedRules  []int64
	UpdatedFields []models.RuleFieldType
}

type RuleEngineResult struct {
	Changesets []RuleChangeset
	Skipped    []models.SkippedResult
}

type RuleEngine struct {
	categories map[int64]models.CategoryResponse
}

func NewRuleEngine(categories []models.CategoryResponse) *RuleEngine {
	categoryMap := make(map[int64]models.CategoryResponse)
	for _, category := range categories {
		categoryMap[category.Id] = category
	}
	
	return &RuleEngine{
		categories: categoryMap,
	}
}

func (e *RuleEngine) ExecuteRules(rules []models.DescribeRuleResponse, transactions []models.TransactionResponse) RuleEngineResult {
	var changesets []RuleChangeset
	var skipped []models.SkippedResult

	for _, transaction := range transactions {
		changeset, skipReason := e.executeRulesOnTransaction(rules, transaction)
		
		if skipReason != "" {
			skipped = append(skipped, models.SkippedResult{
				TransactionId: transaction.Id,
				Reason:        skipReason,
			})
		} else if e.hasChangeset(changeset) {
			changesets = append(changesets, changeset)
		}
	}

	return RuleEngineResult{
		Changesets: changesets,
		Skipped:    skipped,
	}
}

func (e *RuleEngine) executeRulesOnTransaction(rules []models.DescribeRuleResponse, transaction models.TransactionResponse) (RuleChangeset, string) {
	changeset := RuleChangeset{
		TransactionId: transaction.Id,
		CategoryAdds:  []int64{},
		AppliedRules:  []int64{},
		UpdatedFields: []models.RuleFieldType{},
	}

	for _, rule := range rules {
		if rule.Rule.EffectiveFrom.After(transaction.Date) {
			continue
		}

		if e.evaluateRuleConditions(rule.Conditions, transaction) {
			ruleApplied := false
			
			for _, action := range rule.Actions {
				switch action.ActionType {
				case models.RuleFieldName:
					if changeset.NameUpdate == nil {
						changeset.NameUpdate = &action.ActionValue
						changeset.UpdatedFields = append(changeset.UpdatedFields, models.RuleFieldName)
						ruleApplied = true
					}
				case models.RuleFieldDescription:
					if changeset.DescUpdate == nil {
						changeset.DescUpdate = &action.ActionValue
						changeset.UpdatedFields = append(changeset.UpdatedFields, models.RuleFieldDescription)
						ruleApplied = true
					}
				case models.RuleFieldCategory:
					categoryId, err := strconv.ParseInt(action.ActionValue, 10, 64)
					if err != nil {
						logger.Warnf("Invalid category ID in rule action: %s", action.ActionValue)
						continue
					}
					
					if !e.categoryExists(categoryId, transaction.CreatedBy) {
						logger.Warnf("Category %d does not exist for user %d", categoryId, transaction.CreatedBy)
						continue
					}
					
					if !e.transactionHasCategory(transaction, categoryId) && !e.changesetHasCategory(changeset, categoryId) {
						changeset.CategoryAdds = append(changeset.CategoryAdds, categoryId)
						changeset.UpdatedFields = append(changeset.UpdatedFields, models.RuleFieldCategory)
						ruleApplied = true
					}
				}
			}
			
			if ruleApplied {
				changeset.AppliedRules = append(changeset.AppliedRules, rule.Rule.Id)
			}
		}
	}

	changeset.UpdatedFields = e.deduplicateFields(changeset.UpdatedFields)
	return changeset, ""
}

func (e *RuleEngine) evaluateRuleConditions(conditions []models.RuleConditionResponse, transaction models.TransactionResponse) bool {
	if len(conditions) == 0 {
		return false
	}

	for _, condition := range conditions {
		if !e.evaluateCondition(condition, transaction) {
			return false
		}
	}
	return true
}

func (e *RuleEngine) evaluateCondition(condition models.RuleConditionResponse, transaction models.TransactionResponse) bool {
	switch condition.ConditionType {
	case models.RuleFieldAmount:
		return e.evaluateAmountCondition(condition, transaction.Amount)
	case models.RuleFieldName:
		return e.evaluateStringCondition(condition, transaction.Name)
	case models.RuleFieldDescription:
		desc := ""
		if transaction.Description != nil {
			desc = *transaction.Description
		}
		return e.evaluateStringCondition(condition, desc)
	case models.RuleFieldCategory:
		return e.evaluateCategoryCondition(condition, transaction.CategoryIds)
	default:
		return false
	}
}

func (e *RuleEngine) evaluateAmountCondition(condition models.RuleConditionResponse, amount float64) bool {
	conditionAmount, err := strconv.ParseFloat(condition.ConditionValue, 64)
	if err != nil {
		return false
	}

	switch condition.ConditionOperator {
	case models.OperatorEquals:
		return amount == conditionAmount
	case models.OperatorGreater:
		return amount > conditionAmount
	case models.OperatorLower:
		return amount < conditionAmount
	default:
		return false
	}
}

func (e *RuleEngine) evaluateStringCondition(condition models.RuleConditionResponse, value string) bool {
	switch condition.ConditionOperator {
	case models.OperatorEquals:
		return strings.EqualFold(value, condition.ConditionValue)
	case models.OperatorContains:
		return strings.Contains(strings.ToLower(value), strings.ToLower(condition.ConditionValue))
	default:
		return false
	}
}

func (e *RuleEngine) evaluateCategoryCondition(condition models.RuleConditionResponse, categoryIds []int64) bool {
	conditionCategoryId, err := strconv.ParseInt(condition.ConditionValue, 10, 64)
	if err != nil {
		return false
	}

	switch condition.ConditionOperator {
	case models.OperatorEquals:
		for _, categoryId := range categoryIds {
			if categoryId == conditionCategoryId {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func (e *RuleEngine) categoryExists(categoryId int64, userId int64) bool {
	category, exists := e.categories[categoryId]
	return exists && category.CreatedBy == userId
}

func (e *RuleEngine) transactionHasCategory(transaction models.TransactionResponse, categoryId int64) bool {
	for _, id := range transaction.CategoryIds {
		if id == categoryId {
			return true
		}
	}
	return false
}

func (e *RuleEngine) changesetHasCategory(changeset RuleChangeset, categoryId int64) bool {
	for _, id := range changeset.CategoryAdds {
		if id == categoryId {
			return true
		}
	}
	return false
}

func (e *RuleEngine) hasChangeset(changeset RuleChangeset) bool {
	return changeset.NameUpdate != nil || changeset.DescUpdate != nil || len(changeset.CategoryAdds) > 0
}

func (e *RuleEngine) deduplicateFields(fields []models.RuleFieldType) []models.RuleFieldType {
	seen := make(map[models.RuleFieldType]bool)
	var result []models.RuleFieldType
	
	for _, field := range fields {
		if !seen[field] {
			seen[field] = true
			result = append(result, field)
		}
	}
	
	return result
}
