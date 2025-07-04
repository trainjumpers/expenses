package controller_test

import (
	"bytes"
	"encoding/json"
	"expenses/internal/models"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	. "github.com/onsi/gomega"
)

// TestHelper encapsulates helper functions for controller tests
type TestHelper struct {
	Client  *http.Client
	BaseURL string
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
func (h *TestHelper) MakeRequest(method, url string, body interface{}) (*http.Response, map[string]interface{}) {
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
	var cookies []*http.Cookie
	for _, raw := range resp.Header["Set-Cookie"] {
		// Split on ' Secure ' to separate cookies
		parts := strings.Split(raw, " Secure ")
		for _, part := range parts {
			cookieStr := strings.TrimSpace(part)
			// Only take the name and value (before the first semicolon)
			if idx := strings.Index(cookieStr, ";"); idx != -1 {
				cookieStr = cookieStr[:idx]
			}
			if eq := strings.Index(cookieStr, "="); eq != -1 {
				name := strings.TrimSpace(cookieStr[:eq])
				value := strings.TrimSpace(cookieStr[eq+1:])
				cookies = append(cookies, &http.Cookie{
					Name:  name,
					Value: value,
					Path:  "/",
				})
			}
		}
	}
	// Set cookies for the root path and correct host so they are available for all paths
	base, _ := url.Parse(h.BaseURL)
	root := &url.URL{
		Scheme: base.Scheme,
		Host:   base.Host,
		Path:   "/",
	}
	h.Client.Jar.SetCookies(root, cookies)
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	Expect(body["message"]).To(Equal("User logged in successfully"))
}
