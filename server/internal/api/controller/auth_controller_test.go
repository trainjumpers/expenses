package controller_test

import (
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
				resp, response := NewTestHelper(baseURL).MakeRequest(http.MethodPost, "/signup", userInput)

				// Assertions
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				Expect(response["message"]).To(Equal("User signed up successfully"))
				Expect(response["data"]).To(HaveKey("user"))
			})
		})

		Context("with invalid input", func() {
			It("should return bad request for invalid email", func() {
				userInput := models.CreateUserInput{
					Email:    "invalid-email",
					Name:     "Test User",
					Password: "password123",
				}

				resp, _ := NewTestHelper(baseURL).MakeRequest(http.MethodPost, "/signup", userInput)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return bad request for short password", func() {
				userInput := models.CreateUserInput{
					Email:    "test@example.com",
					Name:     "Test User",
					Password: "123", // Too short
				}

				resp, _ := NewTestHelper(baseURL).MakeRequest(http.MethodPost, "/signup", userInput)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return bad request for missing name", func() {
				userInput := models.CreateUserInput{
					Email:    "test@example.com",
					Password: "1234567890",
				}
				resp, _ := NewTestHelper(baseURL).MakeRequest(http.MethodPost, "/signup", userInput)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return bad request for missing email", func() {
				userInput := models.CreateUserInput{
					Name:     "Test User",
					Password: "1234567890",
				}
				resp, _ := NewTestHelper(baseURL).MakeRequest(http.MethodPost, "/signup", userInput)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return bad request for missing password", func() {
				userInput := models.CreateUserInput{
					Email: "test@example.com",
					Name:  "Test User",
				}
				resp, _ := NewTestHelper(baseURL).MakeRequest(http.MethodPost, "/signup", userInput)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return conflict for existing user", func() {
				userInput := models.CreateUserInput{
					Email:    "test1@example.com", // Using email from test seed
					Name:     "Test User",
					Password: "password123",
				}
				resp, response := NewTestHelper(baseURL).MakeRequest(http.MethodPost, "/signup", userInput)
				Expect(resp.StatusCode).To(Equal(http.StatusConflict))
				Expect(response["message"]).To(Equal("user already exists"))
			})

			It("should return bad request for SQL injection in email", func() {
				userInput := models.CreateUserInput{
					Email:    "test@example.com' OR '1'='1",
					Name:     "Test User",
					Password: "password123",
				}
				resp, response := NewTestHelper(baseURL).MakeRequest(http.MethodPost, "/signup", userInput)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("Error:Field validation for 'Email'"))
			})

			It("should work fine for SQL injection in name", func() {
				userInput := models.CreateUserInput{
					Email:    "test@example.com",
					Name:     "Test User'; DROP TABLE users; --",
					Password: "password123",
				}
				resp, _ := NewTestHelper(baseURL).MakeRequest(http.MethodPost, "/signup", userInput)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			})

			It("should return bad request for complex SQL injection attempt", func() {
				userInput := models.CreateUserInput{
					Email:    "test@example.com' UNION SELECT * FROM users; --",
					Name:     "Test User'; INSERT INTO users (email, name, password) VALUES ('hack@example.com', 'Hacker', 'password'); --",
					Password: "password123' OR 1=1; --",
				}
				resp, response := NewTestHelper(baseURL).MakeRequest(http.MethodPost, "/signup", userInput)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("Error:Field validation for 'Email'"))
			})

			It("should return bad request for invalid JSON", func() {
				resp, _ := NewTestHelper(baseURL).MakeRequest(http.MethodPost, "/signup", "{ invalid json }")
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return bad request for empty body", func() {
				resp, _ := NewTestHelper(baseURL).MakeRequest(http.MethodPost, "/signup", "")
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
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

				resp, response := testHelperUser1.MakeRequest(http.MethodPost, "/login", loginInput)

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(response["message"]).To(Equal("User logged in successfully"))
			})
		})

		Context("with invalid credentials", func() {
			It("should return unauthorized for wrong password", func() {
				loginInput := models.LoginInput{
					Email:    "test1@example.com",
					Password: "wrongpassword",
				}
				resp, _ := NewTestHelper(baseURL).MakeRequest(http.MethodPost, "/login", loginInput)
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should return unauthorized for non-existent user", func() {
				loginInput := models.LoginInput{
					Email:    "nonexistent@example.com",
					Password: "password123",
				}
				resp, _ := NewTestHelper(baseURL).MakeRequest(http.MethodPost, "/login", loginInput)
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should return bad request for invalid JSON", func() {
				resp, _ := NewTestHelper(baseURL).MakeRequest(http.MethodPost, "/login", "{ invalid json }")
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return bad request for empty body", func() {
				resp, _ := NewTestHelper(baseURL).MakeRequest(http.MethodPost, "/login", "")
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return bad request for missing email", func() {
				loginInput := models.LoginInput{
					Password: "password123",
				}
				resp, _ := NewTestHelper(baseURL).MakeRequest(http.MethodPost, "/login", loginInput)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return bad request for missing password", func() {
				loginInput := models.LoginInput{
					Email: "test@example.com",
				}
				resp, _ := NewTestHelper(baseURL).MakeRequest(http.MethodPost, "/login", loginInput)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("RefreshToken", func() {
		Context("with valid refresh token", func() {
			It("should refresh token successfully", func() {
				// No need to provide refreshToken, cookies are managed by http.Client
				resp, response := testHelperUser1.MakeRequest(http.MethodPost, "/refresh", nil)

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(response["message"]).To(Equal("Token refreshed successfully"))
			})
		})

		Context("with invalid refresh token", func() {
			It("should return unauthorized for invalid token", func() {
				refreshInput := struct {
					RefreshToken string `json:"refresh_token"`
				}{
					RefreshToken: "invalid-refresh-token",
				}
				resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/refresh", refreshInput)
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should return bad request for missing refresh token", func() {
				refreshInput := struct {
					RefreshToken string `json:"refresh_token"`
				}{
					RefreshToken: "",
				}

				resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/refresh", refreshInput)
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})
	})
})
