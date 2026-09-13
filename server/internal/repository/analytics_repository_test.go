package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"expenses/internal/config"
	"expenses/internal/models"
	databasemanager "expenses/pkg/database/manager"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAnalyticsRepository(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Analytics Repository Suite")
}

type analyticsFixture struct {
	userID       int64
	bankID       int64
	investmentID int64
	foodID       int64
	travelID     int64
	transfersID  int64
}

func analyticsDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

var _ = Describe("AnalyticsRepository", func() {
	var (
		ctx  context.Context
		cfg  *config.Config
		db   databasemanager.DatabaseManager
		repo AnalyticsRepositoryInterface
		f    analyticsFixture
	)

	BeforeEach(func() {
		var err error
		cfg, err = config.NewConfig()
		Expect(err).NotTo(HaveOccurred())

		db, err = databasemanager.NewDatabaseManager(cfg)
		Expect(err).NotTo(HaveOccurred())

		repo = NewAnalyticsRepository(db, cfg)
		ctx = context.Background()
		f = seedAnalyticsFixture(db, ctx, cfg.DBSchema)
	})

	AfterEach(func() {
		cleanupAnalyticsFixture(db, ctx, cfg.DBSchema, f.userID)
		Expect(db.Close()).To(Succeed())
	})

	It("allocates multi-category spending and excludes transfers", func() {
		categories, err := repo.GetInsightsCategories(ctx, f.userID, analyticsDate(2024, 4, 1), analyticsDate(2024, 4, 30))
		Expect(err).NotTo(HaveOccurred())

		byName := make(map[string]float64, len(categories))
		for _, category := range categories {
			byName[category.CategoryName] = category.TotalAmount
		}
		Expect(byName).To(HaveKeyWithValue("Food", 250.0))
		Expect(byName).To(HaveKeyWithValue("Travel", 150.0))
		Expect(byName).NotTo(HaveKey("Transfers"))
	})

	It("counts only expense-side uncategorized transactions", func() {
		count, amount, err := repo.GetInsightsUncategorized(ctx, f.userID, analyticsDate(2024, 4, 1), analyticsDate(2024, 4, 30))
		Expect(err).NotTo(HaveOccurred())
		Expect(count).To(Equal(int64(1)))
		Expect(amount).To(Equal(50.0))
	})

	It("counts multi-category transactions", func() {
		count, err := repo.GetInsightsMultiCategoryCount(ctx, f.userID, analyticsDate(2024, 4, 1), analyticsDate(2024, 4, 30))
		Expect(err).NotTo(HaveOccurred())
		Expect(count).To(Equal(int64(1)))
	})

	It("computes the spending summary for expenses only", func() {
		summary, err := repo.GetInsightsSpendingSummary(ctx, f.userID, analyticsDate(2024, 4, 1), analyticsDate(2024, 4, 30))
		Expect(err).NotTo(HaveOccurred())
		Expect(summary.ExpenseCount).To(Equal(int64(3)))
		Expect(summary.AverageTransaction).To(Equal(150.0))
		Expect(summary.MedianTransaction).To(Equal(100.0))
		Expect(summary.LargestExpense).To(Equal(300.0))
		Expect(summary.ActiveSpendingDays).To(Equal(int64(2)))
	})

	It("groups expenses by weekday", func() {
		days, err := repo.GetInsightsWeekday(ctx, f.userID, analyticsDate(2024, 4, 1), analyticsDate(2024, 4, 30))
		Expect(err).NotTo(HaveOccurred())

		byWeekday := make(map[int]models.InsightsWeekday, len(days))
		for _, day := range days {
			byWeekday[day.Weekday] = day
		}

		wednesday := byWeekday[int(time.Wednesday)]
		Expect(wednesday.Total).To(Equal(400.0))
		Expect(wednesday.Count).To(Equal(int64(2)))
		Expect(wednesday.ActiveDays).To(Equal(int64(1)))
		Expect(byWeekday[int(time.Friday)].Total).To(Equal(50.0))
	})

	It("returns expense-only category totals per month", func() {
		months, err := repo.GetInsightsCategoryMonths(ctx, f.userID, analyticsDate(2024, 4, 1), analyticsDate(2024, 4, 30))
		Expect(err).NotTo(HaveOccurred())

		totals := make(map[string]float64, len(months))
		for _, month := range months {
			totals[month.CategoryName] = month.Total
		}
		Expect(totals).To(HaveKeyWithValue("Food", 250.0))
		Expect(totals).To(HaveKeyWithValue("Travel", 150.0))
		Expect(totals).NotTo(HaveKey("Transfers"))
	})

	It("excludes investment accounts and transfers from monthly figures", func() {
		monthly, err := repo.GetMonthlyAnalytics(ctx, f.userID, analyticsDate(2024, 4, 1), analyticsDate(2024, 4, 30))
		Expect(err).NotTo(HaveOccurred())
		Expect(monthly.TotalExpenses).To(Equal(450.0))
		Expect(monthly.TotalIncome).To(Equal(200.0))
	})

	It("excludes investment ledgers from the cash balance history", func() {
		initial, income, expenses, series, err := repo.GetCashBalanceHistory(ctx, f.userID, analyticsDate(2024, 4, 1), analyticsDate(2024, 4, 30))
		Expect(err).NotTo(HaveOccurred())
		Expect(initial).To(Equal(0.0))
		Expect(income).To(Equal(200.0))
		// Transfers move cash between accounts, so they stay in the cash series;
		// only the investment ledger is excluded.
		Expect(expenses).To(Equal(950.0))
		Expect(series).NotTo(BeEmpty())
	})

	It("filters category analytics by the requested categories", func() {
		resp, err := repo.GetCategoryAnalytics(ctx, f.userID, analyticsDate(2024, 4, 1), analyticsDate(2024, 4, 30), []int64{f.foodID})
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.CategoryTransactions).To(HaveLen(1))
		Expect(resp.CategoryTransactions[0].TotalAmount).To(Equal(250.0))
	})

	It("returns the latest transaction date across accounts", func() {
		latest, err := repo.GetInsightsLatestTransactionDate(ctx, f.userID)
		Expect(err).NotTo(HaveOccurred())
		Expect(latest).NotTo(BeNil())
		Expect(latest.Format("2006-01-02")).To(Equal("2024-04-08"))
	})
})

func seedAnalyticsFixture(db databasemanager.DatabaseManager, ctx context.Context, schema string) analyticsFixture {
	f := analyticsFixture{}

	insertID := func(query string, args ...any) int64 {
		var id int64
		Expect(db.FetchOne(ctx, query, args...).Scan(&id)).To(Succeed())
		return id
	}
	exec := func(query string, args ...any) {
		_, err := db.ExecuteQuery(ctx, query, args...)
		Expect(err).NotTo(HaveOccurred())
	}
	addTransaction := func(name string, amount float64, date time.Time, accountID int64, categoryIDs ...int64) {
		id := insertID(fmt.Sprintf(
			"INSERT INTO %s.transaction (name, description, amount, date, account_id, created_by) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id",
			schema), name, "", amount, date, accountID, f.userID)
		for _, categoryID := range categoryIDs {
			exec(fmt.Sprintf(
				"INSERT INTO %s.transaction_category_mapping (category_id, transaction_id) VALUES ($1,$2)",
				schema), categoryID, id)
		}
	}

	f.userID = insertID(fmt.Sprintf(
		"INSERT INTO %s.user (name, email, password) VALUES ($1,$2,$3) RETURNING id",
		schema), "Analytics Repository", fmt.Sprintf("analytics-repo-%d@example.com", time.Now().UnixNano()), "x")

	f.bankID = insertID(fmt.Sprintf(
		"INSERT INTO %s.account (name, balance, bank_type, currency, created_by) VALUES ($1,$2,$3,$4,$5) RETURNING id",
		schema), "Repository Bank", 1000.0, "sbi", "inr", f.userID)

	f.investmentID = insertID(fmt.Sprintf(
		"INSERT INTO %s.account (name, balance, bank_type, currency, created_by) VALUES ($1,$2,$3,$4,$5) RETURNING id",
		schema), "Repository Investment", 0.0, "investment", "inr", f.userID)

	insertCategory := func(name string) int64 {
		return insertID(fmt.Sprintf(
			"INSERT INTO %s.categories (name, icon, created_by) VALUES ($1,$2,$3) RETURNING id",
			schema), name, "icon", f.userID)
	}
	f.foodID = insertCategory("Food")
	f.travelID = insertCategory("Travel")
	f.transfersID = insertCategory("Transfers")

	addTransaction("Repository Split", 300.0, analyticsDate(2024, 4, 3), f.bankID, f.foodID, f.travelID)
	addTransaction("Repository Food", 100.0, analyticsDate(2024, 4, 3), f.bankID, f.foodID)
	addTransaction("Repository Transfer", 500.0, analyticsDate(2024, 4, 4), f.bankID, f.transfersID)
	addTransaction("Repository Income", -200.0, analyticsDate(2024, 4, 4), f.bankID)
	addTransaction("Repository Uncategorized", 50.0, analyticsDate(2024, 4, 5), f.bankID)
	addTransaction("Repository Investment Debit", 1000.0, analyticsDate(2024, 4, 8), f.investmentID)

	return f
}

func cleanupAnalyticsFixture(db databasemanager.DatabaseManager, ctx context.Context, schema string, userID int64) {
	statements := []string{
		fmt.Sprintf("DELETE FROM %s.transaction_category_mapping WHERE transaction_id IN (SELECT id FROM %s.transaction WHERE created_by=$1)", schema, schema),
		fmt.Sprintf("DELETE FROM %s.transaction WHERE created_by=$1", schema),
		fmt.Sprintf("DELETE FROM %s.categories WHERE created_by=$1", schema),
		fmt.Sprintf("DELETE FROM %s.account WHERE created_by=$1", schema),
		fmt.Sprintf("DELETE FROM %s.user WHERE id=$1", schema),
	}
	for _, statement := range statements {
		_, err := db.ExecuteQuery(ctx, statement, userID)
		Expect(err).NotTo(HaveOccurred())
	}
}
