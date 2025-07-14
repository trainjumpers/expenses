package service

import (
	"expenses/internal/models"
	"expenses/internal/repository"
	"expenses/pkg/logger"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type RuleEngineServiceInterface interface {
	ExecuteRules(c *gin.Context, userId int64, request models.ExecuteRulesRequest) (models.ExecuteRulesResponse, error)
	ExecuteRulesForTransaction(c *gin.Context, transactionId int64, userId int64) (models.ExecuteRulesResponse, error)
	ExecuteRulesForRule(c *gin.Context, ruleId int64, userId int64) (models.ExecuteRulesResponse, error)
}

type ruleEngineService struct {
	ruleRepo        repository.RuleRepositoryInterface
	transactionRepo repository.TransactionRepositoryInterface
	categoryRepo    repository.CategoryRepositoryInterface
}

func NewRuleEngineService(
	ruleRepo repository.RuleRepositoryInterface,
	transactionRepo repository.TransactionRepositoryInterface,
	categoryRepo repository.CategoryRepositoryInterface,
) RuleEngineServiceInterface {
	return &ruleEngineService{
		ruleRepo:        ruleRepo,
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
	}
}

func (s *ruleEngineService) ExecuteRules(c *gin.Context, userId int64, request models.ExecuteRulesRequest) (models.ExecuteRulesResponse, error) {
	logger.Infof("Executing rules for user %d", userId)

	categories, err := s.categoryRepo.ListCategories(c, userId)
	if err != nil {
		return models.ExecuteRulesResponse{}, fmt.Errorf("failed to fetch categories: %w", err)
	}

	rules, err := s.getRulesForExecution(c, userId, request.RuleIds)
	if err != nil {
		return models.ExecuteRulesResponse{}, fmt.Errorf("failed to fetch rules: %w", err)
	}

	if len(rules) == 0 {
		logger.Infof("No rules found for user %d", userId)
		return models.ExecuteRulesResponse{TotalRules: 0, ProcessedTxns: 0}, nil
	}

	pageSize := request.PageSize
	if pageSize <= 0 || pageSize > 1000 {
		pageSize = 100
	}

	var allChangesets []RuleChangeset
	var allSkipped []models.SkippedResult
	totalProcessed := 0

	engine := NewRuleEngine(categories)

	if request.TransactionIds != nil && len(*request.TransactionIds) > 0 {
		transactions, err := s.getSpecificTransactions(c, userId, *request.TransactionIds)
		if err != nil {
			return models.ExecuteRulesResponse{}, fmt.Errorf("failed to fetch specific transactions: %w", err)
		}

		result := engine.ExecuteRules(rules, transactions)
		allChangesets = append(allChangesets, result.Changesets...)
		allSkipped = append(allSkipped, result.Skipped...)
		totalProcessed = len(transactions)
	} else {
		page := 1
		for {
			query := models.TransactionListQuery{
				Page:      page,
				PageSize:  pageSize,
				SortBy:    "date",
				SortOrder: "desc",
			}

			result, err := s.transactionRepo.ListTransactions(c, userId, query)
			fmt.Println("Transactions", result)
			if err != nil {
				return models.ExecuteRulesResponse{}, fmt.Errorf("failed to fetch transactions page %d: %w", page, err)
			}

			if len(result.Transactions) == 0 {
				break
			}

			engineResult := engine.ExecuteRules(rules, result.Transactions)
			allChangesets = append(allChangesets, engineResult.Changesets...)
			allSkipped = append(allSkipped, engineResult.Skipped...)
			totalProcessed += len(result.Transactions)

			if len(result.Transactions) < pageSize {
				break
			}
			page++
		}
	}

	fmt.Println("Changesets", allChangesets)
	modified, err := s.applyChangesets(c, allChangesets)
	if err != nil {
		return models.ExecuteRulesResponse{}, fmt.Errorf("failed to apply changesets: %w", err)
	}

	response := models.ExecuteRulesResponse{
		Modified:      modified,
		Skipped:       allSkipped,
		TotalRules:    len(rules),
		ProcessedTxns: totalProcessed,
	}

	logger.Infof("Rule execution completed for user %d: %d modified, %d skipped, %d total processed",
		userId, len(modified), len(allSkipped), totalProcessed)
	return response, nil
}

func (s *ruleEngineService) ExecuteRulesForTransaction(c *gin.Context, transactionId int64, userId int64) (models.ExecuteRulesResponse, error) {
	logger.Infof("Executing rules for transaction %d, user %d", transactionId, userId)

	transaction, err := s.transactionRepo.GetTransactionById(c, transactionId, userId)
	if err != nil {
		return models.ExecuteRulesResponse{}, fmt.Errorf("failed to fetch transaction: %w", err)
	}

	categories, err := s.categoryRepo.ListCategories(c, userId)
	if err != nil {
		return models.ExecuteRulesResponse{}, fmt.Errorf("failed to fetch categories: %w", err)
	}

	rules, err := s.getRulesForExecution(c, userId, nil)
	if err != nil {
		return models.ExecuteRulesResponse{}, fmt.Errorf("failed to fetch rules: %w", err)
	}

	engine := NewRuleEngine(categories)
	result := engine.ExecuteRules(rules, []models.TransactionResponse{transaction})

	modified, err := s.applyChangesets(c, result.Changesets)
	if err != nil {
		return models.ExecuteRulesResponse{}, fmt.Errorf("failed to apply changesets: %w", err)
	}

	return models.ExecuteRulesResponse{
		Modified:      modified,
		Skipped:       result.Skipped,
		TotalRules:    len(rules),
		ProcessedTxns: 1,
	}, nil
}

func (s *ruleEngineService) ExecuteRulesForRule(c *gin.Context, ruleId int64, userId int64) (models.ExecuteRulesResponse, error) {
	logger.Infof("Executing rule %d for user %d", ruleId, userId)

	ruleIds := []int64{ruleId}
	request := models.ExecuteRulesRequest{
		RuleIds:  &ruleIds,
		PageSize: 100,
	}

	return s.ExecuteRules(c, userId, request)
}

func (s *ruleEngineService) getRulesForExecution(c *gin.Context, userId int64, ruleIds *[]int64) ([]models.DescribeRuleResponse, error) {
	if ruleIds != nil && len(*ruleIds) > 0 {
		var rules []models.DescribeRuleResponse
		for _, ruleId := range *ruleIds {
			rule, err := s.ruleRepo.GetRule(c, ruleId, userId)
			if err != nil {
				logger.Warnf("Rule %d not found for user %d: %v", ruleId, userId, err)
				continue
			}

			actions, err := s.ruleRepo.ListRuleActionsByRuleId(c, ruleId)
			if err != nil {
				logger.Warnf("Failed to get actions for rule %d: %v", ruleId, err)
				continue
			}

			conditions, err := s.ruleRepo.ListRuleConditionsByRuleId(c, ruleId)
			if err != nil {
				logger.Warnf("Failed to get conditions for rule %d: %v", ruleId, err)
				continue
			}

			rules = append(rules, models.DescribeRuleResponse{
				Rule:       rule,
				Actions:    actions,
				Conditions: conditions,
			})
		}
		return rules, nil
	}

	allRules, err := s.ruleRepo.ListRules(c, userId)
	if err != nil {
		return nil, err
	}

	var rules []models.DescribeRuleResponse
	for _, rule := range allRules {
		if rule.EffectiveFrom.After(time.Now()) {
			continue
		}

		actions, err := s.ruleRepo.ListRuleActionsByRuleId(c, rule.Id)
		if err != nil {
			logger.Warnf("Failed to get actions for rule %d: %v", rule.Id, err)
			continue
		}

		conditions, err := s.ruleRepo.ListRuleConditionsByRuleId(c, rule.Id)
		if err != nil {
			logger.Warnf("Failed to get conditions for rule %d: %v", rule.Id, err)
			continue
		}

		rules = append(rules, models.DescribeRuleResponse{
			Rule:       rule,
			Actions:    actions,
			Conditions: conditions,
		})
	}

	return rules, nil
}

func (s *ruleEngineService) getSpecificTransactions(c *gin.Context, userId int64, transactionIds []int64) ([]models.TransactionResponse, error) {
	var transactions []models.TransactionResponse
	for _, txnId := range transactionIds {
		txn, err := s.transactionRepo.GetTransactionById(c, txnId, userId)
		if err != nil {
			logger.Warnf("Transaction %d not found for user %d: %v", txnId, userId, err)
			continue
		}
		transactions = append(transactions, txn)
	}
	return transactions, nil
}

func (s *ruleEngineService) applyChangesets(c *gin.Context, changesets []RuleChangeset) ([]models.ModifiedResult, error) {
	var modified []models.ModifiedResult

	for _, changeset := range changesets {
		err := s.applyChangesetToTransaction(c, changeset)
		if err != nil {
			fmt.Println("Error applying changeset:", err)
			logger.Errorf("Failed to apply changeset to transaction %d: %v", changeset.TransactionId, err)
			continue
		}

		modified = append(modified, models.ModifiedResult{
			TransactionId: changeset.TransactionId,
			AppliedRules:  changeset.AppliedRules,
			UpdatedFields: changeset.UpdatedFields,
		})
	}

	fmt.Println("Modified", modified)
	return modified, nil
}

func (s *ruleEngineService) applyChangesetToTransaction(c *gin.Context, changeset RuleChangeset) error {
	hasBaseUpdates := changeset.NameUpdate != nil || changeset.DescUpdate != nil
	hasCategoryUpdates := len(changeset.CategoryAdds) > 0

	if hasBaseUpdates {
		updateInput := models.UpdateBaseTransactionInput{}
		if changeset.NameUpdate != nil {
			updateInput.Name = *changeset.NameUpdate
		}
		if changeset.DescUpdate != nil {
			updateInput.Description = changeset.DescUpdate
		}

		transaction, err := s.transactionRepo.GetTransactionById(c, changeset.TransactionId, 0)
		fmt.Println("Transaction", transaction)
		if err != nil {
			return fmt.Errorf("failed to get transaction for update: %w", err)
		}

		err = s.transactionRepo.UpdateTransaction(c, changeset.TransactionId, transaction.CreatedBy, updateInput)
		if err != nil {
			return fmt.Errorf("failed to update transaction: %w", err)
		}
	}

	if hasCategoryUpdates {
		transaction, err := s.transactionRepo.GetTransactionById(c, changeset.TransactionId, 0)
		if err != nil {
			return fmt.Errorf("failed to get transaction for category update: %w", err)
		}

		newCategoryIds := append(transaction.CategoryIds, changeset.CategoryAdds...)

		err = s.transactionRepo.UpdateCategoryMapping(c, changeset.TransactionId, transaction.CreatedBy, newCategoryIds)
		if err != nil {
			return fmt.Errorf("failed to update category mapping: %w", err)
		}
	}

	return nil
}
