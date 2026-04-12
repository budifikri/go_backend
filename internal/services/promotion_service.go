package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	applogger "github.com/pos-retail/go_backend/internal/logger"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/repository"
	"github.com/pos-retail/go_backend/internal/types/response"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PromotionService struct {
	db        *gorm.DB
	promoRepo *repository.PromotionRepository
}

func NewPromotionService(db *gorm.DB, promoRepo *repository.PromotionRepository) *PromotionService {
	return &PromotionService{db: db, promoRepo: promoRepo}
}

func (s *PromotionService) GetPromotions(isActive *bool, promoType *string, scope *string, search string, limit, offset int) response.PaginatedResponse {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	promos, total, err := s.promoRepo.FindPromotions(isActive, promoType, scope, search, limit, offset)
	if err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}
	return response.NewPaginatedResponse(promos, total, limit, offset)
}

func (s *PromotionService) GetPromotionByID(id string) response.ApiResponse {
	pid, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Promotion not found")
	}
	promo, err := s.promoRepo.GetPromotionByID(pid)
	if err != nil {
		return response.NewErrorResponse("Promotion not found")
	}
	if promo == nil {
		return response.NewErrorResponse("Promotion not found")
	}

	var products []struct {
		ProductID uuid.UUID `gorm:"column:product_id"`
	}
	var categories []struct {
		CategoryID uuid.UUID `gorm:"column:category_id"`
	}
	var customers []struct {
		CustomerID uuid.UUID `gorm:"column:customer_id"`
	}
	_ = s.db.Table("promotion_products").Select("product_id").Where("promotion_id = ?", pid).Scan(&products).Error
	_ = s.db.Table("promotion_categories").Select("category_id").Where("promotion_id = ?", pid).Scan(&categories).Error
	_ = s.db.Table("promotion_customers").Select("customer_id").Where("promotion_id = ?", pid).Scan(&customers).Error

	associatedProducts := make([]uuid.UUID, 0, len(products))
	for _, p := range products {
		associatedProducts = append(associatedProducts, p.ProductID)
	}
	associatedCategories := make([]uuid.UUID, 0, len(categories))
	for _, c := range categories {
		associatedCategories = append(associatedCategories, c.CategoryID)
	}
	associatedCustomers := make([]uuid.UUID, 0, len(customers))
	for _, c := range customers {
		associatedCustomers = append(associatedCustomers, c.CustomerID)
	}

	data := map[string]interface{}{}
	data["id"] = promo.ID
	data["code"] = promo.Code
	data["name"] = promo.Name
	data["description"] = promo.Description
	data["promotion_type"] = promo.PromotionType
	data["scope"] = promo.Scope
	data["discount_value"] = promo.DiscountValue
	data["min_purchase_amount"] = promo.MinPurchaseAmount
	data["max_discount_amount"] = promo.MaxDiscountAmount
	data["buy_quantity"] = promo.BuyQuantity
	data["get_quantity"] = promo.GetQuantity
	data["start_date"] = promo.StartDate
	data["start_time"] = promo.StartTime
	data["end_date"] = promo.EndDate
	data["end_time"] = promo.EndTime
	data["is_active"] = promo.IsActive
	data["usage_limit"] = promo.UsageLimit
	data["usage_count"] = promo.UsageCount
	data["created_at"] = promo.CreatedAt
	data["associated_products"] = associatedProducts
	data["associated_categories"] = associatedCategories
	data["associated_customers"] = associatedCustomers

	return response.NewSuccessResponse(data, "")
}

type CreatePromotionInput struct {
	Code              string
	Name              string
	Description       *string
	PromotionType     string
	Scope             string
	DiscountValue     float64
	MinPurchaseAmount *float64
	MaxDiscountAmount *float64
	BuyQuantity       *int
	GetQuantity       *int
	StartDate         *time.Time
	StartTime         *time.Time
	EndDate           *time.Time
	EndTime           *time.Time
	UsageLimit        *int
	ProductIDs        []string
	CategoryIDs       []string
	CustomerIDs       []string
}

func (s *PromotionService) CreatePromotion(input CreatePromotionInput) response.ApiResponse {
	var created models.Promotion
	fmt.Printf("[DEBUG] CreatePromotion: Code=%s, Name=%s\n", input.Code, input.Name)
	err := s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		promo := models.Promotion{
			ID:            uuid.New(),
			Code:          input.Code,
			Name:          input.Name,
			Description:   "",
			PromotionType: input.PromotionType,
			Scope:         input.Scope,
			DiscountValue: input.DiscountValue,
			StartDate:     now,
			EndDate:       now.AddDate(0, 1, 0),
			StartTime:     nil,
			EndTime:       nil,
			IsActive:      true,
			UsageCount:    0,
			CreatedAt:     now,
		}
		if input.StartDate != nil && !input.StartDate.IsZero() {
			promo.StartDate = *input.StartDate
		}
		if input.EndDate != nil && !input.EndDate.IsZero() {
			promo.EndDate = *input.EndDate
		}
		if input.Description != nil {
			promo.Description = *input.Description
		}
		if input.MinPurchaseAmount != nil {
			promo.MinPurchaseAmount = *input.MinPurchaseAmount
		}
		if input.MaxDiscountAmount != nil {
			promo.MaxDiscountAmount = input.MaxDiscountAmount
		}
		if input.BuyQuantity != nil && *input.BuyQuantity > 0 {
			promo.BuyQuantity = *input.BuyQuantity
		}
		if input.GetQuantity != nil && *input.GetQuantity > 0 {
			promo.GetQuantity = *input.GetQuantity
		}
		if input.StartTime != nil {
			promo.StartTime = input.StartTime
		}
		if input.EndTime != nil {
			promo.EndTime = input.EndTime
		}
		promo.UsageLimit = input.UsageLimit

		if err := tx.Clauses(clause.Returning{}).Create(&promo).Scan(&created).Error; err != nil {
			return err
		}

		for _, id := range input.ProductIDs {
			pid, err := uuid.Parse(id)
			if err != nil {
				continue
			}
			pp := models.PromotionProduct{ID: uuid.New(), PromotionID: created.ID, ProductID: pid}
			_ = tx.Create(&pp).Error
		}
		for _, id := range input.CategoryIDs {
			cid, err := uuid.Parse(id)
			if err != nil {
				continue
			}
			pc := models.PromotionCategory{ID: uuid.New(), PromotionID: created.ID, CategoryID: cid}
			_ = tx.Create(&pc).Error
		}
		for _, id := range input.CustomerIDs {
			cid, err := uuid.Parse(id)
			if err != nil {
				continue
			}
			pc := models.PromotionCustomer{ID: uuid.New(), PromotionID: created.ID, CustomerID: cid}
			_ = tx.Create(&pc).Error
		}
		return nil
	})
	if err != nil {
		if l := applogger.Default(); l != nil {
			l.LogError(applogger.ActionCreate, "promotions", "", "", created.ID.String(), err)
		}
		return response.NewErrorResponse("Failed to create promotion")
	}
	if l := applogger.Default(); l != nil {
		l.Log(applogger.ActionCreate, "promotions", "", "", created.ID.String(), nil, created)
	}

	return response.NewSuccessResponse(map[string]interface{}{
		"id":                  created.ID,
		"code":                created.Code,
		"name":                created.Name,
		"description":         created.Description,
		"promotion_type":      created.PromotionType,
		"scope":               created.Scope,
		"discount_value":      created.DiscountValue,
		"min_purchase_amount": created.MinPurchaseAmount,
		"max_discount_amount": created.MaxDiscountAmount,
		"buy_quantity":        created.BuyQuantity,
		"get_quantity":        created.GetQuantity,
		"start_date":          created.StartDate,
		"start_time":          created.StartTime,
		"end_date":            created.EndDate,
		"end_time":            created.EndTime,
		"is_active":           created.IsActive,
		"usage_limit":         created.UsageLimit,
		"usage_count":         created.UsageCount,
	}, "")
}

type UpdatePromotionInput struct {
	Code              *string
	Name              *string
	Description       *string
	PromotionType     *string
	Scope             *string
	DiscountValue     *float64
	MinPurchaseAmount *float64
	MaxDiscountAmount *float64
	BuyQuantity       *int
	GetQuantity       *int
	StartDate         *time.Time
	StartTime         *time.Time
	EndDate           *time.Time
	EndTime           *time.Time
	IsActive          *bool
	UsageLimit        *int
	ProductIDs        *[]string
	CategoryIDs       *[]string
	CustomerIDs       *[]string
}

func (s *PromotionService) UpdatePromotion(id string, input UpdatePromotionInput) response.ApiResponse {
	pid, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Promotion not found")
	}

	var oldPromo map[string]interface{}
	_ = s.db.Table("promotions").Where("id = ?", pid).Take(&oldPromo).Error

	err = s.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{}
		if input.Code != nil {
			updates["code"] = *input.Code
		}
		if input.Name != nil {
			updates["name"] = *input.Name
		}
		if input.Description != nil {
			updates["description"] = *input.Description
		}
		if input.DiscountValue != nil {
			updates["discount_value"] = *input.DiscountValue
		}
		if input.MinPurchaseAmount != nil {
			updates["min_purchase_amount"] = *input.MinPurchaseAmount
		}
		if input.MaxDiscountAmount != nil {
			updates["max_discount_amount"] = input.MaxDiscountAmount
		}
		if input.PromotionType != nil {
			updates["promotion_type"] = *input.PromotionType
		}
		if input.Scope != nil {
			updates["scope"] = *input.Scope
		}
		if input.BuyQuantity != nil {
			updates["buy_quantity"] = *input.BuyQuantity
		}
		if input.GetQuantity != nil {
			updates["get_quantity"] = *input.GetQuantity
		}
		if input.StartDate != nil {
			updates["start_date"] = *input.StartDate
		}
		if input.StartTime != nil {
			updates["start_time"] = *input.StartTime
		}
		if input.EndDate != nil {
			updates["end_date"] = *input.EndDate
		}
		if input.EndTime != nil {
			updates["end_time"] = *input.EndTime
		}
		if input.IsActive != nil {
			updates["is_active"] = *input.IsActive
		}
		if input.UsageLimit != nil {
			updates["usage_limit"] = *input.UsageLimit
		}
		if len(updates) == 0 && input.ProductIDs == nil && input.CategoryIDs == nil && input.CustomerIDs == nil {
			return fmt.Errorf("No fields to update")
		}

		if len(updates) > 0 {
			res := tx.Table("promotions").Where("id = ?", pid).Updates(updates)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
		}

		if input.ProductIDs != nil {
			_ = tx.Exec("DELETE FROM promotion_products WHERE promotion_id = ?", pid).Error
			for _, id := range *input.ProductIDs {
				v, err := uuid.Parse(id)
				if err != nil {
					continue
				}
				_ = tx.Create(&models.PromotionProduct{ID: uuid.New(), PromotionID: pid, ProductID: v}).Error
			}
		}
		if input.CategoryIDs != nil {
			_ = tx.Exec("DELETE FROM promotion_categories WHERE promotion_id = ?", pid).Error
			for _, id := range *input.CategoryIDs {
				v, err := uuid.Parse(id)
				if err != nil {
					continue
				}
				_ = tx.Create(&models.PromotionCategory{ID: uuid.New(), PromotionID: pid, CategoryID: v}).Error
			}
		}
		if input.CustomerIDs != nil {
			_ = tx.Exec("DELETE FROM promotion_customers WHERE promotion_id = ?", pid).Error
			for _, id := range *input.CustomerIDs {
				v, err := uuid.Parse(id)
				if err != nil {
					continue
				}
				_ = tx.Create(&models.PromotionCustomer{ID: uuid.New(), PromotionID: pid, CustomerID: v}).Error
			}
		}
		return nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.NewErrorResponse("Promotion not found")
		}
		if l := applogger.Default(); l != nil {
			l.LogError(applogger.ActionUpdate, "promotions", "", "", pid.String(), err)
		}
		return response.NewErrorResponse(err.Error())
	}
	var newPromo map[string]interface{}
	_ = s.db.Table("promotions").Where("id = ?", pid).Take(&newPromo).Error
	if l := applogger.Default(); l != nil {
		l.Log(applogger.ActionUpdate, "promotions", "", "", pid.String(), oldPromo, newPromo)
	}

	return s.GetPromotionByID(id)
}

func (s *PromotionService) DeletePromotion(id string) response.ApiResponse {
	pid, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Promotion not found")
	}
	err = applogger.AuditDelete(s.db, applogger.Default(), "promotions", pid.String(), "", "", func() error {
		res := s.db.Exec("DELETE FROM promotions WHERE id = ?", pid)
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return res.Error
	})
	if err != nil {
		return response.NewErrorResponse("Promotion not found")
	}
	return response.NewSuccessResponse(nil, "Promotion deleted successfully")
}
