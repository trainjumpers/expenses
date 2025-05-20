package controllers

import (
	"errors"
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
		c.JSON(authErr.Status, gin.H{"message": authErr.Message, "error": authErr.Err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong", "error": err.Error()})
	return
}
