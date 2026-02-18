package repository

import (
	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
)

type PromotionRepository struct {
	db *gorm.DB
}

func NewPromotionRepository(db *gorm.DB) *PromotionRepository {
	return &PromotionRepository{db: db}
}

func (r *PromotionRepository) FindPromotions(isActive *bool, promoType *string, scope *string, limit, offset int) ([]models.Promotion, error) {
	query := r.db.Table("promotions")
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}
	if promoType != nil && *promoType != "" {
		query = query.Where("promotion_type = ?", *promoType)
	}
	if scope != nil && *scope != "" {
		query = query.Where("scope = ?", *scope)
	}
	var promos []models.Promotion
	if err := query.Order("start_date DESC").Limit(limit).Offset(offset).Find(&promos).Error; err != nil {
		return nil, err
	}
	return promos, nil
}

func (r *PromotionRepository) GetPromotionByID(id uuid.UUID) (*models.Promotion, error) {
	var promo models.Promotion
	if err := r.db.First(&promo, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &promo, nil
}
