package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/pos-retail/go_backend/internal/models"
)

type PaketRepository struct {
	db *gorm.DB
}

func NewPaketRepository(db *gorm.DB) *PaketRepository {
	return &PaketRepository{db: db}
}

// Create paket with details in transaction
func (r *PaketRepository) Create(paket *models.Paket, details []models.DetailPaket) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(paket).Error; err != nil {
			return err
		}
		for i := range details {
			details[i].IDPaket = paket.ID
			if err := tx.Create(&details[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetByID returns paket with preloaded details and produk info
func (r *PaketRepository) GetByID(id uuid.UUID, companyID string) (*models.Paket, error) {
	var paket models.Paket
	query := r.db.Preload("Details.Produk").Where("id = ?", id)
	if companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}
	err := query.First(&paket).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("paket not found")
		}
		return nil, err
	}
	return &paket, nil
}

// GetAll returns list of paket with pagination and filters
func (r *PaketRepository) GetAll(companyID string, filters map[string]interface{}, limit, offset int) ([]models.Paket, int64, error) {
	var pakets []models.Paket
	var total int64

	query := r.db.Model(&models.Paket{})
	if companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}

	if search, ok := filters["search"].(string); ok && search != "" {
		query = query.Where("kodepaket ILIKE ? OR nm_paket ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if active, ok := filters["active"]; ok {
		query = query.Where("is_active = ?", active)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Order("nm_paket ASC").Find(&pakets).Error; err != nil {
		return nil, 0, err
	}

	return pakets, total, nil
}

// Update paket and replace details in transaction
func (r *PaketRepository) Update(paket *models.Paket, details []models.DetailPaket) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update header
		if err := tx.Save(paket).Error; err != nil {
			return err
		}
		// Delete old details
		if err := tx.Where("id_paket = ?", paket.ID).Delete(&models.DetailPaket{}).Error; err != nil {
			return err
		}
		// Insert new details
		for i := range details {
			details[i].IDPaket = paket.ID
			if err := tx.Create(&details[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Delete paket (details will be cascade deleted)
func (r *PaketRepository) Delete(id uuid.UUID, companyID string) error {
	result := r.db.Where("id = ?", id)
	if companyID != "" {
		result = result.Where("company_id = ?", companyID)
	}
	result = result.Delete(&models.Paket{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("paket not found")
	}
	return nil
}

// CalculateTotalHarga returns sum of harga_jual from all products in the package
func (r *PaketRepository) CalculateTotalHarga(idPaket uuid.UUID) (float64, error) {
	var total float64
	err := r.db.Raw(`
		SELECT COALESCE(SUM(p.retail_price), 0) 
		FROM detail_paket dp 
		JOIN products p ON dp.id_produk = p.id 
		WHERE dp.id_paket = ?
	`, idPaket).Scan(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

// CheckProdukExists checks if a product exists and is active
func (r *PaketRepository) CheckProdukExists(produkID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Product{}).Where("id = ? AND is_active = ?", produkID, true).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
