package controller_test

import (
	"bytes"
	"encoding/json"
	"expenses/internal/models"
	"io"
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

// MakeRequest performs an HTTP request and returns the response and decoded body
func (h *TestHelper) MakeRequest(method, reqUrl string, body interface{}) (*http.Response, map[string]interface{}) {
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

	resp, err := h.Client.Do(req)
	Expect(err).NotTo(HaveOccurred())

	responseBody, err := decodeJSON(resp.Body)
	if err != nil && err != io.EOF { // EOF is fine if body is empty
		Expect(err).NotTo(HaveOccurred())
	}
	resp.Body.Close()

	return resp, responseBody
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

func (h *TestHelper) MakeRequestWithToken(method, reqUrl, token string, body interface{}) (*http.Response, map[string]interface{}) {
	origAccessToken := h.AccessToken
	origRefreshToken := h.RefreshToken
	h.AccessToken = token
	resp, response := h.MakeRequest(method, reqUrl, body)
	h.AccessToken = origAccessToken
	h.RefreshToken = origRefreshToken

	return resp, response
}

// checkMalformedTokens tests endpoints with malformed and bad tokens for auth edge cases
func checkMalformedTokens(helper *TestHelper, method, path string, body interface{}) {
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
