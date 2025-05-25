package controller

import (
	"encoding/json"
	"errors"
	"expenses/internal/config"
	customErrors "expenses/internal/errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

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
				var response map[string]interface{}
				err := json.NewDecoder(recorder.Body).Decode(&response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(Equal("Authentication failed"))
			})
		})

		Context("with generic error", func() {
			It("should handle generic error correctly", func() {
				err := errors.New("something went wrong")
				baseController.HandleError(ctx, err)

				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				var response map[string]interface{}
				err = json.NewDecoder(recorder.Body).Decode(&response)
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
			var response map[string]interface{}
			err := json.NewDecoder(recorder.Body).Decode(&response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Success"))
			Expect(response["data"]).To(HaveKey("key"))
		})

		It("should send success response without data", func() {
			baseController.SendSuccess(ctx, http.StatusOK, "Success", nil)

			Expect(recorder.Code).To(Equal(http.StatusOK))
			var response map[string]interface{}
			err := json.NewDecoder(recorder.Body).Decode(&response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Success"))
			Expect(response).NotTo(HaveKey("data"))
		})
	})

	Describe("SendError", func() {
		It("should send error response", func() {
			baseController.SendError(ctx, http.StatusBadRequest, "Invalid input")

			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			var response map[string]interface{}
			err := json.NewDecoder(recorder.Body).Decode(&response)
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
				ID string `uri:"id"`
			}
			ctx.Params = []gin.Param{{Key: "id", Value: "123"}}

			var testStruct TestStruct
			err := baseController.BindURI(ctx, &testStruct)
			Expect(err).NotTo(HaveOccurred())
			Expect(testStruct.ID).To(Equal("123"))
		})

		It("should return error for missing URI parameters", func() {
			type TestStruct struct {
				ID string `uri:"id" binding:"required"`
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
})
