package services

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/types/response"
	"gorm.io/gorm"
)

type ModuleService struct {
	db *gorm.DB
}

func NewModuleService(db *gorm.DB) *ModuleService {
	return &ModuleService{db: db}
}

type CreateBusinessTypeInput struct {
	Code        string
	Name        string
	Description string
	IsActive    *bool
	IsDefault   *bool
	IsSystem    *bool
	SortOrder   *int
}

type UpdateBusinessTypeInput struct {
	Name        *string
	Description *string
	IsActive    *bool
	IsDefault   *bool
	IsSystem    *bool
	SortOrder   *int
}

type CreateModulePackageInput struct {
	BusinessType string
	Code         string
	Name         string
	Description  string
	IsActive     *bool
	IsDefault    *bool
	IsSystem     *bool
	SortOrder    *int
}

type UpdateModulePackageInput struct {
	BusinessType *string
	Name         *string
	Description  *string
	IsActive     *bool
	IsDefault    *bool
	IsSystem     *bool
	SortOrder    *int
}

type CompanyModuleStatus struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	BusinessType string `json:"business_type"`
	IsDefault    bool   `json:"is_default"`
	IsSystem     bool   `json:"is_system"`
	IsActive     bool   `json:"is_active"`
}

type CompanyModuleSummary struct {
	BusinessType string                `json:"business_type"`
	Modules      []string              `json:"modules"`
	Items        []CompanyModuleStatus `json:"items,omitempty"`
}

func normalizeCode(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func boolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func intValue(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}

func (s *ModuleService) businessTypeExists(code string, onlyActive bool) bool {
	query := s.db.Model(&models.BusinessType{}).Where("code = ?", normalizeCode(code))
	if onlyActive {
		query = query.Where("is_active = ?", true)
	}
	var count int64
	_ = query.Count(&count).Error
	return count > 0
}

func (s *ModuleService) GetBusinessTypes(search string, isActive *bool, limit, offset int) response.PaginatedResponse {
	var items []models.BusinessType
	var total int64
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := s.db.Model(&models.BusinessType{})
	if keyword := strings.TrimSpace(search); keyword != "" {
		like := "%" + strings.ToLower(keyword) + "%"
		query = query.Where("LOWER(code) LIKE ? OR LOWER(name) LIKE ? OR LOWER(description) LIKE ?", like, like, like)
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}
	if err := query.Order("sort_order, name").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}
	return response.NewPaginatedResponse(items, total, limit, offset)
}

func (s *ModuleService) CreateBusinessType(input CreateBusinessTypeInput) response.ApiResponse {
	code := normalizeCode(input.Code)
	if code == "" || strings.TrimSpace(input.Name) == "" {
		return response.NewErrorResponse("Business type code and name are required")
	}
	item := models.BusinessType{
		ID:          uuid.New(),
		Code:        models.BusinessTypeCode(code),
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		IsActive:    boolValue(input.IsActive, true),
		IsDefault:   boolValue(input.IsDefault, false),
		IsSystem:    boolValue(input.IsSystem, false),
		SortOrder:   intValue(input.SortOrder, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.db.Create(&item).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return response.NewErrorResponse("Business type code already exists")
		}
		return response.NewErrorResponse("Failed to create business type")
	}
	return response.NewSuccessResponse(item, "")
}

func (s *ModuleService) UpdateBusinessType(id string, input UpdateBusinessTypeInput) response.ApiResponse {
	businessTypeID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Business type not found")
	}
	updates := map[string]interface{}{}
	if input.Name != nil {
		updates["name"] = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		updates["description"] = strings.TrimSpace(*input.Description)
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}
	if input.IsDefault != nil {
		updates["is_default"] = *input.IsDefault
	}
	if input.IsSystem != nil {
		updates["is_system"] = *input.IsSystem
	}
	if input.SortOrder != nil {
		updates["sort_order"] = *input.SortOrder
	}
	if len(updates) == 0 {
		return response.NewErrorResponse("No fields to update")
	}
	updates["updated_at"] = time.Now()
	res := s.db.Model(&models.BusinessType{}).Where("id = ?", businessTypeID).Updates(updates)
	if res.Error != nil {
		return response.NewErrorResponse("Failed to update business type")
	}
	if res.RowsAffected == 0 {
		return response.NewErrorResponse("Business type not found")
	}
	var item models.BusinessType
	if err := s.db.First(&item, "id = ?", businessTypeID).Error; err != nil {
		return response.NewErrorResponse("Business type not found")
	}
	return response.NewSuccessResponse(item, "")
}

func (s *ModuleService) GetModulePackages(businessType, search string, isActive *bool, limit, offset int) response.PaginatedResponse {
	var items []models.ModulePackage
	var total int64
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	query := s.db.Model(&models.ModulePackage{})
	if businessType = normalizeCode(businessType); businessType != "" {
		query = query.Where("business_type = ?", businessType)
	}
	if keyword := strings.TrimSpace(search); keyword != "" {
		like := "%" + strings.ToLower(keyword) + "%"
		query = query.Where("LOWER(code) LIKE ? OR LOWER(name) LIKE ? OR LOWER(description) LIKE ?", like, like, like)
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}
	if err := query.Order("sort_order, name").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}
	return response.NewPaginatedResponse(items, total, limit, offset)
}

func (s *ModuleService) CreateModulePackage(input CreateModulePackageInput) response.ApiResponse {
	businessType := normalizeCode(input.BusinessType)
	code := normalizeCode(input.Code)
	if businessType == "" || code == "" || strings.TrimSpace(input.Name) == "" {
		return response.NewErrorResponse("Business type, code, and name are required")
	}
	if !s.businessTypeExists(businessType, false) {
		return response.NewErrorResponse("Business type not found")
	}
	item := models.ModulePackage{
		ID:           uuid.New(),
		BusinessType: models.BusinessTypeCode(businessType),
		Code:         code,
		Name:         strings.TrimSpace(input.Name),
		Description:  strings.TrimSpace(input.Description),
		IsActive:     boolValue(input.IsActive, true),
		IsDefault:    boolValue(input.IsDefault, false),
		IsSystem:     boolValue(input.IsSystem, false),
		SortOrder:    intValue(input.SortOrder, 0),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := s.db.Create(&item).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return response.NewErrorResponse("Module package code already exists")
		}
		return response.NewErrorResponse("Failed to create module package")
	}
	return response.NewSuccessResponse(item, "")
}

func (s *ModuleService) UpdateModulePackage(id string, input UpdateModulePackageInput) response.ApiResponse {
	moduleID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Module package not found")
	}
	updates := map[string]interface{}{}
	if input.BusinessType != nil {
		businessType := normalizeCode(*input.BusinessType)
		if !s.businessTypeExists(businessType, false) {
			return response.NewErrorResponse("Business type not found")
		}
		updates["business_type"] = businessType
	}
	if input.Name != nil {
		updates["name"] = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		updates["description"] = strings.TrimSpace(*input.Description)
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}
	if input.IsDefault != nil {
		updates["is_default"] = *input.IsDefault
	}
	if input.IsSystem != nil {
		updates["is_system"] = *input.IsSystem
	}
	if input.SortOrder != nil {
		updates["sort_order"] = *input.SortOrder
	}
	if len(updates) == 0 {
		return response.NewErrorResponse("No fields to update")
	}
	updates["updated_at"] = time.Now()
	res := s.db.Model(&models.ModulePackage{}).Where("id = ?", moduleID).Updates(updates)
	if res.Error != nil {
		return response.NewErrorResponse("Failed to update module package")
	}
	if res.RowsAffected == 0 {
		return response.NewErrorResponse("Module package not found")
	}
	var item models.ModulePackage
	if err := s.db.First(&item, "id = ?", moduleID).Error; err != nil {
		return response.NewErrorResponse("Module package not found")
	}
	return response.NewSuccessResponse(item, "")
}

func (s *ModuleService) GetCompanyModules(companyID string, includeItems bool) response.ApiResponse {
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Company not found")
	}
	return s.getCompanyModulesByUUID(companyUUID, includeItems)
}

func (s *ModuleService) getCompanyModulesByUUID(companyID uuid.UUID, includeItems bool) response.ApiResponse {
	var company models.Company
	if err := s.db.First(&company, "id = ?", companyID).Error; err != nil {
		return response.NewErrorResponse("Company not found")
	}
	var packages []models.ModulePackage
	if err := s.db.Where("business_type = ? AND is_active = ?", company.BusinessType, true).Order("sort_order, name").Find(&packages).Error; err != nil {
		return response.NewErrorResponse("Failed to load module packages")
	}
	var companyModules []models.CompanyModule
	if err := s.db.Where("company_id = ?", companyID).Find(&companyModules).Error; err != nil {
		return response.NewErrorResponse("Failed to load company modules")
	}
	activeByCode := make(map[string]models.CompanyModule, len(companyModules))
	for _, item := range companyModules {
		activeByCode[item.ModuleCode] = item
	}
	summary := CompanyModuleSummary{
		BusinessType: string(company.BusinessType),
		Modules:      []string{},
	}
	if includeItems {
		summary.Items = make([]CompanyModuleStatus, 0, len(packages))
	}
	for _, item := range packages {
		status := activeByCode[item.Code]
		if status.IsActive {
			summary.Modules = append(summary.Modules, item.Code)
		}
		if includeItems {
			summary.Items = append(summary.Items, CompanyModuleStatus{
				Code:         item.Code,
				Name:         item.Name,
				Description:  item.Description,
				BusinessType: string(item.BusinessType),
				IsDefault:    item.IsDefault,
				IsSystem:     item.IsSystem,
				IsActive:     status.IsActive,
			})
		}
	}
	return response.NewSuccessResponse(summary, "")
}

func (s *ModuleService) ToggleCompanyModule(companyID, moduleCode, actorUserID string, isActive bool) response.ApiResponse {
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Company not found")
	}
	var company models.Company
	if err := s.db.First(&company, "id = ?", companyUUID).Error; err != nil {
		return response.NewErrorResponse("Company not found")
	}
	moduleCode = normalizeCode(moduleCode)
	var pkg models.ModulePackage
	if err := s.db.Where("code = ?", moduleCode).First(&pkg).Error; err != nil {
		return response.NewErrorResponse("Module package not found")
	}
	if pkg.BusinessType != company.BusinessType {
		return response.NewErrorResponse("Module package does not match company business type")
	}
	var activatedBy *uuid.UUID
	if actorUserID != "" {
		if userID, parseErr := uuid.Parse(actorUserID); parseErr == nil {
			activatedBy = &userID
		}
	}
	now := time.Now()
	var existing models.CompanyModule
	res := s.db.Where("company_id = ? AND module_code = ?", companyUUID, moduleCode).First(&existing)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		item := models.CompanyModule{
			ID:         uuid.New(),
			CompanyID:  companyUUID,
			ModuleCode: moduleCode,
			IsActive:   isActive,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if isActive {
			item.ActivatedAt = &now
			item.ActivatedBy = activatedBy
		}
		if err := s.db.Create(&item).Error; err != nil {
			return response.NewErrorResponse("Failed to update company module")
		}
		return s.getCompanyModulesByUUID(companyUUID, true)
	}
	if res.Error != nil {
		return response.NewErrorResponse("Failed to update company module")
	}
	updates := map[string]interface{}{
		"is_active":    isActive,
		"updated_at":   now,
		"activated_by": activatedBy,
	}
	if isActive {
		updates["activated_at"] = &now
	} else {
		updates["activated_at"] = nil
		updates["activated_by"] = nil
	}
	if err := s.db.Model(&models.CompanyModule{}).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
		return response.NewErrorResponse("Failed to update company module")
	}
	return s.getCompanyModulesByUUID(companyUUID, true)
}
