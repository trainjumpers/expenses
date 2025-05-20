package api

import (
	controllers "expenses/internal/api/controller"
	database "expenses/internal/database/postgres"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Init() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           6 * time.Hour,
	}))

	// Health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "UP",
			"message": "API is running",
		})
	})
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the expense tracker server",
		})
	})

	authController := controllers.NewAuthController(database.DbPool)
	api := router.Group("/api/v1")
	{
		base := api.Group("")
		base.POST("/signup", authController.Signup)
		base.POST("/login", authController.Login)
		base.POST("/refresh", authController.RefreshToken)
	}

	return router
}
