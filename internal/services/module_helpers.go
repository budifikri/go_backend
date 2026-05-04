package services

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
)

func normalizeModuleCodes(codes []string) []string {
	if len(codes) == 0 {
		return nil
	}
	unique := map[string]struct{}{}
	result := make([]string, 0, len(codes))
	for _, code := range codes {
		normalized := normalizeCode(code)
		if normalized == "" {
			continue
		}
		if _, exists := unique[normalized]; exists {
			continue
		}
		unique[normalized] = struct{}{}
		result = append(result, normalized)
	}
	sort.Strings(result)
	return result
}

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

func syncCompanyModules(tx *gorm.DB, companyID uuid.UUID, businessType string, requestedCodes []string) error {
	businessType = normalizeCode(businessType)
	requestedCodes = normalizeModuleCodes(requestedCodes)

	var packages []models.ModulePackage
	if err := tx.Where("business_type = ? AND is_active = ?", businessType, true).Order("sort_order, code").Find(&packages).Error; err != nil {
		return err
	}
	allowed := make(map[string]models.ModulePackage, len(packages))
	selected := make(map[string]bool, len(requestedCodes))
	for _, pkg := range packages {
		allowed[pkg.Code] = pkg
	}
	if len(requestedCodes) == 0 {
		for _, pkg := range packages {
			if pkg.IsDefault {
				selected[pkg.Code] = true
			}
		}
	} else {
		for _, code := range requestedCodes {
			if _, ok := allowed[code]; !ok {
				return errors.New("module package does not match company business type")
			}
			selected[code] = true
		}
		for _, pkg := range packages {
			if pkg.IsDefault {
				selected[pkg.Code] = true
			}
		}
	}

	var existing []models.CompanyModule
	if err := tx.Where("company_id = ?", companyID).Find(&existing).Error; err != nil {
		return err
	}
	existingByCode := make(map[string]models.CompanyModule, len(existing))
	for _, item := range existing {
		existingByCode[strings.ToLower(item.ModuleCode)] = item
	}
	for code, item := range existingByCode {
		if _, ok := allowed[code]; ok {
			continue
		}
		if err := tx.Delete(&models.CompanyModule{}, "id = ?", item.ID).Error; err != nil {
			return err
		}
		delete(existingByCode, code)
	}
	now := time.Now()
	for _, pkg := range packages {
		shouldBeActive := selected[pkg.Code]
		if existingItem, ok := existingByCode[pkg.Code]; ok {
			updates := map[string]interface{}{
				"is_active":  shouldBeActive,
				"updated_at": now,
			}
			if shouldBeActive {
				updates["activated_at"] = &now
			} else {
				updates["activated_at"] = nil
				updates["activated_by"] = nil
			}
			if err := tx.Model(&models.CompanyModule{}).Where("id = ?", existingItem.ID).Updates(updates).Error; err != nil {
				return err
			}
			continue
		}

		item := models.CompanyModule{
			ID:         uuid.New(),
			CompanyID:  companyID,
			ModuleCode: pkg.Code,
			IsActive:   shouldBeActive,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if shouldBeActive {
			item.ActivatedAt = &now
		}
		if err := tx.Create(&item).Error; err != nil {
			return err
		}
	}

	return nil
}
