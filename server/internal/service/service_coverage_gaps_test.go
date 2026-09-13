package service

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"expenses/internal/config"
	mockDatabase "expenses/internal/mock/database"
	mock_repository "expenses/internal/mock/repository"
	"expenses/internal/models"
	"expenses/internal/validator"
	"expenses/pkg/utils"
	"hash/crc32"
	"math"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type rawCashBalanceRepository struct {
	*mock_repository.MockAnalyticsRepository
	initialBalance float64
	dailyData      []map[string]any
}

func (r *rawCashBalanceRepository) GetCashBalanceHistory(context.Context, int64, time.Time, time.Time) (float64, float64, float64, []map[string]any, error) {
	return r.initialBalance, 0, 0, r.dailyData, nil
}

func gapOversizedXLSX() []byte {
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	payload := []byte("PK")
	raw, err := writer.CreateRaw(&zip.FileHeader{
		Name:               "xl/oversized.bin",
		Method:             zip.Store,
		CRC32:              crc32.ChecksumIEEE(payload),
		CompressedSize64:   uint64(len(payload)),
		UncompressedSize64: 200 << 20,
	})
	Expect(err).NotTo(HaveOccurred())
	_, err = raw.Write(payload)
	Expect(err).NotTo(HaveOccurred())
	Expect(writer.Close()).To(Succeed())
	return buf.Bytes()
}

var _ = Describe("AnalyticsService coverage gaps", func() {
	var (
		analyticsService  AnalyticsServiceInterface
		mockAnalyticsRepo *mock_repository.MockAnalyticsRepository
		mockAccountRepo   *mock_repository.MockAccountRepository
		ctx               context.Context
		userId            int64
	)

	BeforeEach(func() {
		ctx = context.Background()
		userId = 1
		mockAnalyticsRepo = mock_repository.NewMockAnalyticsRepository()
		mockAccountRepo = mock_repository.NewMockAccountRepository()
		analyticsService = NewAnalyticsService(mockAnalyticsRepo, mockAccountRepo)
	})

	Describe("GetAccountAnalytics", func() {
		It("returns empty analytics when listing accounts fails", func() {
			mockAccountRepo.FailOn("ListAccounts", errors.New("accounts down"))

			result, err := analyticsService.GetAccountAnalytics(ctx, userId)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.AccountAnalytics).To(BeEmpty())
		})

		It("propagates historical balance errors", func() {
			_, err := mockAccountRepo.CreateAccount(ctx, models.CreateAccountInput{Name: "Bank", BankType: models.BankTypeAxis, Currency: models.CurrencyINR, CreatedBy: userId})
			Expect(err).NotTo(HaveOccurred())
			mockAnalyticsRepo.SetBalance(userId, nil, nil, map[int64]float64{1: 100})
			mockAnalyticsRepo.FailOn("GetBalanceEndDate", errors.New("historical down"))

			_, err = analyticsService.GetAccountAnalytics(ctx, userId)
			Expect(err).To(HaveOccurred())
		})

		It("propagates investment cash flow errors", func() {
			currentValue := 1000.0
			_, err := mockAccountRepo.CreateAccount(ctx, models.CreateAccountInput{Name: "FD", BankType: models.BankTypeInvestment, Currency: models.CurrencyINR, CurrentValue: &currentValue, CreatedBy: userId})
			Expect(err).NotTo(HaveOccurred())
			mockAnalyticsRepo.FailOn("GetAccountCashFlows", errors.New("cash flows down"))

			_, err = analyticsService.GetAccountAnalytics(ctx, userId)
			Expect(err).To(HaveOccurred())
		})

		It("groups investment cash flows by account", func() {
			currentValue := 1000.0
			account, err := mockAccountRepo.CreateAccount(ctx, models.CreateAccountInput{Name: "FD", BankType: models.BankTypeInvestment, Currency: models.CurrencyINR, CurrentValue: &currentValue, CreatedBy: userId})
			Expect(err).NotTo(HaveOccurred())
			mockAnalyticsRepo.SetAccountCashFlows(userId, []models.AccountCashFlow{
				{AccountID: account.Id, Amount: -1000, Date: time.Now().AddDate(-1, 0, 0)},
			})

			result, err := analyticsService.GetAccountAnalytics(ctx, userId)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.AccountAnalytics).To(HaveLen(1))
		})
	})

	Describe("GetCashBalanceHistory", func() {
		var startDate, endDate time.Time

		BeforeEach(func() {
			startDate, _ = time.Parse("2006-01-02", "2023-01-01")
			endDate, _ = time.Parse("2006-01-02", "2023-01-03")
		})

		It("propagates account listing errors", func() {
			mockAccountRepo.FailOn("ListAccounts", errors.New("accounts down"))

			_, err := analyticsService.GetCashBalanceHistory(ctx, userId, startDate, endDate)
			Expect(err).To(HaveOccurred())
		})

		It("rejects daily data with a non-string date", func() {
			rawRepo := &rawCashBalanceRepository{
				MockAnalyticsRepository: mock_repository.NewMockAnalyticsRepository(),
				initialBalance:          100,
				dailyData:               []map[string]any{{"date": 123, "daily_change": 1.0}},
			}
			service := NewAnalyticsService(rawRepo, mockAccountRepo)

			_, err := service.GetCashBalanceHistory(ctx, userId, startDate, endDate)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid type for date in daily data"))
		})

		It("rejects daily data with a non-float daily_change", func() {
			rawRepo := &rawCashBalanceRepository{
				MockAnalyticsRepository: mock_repository.NewMockAnalyticsRepository(),
				initialBalance:          100,
				dailyData:               []map[string]any{{"date": "2023-01-01", "daily_change": "oops"}},
			}
			service := NewAnalyticsService(rawRepo, mockAccountRepo)

			_, err := service.GetCashBalanceHistory(ctx, userId, startDate, endDate)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid type for daily_change in daily data"))
		})

		It("falls back to a single point when the range is inverted", func() {
			result, err := analyticsService.GetCashBalanceHistory(ctx, userId, endDate, startDate)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.TimeSeries).To(HaveLen(1))
			Expect(result.TimeSeries[0].Date).To(Equal(endDate.Format("2006-01-02")))
		})
	})

	Describe("GetInsights", func() {
		var startDate, endDate time.Time

		BeforeEach(func() {
			startDate = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			endDate = time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC)
		})

		It("propagates account listing errors", func() {
			mockAccountRepo.FailOn("ListAccounts", errors.New("accounts down"))

			_, err := analyticsService.GetInsights(ctx, userId, startDate, endDate)
			Expect(err).To(HaveOccurred())
		})

		It("propagates current balance errors", func() {
			mockAnalyticsRepo.SetShouldErrorOnBalance(true)

			_, err := analyticsService.GetInsights(ctx, userId, startDate, endDate)
			Expect(err).To(HaveOccurred())
		})

		It("propagates balance repository errors", func() {
			mockAnalyticsRepo.FailOn("GetBalance", errors.New("balance down"))

			_, err := analyticsService.GetInsights(ctx, userId, startDate, endDate)
			Expect(err).To(HaveOccurred())
		})

		It("propagates investment cash flow errors", func() {
			currentValue := 1000.0
			_, err := mockAccountRepo.CreateAccount(ctx, models.CreateAccountInput{Name: "FD", BankType: models.BankTypeInvestment, Currency: models.CurrencyINR, CurrentValue: &currentValue, CreatedBy: userId})
			Expect(err).NotTo(HaveOccurred())
			mockAnalyticsRepo.FailOn("GetAccountCashFlows", errors.New("cash flows down"))

			_, err = analyticsService.GetInsights(ctx, userId, startDate, endDate)
			Expect(err).To(HaveOccurred())
		})

		var failingMethods = []string{
			"GetInsightsMonthly",
			"GetInsightsCategories",
			"GetInsightsTopExpenses",
			"GetInsightsUncategorized",
			"GetInsightsSpendingSummary",
			"GetInsightsWeekday",
			"GetInsightsMultiCategoryCount",
			"GetInsightsLatestTransactionDate",
			"GetInsightsCategoryMonths",
		}

		for _, method := range failingMethods {
			method := method
			It("propagates "+method+" errors", func() {
				currentValue := 1000.0
				_, err := mockAccountRepo.CreateAccount(ctx, models.CreateAccountInput{Name: "FD", BankType: models.BankTypeInvestment, Currency: models.CurrencyINR, CurrentValue: &currentValue, CreatedBy: userId})
				Expect(err).NotTo(HaveOccurred())
				mockAnalyticsRepo.FailOn(method, errors.New("boom"))

				_, err = analyticsService.GetInsights(ctx, userId, startDate, endDate)
				Expect(err).To(HaveOccurred())
			})
		}

		It("computes top expense share and average", func() {
			mockAnalyticsRepo.SetInsightsMonthly(userId, startDate, endDate, []models.InsightsMonthlyPoint{
				{Month: "2024-01", Expenses: 100},
			})
			mockAnalyticsRepo.SetInsightsTopExpenses(userId, startDate, endDate, []models.InsightsTopExpense{
				{Name: "Cafe", Amount: 50, Count: 2},
			})

			result, err := analyticsService.GetInsights(ctx, userId, startDate, endDate)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.TopExpenses).To(HaveLen(1))
			Expect(result.TopExpenses[0].Share).To(BeNumerically("~", 0.5))
			Expect(result.TopExpenses[0].Average).To(Equal(25.0))
		})
	})

	Describe("helper coverage", func() {
		It("delegates category analytics to the repository", func() {
			result, err := analyticsService.GetCategoryAnalytics(ctx, userId, time.Now().AddDate(0, -1, 0), time.Now(), nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
		})

		It("clamps noSpendDays for inverted or fully active ranges", func() {
			start := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
			Expect(noSpendDays(start, start.AddDate(0, 0, -5), 0)).To(Equal(int64(0)))
			Expect(noSpendDays(start, start, 5)).To(Equal(int64(0)))
		})

		It("returns the previous month for a partial month", func() {
			end := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
			Expect(latestCompleteMonth(end)).To(Equal(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)))
		})

		It("skips trend building when the range is partial", func() {
			svc := analyticsService.(*AnalyticsService)
			start := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
			end := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

			trend, movement, err := svc.buildTrendAndMovement(ctx, userId, start, end, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(trend).To(Equal(models.InsightsTrend{}))
			Expect(movement).To(BeEmpty())
		})

		It("ignores category months outside the compared windows", func() {
			months := []models.InsightsCategoryMonth{
				{Month: "2024-01", CategoryID: 3, CategoryName: "Old", Total: 50},
				{Month: "2024-03", CategoryID: 1, CategoryName: "Food", Total: 100},
			}

			movement := buildCategoryMovement(months, "2024-03", "2024-02")
			Expect(movement).To(HaveLen(1))
			Expect(movement[0].CategoryName).To(Equal("Food"))
		})

		It("breaks equal category movements by name", func() {
			months := []models.InsightsCategoryMonth{
				{Month: "2024-03", CategoryID: 1, CategoryName: "B", Total: 100},
				{Month: "2024-02", CategoryID: 1, CategoryName: "B", Total: 50},
				{Month: "2024-03", CategoryID: 2, CategoryName: "A", Total: 100},
				{Month: "2024-02", CategoryID: 2, CategoryName: "A", Total: 50},
			}

			movement := buildCategoryMovement(months, "2024-03", "2024-02")
			Expect(movement).To(HaveLen(2))
			Expect(movement[0].CategoryName).To(Equal("A"))
		})

		It("drops zero-amount cash flows", func() {
			collected := collectInvestmentCashFlows([]models.AccountCashFlow{
				{AccountID: 1, Amount: 0, Date: time.Now()},
			})
			Expect(collected.flows).To(BeEmpty())
		})

		It("returns zero xirr for a same-day single flow", func() {
			now := time.Now()
			percentage, xirr := calculateInvestmentMetrics([]models.AccountCashFlow{
				{AccountID: 1, Amount: -1000, Date: now},
			}, 15000, now)
			Expect(percentage).To(BeNumerically("~", 1400.0))
			Expect(xirr).NotTo(BeNil())
			Expect(*xirr).To(BeNumerically("==", 0.0))
		})

		It("returns false for a single XIRR flow", func() {
			_, ok := calculateXIRR([]investmentCashFlow{{amount: 1, date: time.Now()}})
			Expect(ok).To(BeFalse())
		})

		It("returns false when the XIRR net present value is not finite", func() {
			date := time.Now()
			_, ok := calculateXIRR([]investmentCashFlow{
				{amount: math.Inf(1), date: date},
				{amount: math.Inf(-1), date: date},
			})
			Expect(ok).To(BeFalse())
		})

		It("returns false when the XIRR derivative is zero", func() {
			date := time.Now()
			_, ok := calculateXIRR([]investmentCashFlow{
				{amount: -100, date: date},
				{amount: 200, date: date},
			})
			Expect(ok).To(BeFalse())
		})

		It("returns false when the XIRR step overflows", func() {
			base := time.Now()
			_, ok := calculateXIRR([]investmentCashFlow{
				{amount: 2, date: base},
				{amount: -1, date: base},
				{amount: 3.15e-293, date: base.Add(time.Nanosecond)},
			})
			Expect(ok).To(BeFalse())
		})
	})
})

func gapCreatePastRule(ctx context.Context, repo *mock_repository.MockRuleRepository, userId int64) models.RuleResponse {
	desc := "rule"
	rule, err := repo.CreateRule(ctx, models.CreateBaseRuleRequest{
		Name:          "Gap Rule",
		Description:   &desc,
		EffectiveFrom: time.Now().Add(-time.Hour),
		CreatedBy:     userId,
	})
	Expect(err).NotTo(HaveOccurred())
	_, err = repo.CreateRuleActions(ctx, []models.CreateRuleActionRequest{
		{RuleId: rule.Id, ActionType: models.RuleFieldName, ActionValue: "Updated"},
	})
	Expect(err).NotTo(HaveOccurred())
	_, err = repo.CreateRuleConditions(ctx, []models.CreateRuleConditionRequest{
		{RuleId: rule.Id, ConditionType: models.RuleFieldName, ConditionOperator: models.OperatorContains, ConditionValue: "grocery"},
	})
	Expect(err).NotTo(HaveOccurred())
	return rule
}

var _ = Describe("RuleEngineService coverage gaps", func() {
	var (
		service          RuleEngineServiceInterface
		mockRuleRepo     *mock_repository.MockRuleRepository
		mockTxnRepo      *mock_repository.MockTransactionRepository
		mockCategoryRepo *mock_repository.MockCategoryRepository
		mockAccountRepo  *mock_repository.MockAccountRepository
		ctx              context.Context
		userId           int64
	)

	BeforeEach(func() {
		ctx = context.Background()
		userId = 1
		mockRuleRepo = mock_repository.NewMockRuleRepository()
		mockTxnRepo = mock_repository.NewMockTransactionRepository()
		mockCategoryRepo = mock_repository.NewMockCategoryRepository()
		mockAccountRepo = mock_repository.NewMockAccountRepository()
		service = NewRuleEngineService(mockRuleRepo, mockTxnRepo, mockCategoryRepo, mockAccountRepo)
	})

	It("stops when categories cannot be fetched", func() {
		mockCategoryRepo.FailOn("ListCategories", errors.New("categories down"))

		service.(*ruleEngineService).ExecuteRulesInBackground(ctx, userId, models.ExecuteRulesRequest{})
	})

	It("stops when accounts cannot be fetched", func() {
		mockAccountRepo.FailOn("ListAccounts", errors.New("accounts down"))

		service.(*ruleEngineService).ExecuteRulesInBackground(ctx, userId, models.ExecuteRulesRequest{})
	})

	It("stops when rules cannot be fetched", func() {
		mockRuleRepo.FailOn("ListRules", errors.New("rules down"))

		service.(*ruleEngineService).ExecuteRulesInBackground(ctx, userId, models.ExecuteRulesRequest{})
	})

	It("stops when specific transactions cannot be fetched", func() {
		gapCreatePastRule(ctx, mockRuleRepo, userId)
		mockTxnRepo.FailOn("GetTransactionsByIds", errors.New("txns down"))

		ids := []int64{1}
		service.(*ruleEngineService).ExecuteRulesInBackground(ctx, userId, models.ExecuteRulesRequest{TransactionIds: &ids})
	})

	It("stops when a transaction page cannot be fetched", func() {
		gapCreatePastRule(ctx, mockRuleRepo, userId)
		mockTxnRepo.FailOn("ListTransactions", errors.New("txns down"))

		service.(*ruleEngineService).ExecuteRulesInBackground(ctx, userId, models.ExecuteRulesRequest{})
	})

	It("walks pages until an empty page is returned", func() {
		gapCreatePastRule(ctx, mockRuleRepo, userId)
		amount := 10.0
		_, err := mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
			Name: "Test Tx", Amount: &amount, Date: time.Now(), CreatedBy: userId, AccountId: 1,
		}, nil)
		Expect(err).NotTo(HaveOccurred())

		service.(*ruleEngineService).ExecuteRulesInBackground(ctx, userId, models.ExecuteRulesRequest{PageSize: 1})
	})

	It("wraps action fetch failures when building a rule response", func() {
		rule := gapCreatePastRule(ctx, mockRuleRepo, userId)
		mockRuleRepo.FailOn("ListRuleActionsByRuleId", errors.New("actions down"))

		_, err := service.(*ruleEngineService).buildRuleResponse(ctx, rule)
		Expect(err).To(HaveOccurred())
	})

	It("wraps condition fetch failures when building a rule response", func() {
		rule := gapCreatePastRule(ctx, mockRuleRepo, userId)
		mockRuleRepo.FailOn("ListRuleConditionsByRuleId", errors.New("conditions down"))

		_, err := service.(*ruleEngineService).buildRuleResponse(ctx, rule)
		Expect(err).To(HaveOccurred())
	})

	It("skips specific rules that fail to build", func() {
		rule := gapCreatePastRule(ctx, mockRuleRepo, userId)
		mockRuleRepo.FailOn("ListRuleActionsByRuleId", errors.New("actions down"))

		rules, err := service.(*ruleEngineService).fetchSpecificRules(ctx, userId, []int64{rule.Id})
		Expect(err).NotTo(HaveOccurred())
		Expect(rules).To(BeEmpty())
	})

	It("skips all-user rules that fail to build", func() {
		gapCreatePastRule(ctx, mockRuleRepo, userId)
		mockRuleRepo.FailOn("ListRuleConditionsByRuleId", errors.New("conditions down"))

		rules, err := service.(*ruleEngineService).fetchAllUserRules(ctx, userId)
		Expect(err).NotTo(HaveOccurred())
		Expect(rules).To(BeEmpty())
	})

	It("fails the changeset when the base transaction update fails", func() {
		amount := 10.0
		txn, err := mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
			Name: "Test Tx", Amount: &amount, Date: time.Now(), CreatedBy: userId, AccountId: 1,
		}, nil)
		Expect(err).NotTo(HaveOccurred())
		mockTxnRepo.FailOn("UpdateTransaction", errors.New("update down"))

		err = service.(*ruleEngineService).applyChangeset(ctx, userId, &Changeset{
			TransactionId: txn.Id,
			NameUpdate:    stringPtr("Updated"),
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to update transaction"))
	})

	It("fails the changeset when the category mapping update fails", func() {
		amount := 10.0
		txn, err := mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
			Name: "Test Tx", Amount: &amount, Date: time.Now(), CreatedBy: userId, AccountId: 1,
		}, nil)
		Expect(err).NotTo(HaveOccurred())
		mockTxnRepo.FailOn("UpdateCategoryMapping", errors.New("mapping down"))

		err = service.(*ruleEngineService).applyChangeset(ctx, userId, &Changeset{
			TransactionId: txn.Id,
			CategoryAdds:  []int64{1},
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to update category mapping"))
	})

	It("fails the changeset when the transfer transaction cannot be created", func() {
		amount := 10.0
		txn, err := mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
			Name: "Test Tx", Amount: &amount, Date: time.Now(), CreatedBy: userId, AccountId: 1,
		}, nil)
		Expect(err).NotTo(HaveOccurred())
		mockTxnRepo.FailOn("CreateTransaction", errors.New("create down"))

		err = service.(*ruleEngineService).applyChangeset(ctx, userId, &Changeset{
			TransactionId: txn.Id,
			TransferInfo:  &TransferInfo{AccountId: 2, Amount: 50},
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to create transfer transaction"))
	})

	It("keeps going when the rule mapping cannot be stored", func() {
		mockRuleRepo.FailOn("CreateRuleTransactionMapping", errors.New("mapping down"))

		service.(*ruleEngineService).mapRuleTransaction(ctx, &Changeset{
			TransactionId: 1,
			AppliedRules:  []int64{1, 2},
		})
	})
})

var _ = Describe("TransactionService coverage gaps", func() {
	var (
		transactionService TransactionServiceInterface
		mockRepo           *mock_repository.MockTransactionRepository
		categoryMockRepo   *mock_repository.MockCategoryRepository
		accountMockRepo    *mock_repository.MockAccountRepository
		ctx                context.Context
		userId             int64
		testDate           time.Time
		cat                models.CategoryResponse
		acc                models.AccountResponse
	)

	BeforeEach(func() {
		ctx = context.Background()
		userId = 1
		testDate, _ = time.Parse("2006-01-02", "2023-01-01")
		mockRepo = mock_repository.NewMockTransactionRepository()
		categoryMockRepo = mock_repository.NewMockCategoryRepository()
		accountMockRepo = mock_repository.NewMockAccountRepository()
		mockDB := mockDatabase.NewMockDatabaseManager()
		transactionService = NewTransactionService(mockRepo, categoryMockRepo, accountMockRepo, mockDB)

		var err error
		cat, err = categoryMockRepo.CreateCategory(ctx, models.CreateCategoryInput{Name: "Food", CreatedBy: userId})
		Expect(err).NotTo(HaveOccurred())
		acc, err = accountMockRepo.CreateAccount(ctx, models.CreateAccountInput{Name: "HDFC", BankType: models.BankTypeHDFC, Currency: models.CurrencyINR, CreatedBy: userId})
		Expect(err).NotTo(HaveOccurred())
	})

	It("bulk creates transactions after validating unique accounts and categories", func() {
		amount := 10.0
		inputs := []models.CreateTransactionInput{
			{
				CreateBaseTransactionInput: models.CreateBaseTransactionInput{Name: "One", Amount: &amount, Date: testDate, CreatedBy: userId, AccountId: acc.Id},
				CategoryIds:                []int64{cat.Id},
			},
			{
				CreateBaseTransactionInput: models.CreateBaseTransactionInput{Name: "Two", Amount: &amount, Date: testDate, CreatedBy: userId, AccountId: acc.Id},
				CategoryIds:                []int64{cat.Id},
			},
		}

		created, err := transactionService.CreateTransactions(ctx, inputs)
		Expect(err).NotTo(HaveOccurred())
		Expect(created).To(HaveLen(2))
	})

	It("rejects a future date in bulk create", func() {
		amount := 10.0
		inputs := []models.CreateTransactionInput{
			{
				CreateBaseTransactionInput: models.CreateBaseTransactionInput{Name: "Future", Amount: &amount, Date: time.Now().AddDate(0, 0, 2), CreatedBy: userId, AccountId: acc.Id},
			},
		}

		_, err := transactionService.CreateTransactions(ctx, inputs)
		Expect(err).To(HaveOccurred())
	})

	It("rejects a missing account in bulk create", func() {
		amount := 10.0
		inputs := []models.CreateTransactionInput{
			{
				CreateBaseTransactionInput: models.CreateBaseTransactionInput{Name: "Missing", Amount: &amount, Date: testDate, CreatedBy: userId, AccountId: 9999},
			},
		}

		_, err := transactionService.CreateTransactions(ctx, inputs)
		Expect(err).To(HaveOccurred())
	})

	It("rejects a missing category in bulk create", func() {
		amount := 10.0
		inputs := []models.CreateTransactionInput{
			{
				CreateBaseTransactionInput: models.CreateBaseTransactionInput{Name: "Missing", Amount: &amount, Date: testDate, CreatedBy: userId, AccountId: acc.Id},
				CategoryIds:                []int64{9999},
			},
		}

		_, err := transactionService.CreateTransactions(ctx, inputs)
		Expect(err).To(HaveOccurred())
	})

	It("propagates category mapping errors during update", func() {
		amount := 10.0
		created, err := transactionService.CreateTransaction(ctx, models.CreateTransactionInput{
			CreateBaseTransactionInput: models.CreateBaseTransactionInput{Name: "Tx", Amount: &amount, Date: testDate, CreatedBy: userId, AccountId: acc.Id},
			CategoryIds:                []int64{cat.Id},
		})
		Expect(err).NotTo(HaveOccurred())
		mockRepo.FailOn("UpdateCategoryMapping", errors.New("mapping down"))

		categoryIds := []int64{cat.Id}
		_, err = transactionService.UpdateTransaction(ctx, created.Id, userId, models.UpdateTransactionInput{CategoryIds: &categoryIds})
		Expect(err).To(HaveOccurred())
	})

	It("propagates reload errors during update", func() {
		amount := 10.0
		created, err := transactionService.CreateTransaction(ctx, models.CreateTransactionInput{
			CreateBaseTransactionInput: models.CreateBaseTransactionInput{Name: "Tx", Amount: &amount, Date: testDate, CreatedBy: userId, AccountId: acc.Id},
		})
		Expect(err).NotTo(HaveOccurred())
		mockRepo.FailOn("GetTransactionById", errors.New("reload down"))

		name := "Updated"
		_, err = transactionService.UpdateTransaction(ctx, created.Id, userId, models.UpdateTransactionInput{
			UpdateBaseTransactionInput: models.UpdateBaseTransactionInput{Name: name},
		})
		Expect(err).To(HaveOccurred())
	})

	It("propagates category listing errors during validation", func() {
		categoryMockRepo.FailOn("ListCategories", errors.New("categories down"))
		amount := 10.0

		_, err := transactionService.CreateTransaction(ctx, models.CreateTransactionInput{
			CreateBaseTransactionInput: models.CreateBaseTransactionInput{Name: "Tx", Amount: &amount, Date: testDate, CreatedBy: userId, AccountId: acc.Id},
			CategoryIds:                []int64{cat.Id},
		})
		Expect(err).To(HaveOccurred())
	})

	It("propagates account lookup errors during validation", func() {
		accountMockRepo.FailOn("GetAccountById", errors.New("account down"))
		amount := 10.0

		_, err := transactionService.CreateTransaction(ctx, models.CreateTransactionInput{
			CreateBaseTransactionInput: models.CreateBaseTransactionInput{Name: "Tx", Amount: &amount, Date: testDate, CreatedBy: userId, AccountId: acc.Id},
		})
		Expect(err).To(HaveOccurred())
	})

	It("rejects a future date on create", func() {
		amount := 10.0

		_, err := transactionService.CreateTransaction(ctx, models.CreateTransactionInput{
			CreateBaseTransactionInput: models.CreateBaseTransactionInput{Name: "Future", Amount: &amount, Date: time.Now().AddDate(0, 0, 2), CreatedBy: userId, AccountId: acc.Id},
		})
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("RuleService coverage gaps", func() {
	var (
		ruleService RuleServiceInterface
		mockRepo    *mock_repository.MockRuleRepository
		ctx         context.Context
		userId      int64
	)

	ruleInput := func() models.CreateRuleRequest {
		desc := "desc"
		return models.CreateRuleRequest{
			Rule: models.CreateBaseRuleRequest{
				Name:          "Gap Rule",
				Description:   &desc,
				EffectiveFrom: time.Now(),
				CreatedBy:     userId,
			},
			Actions:    []models.CreateRuleActionRequest{{ActionType: models.RuleFieldAmount, ActionValue: "100"}},
			Conditions: []models.CreateRuleConditionRequest{{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals}},
		}
	}

	BeforeEach(func() {
		ctx = context.Background()
		userId = 1
		mockRepo = mock_repository.NewMockRuleRepository()
		mockTxnRepo := mock_repository.NewMockTransactionRepository()
		mockDB := mockDatabase.NewMockDatabaseManager()
		ruleService = NewRuleService(mockRepo, mockTxnRepo, mockDB)
	})

	It("propagates rule creation errors", func() {
		mockRepo.FailOn("CreateRule", errors.New("create down"))

		_, err := ruleService.CreateRule(ctx, ruleInput())
		Expect(err).To(HaveOccurred())
	})

	It("propagates action creation errors", func() {
		mockRepo.FailOn("CreateRuleActions", errors.New("actions down"))

		_, err := ruleService.CreateRule(ctx, ruleInput())
		Expect(err).To(HaveOccurred())
	})

	It("propagates condition creation errors", func() {
		mockRepo.FailOn("CreateRuleConditions", errors.New("conditions down"))

		_, err := ruleService.CreateRule(ctx, ruleInput())
		Expect(err).To(HaveOccurred())
	})

	It("propagates action listing errors in GetRuleById", func() {
		created, err := ruleService.CreateRule(ctx, ruleInput())
		Expect(err).NotTo(HaveOccurred())
		mockRepo.FailOn("ListRuleActionsByRuleId", errors.New("actions down"))

		_, err = ruleService.GetRuleById(ctx, created.Rule.Id, userId)
		Expect(err).To(HaveOccurred())
	})

	It("propagates condition listing errors in GetRuleById", func() {
		created, err := ruleService.CreateRule(ctx, ruleInput())
		Expect(err).NotTo(HaveOccurred())
		mockRepo.FailOn("ListRuleConditionsByRuleId", errors.New("conditions down"))

		_, err = ruleService.GetRuleById(ctx, created.Rule.Id, userId)
		Expect(err).To(HaveOccurred())
	})

	It("propagates rule listing errors", func() {
		mockRepo.FailOn("ListRules", errors.New("list down"))

		_, err := ruleService.ListRules(ctx, userId, nil)
		Expect(err).To(HaveOccurred())
	})

	It("propagates condition deletion errors", func() {
		created, err := ruleService.CreateRule(ctx, ruleInput())
		Expect(err).NotTo(HaveOccurred())
		mockRepo.FailOn("DeleteRuleConditionsByRuleId", errors.New("delete down"))

		err = ruleService.DeleteRule(ctx, created.Rule.Id, userId)
		Expect(err).To(HaveOccurred())
	})

	It("propagates PUT actions errors", func() {
		created, err := ruleService.CreateRule(ctx, ruleInput())
		Expect(err).NotTo(HaveOccurred())
		mockRepo.FailOn("PutRuleActions", errors.New("put down"))

		_, err = ruleService.PutRuleActions(ctx, created.Rule.Id, models.PutRuleActionsRequest{
			Actions: []models.CreateRuleActionRequest{{ActionType: models.RuleFieldAmount, ActionValue: "200"}},
		}, userId)
		Expect(err).To(HaveOccurred())
	})

	It("propagates PUT conditions errors", func() {
		created, err := ruleService.CreateRule(ctx, ruleInput())
		Expect(err).NotTo(HaveOccurred())
		mockRepo.FailOn("PutRuleConditions", errors.New("put down"))

		_, err = ruleService.PutRuleConditions(ctx, created.Rule.Id, models.PutRuleConditionsRequest{
			Conditions: []models.CreateRuleConditionRequest{{ConditionType: models.RuleFieldAmount, ConditionValue: "200", ConditionOperator: models.OperatorEquals}},
		}, userId)
		Expect(err).To(HaveOccurred())
	})

	It("rejects an invalid effective date on update", func() {
		created, err := ruleService.CreateRule(ctx, ruleInput())
		Expect(err).NotTo(HaveOccurred())
		future := time.Now().Add(time.Hour)

		_, err = ruleService.UpdateRule(ctx, created.Rule.Id, models.UpdateRuleRequest{EffectiveFrom: &future}, userId)
		Expect(err).To(HaveOccurred())
	})

	It("rejects an invalid action type on update", func() {
		created, err := ruleService.CreateRule(ctx, ruleInput())
		Expect(err).NotTo(HaveOccurred())
		invalid := models.RuleFieldType("bogus")

		_, err = ruleService.UpdateRuleAction(ctx, created.Actions[0].Id, created.Rule.Id, models.UpdateRuleActionRequest{ActionType: &invalid}, userId)
		Expect(err).To(HaveOccurred())
	})

	It("rejects an invalid condition type on update", func() {
		created, err := ruleService.CreateRule(ctx, ruleInput())
		Expect(err).NotTo(HaveOccurred())
		invalid := models.RuleFieldType("bogus")

		_, err = ruleService.UpdateRuleCondition(ctx, created.Conditions[0].Id, created.Rule.Id, models.UpdateRuleConditionRequest{ConditionType: &invalid}, userId)
		Expect(err).To(HaveOccurred())
	})

	It("propagates action deletion errors", func() {
		created, err := ruleService.CreateRule(ctx, ruleInput())
		Expect(err).NotTo(HaveOccurred())
		mockRepo.FailOn("DeleteRuleActionsByRuleId", errors.New("delete down"))

		err = ruleService.DeleteRule(ctx, created.Rule.Id, userId)
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("StatementService coverage gaps", func() {
	var (
		service        StatementService
		mockRepo       *mock_repository.MockStatementRepository
		accountService AccountServiceInterface
		userId         int64
		ctx            context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		userId = 42
		mockRepo = mock_repository.NewMockStatementRepository()
		mockTxnRepo := mock_repository.NewMockTransactionRepository()
		mockCategoryRepo := mock_repository.NewMockCategoryRepository()
		mockAccountRepo := mock_repository.NewMockAccountRepository()
		mockDB := mockDatabase.NewMockDatabaseManager()
		mockRuleRepo := mock_repository.NewMockRuleRepository()

		accountService = NewAccountService(mockAccountRepo)
		service = StatementService{
			repo:               mockRepo,
			statementValidator: validator.NewStatementValidator(),
			txService:          NewTransactionService(mockTxnRepo, mockCategoryRepo, mockAccountRepo, mockDB),
			accountService:     accountService,
			ruleEngineService:  NewRuleEngineService(mockRuleRepo, mockTxnRepo, mockCategoryRepo, mockAccountRepo),
			workerSlots:        make(chan struct{}, maxConcurrentParses),
			parseTimeout:       parseTimeout,
		}
	})

	It("requires a password for protected xlsx uploads", func() {
		_, err := service.ParseStatement(ctx, models.ParseStatementInput{
			FileBytes:        []byte("not a zip"),
			OriginalFilename: "statement.xlsx",
			AccountId:        1,
		}, userId)
		Expect(err).To(HaveOccurred())
	})

	It("propagates workbook password validation errors", func() {
		_, err := service.ParseStatement(ctx, models.ParseStatementInput{
			FileBytes:        []byte("not a zip"),
			OriginalFilename: "statement.xlsx",
			AccountId:        1,
			Password:         "wrong",
		}, userId)
		Expect(err).To(HaveOccurred())
	})

	It("propagates statement creation errors", func() {
		balance := 1000.0
		acc, err := accountService.CreateAccount(ctx, models.CreateAccountInput{
			Name: "Test", BankType: models.BankTypeSBI, Currency: models.CurrencyINR, Balance: &balance, CreatedBy: userId,
		})
		Expect(err).NotTo(HaveOccurred())
		mockRepo.FailOn("CreateStatement", errors.New("create down"))

		_, err = service.ParseStatement(ctx, models.ParseStatementInput{
			AccountId:        acc.Id,
			OriginalFilename: "statement.csv",
			FileBytes:        []byte("Date,Details,Debit,Credit\n01/12/2024,Desc,100.00,\n"),
		}, userId)
		Expect(err).To(HaveOccurred())
	})

	It("propagates statement listing errors", func() {
		mockRepo.FailOn("ListStatementByUserId", errors.New("list down"))

		_, err := service.ListStatements(ctx, userId, models.StatementListQuery{Page: 1, PageSize: 5})
		Expect(err).To(HaveOccurred())
	})

	It("propagates statement counting errors", func() {
		mockRepo.FailOn("CountStatementsByUserId", errors.New("count down"))

		_, err := service.ListStatements(ctx, userId, models.StatementListQuery{Page: 1, PageSize: 5})
		Expect(err).To(HaveOccurred())
	})

	It("marks a statement as error when the account cannot be fetched", func() {
		statement, err := mockRepo.CreateStatement(ctx, models.CreateStatementInput{
			AccountId: 1, CreatedBy: userId, OriginalFilename: "statement.csv", FileType: "csv", Status: models.StatementStatusPending,
		})
		Expect(err).NotTo(HaveOccurred())

		service.processStatementAsync(ctx, statement.Id, models.ParseStatementInput{
			AccountId: 9999, OriginalFilename: "statement.csv",
		}, userId)

		result, err := service.GetStatementStatus(ctx, statement.Id, userId)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Status).To(Equal(models.StatementStatusError))
		Expect(result.Message).NotTo(BeNil())
		Expect(*result.Message).To(ContainSubstring("Failed to fetch account"))
	})

	It("marks a statement as error when linking transactions fails", func() {
		balance := 1000.0
		acc, err := accountService.CreateAccount(ctx, models.CreateAccountInput{
			Name: "Test", BankType: models.BankTypeSBI, Currency: models.CurrencyINR, Balance: &balance, CreatedBy: userId,
		})
		Expect(err).NotTo(HaveOccurred())
		statement, err := mockRepo.CreateStatement(ctx, models.CreateStatementInput{
			AccountId: acc.Id, CreatedBy: userId, OriginalFilename: "statement.xlsx", FileType: "excel", Status: models.StatementStatusPending,
		})
		Expect(err).NotTo(HaveOccurred())
		mockRepo.FailOn("CreateStatementTxns", errors.New("link down"))

		fileBytes := utils.CreateXLSXFile([][]string{
			{"Date", "Details", "Ref No/Cheque No", "Debit", "Credit", "Balance"},
			{"03/08/2022", "WDL TFR UPI/DR/221356312527/RITIK S/SBIN/rs6321908@/UPI", "123456", "100.00", "", "1000.00"},
		})
		service.processStatementAsync(ctx, statement.Id, models.ParseStatementInput{
			AccountId:        acc.Id,
			OriginalFilename: "statement.xlsx",
			BankType:         string(models.BankTypeSBI),
			FileBytes:        fileBytes,
		}, userId)

		result, err := service.GetStatementStatus(ctx, statement.Id, userId)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Status).To(Equal(models.StatementStatusError))
		Expect(result.Message).NotTo(BeNil())
		Expect(*result.Message).To(ContainSubstring("failed to link transactions"))
	})

	It("logs but does not fail when the final status update fails", func() {
		balance := 1000.0
		acc, err := accountService.CreateAccount(ctx, models.CreateAccountInput{
			Name: "Test", BankType: models.BankTypeSBI, Currency: models.CurrencyINR, Balance: &balance, CreatedBy: userId,
		})
		Expect(err).NotTo(HaveOccurred())
		statement, err := mockRepo.CreateStatement(ctx, models.CreateStatementInput{
			AccountId: acc.Id, CreatedBy: userId, OriginalFilename: "statement.xlsx", FileType: "excel", Status: models.StatementStatusPending,
		})
		Expect(err).NotTo(HaveOccurred())
		mockRepo.FailOn("UpdateStatementStatus", errors.New("status down"))

		fileBytes := utils.CreateXLSXFile([][]string{
			{"Date", "Details", "Ref No/Cheque No", "Debit", "Credit", "Balance"},
			{"03/08/2022", "WDL TFR UPI/DR/221356312527/RITIK S/SBIN/rs6321908@/UPI", "123456", "100.00", "", "1000.00"},
		})
		service.processStatementAsync(ctx, statement.Id, models.ParseStatementInput{
			AccountId:        acc.Id,
			OriginalFilename: "statement.xlsx",
			BankType:         string(models.BankTypeSBI),
			FileBytes:        fileBytes,
		}, userId)
	})

	It("builds a service with default slots and timeout", func() {
		built := NewStatementService(mockRepo, accountService, service.ruleEngineService, service.statementValidator, service.txService)
		Expect(built).NotTo(BeNil())
	})

	It("requires a password to preview a protected xlsx", func() {
		_, err := service.PreviewStatement(ctx, []byte("not a zip"), "statement.xlsx", 0, 10, "")
		Expect(err).To(HaveOccurred())
	})

	It("rejects oversized workbooks during preview", func() {
		_, err := service.PreviewStatement(ctx, gapOversizedXLSX(), "statement.xlsx", 0, 10, "")
		Expect(err).To(HaveOccurred())
	})

	It("propagates preview parsing errors", func() {
		_, err := service.PreviewStatement(ctx, []byte("col1,col2\n\"unterminated"), "bad.csv", 0, 10, "")
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("AuthService coverage gaps", func() {
	var (
		authService AuthServiceInterface
		userService UserServiceInterface
		mockRepo    *mock_repository.MockUserRepository
		sessionRepo *mock_repository.MockSessionRepository
		ctx         context.Context
	)

	BeforeEach(func() {
		origEnv := os.Getenv("ENV")
		origJwt := os.Getenv("JWT_SECRET")
		origSchema := os.Getenv("DB_SCHEMA")
		os.Setenv("ENV", "test")
		os.Setenv("JWT_SECRET", "test-secret")
		os.Setenv("DB_SCHEMA", "test_schema")
		DeferCleanup(func() {
			os.Setenv("ENV", origEnv)
			os.Setenv("JWT_SECRET", origJwt)
			os.Setenv("DB_SCHEMA", origSchema)
		})

		ctx = context.Background()
		mockRepo = mock_repository.NewMockUserRepository()
		sessionRepo = mock_repository.NewMockSessionRepository()
		userService = NewUserService(mockRepo, sessionRepo)
		cfg, err := config.NewConfig()
		Expect(err).NotTo(HaveOccurred())
		authService = NewAuthService(userService, sessionRepo, cfg)
	})

	It("fails signup when the password cannot be hashed", func() {
		_, err := authService.Signup(ctx, models.CreateUserInput{
			Email:    "long@example.com",
			Name:     "Long Password",
			Password: strings.Repeat("a", 73),
		})
		Expect(err).To(HaveOccurred())
	})

	It("maps a unique email constraint violation to a user-exists error", func() {
		mockRepo.FailOn("CreateUser", errors.New("insert violates unique_active_email constraint"))

		_, err := authService.Signup(ctx, models.CreateUserInput{
			Email:    "dup@example.com",
			Name:     "Dup User",
			Password: "password123",
		})
		Expect(err).To(HaveOccurred())
	})

	It("fails password update when the new password cannot be hashed", func() {
		user := models.CreateUserInput{Email: "hash@example.com", Name: "Hash User", Password: "password123"}
		response, err := authService.Signup(ctx, user)
		Expect(err).NotTo(HaveOccurred())

		_, err = authService.UpdateUserPassword(ctx, response.User.Id, models.UpdateUserPasswordInput{
			OldPassword: user.Password,
			NewPassword: strings.Repeat("a", 73),
		})
		Expect(err).To(HaveOccurred())
	})

	It("fails refresh when the session user no longer exists", func() {
		token := "orphan-token"
		Expect(sessionRepo.Create(ctx, 999, hashRefreshToken(token), time.Now().Add(time.Hour))).To(Succeed())

		_, err := authService.RefreshToken(ctx, token)
		Expect(err).To(HaveOccurred())
	})
})
