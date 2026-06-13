package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"lendogo-backend/internal/repositories"
	"lendogo-backend/structures/models"
	"lendogo-backend/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ==========================================
// 1. DATA TRANSFER OBJECTS (DTOs)
// ==========================================

type CreateUserDTO struct {
	FullName     string `json:"full_name"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	MobileNumber string `json:"mobile_number"`
	DOB          string `json:"dob"`
	PanCard      string `json:"pan_card_number"`
	CreditRating string `json:"credit_rating"`
	CreditScore  int    `json:"credit_score"`
	Address      string `json:"address"`
	City         string `json:"city"`
	State        string `json:"state"`
	Pincode      string `json:"pincode"`
}

type CreateStaffDTO struct {
	FullName    string          `json:"full_name"`
	Email       string          `json:"email"`
	Password    string          `json:"password"`
	Role        string          `json:"role"`
	Permissions map[string]bool `json:"permissions"`
}

// ==========================================
// 2. INTERFACE DEFINITION
// ==========================================

type AdminService interface {
	GetAllUsers() ([]models.User, error)
	CreateUserFromAdmin(req CreateUserDTO) (string, error)
	UpdateUser(id string, updates map[string]interface{}) error
	DeleteUser(userIDStr string) error
	UpdateUserStatus(id string, status string) error

	CreateStaff(req CreateStaffDTO) error
	GetAllStaff() ([]models.Staff, error)
	AdminLogin(email string, password string) (*models.Staff, error)
	DeleteStaff(staffIDStr string) error
	UpdateStaffStatus(id string, status string) error
}

type adminServiceImpl struct {
	repo repositories.AdminRepository
}

func NewAdminService(repo repositories.AdminRepository) AdminService {
	return &adminServiceImpl{repo: repo}
}

// ==========================================
// 3. AUTH & STAFF LOGIC
// ==========================================

func (s *adminServiceImpl) AdminLogin(email string, plainPassword string) (*models.Staff, error) {
	staff, err := s.repo.GetStaffByEmail(email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}
	if staff.Status != "Active" {
		return nil, errors.New("this account is inactive")
	}
	if !utils.CheckPasswordHash(plainPassword, staff.Password) {
		return nil, errors.New("invalid email or password")
	}
	return staff, nil
}

func (s *adminServiceImpl) CreateStaff(req CreateStaffDTO) error {
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}

	staffMember := &models.Staff{
		FullName:    req.FullName,
		Email:       req.Email,
		Password:    hashedPassword,
		Role:        req.Role,
		Permissions: req.Permissions,
		Status:      "Active",
	}
	return s.repo.CreateStaff(staffMember)
}

func (s *adminServiceImpl) GetAllStaff() ([]models.Staff, error) {
	return s.repo.GetAllStaff()
}

func (s *adminServiceImpl) DeleteStaff(staffIDStr string) error {
	staffID, err := uuid.Parse(staffIDStr)
	if err != nil {
		return err
	}
	return s.repo.HardDeleteStaff(staffID)
}

func (s *adminServiceImpl) UpdateStaffStatus(id string, status string) error {
	return s.repo.UpdateStaffStatus(id, status)
}

// ==========================================
// 4. USER MANAGEMENT LOGIC
// ==========================================

func (s *adminServiceImpl) GetAllUsers() ([]models.User, error) {
	return s.repo.GetAllUsers()
}

func (s *adminServiceImpl) CreateUserFromAdmin(req CreateUserDTO) (string, error) {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	plainPassword := "Lendo" + hex.EncodeToString(b)[:4] + "@"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)

	userID := uuid.New()
	user := &models.User{
		ID: userID, FullName: req.FullName, Email: req.Email, 
		Password: string(hashedPassword), Role: req.Role, 
		IsEmailVerified: true, Status: "Active",
	}
	profile := &models.UserProfile{
		UserID: userID, PhoneNumber: req.MobileNumber, DateOfBirth: req.DOB,
		Address: req.Address, City: req.City, State: req.State, Pincode: req.Pincode,
		TrustScore: req.CreditScore, PanCardNumber: req.PanCard, CreditRating: req.CreditRating,
	}

	err := s.repo.CreateSystemUserTx(user, profile)
	return plainPassword, err
}

func (s *adminServiceImpl) UpdateUser(id string, updates map[string]interface{}) error {
	return s.repo.UpdateUser(id, updates)
}

func (s *adminServiceImpl) DeleteUser(userIDStr string) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return err
	}
	return s.repo.HardDeleteUser(userID)
}

func (s *adminServiceImpl) UpdateUserStatus(id string, status string) error {
	return s.repo.UpdateUserStatus(id, status)
}