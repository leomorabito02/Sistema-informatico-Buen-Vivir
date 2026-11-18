package dto

import "auth/internal/models"

type RegisterRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2"`
	LastName  string `json:"last_name" validate:"required,min=2"`
	Email     string `json:"email" validate:"required,email"`
	Phone     string `json:"phone" validate:"required"`
	Password  string `json:"password" validate:"required,password"`
	Role      string `json:"role" validate:"required,oneof=admin volunteer participant teacher company"`
}

// CheckCredentialsRequest is used in step 1: verify email and password without selecting a role
type CheckCredentialsRequest struct {
	Email         string `json:"email" validate:"required,email"`
	Password      string `json:"password" validate:"required"`
	TrustedDevice string `json:"-"`
	UserAgent     string `json:"-"`
}

// AvailableAccount represents a user account associated with an email, each with a unique role
type AvailableAccount struct {
	UserID          string `json:"user_id"`
	RoleID          string `json:"role_id"`
	RoleName        string `json:"role_name"`
	Status          string `json:"status"`
	ApprovalStatus  string `json:"approval_status,omitempty"`
	IsEmailVerified bool   `json:"is_email_verified"`
}

// CheckCredentialsResponse is the result of step 1, containing the list of accounts for the email
type CheckCredentialsResponse struct {
	Accounts        []AvailableAccount `json:"accounts,omitempty"`
	LoginNonce      string             `json:"login_nonce,omitempty"`
	IsEmailVerified bool               `json:"is_email_verified"`
	Email           string             `json:"email,omitempty"`
	RequiresTwofa   bool               `json:"requires_2fa"`
	TwofaMethod     string             `json:"twofa_method,omitempty"`
	TwofaNonce      string             `json:"twofa_nonce,omitempty"`
}

// VerifyTwofaRequest verifies the 2FA challenge code during login
type VerifyTwofaRequest struct {
	TwofaNonce  string `json:"twofa_nonce" validate:"required"`
	Code        string `json:"code" validate:"required,min=6,max=6"`
	TrustDevice bool   `json:"trust_device"`
}

// ResendTwofaRequest triggers a new email OTP code for an active 2FA login challenge
type ResendTwofaRequest struct {
	TwofaNonce string `json:"twofa_nonce" validate:"required"`
}

// SetupTwofaRequest initiates the 2FA setup flow in the user profile
type SetupTwofaRequest struct {
	Method string `json:"method" validate:"required,oneof=email totp"`
}

// SetupTwofaResponse contains the setup configuration for TOTP or confirmation message for email
type SetupTwofaResponse struct {
	Method     string `json:"method"`
	Secret     string `json:"secret,omitempty"`
	OtpauthURL string `json:"otpauth_url,omitempty"`
	Message    string `json:"message"`
}

// EnableTwofaRequest confirms and activates 2FA using a valid code
type EnableTwofaRequest struct {
	Method string `json:"method" validate:"required,oneof=email totp"`
	Code   string `json:"code" validate:"required,min=6,max=6"`
}

// DisableTwofaRequest disables 2FA for the user after password or code confirmation
type DisableTwofaRequest struct {
	Password string `json:"password,omitempty"`
	Code     string `json:"code,omitempty"`
}

// LoginRequest is used in step 2: select a specific account/role to receive a token
type LoginRequest struct {
	UserID     string `json:"user_id" validate:"required"`
	RoleID     string `json:"role_id" validate:"required"`
	LoginNonce string `json:"login_nonce" validate:"required"`
}

// GoogleLoginRequest is used in step 1 for Google: verify firebase token, no role needed
type GoogleLoginRequest struct {
	IDToken       string `json:"id_token" validate:"required"`
	GoogleID      string `json:"google_id" validate:"required"`
	TrustedDevice string `json:"-"`
	UserAgent     string `json:"-"`
}

// LinkGoogleRequest is used to link a Google account to an authenticated user profile
type LinkGoogleRequest struct {
	IDToken  string `json:"id_token" validate:"required"`
	GoogleID string `json:"google_id" validate:"required"`
}

// GoogleSelectRoleRequest is used in step 2 for Google: select which account/role to log into
type GoogleSelectRoleRequest struct {
	UserID string `json:"user_id" validate:"required"`
	RoleID string `json:"role_id" validate:"required"`
}

type LoginResponse struct {
	Token          string        `json:"token"`
	// #nosec G117
	RefreshToken   string        `json:"refresh_token,omitempty"`
	User           *models.User  `json:"user"`
	ApprovalStatus string        `json:"approval_status,omitempty"`
	UserStatus     string        `json:"user_status,omitempty"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,password"`
}
