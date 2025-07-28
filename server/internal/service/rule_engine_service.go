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

	// Step 1: Fetch all categories
	categories, err := s.categoryRepo.ListCategories(c, userId)
	if err != nil {
		return models.ExecuteRulesResponse{}, fmt.Errorf("failed to fetch categories: %w", err)
	}

	// Step 2: Fetch rules - use specific rules if provided, otherwise fetch all
	var rules []models.DescribeRuleResponse
	if request.RuleIds != nil && len(*request.RuleIds) > 0 {
		rules, err = s.fetchSpecificRules(c, userId, *request.RuleIds)
	} else {
		rules, err = s.fetchAllUserRules(c, userId)
	}
	if err != nil {
		return models.ExecuteRulesResponse{}, fmt.Errorf("failed to fetch rules: %w", err)
	}

	if len(rules) == 0 {
		return models.ExecuteRulesResponse{TotalRules: 0, ProcessedTxns: 0}, nil
	}

	// Create rule engine with categories and rules
	engine := NewRuleEngine(categories, rules)

	pageSize := request.PageSize
	if pageSize <= 0 || pageSize > 1000 {
		pageSize = 100
	}

	var allChangesets []*Changeset
	totalProcessed := 0

	// Step 3: Process transactions - use specific transactions if provided, otherwise fetch all in pages
	if request.TransactionIds != nil && len(*request.TransactionIds) > 0 {
		transactions, err := s.fetchSpecificTransactions(c, userId, *request.TransactionIds)
		if err != nil {
			return models.ExecuteRulesResponse{}, fmt.Errorf("failed to fetch specific transactions: %w", err)
		}

		changesets := s.processTransactions(engine, transactions)
		allChangesets = append(allChangesets, changesets...)
		totalProcessed = len(transactions)
	} else {
		page := 1
		for {
			transactions, err := s.fetchTransactionPage(c, userId, page, pageSize)
			if err != nil {
				return models.ExecuteRulesResponse{}, fmt.Errorf("failed to fetch transactions page %d: %w", page, err)
			}

			if len(transactions) == 0 {
				break
			}

			changesets := s.processTransactions(engine, transactions)
			allChangesets = append(allChangesets, changesets...)
			totalProcessed += len(transactions)

			if len(transactions) < pageSize {
				break
			}
			page++
		}
	}

	// Step 4: Apply changesets
	modified, err := s.applyChangesets(c, userId, allChangesets)
	if err != nil {
		return models.ExecuteRulesResponse{}, fmt.Errorf("failed to apply changesets: %w", err)
	}

	response := models.ExecuteRulesResponse{
		Modified:      modified,
		Skipped:       []models.SkippedResult{},
		TotalRules:    len(rules),
		ProcessedTxns: totalProcessed,
	}

	logger.Infof("Rule execution completed for user %d: %d modified, %d total processed",
		userId, len(modified), totalProcessed)
	return response, nil
}

func (s *ruleEngineService) buildRuleResponse(c *gin.Context, rule models.RuleResponse) (*models.DescribeRuleResponse, error) {
	actions, err := s.ruleRepo.ListRuleActionsByRuleId(c, rule.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get actions for rule %d: %w", rule.Id, err)
	}

	conditions, err := s.ruleRepo.ListRuleConditionsByRuleId(c, rule.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get conditions for rule %d: %w", rule.Id, err)
	}

	return &models.DescribeRuleResponse{
		Rule:       rule,
		Actions:    actions,
		Conditions: conditions,
	}, nil
}

func (s *ruleEngineService) fetchSpecificRules(c *gin.Context, userId int64, ruleIds []int64) ([]models.DescribeRuleResponse, error) {
	var rules []models.DescribeRuleResponse
	for _, ruleId := range ruleIds {
		rule, err := s.ruleRepo.GetRule(c, ruleId, userId)
		if err != nil {
			logger.Warnf("Rule %d not found for user %d: %v", ruleId, userId, err)
			continue
		}

		ruleResponse, err := s.buildRuleResponse(c, rule)
		if err != nil {
			logger.Warnf("Failed to build rule response for rule %d: %v", rule.Id, err)
			continue
		}

		rules = append(rules, *ruleResponse)
	}
	return rules, nil
}

func (s *ruleEngineService) fetchAllUserRules(c *gin.Context, userId int64) ([]models.DescribeRuleResponse, error) {
	allRules, err := s.ruleRepo.ListRules(c, userId)
	if err != nil {
		return nil, err
	}

	var rules []models.DescribeRuleResponse
	for _, rule := range allRules {
		if rule.EffectiveFrom.After(time.Now()) {
			continue
		}

		ruleResponse, err := s.buildRuleResponse(c, rule)
		if err != nil {
			logger.Warnf("Failed to build rule response for rule %d: %v", rule.Id, err)
			continue
		}

		rules = append(rules, *ruleResponse)
	}

	return rules, nil
}

func (s *ruleEngineService) fetchSpecificTransactions(c *gin.Context, userId int64, transactionIds []int64) ([]models.TransactionResponse, error) {
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

func (s *ruleEngineService) fetchTransactionPage(c *gin.Context, userId int64, page, pageSize int) ([]models.TransactionResponse, error) {
	query := models.TransactionListQuery{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    "date",
		SortOrder: "desc",
	}

	result, err := s.transactionRepo.ListTransactions(c, userId, query)
	if err != nil {
		return nil, err
	}

	return result.Transactions, nil
}

func (s *ruleEngineService) processTransactions(engine *RuleEngine, transactions []models.TransactionResponse) []*Changeset {
	var changesets []*Changeset
	for _, transaction := range transactions {
		if changeset := engine.ProcessTransaction(transaction); changeset != nil {
			changesets = append(changesets, changeset)
		}
	}
	return changesets
}

func (s *ruleEngineService) applyChangesets(c *gin.Context, userId int64, changesets []*Changeset) ([]models.ModifiedResult, error) {
	var modified []models.ModifiedResult

	for _, changeset := range changesets {
		err := s.applyChangeset(c, userId, changeset)
		if err != nil {
			logger.Errorf("Failed to apply changeset to transaction %d: %v", changeset.TransactionId, err)
			continue
		}

		// TODO: Add rule_txn mapping table tracking here
		// s.trackRuleApplication(c, changeset)

		modified = append(modified, models.ModifiedResult{
			TransactionId: changeset.TransactionId,
			AppliedRules:  changeset.AppliedRules,
			UpdatedFields: s.getUpdatedFields(changeset),
		})
	}

	return modified, nil
}

func (s *ruleEngineService) applyChangeset(c *gin.Context, userId int64, changeset *Changeset) error {
	transaction, err := s.transactionRepo.GetTransactionById(c, changeset.TransactionId, userId)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	// Apply base field updates
	if changeset.NameUpdate != nil || changeset.DescUpdate != nil {
		updateInput := models.UpdateBaseTransactionInput{}
		if changeset.NameUpdate != nil {
			updateInput.Name = *changeset.NameUpdate
		}
		if changeset.DescUpdate != nil {
			updateInput.Description = changeset.DescUpdate
		}

		err = s.transactionRepo.UpdateTransaction(c, changeset.TransactionId, transaction.CreatedBy, updateInput)
		if err != nil {
			return fmt.Errorf("failed to update transaction: %w", err)
		}
	}

	// Apply category updates
	if len(changeset.CategoryAdds) > 0 {
		newCategoryIds := append(transaction.CategoryIds, changeset.CategoryAdds...)
		err = s.transactionRepo.UpdateCategoryMapping(c, changeset.TransactionId, transaction.CreatedBy, newCategoryIds)
		if err != nil {
			return fmt.Errorf("failed to update category mapping: %w", err)
		}
	}

	return nil
}

func (s *ruleEngineService) getUpdatedFields(changeset *Changeset) []models.RuleFieldType {
	var fields []models.RuleFieldType
	if changeset.NameUpdate != nil {
		fields = append(fields, models.RuleFieldName)
	}
	if changeset.DescUpdate != nil {
		fields = append(fields, models.RuleFieldDescription)
	}
	if len(changeset.CategoryAdds) > 0 {
		fields = append(fields, models.RuleFieldCategory)
	}
	return fields
}
