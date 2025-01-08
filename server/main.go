package main

import (
	database "expenses/db"
	"expenses/routes"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	database.ConnectDatabase()
	r := routes.Init()

	r.GET("/", func(c *gin.Context) {
		t := time.Now()
		c.String(200, "Hello, you've requested: %s at %s\n", c.Request.URL.Path, t.UTC().Format("2006-01-02 15:04:05.00 -0700 MST"))
	})

	port := os.Getenv("PORT")
	r.Run("localhost:" + port)
}
