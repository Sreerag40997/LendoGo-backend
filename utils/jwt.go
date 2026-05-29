package utils

import (
    "fmt"
    "os"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

// GenerateToken creates a secure JWT for a logged-in user
// Updated: userID is now a string, and we added the 'role' parameter!
func GenerateToken(userID string, role string) (string, error) { // 👈 FIX 1: Added 'role string'
    // 1. Grab the secret key from your .env file
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        return "", fmt.Errorf("JWT_SECRET is missing from .env")
    }

    // 2. Define the payload (claims)
    claims := jwt.MapClaims{
        "user_id": userID,                                    
        "role":    role,                                      // 👈 FIX 2: Bake the role into the token!
        "exp":     time.Now().Add(time.Hour * 72).Unix(),     // Token expires in 72 hours
        "iat":     time.Now().Unix(),                         // Issued at time
    }

    // 3. Create the token using the HS256 hashing algorithm
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    // 4. Sign the token with your secret key and return the string
    return token.SignedString([]byte(secret))
}