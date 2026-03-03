package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/types/response"
	"gorm.io/gorm"
)

type PriceTierService struct {
	db *gorm.DB
}

func NewPriceTierService(db *gorm.DB) *PriceTierService {
	return &PriceTierService{db: db}
}

func (s *PriceTierService) GetPriceTiers(productID *string, limit, offset int) response.PaginatedResponse {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var rows []map[string]interface{}
	query := s.db.Table("price_tiers pt").
		Select("pt.*, p.name AS product_name, p.sku, p.retail_price AS base_price").
		Joins("LEFT JOIN products p ON p.id = pt.product_id").
		Where("pt.is_active = true")
	if productID != nil && *productID != "" {
		query = query.Where("pt.product_id = ?", *productID)
	}

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}

	if err := query.Order("pt.min_quantity ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}

	return response.NewPaginatedResponse(rows, total, limit, offset)
}

func (s *PriceTierService) GetPriceTierByID(id string) response.ApiResponse {
	tid, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Price tier not found")
	}
	var row map[string]interface{}
	err = s.db.Table("price_tiers pt").
		Select("pt.*, p.name AS product_name, p.sku").
		Joins("LEFT JOIN products p ON p.id = pt.product_id").
		Where("pt.id = ?", tid).
		Limit(1).
		Find(&row).Error
	if err != nil {
		return response.NewErrorResponse("Price tier not found")
	}
	if len(row) == 0 {
		return response.NewErrorResponse("Price tier not found")
	}
	return response.NewSuccessResponse(row, "")
}

type CreatePriceTierInput struct {
	ProductID   string
	TierName    string
	MinQuantity int
	MaxQuantity *int
	UnitPrice   float64
}

func (s *PriceTierService) CreatePriceTier(input CreatePriceTierInput) response.ApiResponse {
	pid, err := uuid.Parse(input.ProductID)
	if err != nil {
		return response.NewErrorResponse("Failed to create price tier")
	}
	pt := models.PriceTier{
		ID:          uuid.New(),
		ProductID:   pid,
		TierName:    input.TierName,
		MinQuantity: input.MinQuantity,
		MaxQuantity: input.MaxQuantity,
		UnitPrice:   input.UnitPrice,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}
	if err := s.db.Create(&pt).Error; err != nil {
		return response.NewErrorResponse("Failed to create price tier")
	}
	return response.NewSuccessResponse(pt, "")
}

type UpdatePriceTierInput struct {
	TierName    *string
	MinQuantity *int
	MaxQuantity **int
	UnitPrice   *float64
	IsActive    *bool
}

func (s *PriceTierService) UpdatePriceTier(id string, input UpdatePriceTierInput) response.ApiResponse {
	tid, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Price tier not found")
	}
	updates := map[string]interface{}{}
	if input.TierName != nil {
		updates["tier_name"] = *input.TierName
	}
	if input.MinQuantity != nil {
		updates["min_quantity"] = *input.MinQuantity
	}
	if input.MaxQuantity != nil {
		updates["max_quantity"] = *input.MaxQuantity
	}
	if input.UnitPrice != nil {
		updates["unit_price"] = *input.UnitPrice
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}
	if len(updates) == 0 {
		return response.NewErrorResponse("No fields to update")
	}

	res := s.db.Table("price_tiers").Where("id = ?", tid).Updates(updates)
	if res.Error != nil {
		return response.NewErrorResponse("Price tier not found")
	}
	if res.RowsAffected == 0 {
		return response.NewErrorResponse("Price tier not found")
	}
	return s.GetPriceTierByID(id)
}

func (s *PriceTierService) DeletePriceTier(id string) response.ApiResponse {
	tid, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Price tier not found")
	}
	res := s.db.Exec("DELETE FROM price_tiers WHERE id = ?", tid)
	if res.Error != nil {
		return response.NewErrorResponse("Price tier not found")
	}
	if res.RowsAffected == 0 {
		return response.NewErrorResponse("Price tier not found")
	}
	return response.NewSuccessResponse(nil, "Price tier deleted successfully")
}
