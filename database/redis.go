package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"github.com/redis/go-redis/v9"
)
var RedisClient *redis.Client
func ConnectRedis() error {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")
	if host == "" || port == "" {
		return fmt.Errorf("CRITICAL: Missing Redis environment variables in .env")
	}
	address := fmt.Sprintf("%s:%s", host, port)
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password, 
		DB:       0,        
	})
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}
	RedisClient = client
	log.Println("REDIS CONNECTED: Ready to store OTPs!")
	return nil
}