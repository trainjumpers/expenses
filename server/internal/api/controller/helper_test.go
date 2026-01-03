package controller_test

import (
	"bytes"
	"encoding/json"
	"expenses/internal/models"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"strings"

	. "github.com/onsi/gomega"
)

// TestHelper encapsulates helper functions for controller tests
type TestHelper struct {
	Client       *http.Client
	BaseURL      string
	AccessToken  string
	RefreshToken string
}

// NewTestHelper creates a new TestHelper
func NewTestHelper(baseURL string) *TestHelper {
	jar, err := cookiejar.New(nil)
	Expect(err).NotTo(HaveOccurred())
	return &TestHelper{
		Client:  &http.Client{Jar: jar},
		BaseURL: baseURL,
	}
}

func (h *TestHelper) setCookies(req *http.Request) {
	// Always set access_token and refresh_token in the Cookie header if present
	var cookieHeader string
	if h.AccessToken != "" {
		cookieHeader += "access_token=" + h.AccessToken
	}
	if h.RefreshToken != "" {
		if cookieHeader != "" {
			cookieHeader += "; "
		}
		cookieHeader += "refresh_token=" + h.RefreshToken
	}
	if cookieHeader != "" {
		req.Header.Set("Cookie", cookieHeader)
	}
}

// MakeRequest performs an HTTP request and returns the response and decoded body
func (h *TestHelper) MakeRequest(method, reqUrl string, body any) (*http.Response, map[string]any) {
	var reqBody io.Reader
	if body != nil {
		if str, ok := body.(string); ok {
			reqBody = bytes.NewBuffer([]byte(str))
		} else {
			jsonBody, err := json.Marshal(body)
			Expect(err).NotTo(HaveOccurred())
			reqBody = bytes.NewBuffer(jsonBody)
		}
	}

	req, err := http.NewRequest(method, h.BaseURL+reqUrl, reqBody)
	Expect(err).NotTo(HaveOccurred())

	req.Header.Set("Content-Type", "application/json")
	h.setCookies(req)
	resp, err := h.Client.Do(req)
	Expect(err).NotTo(HaveOccurred())

	responseBody, err := decodeJSON(resp.Body)
	if err != nil && err != io.EOF { // EOF is fine if body is empty
		Expect(err).NotTo(HaveOccurred())
	}
	resp.Body.Close()

	return resp, responseBody
}

// MakeMultipartRequest sends a multipart/form-data request with file and fields
func (h *TestHelper) MakeMultipartRequest(method, url string, fields map[string]any) (*http.Response, map[string]any) {
	// Use provided original_filename if available so uploaded file has correct name
	filename := "file.csv"
	if f, ok := fields["original_filename"]; ok {
		filename = fmt.Sprintf("%v", f)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, val := range fields {
		switch v := val.(type) {
		case []byte:
			part, _ := writer.CreateFormFile(key, filename)
			part.Write(v)
		default:
			writer.WriteField(key, fmt.Sprintf("%v", v))
		}
	}
	writer.Close()
	req, _ := http.NewRequest(method, h.BaseURL+url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	h.setCookies(req)
	resp, _ := h.Client.Do(req)
	response, _ := decodeJSON(resp.Body)
	return resp, response
}

// Login performs a login request and ensures cookies are set
func (h *TestHelper) Login(email, password string) {
	loginInput := models.LoginInput{
		Email:    email,
		Password: password,
	}
	resp, body := h.MakeRequest(http.MethodPost, "/login", loginInput)

	h.AccessToken = ""
	h.RefreshToken = ""
	for _, raw := range resp.Header["Set-Cookie"] {
		parts := strings.Split(raw, " Secure ")
		for _, part := range parts {
			cookieStr := strings.TrimSpace(part)
			if idx := strings.Index(cookieStr, ";"); idx != -1 {
				cookieStr = cookieStr[:idx]
			}
			if eq := strings.Index(cookieStr, "="); eq != -1 {
				name := strings.TrimSpace(cookieStr[:eq])
				value := strings.TrimSpace(cookieStr[eq+1:])
				if name == "access_token" {
					h.AccessToken = value
				}
				if name == "refresh_token" {
					h.RefreshToken = value
				}
			}
		}
	}
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	Expect(body["message"]).To(Equal("User logged in successfully"))
}

func (h *TestHelper) MakeRequestWithToken(method, reqUrl, token string, body any) (*http.Response, map[string]any) {
	origAccessToken := h.AccessToken
	origRefreshToken := h.RefreshToken
	h.AccessToken = token
	resp, response := h.MakeRequest(method, reqUrl, body)
	h.AccessToken = origAccessToken
	h.RefreshToken = origRefreshToken

	return resp, response
}

// checkMalformedTokens tests endpoints with malformed and bad tokens for auth edge cases
func checkMalformedTokens(helper *TestHelper, method, path string, body any) {
	malformedTokens := []string{
		"invalid-token",
		"Bearer",
		"NotBearer validtoken",
		"Bearer invalid.token.format",
		"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid",
		"Bearer ",
	}
	for _, token := range malformedTokens {
		resp, _ := helper.MakeRequestWithToken(method, path, token, body)
		Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized), "Should fail for malformed token: "+token)
	}
}

// checkNetworthValidation tests validation for networth endpoint
func checkNetworthValidation(helper *TestHelper, testCases []map[string]any) {
	for _, tc := range testCases {
		url := fmt.Sprintf("/analytics/networth?start_date=%s&end_date=%s", tc["startDate"], tc["endDate"])
		resp, response := helper.MakeRequest(http.MethodGet, url, nil)
		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		Expect(response["message"]).To(Equal(tc["expectedMessage"]))
	}
}

// checkCategoryValidation tests validation for category analytics endpoint
func checkCategoryValidation(helper *TestHelper, testCases []map[string]any) {
	for _, tc := range testCases {
		url := fmt.Sprintf("/analytics/category?start_date=%s&end_date=%s", tc["startDate"], tc["endDate"])
		resp, response := helper.MakeRequest(http.MethodGet, url, nil)
		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		Expect(response["message"]).To(Equal(tc["expectedMessage"]))
	}
}
