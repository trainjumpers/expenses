package database

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	logger "github.com/sirupsen/logrus"
)

var DbPool *pgxpool.Pool

func ConnectDatabase() {
	host := os.Getenv("PGHOST")
	port, _ := strconv.Atoi(os.Getenv("PGPORT"))
	user := os.Getenv("PGUSER")
	dbname := os.Getenv("PGDBNAME")
	pass := os.Getenv("PGPASSWORD")

	psqlSetup := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable pool_max_conns=10",
		host, port, user, dbname, pass)

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
