package validator

import (
	"expenses/internal/errors"
	"expenses/internal/models"
	"fmt"
	"strconv"
	"time"
)

// RuleValidator centralizes all rule-related validation logic.
type RuleValidator struct{}

// ValidateUpdateAction validates an UpdateRuleActionRequest.
func (v *RuleValidator) ValidateUpdateAction(action models.UpdateRuleActionRequest) error {
	if action.ActionType != nil {
		if err := v.validateActionType(*action.ActionType); err != nil {
			return err
		}
	}
	if action.ActionValue != nil && action.ActionType != nil {
		if err := v.validateRuleType(*action.ActionType, *action.ActionValue); err != nil {
			return err
		}
	}
	return nil
}

// ValidateUpdateCondition validates an UpdateRuleConditionRequest.
func (v *RuleValidator) ValidateUpdateCondition(cond models.UpdateRuleConditionRequest) error {
	if cond.ConditionType != nil {
		if err := v.validateConditionType(*cond.ConditionType); err != nil {
			return err
		}
	}
	if cond.ConditionValue != nil && cond.ConditionType != nil {
		if err := v.validateRuleType(*cond.ConditionType, *cond.ConditionValue); err != nil {
			return err
		}
	}
	if cond.ConditionOperator != nil && cond.ConditionType != nil {
		if err := v.validateOperator(*cond.ConditionOperator, *cond.ConditionType); err != nil {
			return err
		}
	}
	return nil
}

// ValidateUpdate validates an UpdateRuleRequest.
func (v *RuleValidator) ValidateUpdate(rule models.UpdateRuleRequest) error {
	if rule.EffectiveFrom != nil {
		if err := v.validateEffectiveDate(*rule.EffectiveFrom); err != nil {
			return err
		}
	}
	return nil
}

func (v *RuleValidator) Validate(rule models.CreateRuleRequest) error {
	if len(rule.Actions) == 0 {
		return errors.NewRuleNoActionsError(fmt.Errorf("at least one action is required"))
	}
	for _, action := range rule.Actions {
		if err := v.validateAction(action); err != nil {
			return err
		}
	}
	if len(rule.Conditions) == 0 {
		return errors.NewRuleNoConditionsError(fmt.Errorf("at least one condition is required"))
	}
	for _, condition := range rule.Conditions {
		if err := v.validateCondition(condition); err != nil {
			return err
		}
	}
	if err := v.validateEffectiveDate(rule.Rule.EffectiveFrom); err != nil {
		return err
	}
	return nil
}

// ValidateActionType checks if the action type is valid.
func (v *RuleValidator) validateActionType(actionType models.RuleFieldType) error {
	switch actionType {
	case models.RuleFieldName, models.RuleFieldDescription, models.RuleFieldAmount, models.RuleFieldCategory, models.RuleFieldTransfer:
		return nil
	default:
		return errors.NewRuleInvalidActionTypeError(fmt.Errorf("action type %s is not valid", actionType))
	}
}

// ValidateActionValue can be extended for more specific action value checks.
func (v *RuleValidator) validateRuleType(actionType models.RuleFieldType, value string) error {
	switch actionType {
	case models.RuleFieldAmount:
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return errors.NewRuleInvalidConditionValueError(err)
		}
	case models.RuleFieldCategory, models.RuleFieldTransfer:
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			return errors.NewRuleInvalidConditionValueError(fmt.Errorf("value for %s must be a valid ID", actionType))
		}
	case models.RuleFieldName, models.RuleFieldDescription:
		// Already a string, but you could add length or charset checks here if needed.
		if value == "" {
			return errors.NewRuleInvalidConditionValueError(fmt.Errorf("value for %s cannot be empty", actionType))
		}
	default:
		return errors.NewRuleInvalidActionTypeError(fmt.Errorf("action type %s is not valid", actionType))
	}
	return nil
}

// ValidateAction validates a single rule action.
func (v *RuleValidator) validateAction(action models.CreateRuleActionRequest) error {
	if err := v.validateActionType(action.ActionType); err != nil {
		return err
	}
	if err := v.validateRuleType(action.ActionType, action.ActionValue); err != nil {
		return err
	}
	return nil
}

// ValidateConditionType checks if the condition type is valid.
func (v *RuleValidator) validateConditionType(conditionType models.RuleFieldType) error {
	switch conditionType {
	case models.RuleFieldName, models.RuleFieldDescription, models.RuleFieldAmount, models.RuleFieldCategory, models.RuleFieldTransfer:
		return nil
	default:
		return errors.NewRuleInvalidConditionTypeError(fmt.Errorf("condition type %s is not valid", conditionType))
	}
}

// ValidateOperator checks if the operator is valid for the given field type.
func (v *RuleValidator) validateOperator(op models.RuleOperator, fieldType models.RuleFieldType) error {
	switch fieldType {
	case models.RuleFieldAmount:
		if op == models.OperatorEquals || op == models.OperatorGreater || op == models.OperatorLower {
			return nil
		}
	case models.RuleFieldName, models.RuleFieldDescription:
		if op == models.OperatorEquals || op == models.OperatorContains {
			return nil
		}
	case models.RuleFieldCategory, models.RuleFieldTransfer:
		if op == models.OperatorEquals {
			return nil
		}
	}
	return errors.NewRuleInvalidOperatorError(fmt.Errorf("operator %s is not valid for field type %s", op, fieldType))
}

// ValidateCondition validates a single rule condition.
func (v *RuleValidator) validateCondition(cond models.CreateRuleConditionRequest) error {
	if err := v.validateConditionType(cond.ConditionType); err != nil {
		return err
	}
	if err := v.validateRuleType(cond.ConditionType, cond.ConditionValue); err != nil {
		return err
	}
	if err := v.validateOperator(cond.ConditionOperator, cond.ConditionType); err != nil {
		return err
	}
	return nil
}

// ValidateEffectiveDate checks if the effective date is valid (not zero and not in the future).
func (v *RuleValidator) validateEffectiveDate(effectiveFrom time.Time) error {
	if effectiveFrom.IsZero() || effectiveFrom.After(time.Now()) {
		return errors.NewRuleInvalidEffectiveDateError(fmt.Errorf("effective date must be set and cannot be in the future"))
	}
	return nil
}

// ValidatePutActions validates a PutRuleActionsRequest.
func (v *RuleValidator) ValidatePutActions(req models.PutRuleActionsRequest) error {
	if len(req.Actions) == 0 {
		return errors.NewRuleNoActionsError(fmt.Errorf("at least one action is required"))
	}
	for _, action := range req.Actions {
		if err := v.validateAction(action); err != nil {
			return err
		}
	}
	return nil
}

// ValidatePutConditions validates a PutRuleConditionsRequest.
func (v *RuleValidator) ValidatePutConditions(req models.PutRuleConditionsRequest) error {
	if len(req.Conditions) == 0 {
		return errors.NewRuleNoConditionsError(fmt.Errorf("at least one condition is required"))
	}
	for _, condition := range req.Conditions {
		if err := v.validateCondition(condition); err != nil {
			return err
		}
	}
	return nil
}
