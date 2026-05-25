package dto
// ==========================================
// 3-Step OTP Registration Flow
// ==========================================
type SendOTPReq struct {
    FullName string `json:"full_name"`
    Email    string `json:"email"`
}

type VerifyOTPReq struct {
    Email string `json:"email"`
    OTP   string `json:"otp"`
}

type SetPasswordReq struct {
    FullName        string `json:"full_name"` 
    Email           string `json:"email"`
    Password        string `json:"password"`
    ConfirmPassword string `json:"confirmPassword"`
}

// ==========================================
// Standard Auth Flow
// ==========================================
type RegisterReq struct {
    FullName string `json:"full_name"`
    Email    string `json:"email"`
    Password string `json:"password"`
}

type LoginReq struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

// UserRes handles the JSON structure for the user in the response
type UserRes struct {
    ID       string `json:"id"` // Changed to string for UUID compatibility
    FullName string `json:"full_name"`
    Email    string `json:"email"`
}

type AuthRes struct {
    Token string  `json:"token"`
    User  UserRes `json:"user"`
}



type ForgotPasswordReq struct {
    Email string `json:"email" validate:"required,email"`
}

type ResetPasswordReq struct {
    Email           string `json:"email" validate:"required,email"`
    OTP             string `json:"otp" validate:"required,len=6"`
    Password        string `json:"password" validate:"required,min=8"`
    ConfirmPassword string `json:"confirmPassword" validate:"required,min=8"`
}