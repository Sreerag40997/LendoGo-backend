package services

import (
	"errors"
	"lendogo-backend/internal/repositories"
	"lendogo-backend/structures/dto"
	"lendogo-backend/structures/models"
	"lendogo-backend/utils" // 👈 IMPORTANT: Import your utils package!

	"github.com/google/uuid"
)

type UserProfileService interface {
	GetMyProfile(userID string) (map[string]interface{}, error)
	UpdateMyProfile(userID string, req dto.UpdateProfileRequest, imagePath string) error
}

type userProfileServiceImpl struct {
	repo repositories.UserProfileRepository
}

func NewUserProfileService(repo repositories.UserProfileRepository) UserProfileService {
	return &userProfileServiceImpl{repo: repo}
}

func (s *userProfileServiceImpl) GetMyProfile(userID string) (map[string]interface{}, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	profile, user, err := s.repo.GetProfile(userUUID)
	if err != nil {
		return nil, err
	}

	// 👇 NEW: Convert the secure S3 URL into a 15-minute temporary link for React!
	if profile.ProfileImage != "" {
		profile.ProfileImage = utils.GeneratePresignedURL(profile.ProfileImage)
	}

	// Merge Auth Data and Profile Data for the React Frontend
	return map[string]interface{}{
		"email":         user.Email,
		"full_name":     user.FullName,
		"profile_image": profile.ProfileImage, // This is now a secure, clickable link!
		"phone_number":  profile.PhoneNumber,
		"date_of_birth": profile.DateOfBirth,
		"pincode":       profile.Pincode,
		"address":       profile.Address,
	}, nil
}

func (s *userProfileServiceImpl) UpdateMyProfile(userID string, req dto.UpdateProfileRequest, imagePath string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	newProfile := models.UserProfile{
		PhoneNumber:  req.PhoneNumber,
		DateOfBirth:  req.DateOfBirth,
		Pincode:      req.Pincode,
		Address:      req.Address,
		ProfileImage: imagePath, // Might be empty if they didn't upload a new one
	}

	return s.repo.UpsertProfile(userUUID, newProfile, req.FullName)
}