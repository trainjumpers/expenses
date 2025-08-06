package service

import (
	"expenses/internal/models"
	"strconv"
	"strings"
)

type TransferInfo struct {
	AccountId int64
	Amount    float64
}

type Changeset struct {
	TransactionId int64
	NameUpdate    *string
	DescUpdate    *string
	CategoryAdds  []int64
	TransferInfo  *TransferInfo
	AppliedRules  []int64
}

type RuleEngine struct {
	categories map[int64]models.CategoryResponse
	accounts   map[int64]models.AccountResponse
	rules      []models.DescribeRuleResponse
}

func NewRuleEngine(categories []models.CategoryResponse, accounts []models.AccountResponse, rules []models.DescribeRuleResponse) *RuleEngine {
	categoryMap := make(map[int64]models.CategoryResponse)
	for _, category := range categories {
		categoryMap[category.Id] = category
	}

	accountMap := make(map[int64]models.AccountResponse)
	for _, account := range accounts {
		accountMap[account.Id] = account
	}

	return &RuleEngine{
		categories: categoryMap,
		accounts:   accountMap,
		rules:      rules,
	}
}

// ProcessTransaction evaluates a transaction against all rules in the engine.
// It returns a Changeset if any rule applies, otherwise it returns nil.
func (e *RuleEngine) ProcessTransaction(transaction models.TransactionResponse) *Changeset {
	changeset := &Changeset{
		TransactionId: transaction.Id,
		CategoryAdds:  []int64{},
		AppliedRules:  []int64{},
	}

	hasChanges := false

	for _, rule := range e.rules {
		if rule.Rule.EffectiveFrom.After(transaction.Date) {
			continue
		}

		if !e.evaluateConditions(rule, transaction) {
			continue
		}

		ruleApplied := false
		for _, action := range rule.Actions {
			switch action.ActionType {
			case models.RuleFieldName:
				if changeset.NameUpdate == nil {
					changeset.NameUpdate = &action.ActionValue
					ruleApplied = true
					hasChanges = true
				}
			case models.RuleFieldDescription:
				if changeset.DescUpdate == nil {
					changeset.DescUpdate = &action.ActionValue
					ruleApplied = true
					hasChanges = true
				}
			case models.RuleFieldCategory:
				categoryId, err := strconv.ParseInt(action.ActionValue, 10, 64)
				if err != nil {
					continue
				}

				if !e.categoryExists(categoryId, transaction.CreatedBy) {
					continue
				}

				if !e.hasCategory(transaction.CategoryIds, categoryId) && !e.hasCategory(changeset.CategoryAdds, categoryId) {
					changeset.CategoryAdds = append(changeset.CategoryAdds, categoryId)
					ruleApplied = true
					hasChanges = true
				}
			case models.RuleFieldTransfer:
				accountId, err := strconv.ParseInt(action.ActionValue, 10, 64)
				if err != nil {
					continue
				}

				// Validate that the account exists and belongs to the user
				if !e.accountExists(accountId, transaction.CreatedBy) {
					continue
				}

				// Prevent transfer to the same account
				if accountId == transaction.AccountId {
					continue
				}

				// Only apply transfer if no transfer is already planned
				if changeset.TransferInfo == nil {
					changeset.TransferInfo = &TransferInfo{
						AccountId: accountId,
						Amount:    -transaction.Amount, // Negate the amount
					}
					ruleApplied = true
					hasChanges = true
				}
			}
		}

		if ruleApplied {
			changeset.AppliedRules = append(changeset.AppliedRules, rule.Rule.Id)
		}
	}

	if !hasChanges {
		return nil
	}

	return changeset
}

// evaluateConditions dispatches condition evaluation based on the rule's logic.
func (e *RuleEngine) evaluateConditions(rule models.DescribeRuleResponse, transaction models.TransactionResponse) bool {
	if len(rule.Conditions) == 0 {
		return false // A rule must have at least one condition.
	}

	if rule.Rule.ConditionLogic == models.ConditionLogicOr {
		return e.evaluateOrConditions(rule.Conditions, transaction)
	}

	// Default to AND logic
	return e.evaluateAndConditions(rule.Conditions, transaction)
}

// evaluateAndConditions checks if all conditions are met for a transaction.
func (e *RuleEngine) evaluateAndConditions(conditions []models.RuleConditionResponse, transaction models.TransactionResponse) bool {
	for _, condition := range conditions {
		if !e.evaluateCondition(condition, transaction) {
			return false // For AND, if any condition is false, the whole thing is false.
		}
	}
	return true // If loop finishes, all conditions were met.
}

// evaluateOrConditions checks if at least one condition is met for a transaction.
func (e *RuleEngine) evaluateOrConditions(conditions []models.RuleConditionResponse, transaction models.TransactionResponse) bool {
	for _, condition := range conditions {
		if e.evaluateCondition(condition, transaction) {
			return true // For OR, if any condition is true, the whole thing is true.
		}
	}
	return false // If loop finishes, no conditions were met.
}

// evaluateCondition dispatches the evaluation to the correct function based on the condition type.
func (e *RuleEngine) evaluateCondition(condition models.RuleConditionResponse, transaction models.TransactionResponse) bool {
	switch condition.ConditionType {
	case models.RuleFieldCategory:
		return e.evaluateCategoryCondition(condition, transaction.CategoryIds)
	case models.RuleFieldTransfer:
		return e.evaluateTransferCondition(condition, transaction.AccountId)
	default:
		return e.evaluateStandardFieldCondition(condition, transaction)
	}
}

// evaluateStandardFieldCondition handles the evaluation for primitive transaction fields.
func (e *RuleEngine) evaluateStandardFieldCondition(condition models.RuleConditionResponse, transaction models.TransactionResponse) bool {
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
	}
	return false
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
	}
	return false
}

func (e *RuleEngine) evaluateStringCondition(condition models.RuleConditionResponse, value string) bool {
	switch condition.ConditionOperator {
	case models.OperatorEquals:
		return strings.EqualFold(value, condition.ConditionValue)
	case models.OperatorContains:
		return strings.Contains(strings.ToLower(value), strings.ToLower(condition.ConditionValue))
	}
	return false
}

func (e *RuleEngine) evaluateCategoryCondition(condition models.RuleConditionResponse, categoryIds []int64) bool {
	conditionCategoryId, err := strconv.ParseInt(condition.ConditionValue, 10, 64)
	if err != nil {
		return false
	}

	if condition.ConditionOperator == models.OperatorEquals {
		return e.hasCategory(categoryIds, conditionCategoryId)
	}
	return false
}

func (e *RuleEngine) evaluateTransferCondition(condition models.RuleConditionResponse, accountId int64) bool {
	conditionAccountId, err := strconv.ParseInt(condition.ConditionValue, 10, 64)
	if err != nil {
		return false
	}

	if condition.ConditionOperator == models.OperatorEquals {
		return accountId == conditionAccountId
	}
	return false
}

func (e *RuleEngine) categoryExists(categoryId int64, userId int64) bool {
	category, exists := e.categories[categoryId]
	return exists && category.CreatedBy == userId
}

func (e *RuleEngine) accountExists(accountId int64, userId int64) bool {
	account, exists := e.accounts[accountId]
	return exists && account.CreatedBy == userId
}

func (e *RuleEngine) hasCategory(categoryIds []int64, categoryId int64) bool {
	for _, id := range categoryIds {
		if id == categoryId {
			return true
		}
	}
	return false
}
