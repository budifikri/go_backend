package services

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	applogger "github.com/pos-retail/go_backend/internal/logger"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/types/response"
	"github.com/pos-retail/go_backend/internal/utils"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

type CreateUserInput struct {
	Username  string
	Email     string
	Password  string
	FullName  string
	Role      string
	IsActive  *bool
	CompanyID string
}

type UpdateUserInput struct {
	Username *string
	Email    *string
	Password *string
	FullName *string
	Role     *string
	IsActive *bool
}

func (s *UserService) GetUsers(companyID string, search, role string, isActive *bool, limit, offset int) response.PaginatedResponse {
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}

	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := s.db.Model(&models.User{}).Where("company_id = ?", companyUUID)

	if search != "" {
		q := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(username) LIKE ? OR LOWER(email) LIKE ? OR LOWER(full_name) LIKE ?", q, q, q)
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}

	var users []models.User
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}

	return response.NewPaginatedResponse(users, total, limit, offset)
}

func (s *UserService) GetUserByID(id, companyID string) response.ApiResponse {
	userID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("User not found")
	}
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("User not found")
	}

	var user models.User
	if err := s.db.Where("id = ? AND company_id = ?", userID, companyUUID).First(&user).Error; err != nil {
		return response.NewErrorResponse("User not found")
	}

	return response.NewSuccessResponse(user, "")
}

func (s *UserService) CreateUser(input CreateUserInput) response.ApiResponse {
	companyUUID, err := uuid.Parse(input.CompanyID)
	if err != nil {
		return response.NewErrorResponse("Invalid company")
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	role := models.UserRole(input.Role)
	if role == "" {
		role = models.RoleStaff
	}

	status := models.StatusActive
	if !isActive {
		status = models.StatusInactive
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return response.NewErrorResponse("Failed to create user")
	}

	user := models.User{
		ID:        uuid.New(),
		Username:  input.Username,
		Email:     input.Email,
		Password:  hashedPassword,
		FullName:  input.FullName,
		Role:      role,
		Status:    status,
		IsActive:  isActive,
		CompanyID: companyUUID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	logInstance := applogger.Default()
	err = applogger.AuditCreate(s.db, logInstance, "users", user.ID.String(), "", input.CompanyID, func() error {
		return s.db.Create(&user).Error
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return response.NewErrorResponse("Username or email already exists")
		}
		return response.NewErrorResponse("Failed to create user")
	}

	return response.NewSuccessResponse(user, "")
}

func (s *UserService) UpdateUser(id, companyID string, input UpdateUserInput) response.ApiResponse {
	userID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("User not found")
	}
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("User not found")
	}

	updates := map[string]interface{}{}
	if input.Username != nil {
		updates["username"] = *input.Username
	}
	if input.Email != nil {
		updates["email"] = *input.Email
	}
	if input.FullName != nil {
		updates["full_name"] = *input.FullName
	}
	if input.Role != nil {
		updates["role"] = *input.Role
	}
	if input.Password != nil {
		hashedPassword, err := utils.HashPassword(*input.Password)
		if err != nil {
			return response.NewErrorResponse("Failed to update user")
		}
		updates["password"] = hashedPassword
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
		if *input.IsActive {
			updates["status"] = string(models.StatusActive)
		} else {
			updates["status"] = string(models.StatusInactive)
		}
	}

	if len(updates) == 0 {
		return response.NewErrorResponse("No fields to update")
	}
	updates["updated_at"] = time.Now()

	logInstance := applogger.Default()
	err = applogger.AuditUpdate(s.db, logInstance, "users", userID.String(), "", companyID, func() error {
		res := s.db.Model(&models.User{}).Where("id = ? AND company_id = ?", userID, companyUUID).Updates(updates)
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return res.Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewErrorResponse("User not found")
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return response.NewErrorResponse("Username or email already exists")
		}
		return response.NewErrorResponse("Failed to update user")
	}

	var user models.User
	if err := s.db.Where("id = ? AND company_id = ?", userID, companyUUID).First(&user).Error; err != nil {
		return response.NewErrorResponse("User not found")
	}

	return response.NewSuccessResponse(user, "")
}

func (s *UserService) DeleteUser(id, companyID string) response.ApiResponse {
	userID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("User not found")
	}
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("User not found")
	}

	var user models.User
	if err := s.db.Where("id = ? AND company_id = ?", userID, companyUUID).First(&user).Error; err != nil {
		return response.NewErrorResponse("User not found")
	}

	logInstance := applogger.Default()
	err = applogger.AuditDelete(s.db, logInstance, "users", userID.String(), "", companyID, func() error {
		if err := s.db.Where("user_id = ?", userID).Delete(&models.UserSession{}).Error; err != nil {
			return err
		}
		return s.db.Delete(&models.User{}, "id = ? AND company_id = ?", userID, companyUUID).Error
	})
	if err != nil {
		return response.NewErrorResponse("Failed to delete user")
	}

	return response.NewSuccessResponse(nil, "User deleted successfully")
}

func (s *UserService) UpdateUserPassword(id, companyID, password string) response.ApiResponse {
	userID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("User not found")
	}
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("User not found")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return response.NewErrorResponse("Failed to update user password")
	}

	res := s.db.Model(&models.User{}).
		Where("id = ? AND company_id = ?", userID, companyUUID).
		Updates(map[string]interface{}{"password": hashedPassword, "updated_at": time.Now()})
	if res.Error != nil {
		return response.NewErrorResponse("Failed to update user password")
	}
	if res.RowsAffected == 0 {
		return response.NewErrorResponse("User not found")
	}

	return response.NewSuccessResponse(nil, "User password updated successfully")
}
