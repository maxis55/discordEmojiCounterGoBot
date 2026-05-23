package cache

import (
	"context"
	"emoji-counter/utils"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"os"
	"time"
)

var RedisCache *redis.Client
var Ctx = context.Background()

func Connect() {
	RedisCache = redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), utils.GetEnvStrWithFallback("REDIS_PORT", "6379")),
		Password:        os.Getenv("REDIS_PASSWORD"),
		DB:              0,
		MaxRetries:      3,
		MinRetryBackoff: 100 * time.Millisecond,
		MaxRetryBackoff: 2 * time.Second,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
	})

	_, err := RedisCache.Ping(Ctx).Result()

	if err != nil {
		panic(err)
	}
}

func Close() error {
	return RedisCache.Close()
}

func KeyExists(key string) bool {
	res, err := RedisCache.Get(Ctx, key).Result()

	if res != "" && err == nil {
		return true
	}

	return false
}

func RememberKey(key string) {
	err := RedisCache.Set(Ctx, key, "1", time.Hour).Err()

	if err != nil {
		slog.Warn("redis set failed", "key", key, "err", err)
	}
}
