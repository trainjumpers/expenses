package controller_test

import (
	"bytes"
	"encoding/json"
	"expenses/internal/models"
	"io"
	"net/http"

	. "github.com/onsi/gomega"
)

// TestHelper encapsulates helper functions for controller tests
type TestHelper struct {
	Client  *http.Client
	BaseURL string
}

// NewTestHelper creates a new TestHelper
func NewTestHelper(baseURL string) *TestHelper {
	return &TestHelper{
		Client:  &http.Client{},
		BaseURL: baseURL,
	}
}

// MakeRequest performs an HTTP request and returns the response and decoded body
func (h *TestHelper) MakeRequest(method, url, token string, body interface{}) (*http.Response, map[string]interface{}) {
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

	req, err := http.NewRequest(method, h.BaseURL+url, reqBody)
	Expect(err).NotTo(HaveOccurred())

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
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

// Login performs a login request and returns the access and refresh tokens
func (h *TestHelper) Login(email, password string) (string, string) {
	loginInput := models.LoginInput{
		Email:    email,
		Password: password,
	}
	resp, body := h.MakeRequest(http.MethodPost, "/login", "", loginInput)
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	Expect(body["message"]).To(Equal("User logged in successfully"))

	data := body["data"].(map[string]interface{})
	accessToken := data["access_token"].(string)
	refreshToken := data["refresh_token"].(string)
	return accessToken, refreshToken
}
