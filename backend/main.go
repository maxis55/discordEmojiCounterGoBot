package main

import (
	"context"
	"emoji-counter/bot"
	"emoji-counter/cache"
	"emoji-counter/db"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	db.Connect()
	defer db.Close()

	cache.Connect()
	defer cache.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	bot.Run(ctx)

	slog.Info("shutting down")
}
