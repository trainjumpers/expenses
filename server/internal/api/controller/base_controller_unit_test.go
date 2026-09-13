package controller

import (
	"bytes"
	"errors"
	"expenses/internal/config"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

type unitTrimBasic struct {
	Name string
}

type unitTrimSlicePtr struct {
	Items []*unitTrimBasic
}

func newUnitContext(t *testing.T, method, target string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(method, target, nil)
	return ctx, recorder
}

func TestHandleErrorNilDoesNothing(t *testing.T) {
	ctx, recorder := newUnitContext(t, http.MethodGet, "/")
	baseController := NewBaseController(&config.Config{})

	baseController.HandleError(ctx, nil)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected no response written, got status %d", recorder.Code)
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", recorder.Body.String())
	}
}

func TestHandleErrorGenericIncludesDetailsInTestEnv(t *testing.T) {
	ctx, recorder := newUnitContext(t, http.MethodGet, "/")
	baseController := NewBaseController(&config.Config{Environment: config.EnvironmentTest})

	baseController.HandleError(ctx, errors.New("boom"))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
	response, err := decodeJSON(recorder.Body)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response["error"] != "boom" {
		t.Fatalf("expected error detail, got %v", response["error"])
	}
	if _, ok := response["stack"]; !ok {
		t.Fatalf("expected stack in response")
	}
}

func TestTrimStringFieldsNil(t *testing.T) {
	baseController := NewBaseController(&config.Config{})

	baseController.trimStringFields(nil)
}

func TestTrimStringFieldsTypedNilPointer(t *testing.T) {
	baseController := NewBaseController(&config.Config{})
	var value *unitTrimBasic

	baseController.trimStringFields(value)
}

func TestTrimStringFieldsNonStructPointer(t *testing.T) {
	baseController := NewBaseController(&config.Config{})
	value := 5

	baseController.trimStringFields(&value)

	if value != 5 {
		t.Fatalf("expected value unchanged, got %d", value)
	}
}

func TestTrimStringFieldsSliceOfStructPointers(t *testing.T) {
	baseController := NewBaseController(&config.Config{})
	value := &unitTrimSlicePtr{
		Items: []*unitTrimBasic{
			{Name: "  first  "},
			{Name: "  second  "},
		},
	}

	baseController.trimStringFields(value)

	if value.Items[0].Name != "first" {
		t.Fatalf("expected trimmed first, got %q", value.Items[0].Name)
	}
	if value.Items[1].Name != "second" {
		t.Fatalf("expected trimmed second, got %q", value.Items[1].Name)
	}
}

func TestSetAuthCookieUsesCookieDomainInProd(t *testing.T) {
	ctx, recorder := newUnitContext(t, http.MethodGet, "/")
	baseController := NewBaseController(&config.Config{
		Environment:  config.EnvironmentProd,
		CookieDomain: "example.com",
	})

	baseController.setAuthCookie(ctx, "access_token", "token-value", 60)

	header := recorder.Header().Get("Set-Cookie")
	if header == "" {
		t.Fatalf("expected Set-Cookie header")
	}
	if !bytes.Contains([]byte(header), []byte("Domain=example.com")) {
		t.Fatalf("expected domain in cookie, got %q", header)
	}
}

func TestParseBoolQueryParam(t *testing.T) {
	baseController := NewBaseController(&config.Config{})
	cases := map[string]*bool{
		"true":  boolPtr(true),
		"1":     boolPtr(true),
		"false": boolPtr(false),
		"0":     boolPtr(false),
		"":      nil,
		"bogus": nil,
	}
	for raw, expected := range cases {
		ctx, _ := newUnitContext(t, http.MethodGet, "/?flag="+raw)
		got := baseController.parseBoolQueryParam(ctx, "flag")
		if expected == nil {
			if got != nil {
				t.Fatalf("parseBoolQueryParam(%q) = %v, want nil", raw, *got)
			}
			continue
		}
		if got == nil || *got != *expected {
			t.Fatalf("parseBoolQueryParam(%q) = %v, want %v", raw, got, *expected)
		}
	}
}

func boolPtr(v bool) *bool {
	return &v
}

func TestParseCategoryIdsEmptySegment(t *testing.T) {
	if _, err := parseCategoryIds("1,,2"); err == nil {
		t.Fatalf("expected error for empty segment")
	}
}

func TestReadFileFromRequestNilHeader(t *testing.T) {
	controller := NewStatementController(&config.Config{}, nil)

	if _, _, err := controller.readFileFromRequest(nil); err == nil {
		t.Fatalf("expected error for nil file header")
	}
}

func TestReadFileFromRequestOpenError(t *testing.T) {
	controller := NewStatementController(&config.Config{}, nil)
	fileHeader, path := diskBackedFileHeader(t)
	if err := os.Remove(path); err != nil {
		t.Fatalf("failed to remove temp file: %v", err)
	}

	if _, _, err := controller.readFileFromRequest(fileHeader); err == nil {
		t.Fatalf("expected error when file header cannot be opened")
	}
}

func TestReadFileFromRequestReadError(t *testing.T) {
	controller := NewStatementController(&config.Config{}, nil)
	fileHeader, path := diskBackedFileHeader(t)
	if err := os.Remove(path); err != nil {
		t.Fatalf("failed to remove temp file: %v", err)
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	defer os.RemoveAll(path)

	if _, _, err := controller.readFileFromRequest(fileHeader); err == nil {
		t.Fatalf("expected error when file cannot be read")
	}
}

// diskBackedFileHeader parses a multipart file part that exceeds maxMemory so
// the parser stores it on disk, returning the header and its temp file path.
func diskBackedFileHeader(t *testing.T) (*multipart.FileHeader, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "statement.csv")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := part.Write([]byte("id,date,amount\n1,2026-01-01,100\n")); err != nil {
		t.Fatalf("failed to write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(1)
	if err != nil {
		t.Fatalf("failed to read form: %v", err)
	}
	fileHeader := form.File["file"][0]

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
	return fileHeader, path
}
