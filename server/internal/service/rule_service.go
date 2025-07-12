package service

import (
	database "expenses/internal/database/manager"
	"expenses/internal/models"
	"expenses/internal/repository"
	"expenses/internal/validator"
	"expenses/pkg/logger"

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
}

type ruleService struct {
	ruleRepo        repository.RuleRepositoryInterface
	transactionRepo repository.TransactionRepositoryInterface
	db              database.DatabaseManager
	validator       *validator.RuleValidator
}

func NewRuleService(ruleRepo repository.RuleRepositoryInterface, transactionRepo repository.TransactionRepositoryInterface, db database.DatabaseManager) RuleServiceInterface {
	return &ruleService{
		ruleRepo:        ruleRepo,
		transactionRepo: transactionRepo,
		db:              db,
		validator:       &validator.RuleValidator{},
	}
}

func (s *ruleService) CreateRule(c *gin.Context, ruleReq models.CreateRuleRequest) (models.DescribeRuleResponse, error) {
	logger.Debugf("Creating rule for user %d", ruleReq.Rule.CreatedBy)
	var ruleResponse models.DescribeRuleResponse
	if err := s.validator.Validate(ruleReq); err != nil {
		return ruleResponse, err
	}

	err := s.db.WithTxn(c, func(tx pgx.Tx) error {
		rule, err := s.ruleRepo.CreateRule(c, ruleReq.Rule)
		if err != nil {
			return err
		}

		for i := range ruleReq.Actions {
			ruleReq.Actions[i].RuleId = rule.Id
		}
		for i := range ruleReq.Conditions {
			ruleReq.Conditions[i].RuleId = rule.Id
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

	logger.Debugf("Rule created successfully with Id %d", ruleResponse.Rule.Id)
	return ruleResponse, nil
}

func (s *ruleService) GetRuleById(c *gin.Context, id int64, userId int64) (models.DescribeRuleResponse, error) {
	logger.Debugf("Fetching rule %d for user %d", id, userId)
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

	logger.Debugf("Rule %d fetched successfully", id)
	return ruleResponse, nil
}

func (s *ruleService) ListRules(c *gin.Context, userId int64) ([]models.RuleResponse, error) {
	logger.Debugf("Fetching all rules for user %d", userId)
	rules, err := s.ruleRepo.ListRules(c, userId)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Fetched %d rules for user %d", len(rules), userId)
	return rules, nil
}

func (s *ruleService) UpdateRule(c *gin.Context, id int64, ruleReq models.UpdateRuleRequest, userId int64) (models.RuleResponse, error) {
	logger.Debugf("Updating rule %d for user %d", id, userId)
	if err := s.validator.ValidateUpdate(ruleReq); err != nil {
		return models.RuleResponse{}, err
	}
	rule, err := s.ruleRepo.UpdateRule(c, id, userId, ruleReq)
	if err != nil {
		return models.RuleResponse{}, err
	}
	return rule, nil
}

func (s *ruleService) UpdateRuleAction(c *gin.Context, id int64, ruleId int64, ruleReq models.UpdateRuleActionRequest, userId int64) (models.RuleActionResponse, error) {
	logger.Debugf("Updating rule action %d for user %d", id, userId)
	if err := s.validator.ValidateUpdateAction(ruleReq); err != nil {
		return models.RuleActionResponse{}, err
	}
	rule, err := s.ruleRepo.GetRule(c, ruleId, userId)
	if err != nil {
		return models.RuleActionResponse{}, err
	}
	ruleAction, err := s.ruleRepo.UpdateRuleAction(c, id, rule.Id, ruleReq)
	if err != nil {
		return models.RuleActionResponse{}, err
	}
	return ruleAction, nil
}

func (s *ruleService) UpdateRuleCondition(c *gin.Context, id int64, ruleId int64, ruleReq models.UpdateRuleConditionRequest, userId int64) (models.RuleConditionResponse, error) {
	logger.Debugf("Updating rule condition %d for user %d", id, userId)
	if err := s.validator.ValidateUpdateCondition(ruleReq); err != nil {
		return models.RuleConditionResponse{}, err
	}
	rule, err := s.ruleRepo.GetRule(c, ruleId, userId)
	if err != nil {
		return models.RuleConditionResponse{}, err
	}
	ruleCondition, err := s.ruleRepo.UpdateRuleCondition(c, id, rule.Id, ruleReq)
	if err != nil {
		return models.RuleConditionResponse{}, err
	}
	return ruleCondition, nil
}

func (s *ruleService) DeleteRule(c *gin.Context, id int64, userId int64) error {
	logger.Debugf("Deleting rule %d for user %d", id, userId)
	err := s.db.WithTxn(c, func(tx pgx.Tx) error {
		err := s.ruleRepo.DeleteRuleActionsByRuleId(c, id)
		if err != nil {
			return err
		}
		err = s.ruleRepo.DeleteRuleConditionsByRuleId(c, id)
		if err != nil {
			return err
		}
		err = s.ruleRepo.DeleteRule(c, id, userId)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	logger.Debugf("Rule %d deleted successfully", id)
	return nil
}
