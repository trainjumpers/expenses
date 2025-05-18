package database

import (
	"context"
	"expenses/pkg/logger"
	"fmt"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DbPool *pgxpool.Pool

func ConnectDatabase() {
	host := os.Getenv("DB_HOST")
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		logger.Fatal("Invalid database port number")
	}
	user := os.Getenv("DB_USER")
	dbname := os.Getenv("DB_NAME")
	pass := os.Getenv("DB_PASSWORD")

	psqlSetup := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=verify-full",
		user, pass, host, port, dbname)

	logger.Debug("Connecting to database")
	db, err := pgxpool.New(context.Background(), psqlSetup)

	if err != nil {
		logger.Fatal("There is an error while connecting to the database ", err)
	} else {
		DbPool = db
		logger.Debug("Database connected successfully")
	}
}

func CloseDatabase() {
	if DbPool != nil {
		DbPool.Close()
		logger.Debug("Database connection closed")
	}
}
