package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
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

func (s *CompanyService) GetCompanies() response.ApiResponse {
	var companies []models.Company
	if err := s.db.Order("nama").Find(&companies).Error; err != nil {
		return response.NewErrorResponse("Failed to get companies")
	}
	return response.NewSuccessResponse(companies, "")
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
	return response.NewSuccessResponse(company, "")
}

func (s *CompanyService) GetCompanyByUserCompanyID(companyID string) response.ApiResponse {
	// In TS they join users -> companies by userId. In Go JWT already contains companyId.
	return s.GetCompanyByID(companyID)
}

type CreateCompanyInput struct {
	Code            string
	Nama            string
	Email           string
	Address         *string
	Telp            *string
	Website         *string
	TaxID           *string
	BusinessLicense *string
	IsActive        *bool
}

func (s *CompanyService) CreateCompany(input CreateCompanyInput) response.ApiResponse {
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

	if err := s.db.Create(&company).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return response.NewErrorResponse("Company code or email already exists")
		}
		return response.NewErrorResponse("Failed to create company")
	}
	return response.NewSuccessResponse(company, "")
}

type UpdateCompanyInput struct {
	Nama            *string
	Email           *string
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
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
		if *input.IsActive {
			updates["status"] = "active"
		} else {
			updates["status"] = "inactive"
		}
	}

	if len(updates) == 0 {
		return response.NewErrorResponse("No fields to update")
	}
	updates["updated_at"] = time.Now()

	res := s.db.Model(&models.Company{}).Where("id = ?", companyID).Updates(updates)
	if res.Error != nil {
		var pgErr *pgconn.PgError
		if errors.As(res.Error, &pgErr) && pgErr.Code == "23505" {
			return response.NewErrorResponse("Company email already exists")
		}
		return response.NewErrorResponse("Failed to update company")
	}
	if res.RowsAffected == 0 {
		return response.NewErrorResponse("Company not found")
	}

	var company models.Company
	if err := s.db.First(&company, "id = ?", companyID).Error; err != nil {
		return response.NewErrorResponse("Company not found")
	}
	return response.NewSuccessResponse(company, "")
}

func (s *CompanyService) DeleteCompany(id string) response.ApiResponse {
	companyID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Company not found")
	}

	// Check dependent data.
	var dep struct {
		HasDependentData bool `gorm:"column:has_dependent_data"`
	}
	if err := s.db.Raw(`
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
	`, companyID, companyID, companyID, companyID, companyID, companyID, companyID).Scan(&dep).Error; err != nil {
		return response.NewErrorResponse("Failed to delete company")
	}
	if dep.HasDependentData {
		return response.NewErrorResponse("Cannot delete company: Company has dependent data (warehouses, users, products, customers, suppliers, sales, purchase orders)")
	}

	var company models.Company
	if err := s.db.First(&company, "id = ?", companyID).Error; err != nil {
		return response.NewErrorResponse("Company not found")
	}
	if err := s.db.Delete(&models.Company{}, "id = ?", companyID).Error; err != nil {
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
	return response.NewSuccessResponse(company, "")
}
