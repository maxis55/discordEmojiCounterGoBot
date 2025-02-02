package main

import (
	"emoji-counter/bot"
	"emoji-counter/cache"
	"emoji-counter/db"
	_ "github.com/lib/pq"
	"log"
)

func main() {
	db.Connect()
	defer db.Close()

	cache.Connect()
	defer cache.Close()

	bot.Run()

	log.Println("Shutting down")
}
