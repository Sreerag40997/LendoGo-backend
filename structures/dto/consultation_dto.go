package dto

type ConsultationReq struct {
    FullName    string `json:"full_name" validate:"required"`
    Email       string `json:"email" validate:"required,email"`
    PhoneNumber string `json:"phone_number" validate:"required"`
}