package main

import (
	"emoji-counter/bot"
	"emoji-counter/cache"
	"emoji-counter/db"
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

	db.Connect()
	defer db.Close()

	cache.Connect()
	defer cache.Close()

	bot.Run()

	log.Println("Shutting down")
}
