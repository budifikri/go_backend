package services

import (
	"time"

	"github.com/google/uuid"
	applogger "github.com/pos-retail/go_backend/internal/logger"
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

func (s *PriceTierService) GetPriceTiers(productID *string, search string, limit, offset int) response.PaginatedResponse {
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
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("pt.tier_name ILIKE ? OR p.name ILIKE ? OR p.sku ILIKE ?", like, like, like)
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

func (s *PriceTierService) GetPriceTierReportByProduct(companyID *string, search, scope string, categoryID *string, limit, offset int) response.PaginatedResponse {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	if scope == "" {
		scope = "all"
	}

	rankedQuery := s.db.Table("price_tiers pt").
		Select(`
			pt.product_id,
			pt.min_quantity,
			pt.unit_price,
			ROW_NUMBER() OVER (PARTITION BY pt.product_id ORDER BY pt.min_quantity ASC, pt.created_at ASC) AS rn
		`).
		Where("pt.is_active = true")

	tierSummaryQuery := s.db.Table("(?) AS ranked", rankedQuery).
		Select(`
			product_id,
			MAX(CASE WHEN rn = 1 THEN unit_price END) AS grosir_1_price,
			MAX(CASE WHEN rn = 1 THEN min_quantity END) AS grosir_1_qty,
			MAX(CASE WHEN rn = 2 THEN unit_price END) AS grosir_2_price,
			MAX(CASE WHEN rn = 2 THEN min_quantity END) AS grosir_2_qty,
			MAX(CASE WHEN rn = 3 THEN unit_price END) AS grosir_3_price,
			MAX(CASE WHEN rn = 3 THEN min_quantity END) AS grosir_3_qty,
			TRUE AS has_tier
		`).
		Group("product_id")

	query := s.db.Table("products p").
		Select(`
			p.id AS product_id,
			p.sku,
			p.name AS product_name,
			COALESCE(u.name, u.code, '-') AS unit_name,
			COALESCE(c.name, '-') AS category_name,
			p.category_id,
			p.retail_price,
			COALESCE(t.grosir_1_price, 0) AS grosir_1_price,
			COALESCE(t.grosir_1_qty, 0) AS grosir_1_qty,
			COALESCE(t.grosir_2_price, 0) AS grosir_2_price,
			COALESCE(t.grosir_2_qty, 0) AS grosir_2_qty,
			COALESCE(t.grosir_3_price, 0) AS grosir_3_price,
			COALESCE(t.grosir_3_qty, 0) AS grosir_3_qty,
			COALESCE(t.has_tier, FALSE) AS has_tier
		`).
		Joins("LEFT JOIN units_of_measure u ON u.id = p.unit_id").
		Joins("LEFT JOIN categories c ON c.id = p.category_id").
		Joins("LEFT JOIN (?) t ON t.product_id = p.id", tierSummaryQuery).
		Where("p.is_active = true")

	if companyID != nil && *companyID != "" {
		query = query.Where("p.company_id = ?", *companyID)
	}
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("p.name ILIKE ? OR p.sku ILIKE ?", like, like)
	}
	if categoryID != nil && *categoryID != "" {
		query = query.Where("p.category_id = ?", *categoryID)
	}
	if scope == "grosir" {
		query = query.Where("COALESCE(t.has_tier, FALSE) = TRUE")
	}

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}

	var rows []map[string]interface{}
	err := query.
		Order("product_name ASC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error
	if err != nil {
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
	if err := applogger.AuditCreate(s.db, applogger.Default(), "price_tiers", pt.ID.String(), "", "", func() error {
		return s.db.Create(&pt).Error
	}); err != nil {
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

	err = applogger.AuditUpdate(s.db, applogger.Default(), "price_tiers", tid.String(), "", "", func() error {
		res := s.db.Table("price_tiers").Where("id = ?", tid).Updates(updates)
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return res.Error
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.NewErrorResponse("Price tier not found")
		}
		return response.NewErrorResponse("Price tier not found")
	}
	return s.GetPriceTierByID(id)
}

func (s *PriceTierService) DeletePriceTier(id string) response.ApiResponse {
	tid, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Price tier not found")
	}
	err = applogger.AuditDelete(s.db, applogger.Default(), "price_tiers", tid.String(), "", "", func() error {
		res := s.db.Exec("DELETE FROM price_tiers WHERE id = ?", tid)
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return res.Error
	})
	if err != nil {
		return response.NewErrorResponse("Price tier not found")
	}
	return response.NewSuccessResponse(nil, "Price tier deleted successfully")
}

type PriceTierInput struct {
	TierName    string  `json:"tier_name"`
	MinQuantity int     `json:"min_quantity"`
	MaxQuantity *int    `json:"max_quantity,omitempty"`
	UnitPrice   float64 `json:"unit_price"`
}

func (s *PriceTierService) SaveProductPriceTiers(productID string, tiers []PriceTierInput) response.ApiResponse {
	pid, err := uuid.Parse(productID)
	if err != nil {
		return response.NewErrorResponse("Invalid product ID")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("product_id = ?", pid).Delete(&models.PriceTier{}).Error; err != nil {
			return err
		}

		var createdTiers []models.PriceTier
		for _, tier := range tiers {
			pt := models.PriceTier{
				ID:          uuid.New(),
				ProductID:   pid,
				TierName:    tier.TierName,
				MinQuantity: tier.MinQuantity,
				MaxQuantity: tier.MaxQuantity,
				UnitPrice:   tier.UnitPrice,
				IsActive:    true,
				CreatedAt:   time.Now(),
			}
			if err := tx.Create(&pt).Error; err != nil {
				return err
			}
			createdTiers = append(createdTiers, pt)
		}

		return nil
	})

	if err != nil {
		return response.NewErrorResponse("Failed to save price tiers")
	}
	return response.NewSuccessResponse(nil, "Price tiers saved successfully")
}
