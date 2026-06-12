package user_profile_controller

import (
	"fmt"
	"lendogo-backend/internal/services"
	"lendogo-backend/structures/dto"
	"lendogo-backend/structures/responses"
	"lendogo-backend/utils" // 👈 Added your utils package for S3!

	"github.com/gofiber/fiber/v2"
)

type UserProfileController struct {
	service services.UserProfileService
}

func NewUserProfileController(service services.UserProfileService) *UserProfileController {
	return &UserProfileController{service: service}
}

func (c *UserProfileController) GetProfile(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(string)
	if !ok || userID == "" {
		return responses.Error(ctx, 401, "Unauthorized")
	}

	data, err := c.service.GetMyProfile(userID)
	if err != nil {
		return responses.Error(ctx, 500, "Failed to load profile")
	}

	return responses.Success(ctx, 200, "Profile loaded", data)
}

func (c *UserProfileController) UpdateProfile(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(string)
	if !ok || userID == "" {
		return responses.Error(ctx, 401, "Unauthorized")
	}

	var req dto.UpdateProfileRequest
	// Use BodyParser which handles both JSON and Multipart Form Data
	if err := ctx.BodyParser(&req); err != nil {
		return responses.Error(ctx, 400, "Invalid request format")
	}

	// Handle Image Upload (Direct to AWS S3)
	imagePath := ""
	file, err := ctx.FormFile("profile_image")
	if err == nil {
		// 👇 THE NEW S3 FIX: No more local folders or os.MkdirAll!
		s3URL, uploadErr := utils.UploadFileToS3(file)
		if uploadErr != nil {
			fmt.Println("❌ S3 Upload Failed:", uploadErr)
			return responses.Error(ctx, 500, "Failed to upload image to S3")
		}
		
		// The database will now store the raw S3 URL
		imagePath = s3URL 
	}

	if err := c.service.UpdateMyProfile(userID, req, imagePath); err != nil {
		return responses.Error(ctx, 500, "Failed to update profile: "+err.Error())
	}

	return responses.Success(ctx, 200, "Profile updated successfully!", nil)
}