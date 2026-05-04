package database

import (
	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SeedModuleDefaults(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	businessTypes := []models.BusinessType{
		{
			Code:      models.BusinessTypeRetail,
			Name:      "Retail",
			IsActive:  true,
			IsDefault: true,
			IsSystem:  true,
			SortOrder: 1,
		},
		{
			Code:      models.BusinessTypeClinic,
			Name:      "Klinik",
			IsActive:  true,
			IsDefault: true,
			IsSystem:  true,
			SortOrder: 2,
		},
	}
	if err := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoNothing: true}).Create(&businessTypes).Error; err != nil {
		return err
	}
	packages := []models.ModulePackage{
		{BusinessType: models.BusinessTypeRetail, Code: "retail_basic", Name: "Retail Basic", Description: "Default retail modules", IsActive: true, IsDefault: true, IsSystem: true, SortOrder: 1},
		{BusinessType: models.BusinessTypeRetail, Code: "retail_advanced", Name: "Retail Advanced", Description: "Optional retail advanced modules", IsActive: true, IsDefault: false, IsSystem: true, SortOrder: 2},
		{BusinessType: models.BusinessTypeClinic, Code: "clinic_core", Name: "Clinic Core", Description: "Default clinic modules", IsActive: true, IsDefault: true, IsSystem: true, SortOrder: 1},
		{BusinessType: models.BusinessTypeClinic, Code: "clinic_advanced", Name: "Clinic Advanced", Description: "Optional clinic advanced modules", IsActive: true, IsDefault: false, IsSystem: true, SortOrder: 2},
	}
	return db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoNothing: true}).Create(&packages).Error
}
