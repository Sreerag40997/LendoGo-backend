package dto

type UpdateProfileRequest struct {
	FullName    string `form:"full_name" json:"full_name"` // Maps to users table
	PhoneNumber string `form:"phone_number" json:"phone_number"`
	DateOfBirth string `form:"date_of_birth" json:"date_of_birth"`
	Pincode     string `form:"pincode" json:"pincode"`
	Address     string `form:"address" json:"address"`
}