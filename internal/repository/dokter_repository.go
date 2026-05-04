package repository

import (
	"errors"

	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
)

type DokterRepository struct {
	db *gorm.DB
}

func NewDokterRepository(db *gorm.DB) *DokterRepository {
	return &DokterRepository{db: db}
}

func (r *DokterRepository) Create(dokter *models.Dokter) error {
	return r.db.Create(dokter).Error
}

func (r *DokterRepository) GetByID(id, companyID string) (*models.Dokter, error) {
	var dokter models.Dokter
	err := r.db.Where("id = ? AND company_id = ?", id, companyID).First(&dokter).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("dokter not found")
		}
		return nil, err
	}
	return &dokter, nil
}

func (r *DokterRepository) GetAll(companyID string, filters map[string]interface{}, limit, offset int) ([]models.Dokter, int64, error) {
	var dokters []models.Dokter
	var total int64

	query := r.db.Model(&models.Dokter{}).Where("company_id = ?", companyID)

	if search, ok := filters["search"].(string); ok && search != "" {
		query = query.Where("nama ILIKE ? OR email ILIKE ? OR no_telp ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if tipe, ok := filters["tipe"].(string); ok && tipe != "" {
		query = query.Where("tipe = ?", tipe)
	}

	if active, ok := filters["active"]; ok {
		query = query.Where("active = ?", active)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Order("nama ASC").Find(&dokters).Error; err != nil {
		return nil, 0, err
	}

	return dokters, total, nil
}

func (r *DokterRepository) Update(dokter *models.Dokter) error {
	return r.db.Save(dokter).Error
}

func (r *DokterRepository) Delete(id, companyID string) error {
	result := r.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&models.Dokter{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("dokter not found")
	}
	return nil
}

func (r *DokterRepository) CheckDependencies(id, companyID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Dokter{}).Where("id = ? AND company_id = ?", id, companyID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
