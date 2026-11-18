package dto

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ApproveUserRequest struct {
	ApprovalStatus string `json:"approval_status" validate:"omitempty,oneof=approved rejected"`
}

type RegisterResponse struct {
	Message         string `json:"message"`
	UserID          string `json:"user_id"`
	Email           string `json:"email"`
	IsEmailVerified bool   `json:"is_email_verified"`
	ApprovalStatus  string `json:"approval_status"`
}
