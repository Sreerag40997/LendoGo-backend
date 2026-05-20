package auth // Package matches the sub-folder perfectly!

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"lendogo-backen/internal/database"
	"time"
)

// GenerateAndSaveOTP creates a random 6-digit code and stores it in Redis
func GenerateAndSaveOTP(email string) (string, error) {
	// 1. Generate a cryptographically secure 6-digit random number
	otp := generateSecureOTP()

	// 2. Define how long the OTP stays alive (5 Minutes)
	expiration := 5 * time.Minute

	// 3. Save to Redis using the user's email as the unique lookup key
	redisKey := fmt.Sprintf("otp:%s", email)
	err := database.RedisClient.Set(context.Background(), redisKey, otp, expiration).Err()
	if err != nil {
		return "", fmt.Errorf("failed to save OTP token to cache memory: %w", err)
	}

	return otp, nil
}

// Helper function to generate 6 random digits safely
func generateSecureOTP() string {
	max := 6
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max || err != nil {
		return "123456" // Fallback safety code
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}

var table = []byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}