package controllers

import (
	"errors"
	customErrors "expenses/internal/errors"
	"expenses/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	var authErr *customErrors.AuthError
	if errors.As(err, &authErr) {
		logger.Error("Error in controller: ", err)
		logger.Info("Stack trace: ", authErr.Status)
		c.JSON(authErr.Status, gin.H{"message": authErr.Message, "error": authErr.Err.Error()})
		return
	}
	logger.Error("Error in controller: ", err)
	c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong", "error": err.Error()})
	return
}
