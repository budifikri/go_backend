package database

import (
	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SeedSuperUser(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	// Only seed superuser if no superuser exists
	var superuserCount int64
	db.Model(&models.User{}).Where("role = ?", "superuser").Count(&superuserCount)
	if superuserCount > 0 {
		return nil
	}

	// Find or create company for superuser
	var companyID uuid.UUID
	var companyCount int64
	db.Model(&models.Company{}).Count(&companyCount)

	if companyCount == 0 {
		company := models.Company{
			Code:         "company_code",
			Nama:         "company_name",
			Email:        "company_email@gmail.com",
			BusinessType: models.BusinessTypeRetail,
			IsActive:     true,
		}
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&company).Error; err != nil {
			return err
		}
		companyID = company.ID

		// Ensure company modules for retail business type
		if err := ensureCompanyModules(db, companyID, string(models.BusinessTypeRetail)); err != nil {
			return err
		}
	} else {
		var first models.Company
		if err := db.First(&first).Error; err != nil {
			return err
		}
		companyID = first.ID
	}

	hashedPassword, err := utils.HashPassword("budifikri")
	if err != nil {
		return err
	}

	user := models.User{
		Username:  "superuser",
		Password:  hashedPassword,
		FullName:  "Super User",
		Email:     "superuser@system.local",
		Role:      models.RoleSuperuser,
		IsActive:  true,
		CompanyID: companyID,
	}

	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&user).Error
}

// ensureCompanyModules is a local reference to the function in services package.
// We duplicate the minimal logic here to avoid circular imports.
func ensureCompanyModules(tx *gorm.DB, companyID uuid.UUID, businessType string) error {
	var packages []models.ModulePackage
	if err := tx.Where("business_type = ? AND is_active = ?", businessType, true).Find(&packages).Error; err != nil {
		return err
	}

	modules := make([]models.CompanyModule, 0, len(packages))
	for _, pkg := range packages {
		modules = append(modules, models.CompanyModule{
			CompanyID:  companyID,
			ModuleCode: pkg.Code,
			IsActive:   pkg.IsDefault,
		})
	}

	if len(modules) > 0 {
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&modules).Error
	}
	return nil
}
