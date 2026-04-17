package repository

import (
	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
)

type TelegramRepository struct {
	db *gorm.DB
}

func NewTelegramRepository(db *gorm.DB) *TelegramRepository {
	return &TelegramRepository{db: db}
}

func (r *TelegramRepository) GetConfigByCompany(companyID uuid.UUID) (*models.TelegramConfig, error) {
	var config models.TelegramConfig
	err := r.db.Where("company_id = ?", companyID).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

func (r *TelegramRepository) SaveConfig(config *models.TelegramConfig) (*models.TelegramConfig, error) {
	existing, err := r.GetConfigByCompany(config.CompanyID)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		err = r.db.Create(config).Error
		if err != nil {
			return nil, err
		}
		return config, nil
	}

	config.ID = existing.ID
	config.CreatedAt = existing.CreatedAt
	err = r.db.Save(config).Error
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (r *TelegramRepository) DeleteConfig(companyID uuid.UUID) error {
	return r.db.Where("company_id = ?", companyID).Delete(&models.TelegramConfig{}).Error
}
