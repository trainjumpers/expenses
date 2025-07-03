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
	categoryService service.CategoryServiceInterface,
	transactionService service.TransactionServiceInterface,
	ruleService service.RuleServiceInterface,
) *gin.Engine {
	router := gin.New()
	if cfg.Environment != "test" {
		router.Use(gin.Logger()) // Disable logger when running tests
	}
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
	categoryController := controller.NewCategoryController(cfg, categoryService)
	transactionController := controller.NewTransactionController(cfg, transactionService)
	ruleController := controller.NewRuleController(cfg, ruleService)
	api := router.Group("/api/v1")
	{
		base := api.Group("")
		// Auth related routes
		base.POST("/signup", authController.Signup)
		base.POST("/login", authController.Login)
		base.POST("/refresh", authController.RefreshToken)

		// User related routes
		user := base.Group("/user", middleware.ProtectedWithCreatedBy(cfg)...)
		{
			user.GET("", userController.GetUserById)
			user.DELETE("", userController.DeleteUser)
			user.PATCH("", userController.UpdateUser)
			user.POST("/password", userController.UpdateUserPassword)
		}

		// Account routes
		account := base.Group("/account", middleware.ProtectedWithCreatedBy(cfg)...)
		{
			account.GET("", accountController.ListAccounts)
			account.POST("", accountController.CreateAccount)
			account.GET("/:accountId", accountController.GetAccount)
			account.PATCH("/:accountId", accountController.UpdateAccount)
			account.DELETE("/:accountId", accountController.DeleteAccount)
		}

		// Category routes
		category := base.Group("/category", middleware.ProtectedWithCreatedBy(cfg)...)
		{
			category.GET("", categoryController.ListCategories)
			category.POST("", categoryController.CreateCategory)
			category.GET("/:categoryId", categoryController.GetCategory)
			category.PATCH("/:categoryId", categoryController.UpdateCategory)
			category.DELETE("/:categoryId", categoryController.DeleteCategory)
		}

		// Transaction routes
		transaction := base.Group("/transaction", middleware.ProtectedWithCreatedBy(cfg)...)
		{
			transaction.GET("", transactionController.ListTransactions)
			transaction.POST("", transactionController.CreateTransaction)
			transaction.GET("/:transactionId", transactionController.GetTransaction)
			transaction.PATCH("/:transactionId", transactionController.UpdateTransaction)
			transaction.DELETE("/:transactionId", transactionController.DeleteTransaction)
		}

		// Rule routes
		rule := base.Group("/rule").Use(gin.HandlerFunc(middleware.Protected(cfg)))
		{
			rule.GET("", ruleController.ListRules)
			rule.POST("", ruleController.CreateRule)
			rule.GET("/:ruleId", ruleController.GetRuleById)
			rule.PATCH("/:ruleId", ruleController.UpdateRule)
			rule.DELETE("/:ruleId", ruleController.DeleteRule)
			rule.PATCH("/:ruleId/action/:id", ruleController.UpdateRuleAction)
			rule.PATCH("/:ruleId/condition/:id", ruleController.UpdateRuleCondition)
		}
	}

	return router
}
