package service

import (
	database "expenses/internal/database/manager"
	"expenses/internal/errors"
	"expenses/internal/models"
	"expenses/internal/repository"
	"expenses/pkg/logger"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type RuleServiceInterface interface {
	CreateRule(c *gin.Context, ruleReq models.CreateRuleRequest) (models.DescribeRuleResponse, error)
	GetRuleById(c *gin.Context, id int64, userId int64) (models.DescribeRuleResponse, error)
	ListRules(c *gin.Context, userId int64) ([]models.RuleResponse, error)
	UpdateRule(c *gin.Context, id int64, ruleReq models.UpdateRuleRequest, userId int64) (models.RuleResponse, error)
	UpdateRuleAction(c *gin.Context, id int64, ruleId int64, ruleReq models.UpdateRuleActionRequest, userId int64) (models.RuleActionResponse, error)
	UpdateRuleCondition(c *gin.Context, id int64, ruleId int64, ruleReq models.UpdateRuleConditionRequest, userId int64) (models.RuleConditionResponse, error)
	DeleteRule(c *gin.Context, id int64, userId int64) error
	// ExecuteRules(c *gin.Context, userId int64) (models.ExecuteRulesResponse, error)
}

type ruleService struct {
	ruleRepo        repository.RuleRepositoryInterface
	transactionRepo repository.TransactionRepositoryInterface
	db              database.DatabaseManager
}

func NewRuleService(ruleRepo repository.RuleRepositoryInterface, transactionRepo repository.TransactionRepositoryInterface, db database.DatabaseManager) RuleServiceInterface {
	return &ruleService{
		ruleRepo:        ruleRepo,
		transactionRepo: transactionRepo,
		db:              db,
	}
}

func (s *ruleService) CreateRule(c *gin.Context, ruleReq models.CreateRuleRequest) (models.DescribeRuleResponse, error) {
	logger.Infof("Creating rule for user %d", ruleReq.Rule.CreatedBy)
	var ruleResponse models.DescribeRuleResponse
	if err := s.validateCreateRule(ruleReq); err != nil {
		return ruleResponse, err
	}

	err := s.db.WithTxn(c, func(tx pgx.Tx) error {
		rule, err := s.ruleRepo.CreateRule(c, ruleReq.Rule)
		if err != nil {
			return err
		}

		actions, err := s.ruleRepo.CreateRuleActions(c, ruleReq.Actions)
		if err != nil {
			return err
		}

		conditions, err := s.ruleRepo.CreateRuleConditions(c, ruleReq.Conditions)
		if err != nil {
			return err
		}
		ruleResponse.Rule = rule
		ruleResponse.Actions = actions
		ruleResponse.Conditions = conditions
		return nil
	})

	if err != nil {
		return ruleResponse, err
	}

	logger.Infof("Rule created successfully with Id %d", ruleResponse.Rule.Id)
	return ruleResponse, nil
}

func (s *ruleService) GetRuleById(c *gin.Context, id int64, userId int64) (models.DescribeRuleResponse, error) {
	logger.Infof("Fetching rule %d for user %d", id, userId)

	var ruleResponse models.DescribeRuleResponse
	rule, err := s.ruleRepo.GetRule(c, id, userId)
	if err != nil {
		return ruleResponse, err
	}
	ruleResponse.Rule = rule

	actions, err := s.ruleRepo.ListRuleActionsByRuleId(c, id)
	if err != nil {
		return ruleResponse, err
	}
	ruleResponse.Actions = actions

	conditions, err := s.ruleRepo.ListRuleConditionsByRuleId(c, id)
	if err != nil {
		return ruleResponse, err
	}
	ruleResponse.Conditions = conditions

	logger.Infof("Rule %d fetched successfully", id)
	return ruleResponse, nil
}

func (s *ruleService) ListRules(c *gin.Context, userId int64) ([]models.RuleResponse, error) {
	logger.Infof("Fetching all rules for user %d", userId)
	rules, err := s.ruleRepo.ListRules(c, userId)
	if err != nil {
		return nil, err
	}
	logger.Infof("Fetched %d rules for user %d", len(rules), userId)
	return rules, nil
}

func (s *ruleService) UpdateRule(c *gin.Context, id int64, ruleReq models.UpdateRuleRequest, userId int64) (models.RuleResponse, error) {
	logger.Infof("Updating rule %d for user %d", id, userId)
	rule, err := s.ruleRepo.UpdateRule(c, id, userId, ruleReq)
	if err != nil {
		return models.RuleResponse{}, err
	}
	return rule, nil
}

func (s *ruleService) UpdateRuleAction(c *gin.Context, id int64, ruleId int64, ruleReq models.UpdateRuleActionRequest, userId int64) (models.RuleActionResponse, error) {
	logger.Infof("Updating rule action %d for user %d", id, userId)
	rule, err := s.ruleRepo.GetRule(c, ruleId, userId)
	if err != nil {
		return models.RuleActionResponse{}, err
	}
	ruleAction, err := s.ruleRepo.UpdateRuleAction(c, id, ruleId, rule.Id, ruleReq)
	if err != nil {
		return models.RuleActionResponse{}, err
	}
	return ruleAction, nil
}

func (s *ruleService) UpdateRuleCondition(c *gin.Context, id int64, ruleId int64, ruleReq models.UpdateRuleConditionRequest, userId int64) (models.RuleConditionResponse, error) {
	logger.Infof("Updating rule condition %d for user %d", id, userId)
	rule, err := s.ruleRepo.GetRule(c, ruleId, userId)
	if err != nil {
		return models.RuleConditionResponse{}, err
	}
	ruleCondition, err := s.ruleRepo.UpdateRuleCondition(c, id, ruleId, rule.Id, ruleReq)
	if err != nil {
		return models.RuleConditionResponse{}, err
	}
	return ruleCondition, nil
}

func (s *ruleService) DeleteRule(c *gin.Context, id int64, userId int64) error {
	logger.Infof("Deleting rule %d for user %d", id, userId)
	err := s.db.WithTxn(c, func(tx pgx.Tx) error {
		err := s.ruleRepo.DeleteRule(c, id, userId)
		if err != nil {
			return err
		}
		err = s.ruleRepo.DeleteRuleActionsByRuleId(c, id)
		if err != nil {
			return err
		}
		err = s.ruleRepo.DeleteRuleConditionsByRuleId(c, id)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	logger.Infof("Rule %d deleted successfully", id)
	return nil
}

// --- Validation methods ---
func (s *ruleService) validateCreateRule(req models.CreateRuleRequest) error {
	if err := s.validateRuleActions(req.Actions); err != nil {
		return err
	}
	if err := s.validateRuleConditions(req.Conditions); err != nil {
		return err
	}
	if err := s.validateRuleEffectiveDate(req.Rule.EffectiveFrom); err != nil {
		return err
	}
	return nil
}

func (s *ruleService) validateRuleEffectiveDate(effectiveFrom time.Time) error {
	if effectiveFrom.IsZero() {
		return errors.NewRuleInvalidEffectiveDateError(nil)
	}
	if effectiveFrom.After(time.Now()) {
		return errors.NewRuleInvalidEffectiveDateError(nil)
	}
	return nil
}

func (s *ruleService) validateRuleActions(actions []models.CreateRuleActionRequest) error {
	if len(actions) == 0 {
		return errors.NewRuleNoActionsError(nil)
	}
	for _, action := range actions {
		if err := s.validateAction(action); err != nil {
			return err
		}
	}
	return nil
}

func (s *ruleService) validateRuleConditions(conditions []models.CreateRuleConditionRequest) error {
	if len(conditions) == 0 {
		return errors.NewRuleNoConditionsError(nil)
	}
	for _, cond := range conditions {
		if err := s.validateCondition(cond); err != nil {
			return err
		}
	}
	return nil
}

func (s *ruleService) validateAction(action models.CreateRuleActionRequest) error {
	if !s.isValidActionType(action.ActionType) {
		return errors.NewRuleInvalidActionTypeError(nil)
	}
	// Add more specific validation for action.ActionValue if needed
	return nil
}

func (s *ruleService) validateCondition(condition models.CreateRuleConditionRequest) error {
	if !s.isValidConditionType(condition.ConditionType) {
		return errors.NewRuleInvalidConditionTypeError(nil)
	}
	// Add more specific validation for condition values and operators
	if condition.ConditionType == models.RuleFieldAmount {
		if _, err := strconv.ParseFloat(condition.ConditionValue, 64); err != nil {
			return errors.NewRuleInvalidConditionValueError(err)
		}
	}
	if !s.isValidOperator(condition.ConditionOperator, condition.ConditionType) {
		return errors.NewRuleInvalidOperatorError(nil)
	}
	return nil
}

func (s *ruleService) isValidActionType(actionType models.RuleFieldType) bool {
	switch actionType {
	case models.RuleFieldName, models.RuleFieldDescription, models.RuleFieldAmount, models.RuleFieldCategory:
		return true
	default:
		return false
	}
}

func (s *ruleService) isValidConditionType(conditionType models.RuleFieldType) bool {
	switch conditionType {
	case models.RuleFieldName, models.RuleFieldDescription, models.RuleFieldAmount, models.RuleFieldCategory:
		return true
	default:
		return false
	}
}

func (s *ruleService) isValidOperator(op models.RuleOperator, fieldType models.RuleFieldType) bool {
	numericOperators := map[models.RuleOperator]bool{
		models.OperatorEquals: true, models.OperatorGreater: true, models.OperatorLower: true,
	}
	stringOperators := map[models.RuleOperator]bool{
		models.OperatorEquals: true, models.OperatorContains: true,
	}
	idOperators := map[models.RuleOperator]bool{
		models.OperatorEquals: true,
	}

	switch fieldType {
	case models.RuleFieldAmount:
		return numericOperators[op]
	case models.RuleFieldName, models.RuleFieldDescription:
		return stringOperators[op]
	case models.RuleFieldCategory:
		return idOperators[op]
	default:
		return false
	}
}
