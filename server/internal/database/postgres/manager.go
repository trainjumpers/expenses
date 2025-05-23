package database

import (
	"context"
	"expenses/pkg/logger"
	"fmt"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseManager struct {
	pool *pgxpool.Pool
}

func NewDatabaseManager() (*DatabaseManager, error) {
	host := os.Getenv("DB_HOST")
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil, fmt.Errorf("invalid database port number: %w", err)
	}
	user := os.Getenv("DB_USER")
	dbname := os.Getenv("DB_NAME")
	pass := os.Getenv("DB_PASSWORD")

	psqlSetup := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=verify-full",
		user, pass, host, port, dbname)

	logger.Debug("Connecting to database")
	pool, err := pgxpool.New(context.Background(), psqlSetup)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	logger.Debug("Database connected successfully")

	return &DatabaseManager{
		pool: pool,
	}, nil
}

func (dm *DatabaseManager) GetPool() *pgxpool.Pool {
	if dm.pool == nil {
		logger.Fatal("Database connection is not initialized")
	}
	return dm.pool
}

func (dm *DatabaseManager) Close() error {
	dm.pool.Close()
	logger.Debug("Database connection closed")
	return nil
}
