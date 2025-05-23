package controller

import (
	"errors"
	customErrors "expenses/internal/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleError(ctx *gin.Context, stack bool, err error) {
	if err == nil {
		return
	}

	var authErr *customErrors.AuthError
	if errors.As(err, &authErr) {
		response := gin.H{
			"message": authErr.Message,
		}
		if stack {
			response["error"] = authErr.Err.Error()
			response["stack"] = authErr.Stack
		}
		ctx.JSON(authErr.Status, response)
		return
	}

	response := gin.H{
		"message": "Something went wrong",
	}
	if stack {
		response["error"] = err.Error()
	}
	ctx.JSON(http.StatusInternalServerError, response)
}
