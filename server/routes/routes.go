package routes

import (
	"expenses/controllers"
	database "expenses/db"

	"github.com/gin-gonic/gin"
)

func Init() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	expenseController := controllers.NewExpenseController(database.DbPool)
	userController := controllers.NewUserController(database.DbPool)
	authController := controllers.NewAuthController(database.DbPool)

	api := router.Group("/api/v1")
	{
		auth := api.Group("")
		auth.POST("/signup", authController.Signup)
		auth.POST("/login", authController.Login)

		user := api.Group("/user").Use(gin.HandlerFunc(authController.Protected))
		user.GET("", userController.GetUserById)
		user.PATCH("", userController.UpdateUser)
		user.DELETE("", userController.DeleteUser)

		expense := api.Group("/expense").Use(gin.HandlerFunc(authController.Protected))
		expense.GET("", expenseController.GetExpensesOfUser)
		expense.POST("", expenseController.CreateExpense)
		expense.GET("/:expenseID", expenseController.GetExpenseByID)
		expense.PATCH("/:expenseID", expenseController.UpdateExpenseBasic)
		expense.DELETE("/:expenseID", expenseController.DeleteExpense)
	}

	return router
}
