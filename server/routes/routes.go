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

	api := router.Group("/api")
	{
		user := api.Group("/user")
		user.GET("", userController.GetUsers)
		user.POST("/signup", authController.Signup)
		user.POST("/login", authController.Login)
		user.GET("/:userID", userController.GetUserById)
		// user.PUT("/:userID", userController.UpdateUserData)
		user.DELETE("/:userID", userController.DeleteUser)

		expense := api.Group("/expense").Use(gin.HandlerFunc(authController.Protected))
		expense.GET("", expenseController.GetExpensesOfUser)
		expense.POST("", expenseController.CreateExpense)
		// expense.GET("/:expenseID", expenseController.GetExpenseByID)
		// expense.PUT("/:expenseID", expenseController.UpdateExpense)
		expense.DELETE("/:expenseID", expenseController.DeleteExpense)
	}

	return router
}
