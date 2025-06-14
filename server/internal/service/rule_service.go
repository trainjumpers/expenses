package service

import (
	"expenses/internal/models"
	"expenses/internal/repository"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type RuleServiceInterface interface {
	CreateRule(ctx *gin.Context, ruleReq *models.CreateRuleRequest) (*models.RuleResponse, error)
	GetRuleByID(ctx *gin.Context, id int64) (*models.RuleResponse, error)
	ListRules(ctx *gin.Context) ([]*models.RuleResponse, error)
	UpdateRule(ctx *gin.Context, id int64, ruleReq *models.UpdateRuleRequest) error
	DeleteRule(ctx *gin.Context, id int64) error
	ExecuteRules(ctx *gin.Context, userId int64) (*models.ExecuteRulesResponse, error)
}

type ruleService struct {
	ruleRepo        repository.RuleRepositoryInterface
	transactionRepo repository.TransactionRepositoryInterface
}

func NewRuleService(ruleRepo repository.RuleRepositoryInterface, transactionRepo repository.TransactionRepositoryInterface) RuleServiceInterface {
	return &ruleService{
		ruleRepo:        ruleRepo,
		transactionRepo: transactionRepo,
	}
}

func (s *ruleService) CreateRule(ctx *gin.Context, ruleReq *models.CreateRuleRequest) (*models.RuleResponse, error) {
	return s.ruleRepo.CreateRule(ctx, ruleReq)
}

func (s *ruleService) GetRuleByID(ctx *gin.Context, id int64) (*models.RuleResponse, error) {
	return s.ruleRepo.GetRuleByID(ctx, id)
}

func (s *ruleService) ListRules(ctx *gin.Context) ([]*models.RuleResponse, error) {
	return s.ruleRepo.ListRules(ctx)
}

func (s *ruleService) UpdateRule(ctx *gin.Context, id int64, ruleReq *models.UpdateRuleRequest) error {
	return s.ruleRepo.UpdateRule(ctx, id, ruleReq)
}

func (s *ruleService) DeleteRule(ctx *gin.Context, id int64) error {
	return s.ruleRepo.DeleteRule(ctx, id)
}

func (s *ruleService) ExecuteRules(ctx *gin.Context, userId int64) (*models.ExecuteRulesResponse, error) {
	// 1. Fetch all rules
	rules, err := s.ruleRepo.ListRules(ctx)
	if err != nil {
		return nil, err
	}

	// 2. Fetch target transactions
	var transactions []models.TransactionResponse
	transactions, err = s.transactionRepo.ListTransactions(ctx, userId)
	if err != nil {
		return nil, err
	}

	var modified []models.ModifiedResult
	var skipped []models.SkippedResult

	for _, txn := range transactions {
		var appliedRules []int64
		var updatedFields []models.RuleFieldType
		var updateInput models.UpdateBaseTransactionInput
		var updateCategoryIds *[]int64

		for _, rule := range rules {
			if ruleMatchesTransaction(rule, txn) {
				appliedRules = append(appliedRules, rule.ID)
				for _, action := range rule.Actions {
					switch action.ActionType {
					case models.RuleFieldCategory:
						if updateCategoryIds == nil {
							updateCategoryIds = &[]int64{}
						}
						catID, err := strconv.ParseInt(action.ActionValue, 10, 64)
						if err == nil {
							*updateCategoryIds = append(*updateCategoryIds, catID)
							updatedFields = append(updatedFields, models.RuleFieldCategory)
						}
					case models.RuleFieldName:
						updateInput.Name = action.ActionValue
						updatedFields = append(updatedFields, models.RuleFieldName)
					case models.RuleFieldDescription:
						desc := action.ActionValue
						updateInput.Description = &desc
						updatedFields = append(updatedFields, models.RuleFieldDescription)
					case models.RuleFieldAmount:
						amt, err := strconv.ParseFloat(action.ActionValue, 64)
						if err == nil {
							updateInput.Amount = &amt
							updatedFields = append(updatedFields, models.RuleFieldAmount)
						}
					}
				}
			}
		}

		if len(appliedRules) > 0 {
			// 4. Update transaction in DB
			err := s.transactionRepo.UpdateTransaction(ctx, txn.Id, txn.CreatedBy, updateInput)
			if err != nil {
				skipped = append(skipped, models.SkippedResult{
					TransactionID: txn.Id,
					Reason:        "DB update failed: " + err.Error(),
				})
				continue
			}
			if updateCategoryIds != nil {
				_ = s.transactionRepo.UpdateCategoryMapping(ctx, txn.Id, txn.CreatedBy, *updateCategoryIds)
			}
			modified = append(modified, models.ModifiedResult{
				TransactionID: txn.Id,
				AppliedRules:  appliedRules,
				UpdatedFields: updatedFields,
			})
		} else {
			skipped = append(skipped, models.SkippedResult{
				TransactionID: txn.Id,
				Reason:        "No matching rule",
			})
		}
	}

	return &models.ExecuteRulesResponse{
		Modified: modified,
		Skipped:  skipped,
	}, nil
}

func ruleMatchesTransaction(rule *models.RuleResponse, txn models.TransactionResponse) bool {
	for _, cond := range rule.Conditions {
		if !evaluateCondition(cond, txn) {
			return false
		}
	}
	return true
}

func evaluateCondition(cond models.RuleConditionResponse, txn models.TransactionResponse) bool {
	switch cond.ConditionType {
	case models.RuleFieldAmount:
		condVal, err := strconv.ParseFloat(cond.ConditionValue, 64)
		if err != nil {
			return false
		}
		switch cond.ConditionOperator {
		case models.OperatorEquals:
			return txn.Amount == condVal
		case models.OperatorGreater:
			return txn.Amount > condVal
		case models.OperatorLower:
			return txn.Amount < condVal
		}
	case models.RuleFieldName:
		return stringCompare(txn.Name, cond.ConditionValue, cond.ConditionOperator)
	case models.RuleFieldDescription:
		if txn.Description == nil {
			return false
		}
		return stringCompare(*txn.Description, cond.ConditionValue, cond.ConditionOperator)
	case models.RuleFieldCategory:
		for _, catID := range txn.CategoryIds {
			if cond.ConditionOperator == models.OperatorEquals {
				catCond, err := strconv.ParseInt(cond.ConditionValue, 10, 64)
				if err == nil && catID == catCond {
					return true
				}
			}
		}
		return false
	}
	return false
}

func stringCompare(field, condValue string, op models.RuleOperator) bool {
	switch op {
	case models.OperatorEquals:
		return field == condValue
	case models.OperatorContains:
		return strings.Contains(field, condValue)
	}
	return false
}
