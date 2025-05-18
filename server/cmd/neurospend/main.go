package main

import (
	"context"
	"expenses/internal/api"
	database "expenses/internal/database/postgres"
	"expenses/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	database.ConnectDatabase()
	defer database.CloseDatabase()

	router := api.Init()
	port, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
	if err != nil {
		logger.Warn("Invalid or missing server port number, defaulting to 8080")
		port = 8080
	}
	logger.Infof("Starting server on port %d", port)
	server := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		logger.Debug("receive interrupt signal")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
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
