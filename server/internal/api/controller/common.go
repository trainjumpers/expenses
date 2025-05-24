package controller

import (
	"errors"
	customErrors "expenses/internal/errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// cleanStackTrace processes a stack trace string by splitting it into lines,
// trimming whitespace, and removing empty lines
func cleanStackTrace(stack string) []string {
	if stack == "" {
		return nil
	}

	stackLines := strings.Split(stack, "\n")
	nonEmptyLines := make([]string, 0, len(stackLines))
	for _, line := range stackLines {
		line = strings.TrimSpace(line)
		if line != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}
	return nonEmptyLines
}

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
			response["stack"] = cleanStackTrace(authErr.Stack)
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
