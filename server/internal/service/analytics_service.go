package service

import (
	"context"
	"expenses/internal/models"
	"expenses/internal/repository"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type AnalyticsServiceInterface interface {
	GetAccountAnalytics(ctx context.Context, userId int64) (models.AccountAnalyticsListResponse, error)
	GetNetworthTimeSeries(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (models.NetworthTimeSeriesResponse, error)
	GetCategoryAnalytics(ctx context.Context, userId int64, startDate time.Time, endDate time.Time, categoryIds []int64) (*models.CategoryAnalyticsResponse, error)
	GetMonthlyAnalytics(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (*models.MonthlyAnalyticsResponse, error)
	GetInsights(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (*models.AnalyticsInsightsResponse, error)
}

type AnalyticsService struct {
	analyticsRepo repository.AnalyticsRepositoryInterface
	accountRepo   repository.AccountRepositoryInterface
}

func NewAnalyticsService(analyticsRepo repository.AnalyticsRepositoryInterface, accountRepo repository.AccountRepositoryInterface) AnalyticsServiceInterface {
	return &AnalyticsService{
		analyticsRepo: analyticsRepo,
		accountRepo:   accountRepo,
	}
}

func (s *AnalyticsService) GetAccountAnalytics(ctx context.Context, userId int64) (models.AccountAnalyticsListResponse, error) {
	// Get all user accounts to ensure we include accounts with no transactions
	accounts, err := s.accountRepo.ListAccounts(ctx, userId)
	if err != nil {
		// If no accounts found, return empty analytics (not an error)
		return models.AccountAnalyticsListResponse{
			AccountAnalytics: []models.AccountBalanceAnalytics{},
		}, nil
	}

	// Get current balances (all transactions)
	currentBalances, err := s.analyticsRepo.GetBalance(ctx, userId, nil, nil)
	if err != nil {
		return models.AccountAnalyticsListResponse{}, err
	}

	// Calculate one month ago date
	oneMonthAgo := time.Now().AddDate(0, -1, 0)

	// Get balances from one month ago
	historicalBalances, err := s.analyticsRepo.GetBalance(ctx, userId, nil, &oneMonthAgo)
	if err != nil {
		return models.AccountAnalyticsListResponse{}, err
	}

	// Build analytics response ensuring all accounts are included
	investmentAccountIds := make([]int64, 0)
	for _, account := range accounts {
		if account.BankType == models.BankTypeInvestment && account.CurrentValue != nil {
			investmentAccountIds = append(investmentAccountIds, account.Id)
		}
	}

	cashFlowsByAccount := make(map[int64][]models.AccountCashFlow)
	if len(investmentAccountIds) > 0 {
		cashFlows, err := s.analyticsRepo.GetAccountCashFlows(ctx, userId, investmentAccountIds)
		if err != nil {
			return models.AccountAnalyticsListResponse{}, err
		}
		for _, flow := range cashFlows {
			cashFlowsByAccount[flow.AccountID] = append(cashFlowsByAccount[flow.AccountID], flow)
		}
	}

	var accountAnalytics []models.AccountBalanceAnalytics
	now := time.Now()
	for _, account := range accounts {
		currentBalance := currentBalances[account.Id]       // defaults to 0 if not found
		historicalBalance := historicalBalances[account.Id] // defaults to 0 if not found

		analytics := models.AccountBalanceAnalytics{
			AccountID:          account.Id,
			CurrentBalance:     currentBalance,
			BalanceOneMonthAgo: historicalBalance,
		}

		if account.BankType == models.BankTypeInvestment && account.CurrentValue != nil {
			currentValue := *account.CurrentValue
			percentageIncrease, xirr := calculateInvestmentMetrics(cashFlowsByAccount[account.Id], currentValue, now)
			analytics.CurrentValue = &currentValue
			analytics.PercentageIncrease = &percentageIncrease
			analytics.Xirr = xirr
		}

		accountAnalytics = append(accountAnalytics, analytics)
	}

	return models.AccountAnalyticsListResponse{
		AccountAnalytics: accountAnalytics,
	}, nil
}

func (s *AnalyticsService) GetNetworthTimeSeries(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (models.NetworthTimeSeriesResponse, error) {
	accounts, err := s.accountRepo.ListAccounts(ctx, userId)

	if err != nil {
		return models.NetworthTimeSeriesResponse{}, err
	}

	// Get initial balance and daily changes from repository
	initialBalance, totalIncome, totalExpenses, dailyData, err := s.analyticsRepo.GetNetworthTimeSeries(ctx, userId, startDate, endDate)
	totalAccountBalance := 0.0
	if err != nil {
		return models.NetworthTimeSeriesResponse{}, err
	}

	for _, account := range accounts {
		initialBalance += account.Balance
		totalAccountBalance += account.Balance
	}

	var timeSeries []models.NetworthDataPoint
	runningBalance := initialBalance

	// Create a map of dates with daily changes for easy lookup
	dailyChanges := make(map[string]float64)
	for _, data := range dailyData {
		date, ok := data["date"].(string)
		if !ok {
			return models.NetworthTimeSeriesResponse{}, fmt.Errorf("invalid type for date in daily data")
		}
		dailyChange, ok := data["daily_change"].(float64)
		if !ok {
			return models.NetworthTimeSeriesResponse{}, fmt.Errorf("invalid type for daily_change in daily data")
		}
		dailyChanges[date] = dailyChange
	}

	// Generate time series for each day in the range
	currentDate := startDate
	for currentDate.Before(endDate.AddDate(0, 0, 1)) { // Include end date
		dateStr := currentDate.Format("2006-01-02")

		// Add daily change if it exists
		if dailyChange, exists := dailyChanges[dateStr]; exists {
			runningBalance += dailyChange
		}

		if runningBalance == totalAccountBalance {
			// Txn has not changed yet, so we can skip adding this point
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}

		timeSeries = append(timeSeries, models.NetworthDataPoint{
			Date:     dateStr,
			Networth: runningBalance,
		})

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	if len(timeSeries) == 0 {
		timeSeries = append(timeSeries, models.NetworthDataPoint{
			Date:     startDate.Format("2006-01-02"),
			Networth: initialBalance,
		})
	}

	return models.NetworthTimeSeriesResponse{
		InitialBalance: initialBalance, // Initial balance for frontend
		TotalIncome:    totalIncome,
		TotalExpenses:  totalExpenses,
		TimeSeries:     timeSeries,
	}, nil
}

func (s *AnalyticsService) GetCategoryAnalytics(ctx context.Context, userId int64, startDate time.Time, endDate time.Time, categoryIds []int64) (*models.CategoryAnalyticsResponse, error) {
	return s.analyticsRepo.GetCategoryAnalytics(ctx, userId, startDate, endDate, categoryIds)
}

func (s *AnalyticsService) GetMonthlyAnalytics(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (*models.MonthlyAnalyticsResponse, error) {
	// Validate input - endDate should be after or equal to startDate
	if endDate.Before(startDate) {
		return nil, fmt.Errorf("end date must be after or equal to start date")
	}

	return s.analyticsRepo.GetMonthlyAnalytics(ctx, userId, startDate, endDate)
}

// GetInsights returns a household-scoped view of the ledger: transfers and
// investment ledgers are excluded from the period figures, while the hero
// net worth is a point-in-time mark of every account.
func (s *AnalyticsService) GetInsights(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (*models.AnalyticsInsightsResponse, error) {
	if endDate.Before(startDate) {
		return nil, fmt.Errorf("end date must be after or equal to start date")
	}

	accounts, err := s.accountRepo.ListAccounts(ctx, userId)
	if err != nil {
		return nil, err
	}

	currentBalances, err := s.analyticsRepo.GetBalance(ctx, userId, nil, nil)
	if err != nil {
		return nil, err
	}

	netWorth := 0.0
	investmentValue := 0.0
	bankValue := 0.0
	investmentAccountIds := make([]int64, 0)
	for _, account := range accounts {
		if account.BankType == models.BankTypeInvestment && account.CurrentValue != nil {
			value := *account.CurrentValue
			netWorth += value
			investmentValue += value
			investmentAccountIds = append(investmentAccountIds, account.Id)
			continue
		}

		balance := currentBalances[account.Id] + account.Balance
		netWorth += balance
		if account.BankType != models.BankTypeInvestment {
			bankValue += balance
		}
	}

	monthly, err := s.analyticsRepo.GetInsightsMonthly(ctx, userId, startDate, endDate)
	if err != nil {
		return nil, err
	}
	monthly = fillInsightsMonths(monthly, startDate, endDate)

	periodIncome := 0.0
	periodExpenses := 0.0
	for _, point := range monthly {
		periodIncome += point.Income
		periodExpenses += point.Expenses
	}
	periodNet := periodIncome - periodExpenses
	savingsRate := 0.0
	if periodIncome > 0 {
		savingsRate = periodNet / periodIncome
	}

	categories, err := s.analyticsRepo.GetInsightsCategories(ctx, userId, startDate, endDate)
	if err != nil {
		return nil, err
	}

	topExpenses, err := s.analyticsRepo.GetInsightsTopExpenses(ctx, userId, startDate, endDate)
	if err != nil {
		return nil, err
	}

	uncategorizedCount, uncategorizedAmount, err := s.analyticsRepo.GetInsightsUncategorized(ctx, userId, startDate, endDate)
	if err != nil {
		return nil, err
	}

	flowsByAccount := make(map[int64][]models.AccountCashFlow)
	realizedInterest := 0.0
	if len(investmentAccountIds) > 0 {
		cashFlows, err := s.analyticsRepo.GetAccountCashFlows(ctx, userId, investmentAccountIds)
		if err != nil {
			return nil, err
		}
		for _, flow := range cashFlows {
			flowsByAccount[flow.AccountID] = append(flowsByAccount[flow.AccountID], flow)
			if flow.Amount > 0 && isRealizedInterest(flow.Name) && isInDateRange(flow.Date, startDate, endDate) {
				realizedInterest += flow.Amount
			}
		}
	}

	now := time.Now()
	investments := make([]models.InsightsInvestment, 0, len(investmentAccountIds))
	for _, account := range accounts {
		if account.BankType != models.BankTypeInvestment || account.CurrentValue == nil {
			continue
		}
		collected := collectInvestmentCashFlows(flowsByAccount[account.Id])
		percentageIncrease, xirr := calculateInvestmentMetrics(flowsByAccount[account.Id], *account.CurrentValue, now)
		investments = append(investments, models.InsightsInvestment{
			AccountID:          account.Id,
			Name:               account.Name,
			CurrentValue:       *account.CurrentValue,
			Contributed:        collected.contributed,
			Distributed:        collected.distributed,
			RealizedInterest:   collected.realizedInterest,
			Xirr:               xirr,
			PercentageIncrease: &percentageIncrease,
		})
	}

	return &models.AnalyticsInsightsResponse{
		Summary: models.InsightsSummary{
			NetWorth:            netWorth,
			InvestmentValue:     investmentValue,
			BankValue:           bankValue,
			PeriodIncome:        periodIncome,
			PeriodExpenses:      periodExpenses,
			PeriodNet:           periodNet,
			SavingsRate:         savingsRate,
			UncategorizedCount:  uncategorizedCount,
			UncategorizedAmount: uncategorizedAmount,
			RealizedInterest:    realizedInterest,
		},
		Monthly:     monthly,
		Categories:  categories,
		TopExpenses: topExpenses,
		Investments: investments,
	}, nil
}

// fillInsightsMonths inserts zero months so the chart axis stays continuous.
func fillInsightsMonths(points []models.InsightsMonthlyPoint, startDate, endDate time.Time) []models.InsightsMonthlyPoint {
	byMonth := make(map[string]models.InsightsMonthlyPoint, len(points))
	for _, point := range points {
		byMonth[point.Month] = point
	}

	result := make([]models.InsightsMonthlyPoint, 0)
	cursor := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	last := time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	for !cursor.After(last) {
		key := cursor.Format("2006-01")
		if point, ok := byMonth[key]; ok {
			result = append(result, point)
		} else {
			result = append(result, models.InsightsMonthlyPoint{Month: key})
		}
		cursor = cursor.AddDate(0, 1, 0)
	}
	return result
}

func isInDateRange(date, startDate, endDate time.Time) bool {
	return !date.Before(startDate) && !date.After(endDate)
}

type investmentCashFlow struct {
	amount float64
	date   time.Time
}

// collectedInvestmentCashFlows holds the XIRR inputs together with the
// aggregates derived from the exact same rows, so contributed, distributed
// and realized interest cannot drift from what XIRR sees.
type collectedInvestmentCashFlows struct {
	flows            []investmentCashFlow
	contributed      float64
	distributed      float64
	realizedInterest float64
}

// isInterestCredit matches the FD bookkeeping row that mirrors a positive
// interest exit. Keeping it would cancel the realized coupon in XIRR.
func isInterestCredit(name string) bool {
	return strings.Contains(strings.ToLower(name), "interest credit")
}

// isRealizedInterest matches coupon rows, including the FD interest exit
// rows, but never the interest-credit bookkeeping counterpart.
func isRealizedInterest(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "interest") && !strings.Contains(lower, "interest credit")
}

func collectInvestmentCashFlows(flows []models.AccountCashFlow) collectedInvestmentCashFlows {
	result := collectedInvestmentCashFlows{flows: make([]investmentCashFlow, 0, len(flows))}
	for _, flow := range flows {
		if isInterestCredit(flow.Name) {
			continue
		}
		if flow.Amount == 0 {
			continue
		}
		result.flows = append(result.flows, investmentCashFlow{amount: flow.Amount, date: flow.Date})
		if flow.Amount < 0 {
			result.contributed += -flow.Amount
			continue
		}
		result.distributed += flow.Amount
		if isRealizedInterest(flow.Name) {
			result.realizedInterest += flow.Amount
		}
	}
	return result
}

func calculateInvestmentMetrics(flows []models.AccountCashFlow, currentValue float64, now time.Time) (float64, *float64) {
	// If current value is zero or negative, percentage and XIRR should be zero
	if currentValue <= 0 {
		zero := 0.0
		return 0, &zero
	}

	// If there are no flows, XIRR is defined as zero (no history)
	if len(flows) == 0 {
		zero := 0.0
		return 0, &zero
	}

	// Build cash flows keeping the sign semantics from models: investments should be negative, inflows positive
	collected := collectInvestmentCashFlows(flows)
	cashFlows := collected.flows
	totalInvested := collected.contributed

	percentageIncrease := 0.0
	if totalInvested > 0 {
		percentageIncrease = ((currentValue - totalInvested) / totalInvested) * 100
	}

	// Append current value as the final inflow
	cashFlows = append(cashFlows, investmentCashFlow{amount: currentValue, date: now})

	// Special-case: single negative investment flow (one-time investment) -> compute analytically
	negCount := 0
	var negFlow investmentCashFlow
	for _, f := range cashFlows[:len(cashFlows)-1] { // exclude the final current value
		if f.amount < 0 {
			negCount++
			negFlow = f
		}
	}
	if negCount == 1 && len(cashFlows) == 2 {
		days := cashFlows[1].date.Sub(negFlow.date).Hours() / 24
		years := days / 365.0
		if years <= 0 {
			zero := 0.0
			return percentageIncrease, &zero
		}
		if -negFlow.amount <= 0 {
			zero := 0.0
			return percentageIncrease, &zero
		}
		ratio := currentValue / (-negFlow.amount)
		if ratio <= 0 {
			zero := 0.0
			return percentageIncrease, &zero
		}
		rate := math.Pow(ratio, 1.0/years) - 1.0
		// If rate is extremely small, return nil to indicate undefined
		if math.Abs(rate) < 1e-12 {
			return percentageIncrease, nil
		}
		x := rate * 100
		return percentageIncrease, &x
	}

	// If all provided flows (excluding the current value) are on the same date and there are multiple flows, we treat XIRR as zero
	if len(cashFlows) > 2 {
		firstDate := cashFlows[0].date
		allSameDate := true
		for i := 0; i < len(cashFlows)-1; i++ {
			if !sameDay(firstDate, cashFlows[i].date) {
				allSameDate = false
				break
			}
		}
		if allSameDate {
			zero := 0.0
			return percentageIncrease, &zero
		}
	}

	// Calculate XIRR using Newton/bisection fallback
	xirrVal, ok := calculateXIRR(cashFlows)
	if !ok {
		zero := 0.0
		return percentageIncrease, &zero
	}

	// If XIRR is extremely small, treat it as undefined (return nil)
	if math.Abs(xirrVal) < 1e-12 {
		return percentageIncrease, nil
	}

	x := xirrVal * 100
	return percentageIncrease, &x
}

// sameDay compares dates ignoring the time component
func sameDay(a, b time.Time) bool {
	aYear, aMonth, aDay := a.Date()
	bYear, bMonth, bDay := b.Date()
	return aYear == bYear && aMonth == bMonth && aDay == bDay
}

func calculateXIRR(cashFlows []investmentCashFlow) (float64, bool) {
	if len(cashFlows) < 2 {
		return 0, false
	}

	// Sort flows by date (ascending) for deterministic behavior
	sort.Slice(cashFlows, func(i, j int) bool {
		return cashFlows[i].date.Before(cashFlows[j].date)
	})

	hasNegative := false
	hasPositive := false
	baseDate := cashFlows[0].date
	for _, flow := range cashFlows {
		if flow.amount < 0 {
			hasNegative = true
		}
		if flow.amount > 0 {
			hasPositive = true
		}
		if flow.date.Before(baseDate) {
			baseDate = flow.date
		}
	}

	if !hasNegative || !hasPositive {
		return 0, false
	}

	// Try several initial guesses to improve chances of convergence
	initialGuesses := []float64{0.1, 0.5, 1.0, 0.0, -0.5, 0.2}
	for _, g := range initialGuesses {
		guess := g
		for i := 0; i < 200; i++ {
			npv, derivative := xirrNpvAndDerivative(guess, cashFlows, baseDate)
			if math.IsNaN(npv) || math.IsInf(npv, 0) {
				break
			}
			if math.Abs(npv) < 1e-9 {
				return guess, true
			}
			if derivative == 0 {
				break
			}
			next := guess - npv/derivative
			if next <= -0.999999 {
				next = (guess - 0.999999) / 2
			}
			if math.IsNaN(next) || math.IsInf(next, 0) {
				break
			}
			if math.Abs(next-guess) < 1e-9 {
				return next, true
			}
			guess = next
		}
	}

	return 0, false
}

func xirrNpvAndDerivative(rate float64, cashFlows []investmentCashFlow, baseDate time.Time) (float64, float64) {
	npv := 0.0
	derivative := 0.0
	for _, flow := range cashFlows {
		days := flow.date.Sub(baseDate).Hours() / 24
		years := days / 365
		denominator := math.Pow(1+rate, years)
		npv += flow.amount / denominator
		derivative -= (years * flow.amount) / (denominator * (1 + rate))
	}
	return npv, derivative
}
