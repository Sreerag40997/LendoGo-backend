package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var Ctx = context.Background()

// ConnectRedis initializes the Redis key-value memory store
func ConnectRedis() error {
	addr := fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
	if os.Getenv("REDIS_HOST") == "" || os.Getenv("REDIS_PORT") == "" {
		return fmt.Errorf("CRITICAL: Missing Redis environment variables in .env")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // Default alpine redis image has no password set
		DB:       0,  // Use default DB 0
	})

	// Ping Redis to verify connection works
	if err := rdb.Ping(Ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	RedisClient = rdb
	log.Println("✅ Redis cache engine connected successfully!")
	return nil
}