package repository

import (
	"context"
	"time"

	"expenses/internal/config"
	"expenses/internal/models"
	databasemanager "expenses/pkg/database/manager"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The specs below use a database manager whose pool is already closed, so the
// first database call of each method fails. They assert that repository
// methods surface those failures instead of swallowing them.
var _ = Describe("Repository error propagation", func() {
	var (
		ctx context.Context
		cfg *config.Config
	)

	BeforeEach(func() {
		var err error
		cfg, err = config.NewConfig()
		Expect(err).NotTo(HaveOccurred())
		ctx = context.Background()
	})

	closedManager := func() databasemanager.DatabaseManager {
		db, err := databasemanager.NewDatabaseManager(cfg)
		Expect(err).NotTo(HaveOccurred())
		Expect(db.Close()).To(Succeed())
		return db
	}

	assertAllFail := func(calls map[string]func() error) {
		for name, call := range calls {
			Expect(call()).To(HaveOccurred(), "expected %s to fail", name)
		}
	}

	It("AccountRepository surfaces database failures", func() {
		repo := NewAccountRepository(closedManager(), cfg)
		balance := 100.0
		currentValue := 200.0

		assertAllFail(map[string]func() error{
			"CreateAccount": func() error {
				_, err := repo.CreateAccount(ctx, models.CreateAccountInput{Name: "x", BankType: models.BankTypeOthers, Currency: models.CurrencyINR, Balance: &balance, CreatedBy: 1})
				return err
			},
			"GetAccountById": func() error {
				_, err := repo.GetAccountById(ctx, 1, 1)
				return err
			},
			"UpdateAccount": func() error {
				_, err := repo.UpdateAccount(ctx, 1, 1, models.UpdateAccountInput{Balance: &balance, CurrentValue: &currentValue})
				return err
			},
			"DeleteAccount": func() error {
				return repo.DeleteAccount(ctx, 1, 1)
			},
			"ListAccounts": func() error {
				_, err := repo.ListAccounts(ctx, 1)
				return err
			},
		})
	})

	It("CategoryRepository surfaces database failures", func() {
		repo := NewCategoryRepository(closedManager(), cfg)
		name := "Updated"

		assertAllFail(map[string]func() error{
			"CreateCategory": func() error {
				_, err := repo.CreateCategory(ctx, models.CreateCategoryInput{Name: "x", CreatedBy: 1})
				return err
			},
			"GetCategoryById": func() error {
				_, err := repo.GetCategoryById(ctx, 1, 1)
				return err
			},
			"ListCategories": func() error {
				_, err := repo.ListCategories(ctx, 1)
				return err
			},
			"UpdateCategory": func() error {
				_, err := repo.UpdateCategory(ctx, 1, 1, models.UpdateCategoryInput{Name: name})
				return err
			},
			"DeleteCategory": func() error {
				return repo.DeleteCategory(ctx, 1, 1)
			},
		})
	})

	It("UserRepository surfaces database failures", func() {
		repo := NewUserRepository(closedManager(), cfg)

		assertAllFail(map[string]func() error{
			"CreateUser": func() error {
				_, err := repo.CreateUser(ctx, models.CreateUserInput{Email: "x@example.com", Name: "x", Password: "x"})
				return err
			},
			"GetUserByEmailWithPassword": func() error {
				_, err := repo.GetUserByEmailWithPassword(ctx, "x@example.com")
				return err
			},
			"GetUserByIdWithPassword": func() error {
				_, err := repo.GetUserByIdWithPassword(ctx, 1)
				return err
			},
			"GetUserById": func() error {
				_, err := repo.GetUserById(ctx, 1)
				return err
			},
			"DeleteUser": func() error {
				return repo.DeleteUser(ctx, 1)
			},
			"UpdateUser": func() error {
				_, err := repo.UpdateUser(ctx, 1, models.UpdateUserInput{Name: "Updated"})
				return err
			},
			"UpdateUserPassword": func() error {
				_, err := repo.UpdateUserPassword(ctx, 1, "new-password")
				return err
			},
		})
	})

	It("RuleRepository surfaces database failures", func() {
		repo := NewRuleRepository(closedManager(), cfg)
		name := "Updated"
		actionValue := "Food"
		conditionValue := "swiggy"

		assertAllFail(map[string]func() error{
			"CreateRule": func() error {
				_, err := repo.CreateRule(ctx, models.CreateBaseRuleRequest{Name: "x", EffectiveFrom: time.Now(), CreatedBy: 1})
				return err
			},
			"GetRule": func() error {
				_, err := repo.GetRule(ctx, 1, 1)
				return err
			},
			"ListRules": func() error {
				_, err := repo.ListRules(ctx, 1, models.RuleListQuery{Page: 1, PageSize: 10})
				return err
			},
			"ListRuleActionsByRuleId": func() error {
				_, err := repo.ListRuleActionsByRuleId(ctx, 1)
				return err
			},
			"ListRuleConditionsByRuleId": func() error {
				_, err := repo.ListRuleConditionsByRuleId(ctx, 1)
				return err
			},
			"UpdateRule": func() error {
				_, err := repo.UpdateRule(ctx, 1, 1, models.UpdateRuleRequest{Name: &name})
				return err
			},
			"UpdateRuleAction": func() error {
				_, err := repo.UpdateRuleAction(ctx, 1, 1, models.UpdateRuleActionRequest{ActionValue: &actionValue})
				return err
			},
			"UpdateRuleCondition": func() error {
				_, err := repo.UpdateRuleCondition(ctx, 1, 1, models.UpdateRuleConditionRequest{ConditionValue: &conditionValue})
				return err
			},
			"DeleteRuleActionsByRuleId": func() error {
				return repo.DeleteRuleActionsByRuleId(ctx, 1)
			},
			"DeleteRuleConditionsByRuleId": func() error {
				return repo.DeleteRuleConditionsByRuleId(ctx, 1)
			},
			"DeleteRule": func() error {
				return repo.DeleteRule(ctx, 1, 1)
			},
			"PutRuleActions": func() error {
				_, err := repo.PutRuleActions(ctx, 1, []models.CreateRuleActionRequest{{ActionType: models.RuleFieldCategory, ActionValue: "Food"}})
				return err
			},
			"PutRuleConditions": func() error {
				_, err := repo.PutRuleConditions(ctx, 1, []models.CreateRuleConditionRequest{{ConditionType: models.RuleFieldName, ConditionValue: "swiggy", ConditionOperator: models.OperatorContains}})
				return err
			},
			"CreateRuleTransactionMapping": func() error {
				return repo.CreateRuleTransactionMapping(ctx, 1, 1)
			},
		})
	})

	It("TransactionRepository surfaces database failures", func() {
		repo := NewTransactionRepository(closedManager(), cfg)
		amount := 10.0
		input := models.CreateBaseTransactionInput{Name: "x", Amount: &amount, Date: time.Now(), CreatedBy: 1, AccountId: 1}
		name := "Updated"

		assertAllFail(map[string]func() error{
			"CreateTransaction": func() error {
				_, err := repo.CreateTransaction(ctx, input, nil)
				return err
			},
			"CreateTransactions": func() error {
				_, err := repo.CreateTransactions(ctx, []models.CreateBaseTransactionInput{input}, [][]int64{nil})
				return err
			},
			"GetTransactionById": func() error {
				_, err := repo.GetTransactionById(ctx, 1, 1)
				return err
			},
			"GetTransactionsByIds": func() error {
				_, err := repo.GetTransactionsByIds(ctx, []int64{1}, 1)
				return err
			},
			"UpdateTransaction": func() error {
				return repo.UpdateTransaction(ctx, 1, 1, models.UpdateBaseTransactionInput{Name: name})
			},
			"DeleteTransaction": func() error {
				return repo.DeleteTransaction(ctx, 1, 1)
			},
			"UpdateCategoryMapping": func() error {
				return repo.UpdateCategoryMapping(ctx, 1, 1, []int64{1})
			},
			"ListTransactions": func() error {
				_, err := repo.ListTransactions(ctx, 1, models.TransactionListQuery{Page: 1, PageSize: 10})
				return err
			},
		})
	})

	It("StatementRepository surfaces database failures", func() {
		repo := NewStatementRepository(closedManager(), cfg)

		assertAllFail(map[string]func() error{
			"CreateStatement": func() error {
				_, err := repo.CreateStatement(ctx, models.CreateStatementInput{AccountId: 1, CreatedBy: 1, OriginalFilename: "x.csv", FileType: "csv", Status: models.StatementStatusPending})
				return err
			},
			"CreateStatementTxn": func() error {
				return repo.CreateStatementTxn(ctx, 1, 1)
			},
			"CreateStatementTxns": func() error {
				return repo.CreateStatementTxns(ctx, 1, []int64{1})
			},
			"UpdateStatementStatus": func() error {
				_, err := repo.UpdateStatementStatus(ctx, 1, models.UpdateStatementStatusInput{Status: models.StatementStatusDone})
				return err
			},
			"GetStatementByID": func() error {
				_, err := repo.GetStatementByID(ctx, 1, 1)
				return err
			},
			"ListStatementByUserId": func() error {
				_, err := repo.ListStatementByUserId(ctx, 1, 10, 0, models.StatementListQuery{})
				return err
			},
			"CountStatementsByUserId": func() error {
				_, err := repo.CountStatementsByUserId(ctx, 1, models.StatementListQuery{})
				return err
			},
		})
	})

	It("AnalyticsRepository surfaces database failures", func() {
		repo := NewAnalyticsRepository(closedManager(), cfg)
		dateFrom := time.Now().AddDate(0, -1, 0)
		dateTo := time.Now()

		assertAllFail(map[string]func() error{
			"GetBalance": func() error {
				_, err := repo.GetBalance(ctx, 1, &dateFrom, &dateTo)
				return err
			},
			"GetCashBalanceHistory": func() error {
				_, _, _, _, err := repo.GetCashBalanceHistory(ctx, 1, dateFrom, dateTo)
				return err
			},
			"GetCategoryAnalytics": func() error {
				_, err := repo.GetCategoryAnalytics(ctx, 1, dateFrom, dateTo, nil)
				return err
			},
			"GetMonthlyAnalytics": func() error {
				_, err := repo.GetMonthlyAnalytics(ctx, 1, dateFrom, dateTo)
				return err
			},
			"GetAccountCashFlows": func() error {
				_, err := repo.GetAccountCashFlows(ctx, 1, []int64{1})
				return err
			},
			"GetInsightsMonthly": func() error {
				_, err := repo.GetInsightsMonthly(ctx, 1, dateFrom, dateTo)
				return err
			},
			"GetInsightsCategories": func() error {
				_, err := repo.GetInsightsCategories(ctx, 1, dateFrom, dateTo)
				return err
			},
			"GetInsightsTopExpenses": func() error {
				_, err := repo.GetInsightsTopExpenses(ctx, 1, dateFrom, dateTo)
				return err
			},
			"GetInsightsUncategorized": func() error {
				_, _, err := repo.GetInsightsUncategorized(ctx, 1, dateFrom, dateTo)
				return err
			},
			"GetInsightsSpendingSummary": func() error {
				_, err := repo.GetInsightsSpendingSummary(ctx, 1, dateFrom, dateTo)
				return err
			},
			"GetInsightsCategoryMonths": func() error {
				_, err := repo.GetInsightsCategoryMonths(ctx, 1, dateFrom, dateTo)
				return err
			},
			"GetInsightsWeekday": func() error {
				_, err := repo.GetInsightsWeekday(ctx, 1, dateFrom, dateTo)
				return err
			},
			"GetInsightsLatestTransactionDate": func() error {
				_, err := repo.GetInsightsLatestTransactionDate(ctx, 1)
				return err
			},
			"GetInsightsMultiCategoryCount": func() error {
				_, err := repo.GetInsightsMultiCategoryCount(ctx, 1, dateFrom, dateTo)
				return err
			},
		})
	})

	It("SessionRepository surfaces database failures", func() {
		repo := NewSessionRepository(closedManager(), cfg)

		assertAllFail(map[string]func() error{
			"Create": func() error {
				return repo.Create(ctx, 1, "hash", time.Now().Add(time.Hour))
			},
			"GetActiveByHash": func() error {
				_, err := repo.GetActiveByHash(ctx, "hash")
				return err
			},
			"Rotate": func() error {
				return repo.Rotate(ctx, 1, "old", "new", time.Now().Add(time.Hour))
			},
			"RevokeByHash": func() error {
				_, err := repo.RevokeByHash(ctx, "hash")
				return err
			},
			"RevokeAllForUser": func() error {
				return repo.RevokeAllForUser(ctx, 1)
			},
			"ExpireByHash": func() error {
				_, err := repo.ExpireByHash(ctx, "hash")
				return err
			},
			"DeleteExpired": func() error {
				_, err := repo.DeleteExpired(ctx)
				return err
			},
		})
	})
})
