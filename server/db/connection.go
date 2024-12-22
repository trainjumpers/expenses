package database

import (
	"context"
	"fmt"
	"os"
	"strconv"

	logger "expenses/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DbPool *pgxpool.Pool

func ConnectDatabase() {
	host := os.Getenv("PGHOST")
	port, _ := strconv.Atoi(os.Getenv("PGPORT"))
	user := os.Getenv("PGUSER")
	dbname := os.Getenv("PGDBNAME")
	pass := os.Getenv("PGPASSWORD")

	psqlSetup := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=verify-full",
	user, pass, host, port, dbname)

	logger.Info("Connecting to database")
	db, err := pgxpool.New(context.Background(), psqlSetup)

	if err != nil {
		logger.Fatal("There is an error while connecting to the database ", err)
		panic(err)
	} else {
		DbPool = db
		logger.Info("Database connected successfully")
	}
}
