package main

import (
	"expenses/internal/api"
	database "expenses/internal/database/postgres"
	"expenses/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	database.ConnectDatabase()
	defer database.CloseDatabase()

	router := api.Init()
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	logger.Infof("Starting server on port %s", port)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		logger.Debug("receive interrupt signal")
		if err := server.Close(); err != nil {
			logger.Fatal("Server Close:", err)
		}
	}()
	if err := server.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			logger.Debug("Server closed under request")
		} else {
			logger.Fatal("Server closed unexpectedly")
		}
	}
	logger.Debug("Server exited gracefully")
}
