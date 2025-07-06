package controller

import (
	"errors"
	"expenses/internal/config"
	customErrors "expenses/internal/errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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
		if b.cfg.IsDev() || b.cfg.IsTest() {
			response["error"] = authErr.Err.Error()
			response["stack"] = authErr.Stack
		}
		ctx.JSON(authErr.Status, response)
		return
	}

	response := gin.H{
		"message": "Something went wrong",
	}
	unhandledErr := customErrors.New(err.Error())
	if b.cfg.IsDev() || b.cfg.IsTest() {
		response["error"] = unhandledErr.Err.Error()
		response["stack"] = unhandledErr.Stack
	}
	ctx.JSON(http.StatusInternalServerError, response)
}

// SendSuccess sends a successful response with optional data. Note: when statusCode is 204, the message and data parameters are ignored.
func (b *BaseController) SendSuccess(ctx *gin.Context, statusCode int, message string, data interface{}) {
	if statusCode == http.StatusNoContent {
		ctx.Status(http.StatusNoContent)
		return
	}
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
// It also automatically trims whitespace from all string fields and re-validates using Gin's validator
func (b *BaseController) BindJSON(ctx *gin.Context, obj any) error {
	if err := ctx.ShouldBindJSON(obj); err != nil {
		b.SendError(ctx, http.StatusBadRequest, err.Error())
		return err
	}
	b.trimStringFields(obj)
	if err := binding.Validator.ValidateStruct(obj); err != nil {
		b.SendError(ctx, http.StatusBadRequest, err.Error())
		return err
	}

	return nil
}

// trimStringFields recursively trims whitespace from all string fields in a struct
func (b *BaseController) trimStringFields(obj interface{}) {
	if obj == nil {
		return
	}
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if !field.CanSet() {
			continue
		}
		switch field.Kind() {
		case reflect.String:
			trimmed := strings.TrimSpace(field.String())
			field.SetString(trimmed)
		case reflect.Ptr:
			if !field.IsNil() && field.Elem().Kind() == reflect.String {
				trimmed := strings.TrimSpace(field.Elem().String())
				field.Elem().SetString(trimmed)
			} else if !field.IsNil() && field.Elem().Kind() == reflect.Struct {
				b.trimStringFields(field.Interface())
			}
		case reflect.Struct:
			b.trimStringFields(field.Addr().Interface())
		case reflect.Slice:
			for j := 0; j < field.Len(); j++ {
				elem := field.Index(j)
				if elem.Kind() == reflect.Struct {
					b.trimStringFields(elem.Addr().Interface())
				} else if elem.Kind() == reflect.Ptr && !elem.IsNil() && elem.Elem().Kind() == reflect.Struct {
					b.trimStringFields(elem.Interface())
				}
			}
		}
	}
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

// GetAuthenticatedUserId extracts the authenticated user Id from the gin context
func (b *BaseController) GetAuthenticatedUserId(ctx *gin.Context) int64 {
	return ctx.GetInt64("authUserId")
}

// setAuthCookie sets a secure, HTTP-only cookie with SameSite=Lax for auth tokens
func (b *BaseController) setAuthCookie(ctx *gin.Context, name, value string, maxAge int) {
	domain := ""
	if b.cfg.IsProd() {
		domain = b.cfg.CookieDomain
	}
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   domain,
		MaxAge:   maxAge,
		Secure:   b.cfg.IsProd(),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	// Gin's SetCookie does not support SameSite, so use Header directly
	h := cookie.String()
	ctx.Writer.Header().Add("Set-Cookie", h)
}
