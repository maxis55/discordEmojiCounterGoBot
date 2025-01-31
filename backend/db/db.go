package db

import (
	"database/sql"
	"fmt"
	"os"
)

var Connection *sql.DB

func Connect() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("POSTGRES_USER"), os.Getenv("DB_PASSWORD"), "emoji_tracker")

	conn, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		return nil, err
	}

	Connection = conn

	return Connection, nil
}
