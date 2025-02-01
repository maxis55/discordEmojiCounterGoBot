package db

import (
	"database/sql"
	"emoji-counter/utils"
	"fmt"
	"os"
)

var Connection *sql.DB

func Connect() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), utils.GetEnvStrWithFallback("DB_PORT", "5432"), os.Getenv("POSTGRES_USER"), os.Getenv("DB_PASSWORD"), "emoji_tracker")

	conn, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		panic(err)
	}

	Connection = conn

	return Connection, nil
}

func Close() error {
	return Connection.Close()
}
