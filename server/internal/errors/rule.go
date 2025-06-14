package errors

import (
	"net/http"
)

func NewRuleNotFoundError(err error) *AuthError {
	return formatError(http.StatusNotFound, "rule not found", err, "RuleNotFound")
}

func NewRuleActionInsertError(err error) *AuthError {
	return formatError(http.StatusInternalServerError, "failed to insert rule_action", err, "RuleActionInsert")
}

func NewRuleConditionInsertError(err error) *AuthError {
	return formatError(http.StatusInternalServerError, "failed to insert rule_condition", err, "RuleConditionInsert")
}

func NewRuleRepositoryError(msg string, err error) *AuthError {
	return formatError(http.StatusInternalServerError, msg, err, "RuleRepository")
}
