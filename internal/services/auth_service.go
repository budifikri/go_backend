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
	IsActive  bool      `json:"is_active"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	LastLogin time.Time `json:"last_login"`
	Company   struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		BusinessType string `json:"business_type"`
	} `json:"company"`
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

	if !user.IsActive {
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

	// Get company data
	var company models.Company
	companyData := struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		BusinessType string `json:"business_type"`
	}{
		ID:           "",
		Name:         "",
		BusinessType: "",
	}
	if err := s.db.First(&company, user.CompanyID).Error; err == nil {
		companyData.ID = company.ID.String()
		companyData.Name = company.Nama
		companyData.BusinessType = string(company.BusinessType)
	}

	return response.NewSuccessResponse(LoginResponse{
		UserID:    user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		FullName:  user.FullName,
		Role:      string(user.Role),
		IsActive:  user.IsActive,
		Token:     token,
		ExpiresAt: expiresAt,
		LastLogin: now,
		Company:   companyData,
	}, "Login successful")
}

// Register creates a new user
func (s *AuthService) Register(req struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	FullName    string `json:"full_name"`
	Role        string `json:"role"`
	CompanyName string `json:"company_name"`
}) response.ApiResponse {
	var existingUser models.User
	result := s.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser)
	if result.Error == nil {
		return response.NewErrorResponse(ErrUserExists.Error())
	}

	// Create company first
	company := models.Company{
		ID:           uuid.New(),
		Code:         req.Username,
		Nama:         req.CompanyName,
		Email:        req.Email,
		BusinessType: models.BusinessTypeRetail,
		Status:       models.CompanyStatusActive,
		IsActive:     true,
	}

	// Use transaction to ensure consistency
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Create(&company).Error; err != nil {
		tx.Rollback()
		return response.NewErrorResponse("Failed to create company")
	}
	if err := ensureCompanyModules(tx, company.ID, string(company.BusinessType)); err != nil {
		tx.Rollback()
		return response.NewErrorResponse("Failed to create default company modules")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		tx.Rollback()
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
		IsActive:  true,
		CompanyID: company.ID,
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return response.NewErrorResponse("Failed to create user")
	}

	expiresAt := s.jwtUtil.GetTokenExpiry()
	token, err := s.jwtUtil.GenerateToken(utils.JWTPayload{
		UserID:    user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		CompanyID: user.CompanyID.String(),
		Role:      string(user.Role),
	})
	if err != nil {
		tx.Rollback()
		return response.NewErrorResponse("Failed to generate token")
	}

	session := models.UserSession{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	if err := tx.Create(&session).Error; err != nil {
		tx.Rollback()
		return response.NewErrorResponse("Failed to create session")
	}

	if err := tx.Commit().Error; err != nil {
		return response.NewErrorResponse("Transaction failed")
	}

	return response.NewSuccessResponse(LoginResponse{
		UserID:    user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		FullName:  req.FullName,
		Role:      string(user.Role),
		IsActive:  user.IsActive,
		Token:     token,
		ExpiresAt: expiresAt,
		LastLogin: time.Now(),
		Company: struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			BusinessType string `json:"business_type"`
		}{
			ID:           company.ID.String(),
			Name:         company.Nama,
			BusinessType: string(company.BusinessType),
		},
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
