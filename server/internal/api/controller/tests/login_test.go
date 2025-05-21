package tests

import (
	"encoding/json"
	"expenses/internal/models"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	env := SetupTestEnv(t)
	t.Cleanup(func() {
		TeardownTestEnv(t, env)
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/signup", env.AuthCtrl.Signup)
	router.POST("/login", env.AuthCtrl.Login)

	// First create a user to test login with
	w, err := performPostRequest(router, "/signup", env.TestUser)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, w.Code)

	t.Run("Successful Login", func(t *testing.T) {
		loginInput := models.LoginInput{
			Email:    env.TestUser.Email,
			Password: env.TestUser.Password,
		}

		w, err := performPostRequest(router, "/login", loginInput)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Contains(t, response, "message")
		assert.Contains(t, response, "user")
		assert.Contains(t, response, "access_token")
		assert.Contains(t, response, "refresh_token")

		user := response["user"].(map[string]interface{})
		assert.Equal(t, env.TestUser.Email, user["email"])
		assert.Equal(t, env.TestUser.Name, user["name"])
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		invalidLogin := models.LoginInput{
			Email:    env.TestUser.Email,
			Password: "wrongpassword",
		}

		w, err := performPostRequest(router, "/login", invalidLogin)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Input", func(t *testing.T) {
		invalidLogin := models.LoginInput{
			Email:    "invalid-email",
			Password: "",
		}

		w, err := performPostRequest(router, "/login", invalidLogin)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Non-existent User", func(t *testing.T) {
		nonExistentLogin := models.LoginInput{
			Email:    "nonexistent@example.com",
			Password: "somepassword",
		}

		w, err := performPostRequest(router, "/login", nonExistentLogin)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
