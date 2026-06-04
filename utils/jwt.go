package utils

import (
    "os"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

// GenerateToken creates a secure JWT for a logged-in user
// Updated: userID is now a string, and we added the 'role' parameter!
func GenerateToken(userID string, role string, fullName string, email string) (string, error) {
    secret := os.Getenv("JWT_SECRET")
    claims := jwt.MapClaims{
        "user_id":   userID,
        "role":      role,
        "full_name": fullName, // ← ADD
        "email":     email,    // ← ADD
        "exp":       time.Now().Add(time.Hour * 72).Unix(),
        "iat":       time.Now().Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}