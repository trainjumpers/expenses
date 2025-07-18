package errors

import (
	"fmt"
	"net/http"
)

func NewRuleNotFoundError(err error) *AuthError {
	return formatError(http.StatusNotFound, "the requested rule was not found.", err, "RuleNotFound")
}

func NewRuleActionNotFoundError(err error) *AuthError {
	return formatError(http.StatusNotFound, "the requested rule action was not found.", err, "RuleActionNotFound")
}

func NewRuleConditionNotFoundError(err error) *AuthError {
	return formatError(http.StatusNotFound, "the requested rule condition was not found.", err, "RuleConditionNotFound")
}

func NewRuleInvalidEffectiveDateError(err error) *AuthError {
	return formatError(http.StatusBadRequest, "the effective date for the rule is invalid or in the past.", err, "InvalidEffectiveDate")
}

func NewRuleNoActionsError(err error) *AuthError {
	return formatError(http.StatusBadRequest, "a rule must have at least one action.", err, "RuleNoActions")
}

func NewRuleNoConditionsError(err error) *AuthError {
	return formatError(http.StatusBadRequest, "a rule must have at least one condition.", err, "RuleNoConditions")
}

func NewRuleInvalidActionTypeError(err error) *AuthError {
	return formatError(http.StatusBadRequest, "the provided action type is invalid.", err, "InvalidActionType")
}

func NewRuleInvalidConditionTypeError(err error) *AuthError {
	return formatError(http.StatusBadRequest, "the provided condition type is invalid.", err, "InvalidConditionType")
}

func NewRuleInvalidConditionValueError(err error) *AuthError {
	return formatError(http.StatusBadRequest, fmt.Sprintf("the condition value is invalid for its type: %v", err), err, "InvalidConditionValue")
}

func NewRuleInvalidOperatorError(err error) *AuthError {
	return formatError(http.StatusBadRequest, "the operator is not valid for the given condition type.", err, "InvalidOperator")
}

func NewRuleActionInsertError(err error) *AuthError {
	return formatError(http.StatusInternalServerError, "failed to insert rule_action", err, "RuleActionInsert")
}

func NewRuleConditionInsertError(err error) *AuthError {
	return formatError(http.StatusInternalServerError, "failed to insert rule_condition", err, "RuleConditionInsert")
}

func NewRuleRepositoryError(msg string, err error) *AuthError {
	return formatError(http.StatusInternalServerError, msg, err, "ruleRepository")
}
