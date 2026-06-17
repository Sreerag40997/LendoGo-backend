package middlewares

import (
    "fmt"
    "os"
    "strings"
    "github.com/gofiber/fiber/v2"
    "github.com/golang-jwt/jwt/v5"
)

func Protected() fiber.Handler {
    return func(ctx *fiber.Ctx) error {
        var tokenString string

        // 1. Prioritize the Authorization header (Bearer token)
        authHeader := ctx.Get("Authorization")
        if strings.HasPrefix(authHeader, "Bearer ") {
            tokenString = strings.TrimPrefix(authHeader, "Bearer ")
        }

        // 2. Fall back to the HTTP-Only cookie if no header is provided
        if tokenString == "" {
            tokenString = ctx.Cookies("access_token")
        }

        if tokenString == "" {
            return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Unauthorized: Missing token",
            })
        }

        secret := os.Getenv("JWT_SECRET")
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return []byte(secret), nil
        })

        if err != nil || !token.Valid {
            return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Unauthorized: Invalid or expired token",
            })
        }

        if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
            ctx.Locals("user_id", claims["user_id"])
            ctx.Locals("name", claims["name"])
            ctx.Locals("role", claims["role"]) 
            ctx.Locals("email", claims["email"])
            return ctx.Next()
        }

        return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Unauthorized: Failed to process token claims",
        })
    }
}