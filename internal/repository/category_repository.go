package repository

import (
	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) FindAll(companyID *uuid.UUID, isActive *bool, search string, limit, offset int) ([]models.Category, int64, error) {
	var categories []models.Category
	var total int64
	query := r.db.Model(&models.Category{}).Order("name ASC")
	// If isActive is nil: include both active and inactive.
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if companyID != nil {
		query = query.Where("company_id = ?", companyID)
	}
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ? OR description ILIKE ?", like, like, like)
	}

	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset >= 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&categories).Error; err != nil {
		return nil, 0, err
	}
	return categories, total, nil
}

func (r *CategoryRepository) FindByID(id uuid.UUID) (*models.Category, error) {
	var category models.Category
	if err := r.db.First(&category, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) FindByCode(code string) (*models.Category, error) {
	var category models.Category
	if err := r.db.Where("code = ?", code).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) Create(category *models.Category) error {
	return r.db.Create(category).Error
}

func (r *CategoryRepository) Update(category *models.Category) error {
	return r.db.Save(category).Error
}

func (r *CategoryRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Category{}, "id = ?", id).Error
}

type UnitRepository struct {
	db *gorm.DB
}

func NewUnitRepository(db *gorm.DB) *UnitRepository {
	return &UnitRepository{db: db}
}

func (r *UnitRepository) FindAll(companyID *uuid.UUID) ([]models.Unit, error) {
	var units []models.Unit
	query := r.db.Model(&models.Unit{}).Order("code ASC")
	if companyID != nil {
		query = query.Where("company_id = ?", companyID)
	}
	if err := query.Where("is_active = ?", true).Find(&units).Error; err != nil {
		return nil, err
	}
	return units, nil
}

func (r *UnitRepository) FindAllWithQuery(companyID *uuid.UUID, search *string, isActive *bool, limit, offset *int) ([]models.Unit, error) {
	var units []models.Unit
	q := r.db.Model(&models.Unit{}).Order("code ASC")
	// If isActive is nil: include both active and inactive.
	if isActive != nil {
		q = q.Where("is_active = ?", *isActive)
	}
	if companyID != nil {
		q = q.Where("company_id = ?", companyID)
	}
	if search != nil && *search != "" {
		like := "%%" + *search + "%%"
		q = q.Where("name ILIKE ? OR description ILIKE ?", like, like)
	}
	if limit != nil {
		q = q.Limit(*limit)
	}
	if offset != nil {
		q = q.Offset(*offset)
	}
	if err := q.Find(&units).Error; err != nil {
		return nil, err
	}
	return units, nil
}

func (r *UnitRepository) Count(companyID *uuid.UUID, search *string, isActive *bool) (int64, error) {
	q := r.db.Model(&models.Unit{})
	if companyID != nil {
		q = q.Where("company_id = ?", companyID)
	}
	if isActive != nil {
		q = q.Where("is_active = ?", *isActive)
	}
	if search != nil && *search != "" {
		like := "%%" + *search + "%%"
		q = q.Where("name ILIKE ? OR description ILIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *UnitRepository) FindByID(id uuid.UUID) (*models.Unit, error) {
	var unit models.Unit
	if err := r.db.First(&unit, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &unit, nil
}

func (r *UnitRepository) FindByCode(code string) (*models.Unit, error) {
	var unit models.Unit
	if err := r.db.Where("code = ?", code).First(&unit).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &unit, nil
}

func (r *UnitRepository) Create(unit *models.Unit) error {
	return r.db.Create(unit).Error
}

func (r *UnitRepository) Update(unit *models.Unit) error {
	return r.db.Save(unit).Error
}

func (r *UnitRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Unit{}, "id = ?", id).Error
}
