package db

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func ConnectDB() {
	connStr := os.Getenv("DATABASE_URL")
	var err error

	Pool, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	err = Pool.Ping(context.Background())
	if err != nil {
		log.Fatalf("Database didn't respond to ping: %v\n", err)
	}

	log.Println("Database connection established!")
}
