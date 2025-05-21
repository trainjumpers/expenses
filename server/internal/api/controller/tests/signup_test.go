// tests/signup_test.go
package tests

import (
	"bytes"
	"encoding/json"
	"expenses/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func performPostRequest(router *gin.Engine, url string, body interface{}) (*httptest.ResponseRecorder, error) {
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req := httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w, nil
}

func TestSignup(t *testing.T) {
	env := SetupTestEnv(t)
	t.Cleanup(func() {
		TeardownTestEnv(t, env)
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/signup", env.AuthCtrl.Signup)

	t.Run("Successful Signup", func(t *testing.T) {
		w, err := performPostRequest(router, "/signup", env.TestUser2)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Contains(t, response, "user")
		assert.Contains(t, response, "access_token")
		assert.Contains(t, response, "refresh_token")
		user := response["user"].(map[string]interface{})
		assert.Equal(t, env.TestUser2.Email, user["email"])
		assert.Equal(t, env.TestUser2.Name, user["name"])
	})

	t.Run("Duplicate Email Signup", func(t *testing.T) {
		w, err := performPostRequest(router, "/signup", env.TestUser2)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Invalid Input", func(t *testing.T) {
		invalidUser := models.CreateUserInput{
			Email:    "invalid-email",
			Password: "short",
			Name:     "",
		}

		w, err := performPostRequest(router, "/signup", invalidUser)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
