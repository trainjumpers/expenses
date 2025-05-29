package api

import (
	"expenses/internal/api/controller"
	"expenses/internal/api/middleware"
	"expenses/internal/config"
	"expenses/internal/service"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Init(
	cfg *config.Config,
	authService service.AuthServiceInterface,
	userService service.UserServiceInterface,
	accountService service.AccountServiceInterface,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://neurospend.vercel.app"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           6 * time.Hour,
	}))

	// Health check route
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"status":  "UP",
			"message": "API is running",
		})
	})
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "Welcome to the expense tracker server",
		})
	})

	authController := controller.NewAuthController(cfg, authService)
	userController := controller.NewUserController(cfg, userService, authService)
	accountController := controller.NewAccountController(cfg, accountService)
	api := router.Group("/api/v1")
	{
		base := api.Group("")
		// Auth related routes
		base.POST("/signup", authController.Signup)
		base.POST("/login", authController.Login)
		base.POST("/refresh", authController.RefreshToken)

		// User related routes
		user := base.Group("/user").Use(gin.HandlerFunc(middleware.Protected(cfg)))
		{
			user.GET("", userController.GetUserById)
			user.DELETE("", userController.DeleteUser)
			user.PATCH("", userController.UpdateUser)
			user.POST("/password", userController.UpdateUserPassword)

			account := base.Group("/account").Use(gin.HandlerFunc(middleware.Protected(cfg)))
			account.GET("", accountController.ListAccounts)
			account.POST("", accountController.CreateAccount)
			account.GET("/:accountId", accountController.GetAccount)
			account.PATCH("/:accountId", accountController.UpdateAccount)
			account.DELETE("/:accountId", accountController.DeleteAccount)
		}
	}

	return router
}
