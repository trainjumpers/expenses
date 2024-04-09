package routes

import (
	"expenses/controllers"

	"github.com/gin-gonic/gin"
)

func Init() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	expenseController := new(controllers.ExpenseController)
	userController := new(controllers.UserController)
	authController := new(controllers.AuthController)

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
		expense.GET("", expenseController.GetExpenses)
		expense.POST("", expenseController.CreateExpense)
		expense.GET("/:expenseID", expenseController.GetExpenseByID)
		// expense.PUT("/:expenseID", expenseController.UpdateExpense)
		expense.DELETE("/:expenseID", expenseController.DeleteExpense)
	}

	return router
}
