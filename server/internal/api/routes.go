package api

import (
	"expenses/internal/api/controller"
	"expenses/internal/api/middleware"
	"expenses/internal/config"
	"expenses/internal/service"
	"expenses/internal/validator"
	"expenses/pkg/logger"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const multipartOverheadBytes = 1 << 20

func Init(
	cfg *config.Config,
	authService service.AuthServiceInterface,
	userService service.UserServiceInterface,
	accountService service.AccountServiceInterface,
	categoryService service.CategoryServiceInterface,
	transactionService service.TransactionServiceInterface,
	ruleService service.RuleServiceInterface,
	ruleEngineService service.RuleEngineServiceInterface,
	statementService service.StatementServiceInterface,
	analyticsService service.AnalyticsServiceInterface,
) *gin.Engine {
	router := gin.New()
	if !cfg.IsTest() || cfg.LoggingLevel != "" {
		router.Use(gin.Logger()) // Disable logger when running tests and logging level is not set
	}
	router.Use(gin.Recovery())
	if err := router.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		logger.Warnf("invalid TRUSTED_PROXIES, disabling proxy header trust: %v", err)
		_ = router.SetTrustedProxies(nil)
	}
	router.MaxMultipartMemory = 1 << 20 // spill uploads to disk past 1MB; the body cap bounds total size
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
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
	ruleController := controller.NewRuleController(cfg, ruleService, ruleEngineService)
	statementController := controller.NewStatementController(cfg, statementService)
	analyticsController := controller.NewAnalyticsController(cfg, analyticsService)

	api := router.Group("/api/v1")
	{
		base := api.Group("")

		// Auth endpoints are the most exposed surface. They are rate limited per
		// email (login, the brute-force target) and per IP. The in-process e2e
		// suite hammers these routes from a single address, so throttling is off
		// in the test environment; the middleware itself is unit tested.
		authRateLimiters := map[string][]gin.HandlerFunc{}
		if !cfg.IsTest() {
			authRateLimiters["login"] = []gin.HandlerFunc{
				middleware.RateLimit(middleware.NewKeyedLimiter(60, 20), middleware.ClientIPKey),
				middleware.RateLimit(middleware.NewKeyedLimiter(10, 5), middleware.LoginEmailKey),
			}
			authRateLimiters["signup"] = []gin.HandlerFunc{
				middleware.RateLimit(middleware.NewKeyedLimiter(20, 10), middleware.ClientIPKey),
			}
			authRateLimiters["refresh"] = []gin.HandlerFunc{
				middleware.RateLimit(middleware.NewKeyedLimiter(30, 15), middleware.ClientIPKey),
			}
		}

		// Auth related routes
		base.POST("/signup", append(authRateLimiters["signup"], authController.Signup)...)
		base.POST("/login", append(authRateLimiters["login"], authController.Login)...)
		base.POST("/refresh", append(authRateLimiters["refresh"], authController.RefreshToken)...)
		base.POST("/logout", authController.Logout)

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

		// Statement routes
		statement := base.Group("/statement", middleware.Protected(cfg))
		{
			statement.POST("", middleware.MaxBodySize(validator.MaxStatementFileBytes+multipartOverheadBytes), statementController.CreateStatement)
			statement.POST("/preview", middleware.MaxBodySize(validator.MaxStatementFileBytes+multipartOverheadBytes), statementController.PreviewStatement)
			statement.GET("", statementController.GetStatements)
			statement.GET("/:id", statementController.GetStatementStatus)
		}

		// Rule routes
		rule := base.Group("/rule", middleware.ProtectedWithCreatedBy(cfg)...)
		{
			rule.GET("", ruleController.ListRules)
			rule.POST("", ruleController.CreateRule)
			rule.POST("/execute", ruleController.ExecuteRules)
			rule.GET("/:ruleId", ruleController.GetRuleById)
			rule.PATCH("/:ruleId", ruleController.UpdateRule)
			rule.DELETE("/:ruleId", ruleController.DeleteRule)
			rule.PATCH("/:ruleId/action/:id", ruleController.UpdateRuleAction)
			rule.PATCH("/:ruleId/condition/:id", ruleController.UpdateRuleCondition)
			rule.PUT("/:ruleId/actions", ruleController.PutRuleActions)
			rule.PUT("/:ruleId/conditions", ruleController.PutRuleConditions)
		}

		// Analytics routes
		analytics := base.Group("/analytics", middleware.ProtectedWithCreatedBy(cfg)...)
		{
			analytics.GET("/account", analyticsController.GetAccountAnalytics)
			analytics.GET("/cash-balance", analyticsController.GetCashBalanceHistory)
			analytics.GET("/category", analyticsController.GetCategoryAnalytics)
			analytics.GET("/monthly", analyticsController.GetMonthlyAnalytics)
			analytics.GET("/insights", analyticsController.GetInsights)
		}
	}

	return router
}
