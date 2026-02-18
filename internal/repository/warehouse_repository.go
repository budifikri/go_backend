package repository

import (
	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
)

type WarehouseRepository struct {
	db *gorm.DB
}

func NewWarehouseRepository(db *gorm.DB) *WarehouseRepository {
	return &WarehouseRepository{db: db}
}

func (r *WarehouseRepository) FindAll(filters map[string]interface{}) ([]models.Warehouse, error) {
	var warehouses []models.Warehouse
	query := r.db.Where("status = ?", "active")

	if companyID, ok := filters["company_id"].(string); ok && companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}

	if err := query.Order("name ASC").Find(&warehouses).Error; err != nil {
		return nil, err
	}
	return warehouses, nil
}

func (r *WarehouseRepository) FindByID(id uuid.UUID) (*models.Warehouse, error) {
	var warehouse models.Warehouse
	if err := r.db.First(&warehouse, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &warehouse, nil
}

func (r *WarehouseRepository) FindByCode(code string) (*models.Warehouse, error) {
	var warehouse models.Warehouse
	if err := r.db.Where("code = ?", code).First(&warehouse).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &warehouse, nil
}

func (r *WarehouseRepository) Create(warehouse *models.Warehouse) error {
	return r.db.Create(warehouse).Error
}

func (r *WarehouseRepository) Update(warehouse *models.Warehouse) error {
	return r.db.Save(warehouse).Error
}

func (r *WarehouseRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Warehouse{}, "id = ?", id).Error
}
