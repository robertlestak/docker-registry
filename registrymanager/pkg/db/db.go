package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// Connect connects to the database and returns the current DB
func Connect() (*sql.DB, error) {
	var e error
	connURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)
	db, err := sql.Open("postgres", connURL)
	DB = db
	if err != nil {
		return DB, err
	}
	return DB, e
}
