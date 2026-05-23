package db

import (
	"context"
	"database/sql"
	"emoji-counter/utils"
	"fmt"
	"os"
	"time"
)

var Connection *sql.DB

const defaultQueryTimeout = 5 * time.Second

func Connect() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), utils.GetEnvStrWithFallback("DB_PORT", "5432"), os.Getenv("POSTGRES_USER"), os.Getenv("DB_PASSWORD"), "emoji_tracker")

	conn, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		panic(err)
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)
	conn.SetConnMaxIdleTime(1 * time.Minute)

	Connection = conn

	return Connection, nil
}

func Close() error {
	return Connection.Close()
}

// Exec runs ExecContext against conn with a 5s timeout. Used for writes
// where a hung query would otherwise block a goroutine indefinitely.
func Exec(conn *sql.DB, query string, args ...any) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout)
	defer cancel()
	return conn.ExecContext(ctx, query, args...)
}
