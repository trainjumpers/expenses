package controller

import (
	"errors"
	customErrors "expenses/internal/errors"
	"net/http"
	"strings"

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
			if authErr.Stack != "" {
				stackLines := strings.Split(authErr.Stack, "\n")
				nonEmptyLines := make([]string, 0, len(stackLines))
				for _, line := range stackLines {
					line = strings.ReplaceAll(line, "\t", "")
					if line != "" {
						nonEmptyLines = append(nonEmptyLines, line)
					}
				}
				response["stack"] = nonEmptyLines
			}
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
