package repository

import (
	"log"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) FindAll(filters map[string]interface{}, limit, offset int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	query := r.db.Model(&models.Product{})

	// If is_active is not provided: include both active and inactive.
	if v, ok := filters["is_active"].(bool); ok {
		query = query.Where("is_active = ?", v)
	}
	if categoryID, ok := filters["category_id"].(string); ok && categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if companyID, ok := filters["company_id"].(string); ok && companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}
	if search, ok := filters["search"].(string); ok && search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR sku ILIKE ? OR barcode ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Category").Preload("Unit").Limit(limit).Offset(offset).Order("created_at DESC").Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *ProductRepository) FindByID(id uuid.UUID) (*models.Product, error) {
	var product models.Product
	if err := r.db.Preload("Category").Preload("Unit").First(&product, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) FindBySKU(sku string) (*models.Product, error) {
	var product models.Product
	if err := r.db.Where("sku = ?", sku).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) FindByBarcode(barcode string) (*models.Product, error) {
	var product models.Product
	if err := r.db.Where("barcode = ?", barcode).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *ProductRepository) Update(product *models.Product) error {
	log.Printf("[DEBUG] ProductRepo.Update: id=%s, UnitID=%s", product.ID, product.UnitID)
	err := r.db.Model(product).Select("sku", "barcode", "unit_id", "category_id", "name", "description", "cost_price", "retail_price", "tax_rate", "reorder_point", "is_active", "status", "updated_at").Updates(product).Error
	if err != nil {
		log.Printf("[DEBUG] ProductRepo.Update: error=%v", err)
	}
	return err
}

func (r *ProductRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Product{}, "id = ?", id).Error
}

func (r *ProductRepository) FindPriceTiers(productID uuid.UUID) ([]models.PriceTier, error) {
	var tiers []models.PriceTier
	if err := r.db.Where("product_id = ?", productID).Order("min_quantity ASC").Find(&tiers).Error; err != nil {
		return nil, err
	}
	return tiers, nil
}

func (r *ProductRepository) CreatePriceTier(tier *models.PriceTier) error {
	return r.db.Create(tier).Error
}

func (r *ProductRepository) DeletePriceTier(id uuid.UUID) error {
	return r.db.Delete(&models.PriceTier{}, "id = ?", id).Error
}
