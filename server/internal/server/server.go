package server

import (
	"context"
	"errors"
	"expenses/internal/wire"
	"expenses/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var httpServer *http.Server

func Start() {
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
	defer func(provider *wire.Provider) {
		err := provider.Close()
		if err != nil {
			logger.Error("Failed to close the provider", err)
		}
	}(provider)

	httpServer = &http.Server{
		Addr:              ":" + strconv.Itoa(port),
		Handler:           provider.Handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Graceful shutdown handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		logger.Debug("received interrupt signal")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			logger.Fatal("Server Close:", err)
		}
	}()

	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal("Server closed unexpectedly:", err)
	}
	logger.Info("Server exited successfully")
}

// StartAsync allows testing to start the server in background
func StartAsync() {
	go Start()
}
