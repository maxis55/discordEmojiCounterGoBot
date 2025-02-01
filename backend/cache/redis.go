package cache

import (
	"context"
	"emoji-counter/utils"
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
)

var RedisCache *redis.Client

func Connect() {
	RedisCache = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), utils.GetEnvStrWithFallback("REDIS_PORT", "6379")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	ctx := context.Background()

	_, err := RedisCache.Ping(ctx).Result()

	if err != nil {
		panic(err)
	}
}

func Close() error {
	return RedisCache.Close()
}
