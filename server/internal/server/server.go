package server

import (
	"context"
	"errors"
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

var httpServer *http.Server

func Start() {
	gin.SetMode(gin.ReleaseMode)
	port, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
	if err != nil {
		logger.Warnf("Invalid or missing server port number, defaulting to 8080")
		port = 8080
	}
	logger.Infof("Starting server on port %d", port)
	provider, err := wire.InitializeApplication()
	if err != nil {
		logger.Fatalf("Failed to initialize application: %v", err)
	}
	defer func(provider *wire.Provider) {
		err := provider.Close()
		if err != nil {
			logger.Errorf("Failed to close the provider: %v", err)
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
		logger.Debugf("Received interrupt signal")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			logger.Fatalf("Server close error: %v", err)
		}
	}()

	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatalf("Server closed unexpectedly: %v", err)
	}
	logger.Infof("Server exited successfully")
}

// StartAsync allows testing to start the server in background
func StartAsync() {
	go Start()
}
