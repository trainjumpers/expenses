package models

import (
	"time"
)

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

type CreateBaseRuleRequest struct {
	Name          string    `json:"name" binding:"required,min=1,max=100"`
	Description   *string   `json:"description,omitempty" binding:"omitempty,max=255"`
	EffectiveFrom time.Time `json:"effective_from" binding:"required"`
	CreatedBy     int64     `json:"created_by"`
}

type CreateRuleRequest struct {
	Rule       CreateBaseRuleRequest        `json:"rule" binding:"required"`
	Actions    []CreateRuleActionRequest    `json:"actions" binding:"required,min=1"`
	Conditions []CreateRuleConditionRequest `json:"conditions" binding:"required,min=1"`
}

type UpdateRuleRequest struct {
	Name          *string    `json:"name,omitempty" binding:"omitempty,max=100"`
	Description   *string    `json:"description,omitempty" binding:"omitempty,max=255"`
	EffectiveFrom *time.Time `json:"effective_from,omitempty"`
}

type RuleResponse struct {
	Id            int64     `json:"id"`
	Name          string    `json:"name"`
	Description   *string   `json:"description"`
	EffectiveFrom time.Time `json:"effective_from"`
	CreatedBy     int64     `json:"created_by"`
}

type DescribeRuleResponse struct {
	Rule       RuleResponse            `json:"rule"`
	Actions    []RuleActionResponse    `json:"actions"`
	Conditions []RuleConditionResponse `json:"conditions"`
}

type CreateRuleActionRequest struct {
	ActionType  RuleFieldType `json:"action_type" binding:"required"`
	ActionValue string        `json:"action_value" binding:"required,min=1,max=100"`
	RuleId      int64         `json:"rule_id"`
}

type UpdateRuleActionRequest struct {
	ActionType  *RuleFieldType `json:"action_type"`
	ActionValue *string        `json:"action_value"`
}

type RuleActionResponse struct {
	Id          int64         `json:"id"`
	RuleId      int64         `json:"rule_id"`
	ActionType  RuleFieldType `json:"action_type"`
	ActionValue string        `json:"action_value"`
}

type CreateRuleConditionRequest struct {
	ConditionType     RuleFieldType `json:"condition_type" binding:"required"`
	ConditionValue    string        `json:"condition_value" binding:"required,min=1,max=100"`
	ConditionOperator RuleOperator  `json:"condition_operator" binding:"required"`
	RuleId            int64         `json:"rule_id"`
}

type UpdateRuleConditionRequest struct {
	ConditionType     *RuleFieldType `json:"condition_type"`
	ConditionValue    *string        `json:"condition_value"`
	ConditionOperator *RuleOperator  `json:"condition_operator"`
}

type RuleConditionResponse struct {
	Id                int64         `json:"id"`
	RuleId            int64         `json:"rule_id"`
	ConditionType     RuleFieldType `json:"condition_type"`
	ConditionValue    string        `json:"condition_value"`
	ConditionOperator RuleOperator  `json:"condition_operator"`
}

// Execute
type ExecuteRulesResponse struct {
	Modified []ModifiedResult `json:"modified"`
	Skipped  []SkippedResult  `json:"skipped"`
}

type ModifiedResult struct {
	TransactionId int64           `json:"transaction_id"`
	AppliedRules  []int64         `json:"applied_rules"`
	UpdatedFields []RuleFieldType `json:"updated_fields"`
}

type SkippedResult struct {
	TransactionId int64  `json:"transaction_id"`
	Reason        string `json:"reason"`
}
