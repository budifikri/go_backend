package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/types/response"
	"github.com/pos-retail/go_backend/internal/utils"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("username or email already exists")
	ErrUserInactive       = errors.New("user account is inactive")
)

// AuthService handles authentication logic
type AuthService struct {
	db      *gorm.DB
	jwtUtil *utils.JWTUtil
}

// NewAuthService creates a new auth service
func NewAuthService(db *gorm.DB, jwtUtil *utils.JWTUtil) *AuthService {
	return &AuthService{
		db:      db,
		jwtUtil: jwtUtil,
	}
}

// LoginResponse represents login response
type LoginResponse struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	LastLogin time.Time `json:"last_login"`
}

// Login authenticates user
func (s *AuthService) Login(username, password string) response.ApiResponse {
	var user models.User

	result := s.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return response.NewErrorResponse(ErrInvalidCredentials.Error())
		}
		return response.NewErrorResponse("Login failed")
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return response.NewErrorResponse(ErrInvalidCredentials.Error())
	}

	if user.Status != models.StatusActive {
		return response.NewErrorResponse(ErrUserInactive.Error())
	}

	token, err := s.jwtUtil.GenerateToken(utils.JWTPayload{
		UserID:    user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		CompanyID: user.CompanyID.String(),
		Role:      string(user.Role),
	})

	if err != nil {
		return response.NewErrorResponse("Failed to generate token")
	}

	expiresAt := s.jwtUtil.GetTokenExpiry()

	session := models.UserSession{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	if err := s.db.Create(&session).Error; err != nil {
		return response.NewErrorResponse("Failed to create session")
	}

	now := time.Now()
	s.db.Model(&user).Update("last_login", now)

	return response.NewSuccessResponse(LoginResponse{
		UserID:    user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		FullName:  user.FullName,
		Role:      string(user.Role),
		Status:    string(user.Status),
		Token:     token,
		ExpiresAt: expiresAt,
		LastLogin: now,
	}, "Login successful")
}

// Register creates a new user
func (s *AuthService) Register(req struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}) response.ApiResponse {
	var existingUser models.User
	result := s.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser)
	if result.Error == nil {
		return response.NewErrorResponse(ErrUserExists.Error())
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return response.NewErrorResponse("Failed to hash password")
	}

	role := models.UserRole(req.Role)
	if role == "" {
		role = models.RoleStaff
	}

	user := models.User{
		ID:        uuid.New(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		FullName:  req.FullName,
		Role:      role,
		Status:    models.StatusActive,
		CompanyID: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
	}

	if err := s.db.Create(&user).Error; err != nil {
		return response.NewErrorResponse("Failed to create user")
	}

	expiresAt := s.jwtUtil.GetTokenExpiry()
	token, _ := s.jwtUtil.GenerateToken(utils.JWTPayload{
		UserID:    user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		CompanyID: user.CompanyID.String(),
		Role:      string(user.Role),
	})

	session := models.UserSession{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	s.db.Create(&session)

	return response.NewSuccessResponse(LoginResponse{
		UserID:    user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		FullName:  user.FullName,
		Role:      string(user.Role),
		Status:    string(user.Status),
		Token:     token,
		ExpiresAt: expiresAt,
		LastLogin: time.Now(),
	}, "Registration successful")
}

// Logout invalidates user session
func (s *AuthService) Logout(token string) response.ApiResponse {
	result := s.db.Where("token = ?", token).Delete(&models.UserSession{})
	if result.Error != nil {
		return response.NewErrorResponse("Logout failed")
	}

	return response.NewSuccessResponse(nil, "Logged out successfully")
}
