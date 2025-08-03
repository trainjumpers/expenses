package service

import (
	"context"
	"expenses/internal/models"
	"expenses/internal/repository"
	"expenses/internal/validator"
	database "expenses/pkg/database/manager"
	"expenses/pkg/logger"
)

type RuleServiceInterface interface {
	CreateRule(ctx context.Context, ruleReq models.CreateRuleRequest) (models.DescribeRuleResponse, error)
	GetRuleById(ctx context.Context, id int64, userId int64) (models.DescribeRuleResponse, error)
	ListRules(ctx context.Context, userId int64) ([]models.RuleResponse, error)
	UpdateRule(ctx context.Context, id int64, ruleReq models.UpdateRuleRequest, userId int64) (models.RuleResponse, error)
	UpdateRuleAction(ctx context.Context, id int64, ruleId int64, ruleReq models.UpdateRuleActionRequest, userId int64) (models.RuleActionResponse, error)
	UpdateRuleCondition(ctx context.Context, id int64, ruleId int64, ruleReq models.UpdateRuleConditionRequest, userId int64) (models.RuleConditionResponse, error)
	DeleteRule(ctx context.Context, id int64, userId int64) error
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

func (s *ruleService) CreateRule(ctx context.Context, ruleReq models.CreateRuleRequest) (models.DescribeRuleResponse, error) {
	logger.Debugf("Creating rule for user %d", ruleReq.Rule.CreatedBy)
	var ruleResponse models.DescribeRuleResponse
	if err := s.validator.Validate(ruleReq); err != nil {
		return ruleResponse, err
	}

	err := s.db.WithTxn(ctx, func(txCtx context.Context) error {
		rule, err := s.ruleRepo.CreateRule(txCtx, ruleReq.Rule)
		if err != nil {
			return err
		}

		for i := range ruleReq.Actions {
			ruleReq.Actions[i].RuleId = rule.Id
		}
		for i := range ruleReq.Conditions {
			ruleReq.Conditions[i].RuleId = rule.Id
		}

		actions, err := s.ruleRepo.CreateRuleActions(txCtx, ruleReq.Actions)
		if err != nil {
			return err
		}

		conditions, err := s.ruleRepo.CreateRuleConditions(txCtx, ruleReq.Conditions)
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

func (s *ruleService) GetRuleById(ctx context.Context, id int64, userId int64) (models.DescribeRuleResponse, error) {
	logger.Debugf("Fetching rule %d for user %d", id, userId)
	var ruleResponse models.DescribeRuleResponse
	rule, err := s.ruleRepo.GetRule(ctx, id, userId)
	if err != nil {
		return ruleResponse, err
	}
	ruleResponse.Rule = rule

	actions, err := s.ruleRepo.ListRuleActionsByRuleId(ctx, id)
	if err != nil {
		return ruleResponse, err
	}
	ruleResponse.Actions = actions

	conditions, err := s.ruleRepo.ListRuleConditionsByRuleId(ctx, id)
	if err != nil {
		return ruleResponse, err
	}
	ruleResponse.Conditions = conditions

	logger.Debugf("Rule %d fetched successfully", id)
	return ruleResponse, nil
}

func (s *ruleService) ListRules(ctx context.Context, userId int64) ([]models.RuleResponse, error) {
	logger.Debugf("Fetching all rules for user %d", userId)
	rules, err := s.ruleRepo.ListRules(ctx, userId)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Fetched %d rules for user %d", len(rules), userId)
	return rules, nil
}

func (s *ruleService) UpdateRule(ctx context.Context, id int64, ruleReq models.UpdateRuleRequest, userId int64) (models.RuleResponse, error) {
	logger.Debugf("Updating rule %d for user %d", id, userId)
	if err := s.validator.ValidateUpdate(ruleReq); err != nil {
		return models.RuleResponse{}, err
	}
	rule, err := s.ruleRepo.UpdateRule(ctx, id, userId, ruleReq)
	if err != nil {
		return models.RuleResponse{}, err
	}
	return rule, nil
}

func (s *ruleService) UpdateRuleAction(ctx context.Context, id int64, ruleId int64, ruleReq models.UpdateRuleActionRequest, userId int64) (models.RuleActionResponse, error) {
	logger.Debugf("Updating rule action %d for user %d", id, userId)
	if err := s.validator.ValidateUpdateAction(ruleReq); err != nil {
		return models.RuleActionResponse{}, err
	}
	rule, err := s.ruleRepo.GetRule(ctx, ruleId, userId)
	if err != nil {
		return models.RuleActionResponse{}, err
	}
	ruleAction, err := s.ruleRepo.UpdateRuleAction(ctx, id, rule.Id, ruleReq)
	if err != nil {
		return models.RuleActionResponse{}, err
	}
	return ruleAction, nil
}

func (s *ruleService) UpdateRuleCondition(ctx context.Context, id int64, ruleId int64, ruleReq models.UpdateRuleConditionRequest, userId int64) (models.RuleConditionResponse, error) {
	logger.Debugf("Updating rule condition %d for user %d", id, userId)
	if err := s.validator.ValidateUpdateCondition(ruleReq); err != nil {
		return models.RuleConditionResponse{}, err
	}
	rule, err := s.ruleRepo.GetRule(ctx, ruleId, userId)
	if err != nil {
		return models.RuleConditionResponse{}, err
	}
	ruleCondition, err := s.ruleRepo.UpdateRuleCondition(ctx, id, rule.Id, ruleReq)
	if err != nil {
		return models.RuleConditionResponse{}, err
	}
	return ruleCondition, nil
}

func (s *ruleService) DeleteRule(ctx context.Context, id int64, userId int64) error {
	logger.Debugf("Deleting rule %d for user %d", id, userId)
	err := s.db.WithTxn(ctx, func(txCtx context.Context) error {
		err := s.ruleRepo.DeleteRuleActionsByRuleId(txCtx, id)
		if err != nil {
			return err
		}
		err = s.ruleRepo.DeleteRuleConditionsByRuleId(txCtx, id)
		if err != nil {
			return err
		}
		err = s.ruleRepo.DeleteRule(txCtx, id, userId)
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
