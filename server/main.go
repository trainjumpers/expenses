package main

import (
	database "expenses/db"
	"expenses/routes"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
	}

	database.ConnectDatabase()
	r := routes.Init()

	r.GET("/", func(c *gin.Context) {
		t := time.Now()
		c.String(200, "Hello, you've requested: %s at %s\n", c.Request.URL.Path, t.UTC().Format("2006-01-02 15:04:05.00 -0700 MST"))
	})

	r.Run(":8080")
}
