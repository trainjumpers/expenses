package models

import (
	"time"
)

// --- Enums for Rule Types and Operators ---
type RuleFieldType string

type RuleOperator string

const (
	RuleFieldAmount      RuleFieldType = "amount"
	RuleFieldName        RuleFieldType = "name"
	RuleFieldDescription RuleFieldType = "description"
	RuleFieldCategory    RuleFieldType = "category"
)

const (
	OperatorEquals   RuleOperator = "equals"
	OperatorContains RuleOperator = "contains"
	OperatorGreater  RuleOperator = "greater"
	OperatorLower    RuleOperator = "lower"
)

// --- Base Models ---
type BaseRule struct {
	Name          string    `json:"name" binding:"required,min=1,max=100"`
	Description   *string   `json:"description,omitempty" binding:"max=255"`
	EffectiveFrom time.Time `json:"effective_from" binding:"required"`
}

type BaseRuleAction struct {
	ActionType  RuleFieldType `json:"action_type" binding:"required"`
	ActionValue string        `json:"action_value" binding:"required,min=1,max=100"`
}

type BaseRuleCondition struct {
	ConditionType     RuleFieldType `json:"condition_type" binding:"required"`
	ConditionValue    string        `json:"condition_value" binding:"required,min=1,max=100"`
	ConditionOperator RuleOperator  `json:"condition_operator" binding:"required"`
}

// --- Request Models ---
type CreateRuleRequest struct {
	BaseRule
	CreatedBy  int64                        `json:"created_by"`
	Actions    []CreateRuleActionRequest    `json:"actions"`
	Conditions []CreateRuleConditionRequest `json:"conditions"`
}

type CreateRuleActionRequest struct {
	BaseRuleAction
}

type CreateRuleConditionRequest struct {
	BaseRuleCondition
}

type UpdateRuleRequest struct {
	BaseRule
	ID         int64                        `json:"id" binding:"required"`
	Actions    []CreateRuleActionRequest    `json:"actions"`
	Conditions []CreateRuleConditionRequest `json:"conditions"`
}

// --- Response Models ---
type RuleResponse struct {
	ID int64 `json:"id" db:"id"`
	BaseRule
	CreatedBy  int64                   `json:"created_by"`
	Actions    []RuleActionResponse    `json:"actions"`
	Conditions []RuleConditionResponse `json:"conditions"`
}

type RuleActionResponse struct {
	BaseRuleAction
	ID     int64 `json:"id"`
	RuleID int64 `json:"rule_id"`
}

type RuleConditionResponse struct {
	BaseRuleCondition
	ID     int64 `json:"id"`
	RuleID int64 `json:"rule_id"`
}

// --- DB Models (for internal use) ---
type Rule struct {
	ID int64 `json:"id"`
	BaseRule
	CreatedBy  int64           `json:"created_by"`
	Actions    []RuleAction    `json:"actions"`
	Conditions []RuleCondition `json:"conditions"`
}

type RuleAction struct {
	ID     int64 `json:"id"`
	RuleID int64 `json:"rule_id"`
	BaseRuleAction
}

type RuleCondition struct {
	ID     int64 `json:"id"`
	RuleID int64 `json:"rule_id"`
	BaseRuleCondition
}

type ExecuteRulesResponse struct {
	Modified []ModifiedResult `json:"modified"`
	Skipped  []SkippedResult  `json:"skipped"`
}

type ModifiedResult struct {
	TransactionID int64           `json:"transaction_id"`
	AppliedRules  []int64         `json:"applied_rules"`
	UpdatedFields []RuleFieldType `json:"updated_fields"`
}

type SkippedResult struct {
	TransactionID int64  `json:"transaction_id"`
	Reason        string `json:"reason"`
}
