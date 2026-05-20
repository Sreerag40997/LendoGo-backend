package models

import "time"

// 1. The Database Table Schema
type User struct {
    ID              string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Username        string    `json:"username" gorm:"unique;not null"`
    Email           string    `json:"email" gorm:"unique;not null"`
    Password        string    `json:"-" gorm:"not null"` 
    IsEmailVerified bool      `json:"is_email_verified" gorm:"default:false"`
    CreatedAt       time.Time `json:"created_at"`
}

// 2. The Incoming Request Payload for Signing Up
type SignupRequest struct {
    Username string `json:"username"`
    Email    string `json:"email"`
    Password string `json:"password"`
}

// 3. The Incoming Request Payload for OTP Verification (ADD THIS RIGHT HERE!)
type VerifyOTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}