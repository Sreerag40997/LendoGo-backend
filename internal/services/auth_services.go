package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"lendogo-backend/database"
	"lendogo-backend/internal/repositories"
	"lendogo-backend/structures/dto"
	"lendogo-backend/structures/models"
	"lendogo-backend/utils"
)

type AuthService interface {
	Register(req dto.RegisterReq) error
	Login(req dto.LoginReq) (*dto.AuthRes, error)

	// New Forgot Password Methods
	SendForgotPasswordOTP(email string) error
	ResetPassword(req dto.ResetPasswordReq) error
}

type authServiceImpl struct {
	userRepo repositories.UserRepository
}

func NewAuthService(repo repositories.UserRepository) AuthService {
	return &authServiceImpl{userRepo: repo}
}

// ==========================================
// 1. REGISTER
// ==========================================
func (s *authServiceImpl) Register(req dto.RegisterReq) error {
	existingUser, _ := s.userRepo.FindByEmail(req.Email)
	if existingUser != nil {
		return errors.New("email already in use")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return errors.New("failed to secure password")
	}

	newUser := &models.User{
		FullName: req.FullName,
		Email:    req.Email,
		Password: hashedPassword,
	}

	return s.userRepo.CreateUser(newUser)
}

// ==========================================
// 2. LOGIN
// ==========================================
func (s *authServiceImpl) Login(req dto.LoginReq) (*dto.AuthRes, error) {
	// Try standard users first
	user, err := s.userRepo.FindByEmail(req.Email)
	if err == nil {
		isValid := utils.CheckPasswordHash(req.Password, user.Password)
		if !isValid {
			return nil, errors.New("invalid email or password")
		}

		effectiveRole := user.Role
		if user.Email == "admin@gmail.com" || user.Email == "admin.flow@lendogo.com" {
			effectiveRole = "admin"
		}

		token, err := utils.GenerateToken(user.ID.String(), effectiveRole, user.FullName, user.Email)
		if err != nil {
			return nil, errors.New("failed to generate login token")
		}

		res := &dto.AuthRes{
			Token: token,
			User: dto.UserRes{
				ID:       user.ID.String(),
				FullName: user.FullName,
				Email:    user.Email,
				Role:     effectiveRole,
				Status:   user.Status,
			},
		}

		return res, nil
	}

	// If not found in users, check staffs table
	var staff models.Staff
	if err := database.DB.Where("email = ?", req.Email).First(&staff).Error; err == nil {
		isValid := utils.CheckPasswordHash(req.Password, staff.Password)
		if !isValid {
			return nil, errors.New("invalid email or password")
		}

		if staff.Status == "Blocked" {
			return nil, errors.New("account suspended by an administrator")
		}

		token, err := utils.GenerateToken(staff.ID.String(), staff.Role, staff.FullName, staff.Email)
		if err != nil {
			return nil, errors.New("failed to generate login token")
		}

		res := &dto.AuthRes{
			Token: token,
			User: dto.UserRes{
				ID:          staff.ID.String(),
				FullName:    staff.FullName,
				Email:       staff.Email,
				Role:        staff.Role,
				Status:      staff.Status,
				Permissions: staff.Permissions,
			},
		}

		return res, nil
	}

	return nil, errors.New("invalid email or password")
}

// ==========================================
// 3. FORGOT PASSWORD: Send OTP (with Rate Limit)
// ==========================================
func (s *authServiceImpl) SendForgotPasswordOTP(email string) error {
	ctx := context.Background()
	cooldownKey := "cooldown:forgot:" + email
	otpKey := "forgot_otp:" + email

	// STEP A: Check Rate Limit (2 minutes)
	exists, _ := database.RedisClient.Exists(ctx, cooldownKey).Result()
	if exists > 0 {
		return errors.New("please wait 2 minutes before requesting another OTP")
	}

	// STEP B: Check if the email actually exists in the database
	_, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return errors.New("no account found with this email address")
	}

	// STEP C: Generate a 6-digit OTP
	rand.Seed(time.Now().UnixNano())
	otp := fmt.Sprintf("%06d", rand.Intn(1000000))

	// STEP D: Save OTP to Redis (Expires in 10 minutes)
	err = database.RedisClient.Set(ctx, otpKey, otp, 10*time.Minute).Err()
	if err != nil {
		return errors.New("failed to generate OTP")
	}

	// STEP E: Set Cooldown in Redis (Locks them out for 2 minutes)
	database.RedisClient.Set(ctx, cooldownKey, "locked", 2*time.Minute)

	// STEP F: Send the Email
	err = utils.SendOTPEmail(email, otp)
	if err != nil {
		return errors.New("failed to send email")
	}

	return nil
}

// ==========================================
// 4. FORGOT PASSWORD: Reset Password
// ==========================================
func (s *authServiceImpl) ResetPassword(req dto.ResetPasswordReq) error {
	ctx := context.Background()
	otpKey := "forgot_otp:" + req.Email

	// STEP A: Verify OTP from Redis
	storedOTP, err := database.RedisClient.Get(ctx, otpKey).Result()
	if err != nil || storedOTP != req.OTP {
		return errors.New("invalid or expired OTP")
	}

	// STEP B: Hash the new password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return errors.New("failed to secure new password")
	}

	// STEP C: Update Database
	err = s.userRepo.UpdatePassword(req.Email, hashedPassword)
	if err != nil {
		return errors.New("failed to update password in database")
	}

	// STEP D: Delete the OTP from Redis so it cannot be reused
	database.RedisClient.Del(ctx, otpKey)

	return nil
}