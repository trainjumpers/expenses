package main

import (
	"context"
	"expenses/internal/wire"
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
	port, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
	if err != nil {
		logger.Warn("Invalid or missing server port number, defaulting to 8080")
		port = 8080
	}
	logger.Infof("Starting server on port %d", port)
	provider, err := wire.InitializeApplication()
	if err != nil {
		logger.Fatal("Failed to initialize application:", err)
	}
	defer provider.Close()
	server := &http.Server{
		Addr:              ":" + strconv.Itoa(port),
		Handler:           provider.Handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		logger.Debug("received interrupt signal")
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
	logger.Info("Server exited successfully")
}
