package request

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email,max=100"`
	Password string `json:"password" validate:"required,min=6,max=128"`
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
	Role     string `json:"role" validate:"omitempty,oneof=admin manager cashier staff"`
	IsActive *bool  `json:"is_active" validate:"omitempty"`
}

type UpdateUserRequest struct {
	Username *string `json:"username" validate:"omitempty,min=3,max=50"`
	Email    *string `json:"email" validate:"omitempty,email,max=100"`
	Password *string `json:"password" validate:"omitempty,min=6,max=128"`
	FullName *string `json:"full_name" validate:"omitempty,min=2,max=100"`
	Role     *string `json:"role" validate:"omitempty,oneof=admin manager cashier staff"`
	IsActive *bool   `json:"is_active" validate:"omitempty"`
}

type UpdateUserPasswordRequest struct {
	Password string `json:"password" validate:"required,min=6,max=128"`
}
