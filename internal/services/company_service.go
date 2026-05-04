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
	"gorm.io/gorm"
)

type CompanyService struct {
	db *gorm.DB
}

func NewCompanyService(db *gorm.DB) *CompanyService {
	return &CompanyService{db: db}
}

type CompanyResponse struct {
	ID              uuid.UUID               `json:"id"`
	Code            string                  `json:"code"`
	Nama            string                  `json:"nama"`
	Email           string                  `json:"email"`
	BusinessType    models.BusinessTypeCode `json:"business_type"`
	Address         *string                 `json:"address,omitempty"`
	Telp            *string                 `json:"telp,omitempty"`
	Logo            *string                 `json:"logo,omitempty"`
	Website         *string                 `json:"website,omitempty"`
	TaxID           *string                 `json:"tax_id,omitempty"`
	BusinessLicense *string                 `json:"business_license,omitempty"`
	IsActive        bool                    `json:"is_active"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
	Modules         []string                `json:"modules,omitempty"`
}

func buildCompanyResponse(company models.Company, modules []string) CompanyResponse {
	return CompanyResponse{
		ID:              company.ID,
		Code:            company.Code,
		Nama:            company.Nama,
		Email:           company.Email,
		BusinessType:    company.BusinessType,
		Address:         company.Address,
		Telp:            company.Telp,
		Logo:            company.Logo,
		Website:         company.Website,
		TaxID:           company.TaxID,
		BusinessLicense: company.BusinessLicense,
		IsActive:        company.IsActive,
		CreatedAt:       company.CreatedAt,
		UpdatedAt:       company.UpdatedAt,
		Modules:         modules,
	}
}

func (s *CompanyService) loadActiveModuleCodes(companyID uuid.UUID) []string {
	var moduleCodes []string
	_ = s.db.Model(&models.CompanyModule{}).
		Where("company_id = ? AND is_active = ?", companyID, true).
		Order("module_code").
		Pluck("module_code", &moduleCodes).Error
	return moduleCodes
}

func (s *CompanyService) companyHasDependentData(tx *gorm.DB, companyID uuid.UUID) (bool, error) {
	var dep struct {
		HasDependentData bool `gorm:"column:has_dependent_data"`
	}
	err := tx.Raw(`
		SELECT EXISTS(
			SELECT 1 FROM warehouses WHERE company_id = ?
			UNION ALL
			SELECT 1 FROM users WHERE company_id = ?
			UNION ALL
			SELECT 1 FROM products WHERE company_id = ?
			UNION ALL
			SELECT 1 FROM customers WHERE company_id = ?
			UNION ALL
			SELECT 1 FROM suppliers WHERE company_id = ?
			UNION ALL
			SELECT 1 FROM sales WHERE company_id = ?
			UNION ALL
			SELECT 1 FROM purchase_orders WHERE company_id = ?
		) AS has_dependent_data
	`, companyID, companyID, companyID, companyID, companyID, companyID, companyID).Scan(&dep).Error
	if err != nil {
		return false, err
	}
	return dep.HasDependentData, nil
}

func (s *CompanyService) GetCompanies(companyID *uuid.UUID, search string, limit, offset int) response.PaginatedResponse {
	var companies []models.Company
	var total int64

	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := s.db.Model(&models.Company{})
	if companyID != nil {
		query = query.Where("id = ?", companyID)
	}
	if search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(code) LIKE ? OR LOWER(nama) LIKE ? OR LOWER(email) LIKE ?", like, like, like)
	}
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}
	if err := query.Order("nama").Limit(limit).Offset(offset).Find(&companies).Error; err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}

	items := make([]CompanyResponse, 0, len(companies))
	for _, company := range companies {
		items = append(items, buildCompanyResponse(company, s.loadActiveModuleCodes(company.ID)))
	}

	return response.NewPaginatedResponse(items, total, limit, offset)
}

func (s *CompanyService) GetCompanyByID(id string) response.ApiResponse {
	companyID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Company not found")
	}
	var company models.Company
	if err := s.db.First(&company, "id = ?", companyID).Error; err != nil {
		return response.NewErrorResponse("Company not found")
	}
	return response.NewSuccessResponse(buildCompanyResponse(company, s.loadActiveModuleCodes(company.ID)), "")
}

func (s *CompanyService) GetCompanyByUserCompanyID(companyID string) response.ApiResponse {
	return s.GetCompanyByID(companyID)
}

type CreateCompanyInput struct {
	Code            string
	Nama            string
	Email           string
	BusinessType    string
	ModuleCodes     []string
	Address         *string
	Telp            *string
	Website         *string
	TaxID           *string
	BusinessLicense *string
	IsActive        *bool
}

func (s *CompanyService) CreateCompany(input CreateCompanyInput) response.ApiResponse {
	businessType := normalizeCode(input.BusinessType)
	if businessType == "" {
		businessType = string(models.BusinessTypeRetail)
	}
	if err := validateBusinessType(s.db, businessType); err != nil {
		return response.NewErrorResponse("Business type not found")
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}
	status := models.CompanyStatusActive
	if !isActive {
		status = models.CompanyStatusInactive
	}

	company := models.Company{
		ID:              uuid.New(),
		Code:            input.Code,
		Nama:            input.Nama,
		Email:           input.Email,
		BusinessType:    models.BusinessTypeCode(businessType),
		Address:         input.Address,
		Telp:            input.Telp,
		Website:         input.Website,
		TaxID:           input.TaxID,
		BusinessLicense: input.BusinessLicense,
		Status:          status,
		IsActive:        isActive,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return response.NewErrorResponse("Failed to create company")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	logInstance := applogger.Default()
	err := applogger.AuditCreate(tx, logInstance, "companies", company.ID.String(), "", "", func() error {
		if err := tx.Create(&company).Error; err != nil {
			return err
		}
		if err := ensureCompanyModules(tx, company.ID, businessType); err != nil {
			return err
		}
		return syncCompanyModules(tx, company.ID, businessType, input.ModuleCodes)
	})
	if err != nil {
		tx.Rollback()
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return response.NewErrorResponse("Company code or email already exists")
		}
		return response.NewErrorResponse("Failed to create company")
	}
	if err := tx.Commit().Error; err != nil {
		return response.NewErrorResponse("Failed to create company")
	}
	return response.NewSuccessResponse(buildCompanyResponse(company, s.loadActiveModuleCodes(company.ID)), "")
}

type UpdateCompanyInput struct {
	Nama            *string
	Email           *string
	BusinessType    *string
	ModuleCodes     *[]string
	Address         *string
	Telp            *string
	Logo            *string
	Website         *string
	TaxID           *string
	BusinessLicense *string
	IsActive        *bool
}

func (s *CompanyService) UpdateCompany(id string, input UpdateCompanyInput) response.ApiResponse {
	companyID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Company not found")
	}

	var company models.Company
	if err := s.db.First(&company, "id = ?", companyID).Error; err != nil {
		return response.NewErrorResponse("Company not found")
	}
	nextBusinessType := string(company.BusinessType)
	businessTypeChanged := false
	if input.BusinessType != nil {
		normalizedBusinessType := normalizeCode(*input.BusinessType)
		if normalizedBusinessType != "" && normalizedBusinessType != string(company.BusinessType) {
			if err := validateBusinessType(s.db, normalizedBusinessType); err != nil {
				return response.NewErrorResponse("Business type not found")
			}
			hasDependentData, err := s.companyHasDependentData(s.db, companyID)
			if err != nil {
				return response.NewErrorResponse("Failed to update company")
			}
			if hasDependentData {
				return response.NewErrorResponse("Business type cannot be changed because company still has dependent data")
			}
			nextBusinessType = normalizedBusinessType
			businessTypeChanged = true
		}
	}

	updates := map[string]interface{}{}
	if input.Nama != nil {
		updates["nama"] = *input.Nama
	}
	if input.Email != nil {
		updates["email"] = *input.Email
	}
	if input.Address != nil {
		updates["address"] = input.Address
	}
	if input.Telp != nil {
		updates["telp"] = input.Telp
	}
	if input.Logo != nil {
		updates["logo"] = input.Logo
	}
	if input.Website != nil {
		updates["website"] = input.Website
	}
	if input.TaxID != nil {
		updates["tax_id"] = input.TaxID
	}
	if input.BusinessLicense != nil {
		updates["business_license"] = input.BusinessLicense
	}
	if businessTypeChanged {
		updates["business_type"] = nextBusinessType
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
		if *input.IsActive {
			updates["status"] = "active"
		} else {
			updates["status"] = "inactive"
		}
	}
	shouldUpdateCompany := len(updates) > 0
	if !shouldUpdateCompany && input.ModuleCodes == nil {
		return response.NewErrorResponse("No fields to update")
	}
	if shouldUpdateCompany {
		updates["updated_at"] = time.Now()
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return response.NewErrorResponse("Failed to update company")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	logInstance := applogger.Default()
	err = applogger.AuditUpdate(tx, logInstance, "companies", companyID.String(), "", "", func() error {
		if shouldUpdateCompany {
			res := tx.Model(&models.Company{}).Where("id = ?", companyID).Updates(updates)
			if res.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
			if res.Error != nil {
				return res.Error
			}
		}
		if businessTypeChanged {
			if err := ensureCompanyModules(tx, companyID, nextBusinessType); err != nil {
				return err
			}
		}
		if input.ModuleCodes != nil {
			return syncCompanyModules(tx, companyID, nextBusinessType, *input.ModuleCodes)
		}
		if businessTypeChanged {
			return syncCompanyModules(tx, companyID, nextBusinessType, nil)
		}
		return nil
	})
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewErrorResponse("Company not found")
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return response.NewErrorResponse("Company email already exists")
		}
		return response.NewErrorResponse("Failed to update company")
	}
	if err := tx.Commit().Error; err != nil {
		return response.NewErrorResponse("Failed to update company")
	}

	if err := s.db.First(&company, "id = ?", companyID).Error; err != nil {
		return response.NewErrorResponse("Company not found")
	}
	return response.NewSuccessResponse(buildCompanyResponse(company, s.loadActiveModuleCodes(company.ID)), "")
}

func (s *CompanyService) DeleteCompany(id string) response.ApiResponse {
	companyID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Company not found")
	}

	hasDependentData, err := s.companyHasDependentData(s.db, companyID)
	if err != nil {
		return response.NewErrorResponse("Failed to delete company")
	}
	if hasDependentData {
		return response.NewErrorResponse("Cannot delete company: Company has dependent data (warehouses, users, products, customers, suppliers, sales, purchase orders)")
	}

	var company models.Company
	if err := s.db.First(&company, "id = ?", companyID).Error; err != nil {
		return response.NewErrorResponse("Company not found")
	}
	logInstance := applogger.Default()
	if err := applogger.AuditDelete(s.db, logInstance, "companies", companyID.String(), "", "", func() error {
		return s.db.Delete(&models.Company{}, "id = ?", companyID).Error
	}); err != nil {
		return response.NewErrorResponse("Failed to delete company")
	}
	return response.NewSuccessResponse(nil, "Company deleted successfully")
}

func (s *CompanyService) UploadCompanyLogo(id string, logoPath string) response.ApiResponse {
	companyID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Company not found")
	}
	res := s.db.Model(&models.Company{}).Where("id = ?", companyID).
		Updates(map[string]interface{}{"logo": logoPath, "updated_at": time.Now()})
	if res.Error != nil {
		return response.NewErrorResponse("Failed to update company")
	}
	if res.RowsAffected == 0 {
		return response.NewErrorResponse("Company not found")
	}
	var company models.Company
	if err := s.db.First(&company, "id = ?", companyID).Error; err != nil {
		return response.NewErrorResponse("Company not found")
	}
	return response.NewSuccessResponse(buildCompanyResponse(company, s.loadActiveModuleCodes(company.ID)), "")
}
