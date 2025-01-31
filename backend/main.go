package main

import (
	"discordEmojiCounterBot/bot"
	"discordEmojiCounterBot/db"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(fmt.Sprintf("Error loading .env file because '%s'", err.Error()))
	}

	dbcon, _ := db.Connect()
	defer dbcon.Close()
	bot.Run()

	log.Println("Shutting down")
}
