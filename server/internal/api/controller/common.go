package controller

import (
	"errors"
	customErrors "expenses/internal/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleError(ctx *gin.Context, isDev bool, err error) {
	if err == nil {
		return
	}

	var authErr *customErrors.AuthError
	if errors.As(err, &authErr) {
		response := gin.H{
			"message": authErr.Message,
		}
		if isDev {
			response["error"] = authErr.Err.Error()
			response["stack"] = authErr.Stack
		}
		ctx.JSON(authErr.Status, response)
		return
	}

	response := gin.H{
		"message": "Something went wrong",
	}
	if isDev {
		response["error"] = err.Error()
	}
	ctx.JSON(http.StatusInternalServerError, response)
}
