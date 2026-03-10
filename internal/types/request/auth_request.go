package request

// LoginRequest represents login request body
type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6,max=128"`
}

// RegisterRequest represents registration request body
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email,max=100"`
	Password string `json:"password" validate:"required,min=6,max=128"`
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
	Role     string `json:"role" validate:"omitempty,oneof=admin manager cashier staff"`
	CompanyName string `json:"company_name" validate:"required,min=1,max=200"`
}

// ForgotPasswordRequest represents forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest represents password reset request
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required,min=64,max=64"`
	NewPassword string `json:"new_password" validate:"required,min=6,max=128"`
}

// VerifyEmailRequest represents email verification request
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required,min=64,max=64"`
}
