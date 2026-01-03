package service

import (
	"context"
	"expenses/internal/models"
	"expenses/internal/repository"
	"expenses/pkg/logger"
	"fmt"
	"time"
)

type RuleEngineServiceInterface interface {
	ExecuteRules(ctx context.Context, userId int64, request models.ExecuteRulesRequest) (models.ExecuteRulesResponse, error)
	ExecuteRulesInBackground(ctx context.Context, userId int64, request models.ExecuteRulesRequest)
}

type ruleEngineService struct {
	ruleRepo        repository.RuleRepositoryInterface
	transactionRepo repository.TransactionRepositoryInterface
	categoryRepo    repository.CategoryRepositoryInterface
	accountRepo     repository.AccountRepositoryInterface
}

func NewRuleEngineService(
	ruleRepo repository.RuleRepositoryInterface,
	transactionRepo repository.TransactionRepositoryInterface,
	categoryRepo repository.CategoryRepositoryInterface,
	accountRepo repository.AccountRepositoryInterface,
) RuleEngineServiceInterface {
	return &ruleEngineService{
		ruleRepo:        ruleRepo,
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
		accountRepo:     accountRepo,
	}
}

func (s *ruleEngineService) ExecuteRules(ctx context.Context, userId int64, request models.ExecuteRulesRequest) (models.ExecuteRulesResponse, error) {
	go s.ExecuteRulesInBackground(context.Background(), userId, request)
	logger.Infof("Rule execution started in background for user %d", userId)
	return models.ExecuteRulesResponse{}, nil
}

func (s *ruleEngineService) ExecuteRulesInBackground(ctx context.Context, userId int64, request models.ExecuteRulesRequest) {
	logger.Infof("Executing rules for user %d", userId)

	// Step 1: Fetch all categories
	categories, err := s.categoryRepo.ListCategories(ctx, userId)
	if err != nil {
		logger.Errorf("Rule execution for user %d failed to fetch categories: %v", userId, err)
		return
	}

	// Step 1.5: Fetch all accounts
	accounts, err := s.accountRepo.ListAccounts(ctx, userId)
	if err != nil {
		logger.Errorf("Rule execution for user %d failed to fetch accounts: %v", userId, err)
		return
	}

	// Step 2: Fetch rules - use specific rules if provided, otherwise fetch all
	var rules []models.DescribeRuleResponse
	if request.RuleIds != nil && len(*request.RuleIds) > 0 {
		rules, err = s.fetchSpecificRules(ctx, userId, *request.RuleIds)
	} else {
		rules, err = s.fetchAllUserRules(ctx, userId)
	}
	if err != nil {
		logger.Errorf("Rule execution for user %d failed to fetch rules: %v", userId, err)
		return
	}

	if len(rules) == 0 {
		logger.Infof("No rules found for user %d, skipping execution.", userId)
		return
	}

	// Create rule engine with categories, accounts and rules
	engine := NewRuleEngine(categories, accounts, rules)

	pageSize := request.PageSize
	if pageSize <= 0 || pageSize > 1000 {
		pageSize = 100
	}

	var allChangesets []*Changeset
	totalProcessed := 0

	// Step 3: Process transactions - use specific transactions if provided, otherwise fetch all in pages
	if request.TransactionIds != nil && len(*request.TransactionIds) > 0 {
		transactions, err := s.fetchSpecificTransactions(ctx, userId, *request.TransactionIds)
		if err != nil {
			logger.Errorf("Rule execution for user %d failed to fetch specific transactions: %v", userId, err)
			return
		}

		changesets := s.processTransactions(engine, transactions)
		allChangesets = append(allChangesets, changesets...)
		totalProcessed = len(transactions)
	} else {
		page := 1
		for {
			transactions, err := s.fetchTransactionPage(ctx, userId, page, pageSize)
			if err != nil {
				logger.Errorf("Rule execution for user %d failed to fetch transactions page %d: %v", userId, page, err)
				return
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
	modified, err := s.applyChangesets(ctx, userId, allChangesets)
	if err != nil {
		logger.Errorf("Rule execution for user %d failed to apply changesets: %v", userId, err)
		return
	}

	logger.Infof("Rule execution completed for user %d: %d modified, %d total processed",
		userId, len(modified), totalProcessed)
}

func (s *ruleEngineService) buildRuleResponse(ctx context.Context, rule models.RuleResponse) (*models.DescribeRuleResponse, error) {
	actions, err := s.ruleRepo.ListRuleActionsByRuleId(ctx, rule.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get actions for rule %d: %w", rule.Id, err)
	}

	conditions, err := s.ruleRepo.ListRuleConditionsByRuleId(ctx, rule.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get conditions for rule %d: %w", rule.Id, err)
	}

	return &models.DescribeRuleResponse{
		Rule:       rule,
		Actions:    actions,
		Conditions: conditions,
	}, nil
}

func (s *ruleEngineService) fetchSpecificRules(ctx context.Context, userId int64, ruleIds []int64) ([]models.DescribeRuleResponse, error) {
	var rules []models.DescribeRuleResponse
	for _, ruleId := range ruleIds {
		rule, err := s.ruleRepo.GetRule(ctx, ruleId, userId)
		if err != nil {
			logger.Warnf("Rule %d not found for user %d: %v", ruleId, userId, err)
			continue
		}

		ruleResponse, err := s.buildRuleResponse(ctx, rule)
		if err != nil {
			logger.Warnf("Failed to build rule response for rule %d: %v", rule.Id, err)
			continue
		}

		rules = append(rules, *ruleResponse)
	}
	return rules, nil
}

func (s *ruleEngineService) fetchAllUserRules(ctx context.Context, userId int64) ([]models.DescribeRuleResponse, error) {
	allRulesResponse, err := s.ruleRepo.ListRules(ctx, userId, models.RuleListQuery{})
	if err != nil {
		return nil, err
	}

	var rules []models.DescribeRuleResponse
	for _, rule := range allRulesResponse.Rules {
		if rule.EffectiveFrom.After(time.Now()) {
			continue
		}

		ruleResponse, err := s.buildRuleResponse(ctx, rule)
		if err != nil {
			logger.Warnf("Failed to build rule response for rule %d: %v", rule.Id, err)
			continue
		}

		rules = append(rules, *ruleResponse)
	}

	return rules, nil
}

func (s *ruleEngineService) fetchSpecificTransactions(ctx context.Context, userId int64, transactionIds []int64) ([]models.TransactionResponse, error) {
	// Fetch all transactions in bulk instead of one by one
	transactions, err := s.transactionRepo.GetTransactionsByIds(ctx, transactionIds, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}
	return transactions, nil
}

func (s *ruleEngineService) fetchTransactionPage(ctx context.Context, userId int64, page, pageSize int) ([]models.TransactionResponse, error) {
	query := models.TransactionListQuery{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    "date",
		SortOrder: "desc",
	}

	result, err := s.transactionRepo.ListTransactions(ctx, userId, query)
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

func (s *ruleEngineService) applyChangesets(ctx context.Context, userId int64, changesets []*Changeset) ([]models.ModifiedResult, error) {
	var modified []models.ModifiedResult

	for _, changeset := range changesets {
		err := s.applyChangeset(ctx, userId, changeset)
		if err != nil {
			logger.Errorf("Failed to apply changeset to transaction %d: %v", changeset.TransactionId, err)
			continue
		}

		// map rule transaction in mapping table
		s.mapRuleTransaction(ctx, changeset)

		modified = append(modified, models.ModifiedResult{
			TransactionId: changeset.TransactionId,
			AppliedRules:  changeset.AppliedRules,
			UpdatedFields: s.getUpdatedFields(changeset),
		})
	}

	return modified, nil
}

func (s *ruleEngineService) applyChangeset(ctx context.Context, userId int64, changeset *Changeset) error {
	transaction, err := s.transactionRepo.GetTransactionById(ctx, changeset.TransactionId, userId)
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

		err = s.transactionRepo.UpdateTransaction(ctx, changeset.TransactionId, transaction.CreatedBy, updateInput)
		if err != nil {
			return fmt.Errorf("failed to update transaction: %w", err)
		}
	}

	// Apply category updates
	if len(changeset.CategoryAdds) > 0 {
		newCategoryIds := append(transaction.CategoryIds, changeset.CategoryAdds...)
		err = s.transactionRepo.UpdateCategoryMapping(ctx, changeset.TransactionId, transaction.CreatedBy, newCategoryIds)
		if err != nil {
			return fmt.Errorf("failed to update category mapping: %w", err)
		}
	}

	// Apply transfer updates
	if changeset.TransferInfo != nil {
		err = s.createTransferTransaction(ctx, userId, transaction, changeset.TransferInfo)
		if err != nil {
			return fmt.Errorf("failed to create transfer transaction: %w", err)
		}
	}

	return nil
}

func (s *ruleEngineService) createTransferTransaction(ctx context.Context, userId int64, originalTransaction models.TransactionResponse, transferInfo *TransferInfo) error {
	// Create the transfer transaction input
	transferInput := models.CreateTransactionInput{
		CreateBaseTransactionInput: models.CreateBaseTransactionInput{
			Name:        fmt.Sprintf("Transfer from %s", originalTransaction.Name),
			Description: fmt.Sprintf("Transfer from transaction: %s", originalTransaction.Name),
			Amount:      &transferInfo.Amount,
			Date:        originalTransaction.Date,
			CreatedBy:   userId,
			AccountId:   transferInfo.AccountId,
		},
		CategoryIds: originalTransaction.CategoryIds, // Inherit categories from original transaction
	}

	// Create the transfer transaction
	_, err := s.transactionRepo.CreateTransaction(ctx, transferInput.CreateBaseTransactionInput, transferInput.CategoryIds)
	if err != nil {
		return fmt.Errorf("failed to create transfer transaction: %w", err)
	}

	logger.Infof("Created transfer transaction for user %d: amount %.2f to account %d", userId, transferInfo.Amount, transferInfo.AccountId)
	return nil
}

func (s *ruleEngineService) mapRuleTransaction(ctx context.Context, changeset *Changeset) {
	for _, ruleId := range changeset.AppliedRules {
		err := s.ruleRepo.CreateRuleTransactionMapping(ctx, ruleId, changeset.TransactionId)
		if err != nil {
			logger.Errorf("Failed to map rule application for rule %d and transaction %d: %v",
				ruleId, changeset.TransactionId, err)
		}
	}
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
	if changeset.TransferInfo != nil {
		fields = append(fields, models.RuleFieldTransfer)
	}
	return fields
}
