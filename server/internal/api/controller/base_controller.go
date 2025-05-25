package controller

import (
	"errors"
	"expenses/internal/config"
	customErrors "expenses/internal/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// BaseController provides common functionality for all controllers
type BaseController struct {
	cfg *config.Config
}

// NewBaseController creates a new base controller instance
func NewBaseController(cfg *config.Config) *BaseController {
	return &BaseController{
		cfg: cfg,
	}
}

// GetConfig returns the configuration instance
func (b *BaseController) GetConfig() *config.Config {
	return b.cfg
}

// HandleError handles errors in a consistent way across all controllers
func (b *BaseController) HandleError(ctx *gin.Context, err error) {
	if err == nil {
		return
	}

	var authErr *customErrors.AuthError
	if errors.As(err, &authErr) {
		response := gin.H{
			"message": authErr.Message,
		}
		if b.cfg.IsDev() {
			response["error"] = authErr.Err.Error()
			response["stack"] = authErr.Stack
		}
		ctx.JSON(authErr.Status, response)
		return
	}

	response := gin.H{
		"message": "Something went wrong",
	}
	if b.cfg.IsDev() {
		response["error"] = err.Error()
	}
	ctx.JSON(http.StatusInternalServerError, response)
}

// SendSuccess sends a successful response with optional data
func (b *BaseController) SendSuccess(ctx *gin.Context, statusCode int, message string, data interface{}) {
	response := gin.H{
		"message": message,
	}
	if data != nil {
		response["data"] = data
	}
	ctx.JSON(statusCode, response)
}

// SendError sends an error response
func (b *BaseController) SendError(ctx *gin.Context, statusCode int, message string) {
	ctx.JSON(statusCode, gin.H{
		"message": message,
	})
}

// BindJSON binds JSON request body to the provided struct and handles errors
func (b *BaseController) BindJSON(ctx *gin.Context, obj interface{}) error {
	if err := ctx.ShouldBindJSON(obj); err != nil {
		b.SendError(ctx, http.StatusBadRequest, err.Error())
		return err
	}
	return nil
}

// BindQuery binds query parameters to the provided struct and handles errors
func (b *BaseController) BindQuery(ctx *gin.Context, obj interface{}) error {
	if err := ctx.ShouldBindQuery(obj); err != nil {
		b.SendError(ctx, http.StatusBadRequest, err.Error())
		return err
	}
	return nil
}

// BindURI binds URI parameters to the provided struct and handles errors
func (b *BaseController) BindURI(ctx *gin.Context, obj interface{}) error {
	if err := ctx.ShouldBindUri(obj); err != nil {
		b.SendError(ctx, http.StatusBadRequest, err.Error())
		return err
	}
	return nil
}

// BindForm binds form data to the provided struct and handles errors
func (b *BaseController) BindForm(ctx *gin.Context, obj interface{}) error {
	if err := ctx.ShouldBind(obj); err != nil {
		b.SendError(ctx, http.StatusBadRequest, err.Error())
		return err
	}
	return nil
}
