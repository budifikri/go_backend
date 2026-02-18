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

func (r *CategoryRepository) FindAll(companyID *uuid.UUID) ([]models.Category, error) {
	var categories []models.Category
	query := r.db.Where("is_active = ?", true).Order("name ASC")

	if companyID != nil {
		query = query.Where("company_id = ? OR company_id IS NULL", companyID)
	}

	if err := query.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
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

func (r *UnitRepository) FindAll() ([]models.Unit, error) {
	var units []models.Unit
	if err := r.db.Where("is_active = ?", true).Order("name ASC").Find(&units).Error; err != nil {
		return nil, err
	}
	return units, nil
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
