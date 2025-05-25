package controller

import (
	"bytes"
	"encoding/json"
	"expenses/internal/models"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AuthController", func() {
	Describe("Signup", func() {
		Context("with valid input", func() {
			It("should create a new user successfully", func() {
				// Prepare test data
				userInput := models.CreateUserInput{
					Email:    "random@example.com",
					Name:     "Test User",
					Password: "password123",
				}

				// Create request
				body, _ := json.Marshal(userInput)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/signup", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")

				// Perform request
				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				// Assertions
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				response, err := decodeJSONResponse(resp)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(Equal("User signed up successfully"))
				Expect(response["data"]).To(HaveKey("user"))
				Expect(response["data"]).To(HaveKey("access_token"))
				Expect(response["data"]).To(HaveKey("refresh_token"))
			})
		})

		Context("with invalid input", func() {
			It("should return bad request for invalid email", func() {
				userInput := models.CreateUserInput{
					Email:    "invalid-email",
					Name:     "Test User",
					Password: "password123",
				}

				body, _ := json.Marshal(userInput)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/signup", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return bad request for short password", func() {
				userInput := models.CreateUserInput{
					Email:    "test@example.com",
					Name:     "Test User",
					Password: "123", // Too short
				}

				body, _ := json.Marshal(userInput)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/signup", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return conflict for existing user", func() {
				userInput := models.CreateUserInput{
					Email:    "test1@example.com", // Using email from test seed
					Name:     "Test User",
					Password: "password123",
				}

				body, _ := json.Marshal(userInput)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/signup", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusConflict))
				response, err := decodeJSONResponse(resp)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(Equal("user already exists"))
			})
		})
	})

	Describe("Login", func() {
		Context("with valid credentials", func() {
			It("should login successfully", func() {
				loginInput := models.LoginInput{
					Email:    "test1@example.com",
					Password: "password",
				}

				body, _ := json.Marshal(loginInput)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/login", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				response, err := decodeJSONResponse(resp)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(Equal("User logged in successfully"))
				Expect(response["data"]).To(HaveKey("access_token"))
				Expect(response["data"]).To(HaveKey("refresh_token"))
			})
		})

		Context("with invalid credentials", func() {
			It("should return unauthorized for wrong password", func() {
				loginInput := models.LoginInput{
					Email:    "test1@example.com",
					Password: "wrongpassword",
				}

				body, _ := json.Marshal(loginInput)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/login", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should return unauthorized for non-existent user", func() {
				loginInput := models.LoginInput{
					Email:    "nonexistent@example.com",
					Password: "password123",
				}

				body, _ := json.Marshal(loginInput)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/login", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("RefreshToken", func() {
		Context("with valid refresh token", func() {
			It("should refresh token successfully", func() {
				refreshInput := struct {
					RefreshToken string `json:"refresh_token"`
				}{
					RefreshToken: refreshToken,
				}

				body, _ := json.Marshal(refreshInput)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/refresh", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				response, err := decodeJSONResponse(resp)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(Equal("Token refreshed successfully"))
				Expect(response["data"]).To(HaveKey("access_token"))
				Expect(response["data"]).To(HaveKey("refresh_token"))

				// Verify the new tokens are different
				data := response["data"].(map[string]interface{})
				newAccessToken := data["access_token"].(string)
				newRefreshToken := data["refresh_token"].(string)
				Expect(newAccessToken).NotTo(BeEmpty())
				Expect(newRefreshToken).NotTo(BeEmpty())
				Expect(newAccessToken).NotTo(Equal(accessToken))
				Expect(newRefreshToken).NotTo(Equal(refreshToken))
			})
		})

		Context("with invalid refresh token", func() {
			It("should return unauthorized for invalid token", func() {
				refreshInput := struct {
					RefreshToken string `json:"refresh_token"`
				}{
					RefreshToken: "invalid-refresh-token",
				}

				body, _ := json.Marshal(refreshInput)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/refresh", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})
	})
})
