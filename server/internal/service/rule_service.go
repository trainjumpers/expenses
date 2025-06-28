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

// func (s *ruleService) ExecuteRules(c *gin.Context, userId int64) (*models.ExecuteRulesResponse, error) {
// 	logger.Infof("Executing rules for user %d", userId)

// 	rules, err := s.ListRules(c, userId)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if len(rules) == 0 {
// 		logger.Infof("No rules found for user %d", userId)
// 		return &models.ExecuteRulesResponse{Modified: []models.ModifiedResult{}, Skipped: []models.SkippedResult{}}, nil
// 	}

// 	effectiveRules := s.filterEffectiveRules(rules)
// 	if len(effectiveRules) == 0 {
// 		logger.Infof("No effective rules found for user %d", userId)
// 		return &models.ExecuteRulesResponse{Modified: []models.ModifiedResult{}, Skipped: []models.SkippedResult{}}, nil
// 	}

// 	for i := range effectiveRules {
// 		rule, err := s.GetRuleById(c, effectiveRules[i].Id, userId)
// 		if err != nil {
// 			logger.Warnf("Could not fetch details for rule %d, skipping: %v", effectiveRules[i].Id, err)
// 			continue
// 		}
// 		effectiveRules[i] = rule
// 	}

// 	transactions, err := s.transactionRepo.ListTransactions(c, userId, models.TransactionListQuery{})
// 	if err != nil {
// 		return nil, err
// 	}

// 	var modified []models.ModifiedResult
// 	var skipped []models.SkippedResult

// 	for _, txn := range transactions.Transactions {
// 		result := s.processTransactionWithRules(c, txn, effectiveRules)
// 		if result.Modified {
// 			modified = append(modified, result.ModifiedResult)
// 		} else {
// 			skipped = append(skipped, result.SkippedResult)
// 		}
// 	}

// 	logger.Infof("Rules execution completed for user %d. Modified: %d, Skipped: %d",
// 		userId, len(modified), len(skipped))

// 	return &models.ExecuteRulesResponse{
// 		Modified: modified,
// 		Skipped:  skipped,
// 	}, nil
// }

// func (s *ruleService) filterEffectiveRules(rules []models.RuleResponse) []models.RuleResponse {
// 	var effectiveRules []models.RuleResponse
// 	now := time.Now()

// 	for _, rule := range rules {
// 		if !rule.EffectiveFrom.After(now) {
// 			effectiveRules = append(effectiveRules, rule)
// 		} else {
// 			logger.Debugf("Skipping rule %d as it's not yet effective (effective from: %s)",
// 				rule.Id, rule.EffectiveFrom.Format("2006-01-02"))
// 		}
// 	}
// 	return effectiveRules
// }

// func (s *ruleService) processTransactionWithRules(c *gin.Context, txn models.TransactionResponse, rules []models.RuleResponse) ruleExecutionResult {
// 	var appliedRules []int64
// 	var updatedFields []models.RuleFieldType
// 	var updateInput models.UpdateBaseTransactionInput
// 	var updateCategoryIds *[]int64

// 	for _, rule := range rules {
// 		if s.ruleMatchesTransaction(&rule, txn) {
// 			appliedRules = append(appliedRules, rule.Id)
// 			s.applyRuleActions(rule.Actions, &updateInput, &updateCategoryIds, &updatedFields)
// 		}
// 	}

// 	if len(appliedRules) == 0 {
// 		return ruleExecutionResult{
// 			Modified:      false,
// 			SkippedResult: models.SkippedResult{TransactionId: txn.Id, Reason: "No matching rule"},
// 		}
// 	}
// 	err := s.transactionRepo.UpdateTransaction(c, txn.Id, txn.CreatedBy, updateInput)
// 	if err != nil {
// 		return ruleExecutionResult{
// 			Modified:      false,
// 			SkippedResult: models.SkippedResult{TransactionId: txn.Id, Reason: fmt.Sprintf("Failed to update transaction: %v", err)},
// 		}
// 	}

// 	if updateCategoryIds != nil {
// 		err := s.transactionRepo.UpdateCategoryMapping(c, txn.Id, txn.CreatedBy, *updateCategoryIds)
// 		if err != nil {
// 			logger.Warnf("Failed to update category mapping for transaction %d: %v", txn.Id, err)
// 			// Decide if this should be a skipped result
// 		}
// 	}

// 	return ruleExecutionResult{
// 		Modified: true,
// 		ModifiedResult: models.ModifiedResult{
// 			TransactionId: txn.Id,
// 			AppliedRules:  appliedRules,
// 			UpdatedFields: updatedFields,
// 		},
// 	}
// }

// type ruleExecutionResult struct {
// 	Modified       bool
// 	ModifiedResult models.ModifiedResult
// 	SkippedResult  models.SkippedResult
// }

// func (s *ruleService) applyRuleActions(actions []models.RuleActionResponse, updateInput *models.UpdateBaseTransactionInput, updateCategoryIds **[]int64, updatedFields *[]models.RuleFieldType) {
// 	for _, action := range actions {
// 		*updatedFields = append(*updatedFields, action.ActionType)
// 		switch action.ActionType {
// 		case models.RuleFieldName:
// 			updateInput.Name = action.ActionValue
// 		case models.RuleFieldDescription:
// 			updateInput.Description = &action.ActionValue
// 		case models.RuleFieldAmount:
// 			if amount, err := strconv.ParseFloat(action.ActionValue, 64); err == nil {
// 				updateInput.Amount = &amount
// 			}
// 		case models.RuleFieldCategory:
// 			if *updateCategoryIds == nil {
// 				*updateCategoryIds = new([]int64)
// 			}
// 			if catId, err := strconv.ParseInt(action.ActionValue, 10, 64); err == nil {
// 				**updateCategoryIds = append(**updateCategoryIds, catId)
// 			}
// 		}
// 	}
// }

// func (s *ruleService) ruleMatchesTransaction(rule *models.Rule, txn models.TransactionResponse) bool {
// 	for _, cond := range rule.Conditions {
// 		if !s.evaluateCondition(cond, txn) {
// 			return false
// 		}
// 	}
// 	return true
// }

// func (s *ruleService) evaluateCondition(cond models.RuleCondition, txn models.TransactionResponse) bool {
// 	switch cond.ConditionType {
// 	case models.RuleFieldAmount:
// 		return s.evaluateAmountCondition(cond, txn.Amount)
// 	case models.RuleFieldName:
// 		return s.evaluateStringCondition(txn.Name, cond.ConditionValue, cond.ConditionOperator)
// 	case models.RuleFieldDescription:
// 		if txn.Description != nil {
// 			return s.evaluateStringCondition(*txn.Description, cond.ConditionValue, cond.ConditionOperator)
// 		}
// 		return false
// 	case models.RuleFieldCategory:
// 		return s.evaluateCategoryCondition(cond, txn.CategoryIds)
// 	}
// 	return false
// }

// func (s *ruleService) evaluateAmountCondition(cond models.RuleCondition, amount float64) bool {
// 	condAmount, err := strconv.ParseFloat(cond.ConditionValue, 64)
// 	if err != nil {
// 		return false
// 	}
// 	switch cond.ConditionOperator {
// 	case models.OperatorEquals:
// 		return amount == condAmount
// 	case models.OperatorGreater:
// 		return amount > condAmount
// 	case models.OperatorLower:
// 		return amount < condAmount
// 	}
// 	return false
// }

// func (s *ruleService) evaluateStringCondition(field, condValue string, op models.RuleOperator) bool {
// 	switch op {
// 	case models.OperatorEquals:
// 		return strings.EqualFold(field, condValue)
// 	case models.OperatorContains:
// 		return strings.Contains(strings.ToLower(field), strings.ToLower(condValue))
// 	}
// 	return false
// }

// func (s *ruleService) evaluateCategoryCondition(cond models.RuleCondition, categoryIds []int64) bool {
// 	condCatId, err := strconv.ParseInt(cond.ConditionValue, 10, 64)
// 	if err != nil {
// 		return false
// 	}
// 	for _, catId := range categoryIds {
// 		if catId == condCatId {
// 			return true
// 		}
// 	}
// 	return false
// }

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
