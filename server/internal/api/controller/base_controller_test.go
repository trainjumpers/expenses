package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"expenses/internal/config"
	customErrors "expenses/internal/errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// decodeJSON is a helper function to decode JSON from any io.Reader
func decodeJSON(reader io.Reader) (map[string]interface{}, error) {
	var response map[string]interface{}
	err := json.NewDecoder(reader).Decode(&response)
	return response, err
}

// Test structs for whitespace trimming validation
type TestBasicStruct struct {
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Description string `json:"description"`
}

type TestNestedStruct struct {
	User    TestBasicStruct `json:"user" binding:"required"`
	Company string          `json:"company" binding:"required"`
}

type TestPointerStruct struct {
	Name        *string `json:"name" binding:"required"`
	Description *string `json:"description"`
}

type TestSliceStruct struct {
	Users []TestBasicStruct `json:"users" binding:"required"`
	Tags  []string          `json:"tags"`
}

var _ = Describe("BaseController", func() {
	var (
		baseController *BaseController
		recorder       *httptest.ResponseRecorder
		ctx            *gin.Context
	)

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		recorder = httptest.NewRecorder()
		ctx, _ = gin.CreateTestContext(recorder)
		baseController = NewBaseController(&config.Config{})
	})

	Describe("HandleError", func() {
		Context("with AuthError", func() {
			It("should handle auth error correctly", func() {
				authErr := &customErrors.AuthError{
					Message: "Authentication failed",
					Status:  http.StatusUnauthorized,
				}
				baseController.HandleError(ctx, authErr)

				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
				response, err := decodeJSON(recorder.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(Equal("Authentication failed"))
			})
		})

		Context("with generic error", func() {
			It("should handle generic error correctly", func() {
				err := errors.New("something went wrong")
				baseController.HandleError(ctx, err)

				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				response, err := decodeJSON(recorder.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(Equal("Something went wrong"))
			})
		})
	})

	Describe("SendSuccess", func() {
		It("should send success response with data", func() {
			data := map[string]string{"key": "value"}
			baseController.SendSuccess(ctx, http.StatusOK, "Success", data)

			Expect(recorder.Code).To(Equal(http.StatusOK))
			response, err := decodeJSON(recorder.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Success"))
			Expect(response["data"]).To(HaveKey("key"))
		})

		It("should send success response without data", func() {
			baseController.SendSuccess(ctx, http.StatusOK, "Success", nil)

			Expect(recorder.Code).To(Equal(http.StatusOK))
			response, err := decodeJSON(recorder.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Success"))
			Expect(response).NotTo(HaveKey("data"))
		})
	})

	Describe("SendError", func() {
		It("should send error response", func() {
			baseController.SendError(ctx, http.StatusBadRequest, "Invalid input")

			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			response, err := decodeJSON(recorder.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Invalid input"))
		})
	})

	Describe("BindJSON", func() {
		It("should bind JSON successfully", func() {
			type TestStruct struct {
				Name string `json:"name"`
			}
			ctx.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{"name": "test"}`))
			ctx.Request.Header.Set("Content-Type", "application/json")

			var testStruct TestStruct
			err := baseController.BindJSON(ctx, &testStruct)
			Expect(err).NotTo(HaveOccurred())
			Expect(testStruct.Name).To(Equal("test"))
		})

		It("should return error for invalid JSON", func() {
			ctx.Request = httptest.NewRequest("POST", "/", strings.NewReader(`invalid json`))
			ctx.Request.Header.Set("Content-Type", "application/json")

			var testStruct struct{}
			err := baseController.BindJSON(ctx, &testStruct)
			Expect(err).To(HaveOccurred())
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Describe("BindQuery", func() {
		It("should bind query parameters successfully", func() {
			type TestStruct struct {
				Name string `form:"name"`
			}
			ctx.Request = httptest.NewRequest("GET", "/?name=test", nil)

			var testStruct TestStruct
			err := baseController.BindQuery(ctx, &testStruct)
			Expect(err).NotTo(HaveOccurred())
			Expect(testStruct.Name).To(Equal("test"))
		})

		It("should return error for invalid query parameters", func() {
			type TestStruct struct {
				Name int `form:"name"`
			}
			ctx.Request = httptest.NewRequest("GET", "/?name=invalid", nil)

			var testStruct TestStruct
			err := baseController.BindQuery(ctx, &testStruct)
			Expect(err).To(HaveOccurred())
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Describe("BindURI", func() {
		It("should bind URI parameters successfully", func() {
			type TestStruct struct {
				Id string `uri:"id"`
			}
			ctx.Params = []gin.Param{{Key: "id", Value: "123"}}

			var testStruct TestStruct
			err := baseController.BindURI(ctx, &testStruct)
			Expect(err).NotTo(HaveOccurred())
			Expect(testStruct.Id).To(Equal("123"))
		})

		It("should return error for missing URI parameters", func() {
			type TestStruct struct {
				Id string `uri:"id" binding:"required"`
			}
			ctx.Params = []gin.Param{}

			var testStruct TestStruct
			err := baseController.BindURI(ctx, &testStruct)
			Expect(err).To(HaveOccurred())
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Describe("BindForm", func() {
		It("should bind form data successfully", func() {
			type TestStruct struct {
				Name string `form:"name"`
			}
			ctx.Request = httptest.NewRequest("POST", "/", strings.NewReader("name=test"))
			ctx.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			var testStruct TestStruct
			err := baseController.BindForm(ctx, &testStruct)
			Expect(err).NotTo(HaveOccurred())
			Expect(testStruct.Name).To(Equal("test"))
		})

		It("should return error for invalid form data", func() {
			type TestStruct struct {
				Name int `form:"name" binding:"required"`
			}
			ctx.Request = httptest.NewRequest("POST", "/", strings.NewReader("name=invalid"))
			ctx.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			var testStruct TestStruct
			err := baseController.BindForm(ctx, &testStruct)
			Expect(err).To(HaveOccurred())
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Describe("Whitespace Trimming Functionality", func() {
		Context("trimStringFields method", func() {
			It("should trim whitespace from basic string fields", func() {
				testStruct := &TestBasicStruct{
					Name:        "  John Doe  ",
					Email:       " john@example.com ",
					Description: "   Some description   ",
				}

				baseController.trimStringFields(testStruct)

				Expect(testStruct.Name).To(Equal("John Doe"))
				Expect(testStruct.Email).To(Equal("john@example.com"))
				Expect(testStruct.Description).To(Equal("Some description"))
			})

			It("should trim whitespace from nested struct fields", func() {
				testStruct := &TestNestedStruct{
					User: TestBasicStruct{
						Name:        "  Jane Doe  ",
						Email:       " jane@example.com ",
						Description: "   User description   ",
					},
					Company: "  ACME Corp  ",
				}

				baseController.trimStringFields(testStruct)

				Expect(testStruct.User.Name).To(Equal("Jane Doe"))
				Expect(testStruct.User.Email).To(Equal("jane@example.com"))
				Expect(testStruct.User.Description).To(Equal("User description"))
				Expect(testStruct.Company).To(Equal("ACME Corp"))
			})

			It("should trim whitespace from pointer string fields", func() {
				name := "  John Pointer  "
				description := "   Pointer description   "
				testStruct := &TestPointerStruct{
					Name:        &name,
					Description: &description,
				}

				baseController.trimStringFields(testStruct)

				Expect(*testStruct.Name).To(Equal("John Pointer"))
				Expect(*testStruct.Description).To(Equal("Pointer description"))
			})

			It("should handle nil pointers without panicking", func() {
				testStruct := &TestPointerStruct{
					Name:        nil,
					Description: nil,
				}

				Expect(func() {
					baseController.trimStringFields(testStruct)
				}).ToNot(Panic())

				Expect(testStruct.Name).To(BeNil())
				Expect(testStruct.Description).To(BeNil())
			})

			It("should trim whitespace from slice of structs", func() {
				testStruct := &TestSliceStruct{
					Users: []TestBasicStruct{
						{
							Name:        "  User One  ",
							Email:       " user1@example.com ",
							Description: "   First user   ",
						},
						{
							Name:        "  User Two  ",
							Email:       " user2@example.com ",
							Description: "   Second user   ",
						},
					},
					Tags: []string{"  tag1  ", " tag2 ", "   tag3   "},
				}

				baseController.trimStringFields(testStruct)

				Expect(testStruct.Users[0].Name).To(Equal("User One"))
				Expect(testStruct.Users[0].Email).To(Equal("user1@example.com"))
				Expect(testStruct.Users[0].Description).To(Equal("First user"))

				Expect(testStruct.Users[1].Name).To(Equal("User Two"))
				Expect(testStruct.Users[1].Email).To(Equal("user2@example.com"))
				Expect(testStruct.Users[1].Description).To(Equal("Second user"))

				// Note: Direct string slices are not trimmed in current implementation
				Expect(testStruct.Tags[0]).To(Equal("  tag1  "))
			})

			It("should handle empty and whitespace-only strings", func() {
				testStruct := &TestBasicStruct{
					Name:        "   ",
					Email:       "",
					Description: "  \t\n  ",
				}

				baseController.trimStringFields(testStruct)

				Expect(testStruct.Name).To(Equal(""))
				Expect(testStruct.Email).To(Equal(""))
				Expect(testStruct.Description).To(Equal(""))
			})

			It("should handle edge case whitespace characters", func() {
				testStruct := &TestBasicStruct{
					Name:        "\t  John Doe  \n",
					Email:       " \r john@example.com \t ",
					Description: "\n   Some description \r\n   ",
				}

				baseController.trimStringFields(testStruct)

				Expect(testStruct.Name).To(Equal("John Doe"))
				Expect(testStruct.Email).To(Equal("john@example.com"))
				Expect(testStruct.Description).To(Equal("Some description"))
			})
		})

		Context("BindJSON with whitespace handling", func() {
			var (
				req *http.Request
				w   *httptest.ResponseRecorder
				ctx *gin.Context
			)

			BeforeEach(func() {
				w = httptest.NewRecorder()
				ctx, _ = gin.CreateTestContext(w)
			})

			It("should successfully bind and trim valid input", func() {
				requestBody := map[string]interface{}{
					"name":        "  John Doe  ",
					"email":       "test@example.com",
					"description": "   Some description   ",
				}
				jsonBody, _ := json.Marshal(requestBody)

				req, _ = http.NewRequest("POST", "/test", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				ctx.Request = req

				var testStruct TestBasicStruct
				err := baseController.BindJSON(ctx, &testStruct)

				Expect(err).ToNot(HaveOccurred())
				Expect(testStruct.Name).To(Equal("John Doe"))
				Expect(testStruct.Email).To(Equal("test@example.com"))
				Expect(testStruct.Description).To(Equal("Some description"))
			})

			It("should fail validation when required field becomes empty after trimming", func() {
				requestBody := map[string]interface{}{
					"name":        "   ", // This will become empty after trimming
					"email":       " john@example.com ",
					"description": "   Some description   ",
				}
				jsonBody, _ := json.Marshal(requestBody)

				req, _ = http.NewRequest("POST", "/test", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				ctx.Request = req

				var testStruct TestBasicStruct
				err := baseController.BindJSON(ctx, &testStruct)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Error:Field validation"))
				Expect(w.Code).To(Equal(http.StatusBadRequest))
			})

			It("should fail validation when email becomes invalid after trimming", func() {
				requestBody := map[string]interface{}{
					"name":        "John Doe",
					"email":       " invalid-email ", // Still invalid after trimming
					"description": "Some description",
				}
				jsonBody, _ := json.Marshal(requestBody)

				req, _ = http.NewRequest("POST", "/test", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				ctx.Request = req

				var testStruct TestBasicStruct
				err := baseController.BindJSON(ctx, &testStruct)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("email"))
				Expect(w.Code).To(Equal(http.StatusBadRequest))
			})

			It("should fail validation when email field becomes empty after trimming", func() {
				requestBody := map[string]interface{}{
					"name":        "John Doe",
					"email":       "   ", // This will become empty after trimming, triggering email validation
					"description": "Some description",
				}
				jsonBody, _ := json.Marshal(requestBody)

				req, _ = http.NewRequest("POST", "/test", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				ctx.Request = req

				var testStruct TestBasicStruct
				err := baseController.BindJSON(ctx, &testStruct)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("email"))
				Expect(w.Code).To(Equal(http.StatusBadRequest))
			})

			It("should handle nested struct validation with whitespace", func() {
				requestBody := map[string]interface{}{
					"user": map[string]interface{}{
						"name":        "   ", // Will become empty after trimming
						"email":       " user@example.com ",
						"description": "User description",
					},
					"company": "  ACME Corp  ",
				}
				jsonBody, _ := json.Marshal(requestBody)

				req, _ = http.NewRequest("POST", "/test", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				ctx.Request = req

				var testStruct TestNestedStruct
				err := baseController.BindJSON(ctx, &testStruct)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Error:Field validation"))
				Expect(w.Code).To(Equal(http.StatusBadRequest))
			})

			It("should handle nested struct email validation with whitespace", func() {
				requestBody := map[string]interface{}{
					"user": map[string]interface{}{
						"name":        "John Doe",
						"email":       "   ", // Will become empty after trimming, triggering email validation
						"description": "User description",
					},
					"company": "  ACME Corp  ",
				}
				jsonBody, _ := json.Marshal(requestBody)

				req, _ = http.NewRequest("POST", "/test", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				ctx.Request = req

				var testStruct TestNestedStruct
				err := baseController.BindJSON(ctx, &testStruct)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("email"))
				Expect(w.Code).To(Equal(http.StatusBadRequest))
			})

			It("should handle invalid JSON gracefully", func() {
				req, _ = http.NewRequest("POST", "/test", bytes.NewBuffer([]byte("invalid json")))
				req.Header.Set("Content-Type", "application/json")
				ctx.Request = req

				var testStruct TestBasicStruct
				err := baseController.BindJSON(ctx, &testStruct)

				Expect(err).To(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusBadRequest))
			})

			It("should trim all string fields in complex nested structure", func() {
				requestBody := map[string]interface{}{
					"user": map[string]interface{}{
						"name":        "  Jane Doe  ",
						"email":       "jane@example.com",
						"description": "   User description   ",
					},
					"company": "  ACME Corp  ",
				}
				jsonBody, _ := json.Marshal(requestBody)

				req, _ = http.NewRequest("POST", "/test", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				ctx.Request = req

				var testStruct TestNestedStruct
				err := baseController.BindJSON(ctx, &testStruct)

				Expect(err).ToNot(HaveOccurred())
				Expect(testStruct.User.Name).To(Equal("Jane Doe"))
				Expect(testStruct.User.Email).To(Equal("jane@example.com"))
				Expect(testStruct.User.Description).To(Equal("User description"))
				Expect(testStruct.Company).To(Equal("ACME Corp"))
			})
		})
	})
})
