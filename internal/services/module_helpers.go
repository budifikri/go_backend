package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
)

func validateBusinessType(tx *gorm.DB, businessType string) error {
	businessType = normalizeCode(businessType)
	var count int64
	if err := tx.Model(&models.BusinessType{}).Where("code = ? AND is_active = ?", businessType, true).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("business type not found")
	}
	return nil
}

func ensureCompanyModules(tx *gorm.DB, companyID uuid.UUID, businessType string) error {
	var packages []models.ModulePackage
	if err := tx.Where("business_type = ? AND is_active = ?", normalizeCode(businessType), true).Find(&packages).Error; err != nil {
		return err
	}
	now := time.Now()
	for _, pkg := range packages {
		var existing models.CompanyModule
		err := tx.Where("company_id = ? AND module_code = ?", companyID, pkg.Code).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			item := models.CompanyModule{
				ID:         uuid.New(),
				CompanyID:  companyID,
				ModuleCode: pkg.Code,
				IsActive:   pkg.IsDefault,
				CreatedAt:  now,
				UpdatedAt:  now,
			}
			if pkg.IsDefault {
				item.ActivatedAt = &now
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}
