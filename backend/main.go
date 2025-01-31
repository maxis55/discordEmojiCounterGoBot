package main

import (
	"database/sql"
	"discordEmojiCounterBot/bot"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"os"
)

func connect() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("POSTGRES_USER"), os.Getenv("DB_PASSWORD"), "emoji_tracker")

	fmt.Println(psqlInfo)

	return sql.Open("postgres", psqlInfo)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(fmt.Sprintf("Error loading .env file because '%s'", err.Error()))
	}

	bot.Token = os.Getenv("DISCORD_KEY")
	db, _ := connect()
	defer db.Close()
	bot.Run(db)

	log.Println("Shutting down")
}
