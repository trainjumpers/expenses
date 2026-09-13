package controller

import (
	"bytes"
	"context"
	"errors"
	"expenses/internal/config"
	"expenses/internal/models"
	"expenses/internal/service"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type failingAccountService struct {
	service.AccountServiceInterface
}

func (failingAccountService) CreateAccount(context.Context, models.CreateAccountInput) (models.AccountResponse, error) {
	return models.AccountResponse{}, errors.New("boom")
}

func (failingAccountService) ListAccounts(context.Context, int64) ([]models.AccountResponse, error) {
	return nil, errors.New("boom")
}

type failingAnalyticsService struct {
	service.AnalyticsServiceInterface
}

func (failingAnalyticsService) GetAccountAnalytics(context.Context, int64) (models.AccountAnalyticsListResponse, error) {
	return models.AccountAnalyticsListResponse{}, errors.New("boom")
}

func (failingAnalyticsService) GetCashBalanceHistory(context.Context, int64, time.Time, time.Time) (models.CashBalanceHistoryResponse, error) {
	return models.CashBalanceHistoryResponse{}, errors.New("boom")
}

func (failingAnalyticsService) GetCategoryAnalytics(context.Context, int64, time.Time, time.Time, []int64) (*models.CategoryAnalyticsResponse, error) {
	return nil, errors.New("boom")
}

func (failingAnalyticsService) GetMonthlyAnalytics(context.Context, int64, time.Time, time.Time) (*models.MonthlyAnalyticsResponse, error) {
	return nil, errors.New("boom")
}

func (failingAnalyticsService) GetInsights(context.Context, int64, time.Time, time.Time) (*models.AnalyticsInsightsResponse, error) {
	return nil, errors.New("boom")
}

type failingStatementService struct {
	service.StatementServiceInterface
}

func (failingStatementService) ListStatements(context.Context, int64, models.StatementListQuery) (models.PaginatedStatementResponse, error) {
	return models.PaginatedStatementResponse{}, errors.New("boom")
}

type failingRuleService struct{ service.RuleServiceInterface }

func (failingRuleService) ListRules(context.Context, int64, *models.RuleListQuery) (models.PaginatedRulesResponse, error) {
	return models.PaginatedRulesResponse{}, errors.New("boom")
}

type failingRuleEngineService struct {
	service.RuleEngineServiceInterface
}

func (failingRuleEngineService) ExecuteRules(context.Context, int64, models.ExecuteRulesRequest) (models.ExecuteRulesResponse, error) {
	return models.ExecuteRulesResponse{}, errors.New("boom")
}

type failingAuthService struct{ service.AuthServiceInterface }

func (failingAuthService) Logout(context.Context, string) error {
	return errors.New("boom")
}

type failingTransactionService struct {
	service.TransactionServiceInterface
}

func (failingTransactionService) ListTransactions(context.Context, int64, models.TransactionListQuery) (models.PaginatedTransactionsResponse, error) {
	return models.PaginatedTransactionsResponse{}, errors.New("boom")
}

type failingCategoryService struct {
	service.CategoryServiceInterface
}

func (failingCategoryService) ListCategories(context.Context, int64) ([]models.CategoryResponse, error) {
	return nil, errors.New("boom")
}

func newControllerContext(t *testing.T, method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(method, target, bytes.NewReader([]byte(body)))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("authUserId", int64(1))
	return ctx, recorder
}

func testConfig() *config.Config {
	return &config.Config{Environment: config.EnvironmentTest}
}

func requireStatus(t *testing.T, recorder *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if recorder.Code != expected {
		t.Fatalf("expected status %d, got %d (%s)", expected, recorder.Code, recorder.Body.String())
	}
}

func TestAnalyticsControllerServiceErrors(t *testing.T) {
	controller := NewAnalyticsController(testConfig(), failingAnalyticsService{})
	dateRange := "start_date=2026-01-01&end_date=2026-01-31"

	cases := []struct {
		name   string
		target string
		call   func(*gin.Context)
	}{
		{"GetAccountAnalytics", "/analytics/accounts", controller.GetAccountAnalytics},
		{"GetCashBalanceHistory", "/analytics/cash-balance?" + dateRange, controller.GetCashBalanceHistory},
		{"GetCategoryAnalytics", "/analytics/category?" + dateRange + "&category_ids=1,2", controller.GetCategoryAnalytics},
		{"GetMonthlyAnalytics", "/analytics/monthly?" + dateRange, controller.GetMonthlyAnalytics},
		{"GetInsights", "/analytics/insights?" + dateRange, controller.GetInsights},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, recorder := newControllerContext(t, http.MethodGet, tc.target, "")
			tc.call(ctx)
			requireStatus(t, recorder, http.StatusInternalServerError)
		})
	}
}

func TestAccountControllerServiceErrors(t *testing.T) {
	controller := NewAccountController(testConfig(), failingAccountService{})

	createCtx, createRecorder := newControllerContext(t, http.MethodPost, "/accounts", `{"name":"Cash","bank_type":"hdfc","currency":"inr"}`)
	controller.CreateAccount(createCtx)
	requireStatus(t, createRecorder, http.StatusInternalServerError)

	listCtx, listRecorder := newControllerContext(t, http.MethodGet, "/accounts", "")
	controller.ListAccounts(listCtx)
	requireStatus(t, listRecorder, http.StatusInternalServerError)
}

func TestStatementControllerServiceError(t *testing.T) {
	controller := NewStatementController(testConfig(), failingStatementService{})

	ctx, recorder := newControllerContext(t, http.MethodGet, "/statements", "")
	controller.GetStatements(ctx)
	requireStatus(t, recorder, http.StatusInternalServerError)
}

func TestStatementControllerReadFileErrors(t *testing.T) {
	controller := NewStatementController(testConfig(), nil)

	createCtx, createRecorder := newControllerContext(t, http.MethodPost, "/statements", "")
	createCtx.Request = brokenMultipartRequest(t, map[string]string{"account_id": "1"})
	controller.CreateStatement(createCtx)
	requireStatus(t, createRecorder, http.StatusBadRequest)

	previewCtx, previewRecorder := newControllerContext(t, http.MethodPost, "/statements/preview", "")
	previewCtx.Request = brokenMultipartRequest(t, nil)
	controller.PreviewStatement(previewCtx)
	requireStatus(t, previewRecorder, http.StatusInternalServerError)
}

func TestRuleControllerServiceErrors(t *testing.T) {
	controller := NewRuleController(testConfig(), failingRuleService{}, failingRuleEngineService{})

	listCtx, listRecorder := newControllerContext(t, http.MethodGet, "/rules", "")
	controller.ListRules(listCtx)
	requireStatus(t, listRecorder, http.StatusInternalServerError)

	executeCtx, executeRecorder := newControllerContext(t, http.MethodPost, "/rules/execute", `{}`)
	controller.ExecuteRules(executeCtx)
	requireStatus(t, executeRecorder, http.StatusInternalServerError)
}

func TestAuthControllerRefreshTokenMissingCookie(t *testing.T) {
	controller := NewAuthController(testConfig(), failingAuthService{})

	ctx, recorder := newControllerContext(t, http.MethodPost, "/refresh-token", "")
	controller.RefreshToken(ctx)
	requireStatus(t, recorder, http.StatusInternalServerError)
}

func TestAuthControllerLogoutServiceError(t *testing.T) {
	controller := NewAuthController(testConfig(), failingAuthService{})

	ctx, recorder := newControllerContext(t, http.MethodPost, "/logout", "")
	controller.Logout(ctx)
	requireStatus(t, recorder, http.StatusOK)
}

func TestTransactionControllerServiceError(t *testing.T) {
	controller := NewTransactionController(testConfig(), failingTransactionService{})

	ctx, recorder := newControllerContext(t, http.MethodGet, "/transactions", "")
	controller.ListTransactions(ctx)
	requireStatus(t, recorder, http.StatusInternalServerError)
}

func TestCategoryControllerServiceError(t *testing.T) {
	controller := NewCategoryController(testConfig(), failingCategoryService{})

	ctx, recorder := newControllerContext(t, http.MethodGet, "/categories", "")
	controller.ListCategories(ctx)
	requireStatus(t, recorder, http.StatusInternalServerError)
}

// brokenMultipartRequest parses a multipart request whose file part is stored
// on disk and then removes the temp file, so the controller's file read fails.
func brokenMultipartRequest(t *testing.T, fields map[string]string) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("failed to write field: %v", err)
		}
	}
	part, err := writer.CreateFormFile("file", "statement.csv")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := part.Write([]byte("id,date\n1,2026-01-01\n")); err != nil {
		t.Fatalf("failed to write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/statements", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(1); err != nil {
		t.Fatalf("failed to parse multipart form: %v", err)
	}

	fileHeader := req.MultipartForm.File["file"][0]
	file, err := fileHeader.Open()
	if err != nil {
		t.Fatalf("failed to open file header: %v", err)
	}
	osFile, ok := file.(*os.File)
	if !ok {
		t.Fatalf("expected disk-backed file, got %T", file)
	}
	path := osFile.Name()
	osFile.Close()
	if err := os.Remove(path); err != nil {
		t.Fatalf("failed to remove temp file: %v", err)
	}
	return req
}
