package controllers

import (
	"errors"
	"expenses/internal/config"
	customErrors "expenses/internal/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	var authErr *customErrors.AuthError
	if errors.As(err, &authErr) {
		response := gin.H{
			"message": authErr.Message,
		}
		if config.IsDev() {
			response["error"] = authErr.Err.Error()
			response["stack"] = authErr.Stack
		}
		c.JSON(authErr.Status, response)
		return
	}

	response := gin.H{
		"message": "Something went wrong",
	}
	if config.IsDev() {
		response["error"] = err.Error()
	}
	c.JSON(http.StatusInternalServerError, response)
	return
}
