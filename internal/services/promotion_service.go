package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
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

func (s *PromotionService) GetPromotions(isActive *bool, promoType *string, scope *string, limit, offset int) response.ApiResponse {
	promos, err := s.promoRepo.FindPromotions(isActive, promoType, scope, limit, offset)
	if err != nil {
		return response.NewErrorResponse("Failed to get promotions")
	}
	return response.NewSuccessResponse(promos, "")
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
	data["start_date"] = promo.StartDate
	data["end_date"] = promo.EndDate
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
	StartDate         time.Time
	EndDate           time.Time
	UsageLimit        *int
	ProductIDs        []string
	CategoryIDs       []string
	CustomerIDs       []string
}

func (s *PromotionService) CreatePromotion(input CreatePromotionInput) response.ApiResponse {
	var created models.Promotion
	err := s.db.Transaction(func(tx *gorm.DB) error {
		promo := models.Promotion{
			ID:            uuid.New(),
			Code:          input.Code,
			Name:          input.Name,
			Description:   "",
			PromotionType: input.PromotionType,
			Scope:         input.Scope,
			DiscountValue: input.DiscountValue,
			StartDate:     input.StartDate,
			EndDate:       input.EndDate,
			IsActive:      true,
			UsageCount:    0,
			CreatedAt:     time.Now(),
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
		return response.NewErrorResponse("Failed to create promotion")
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
		"start_date":          created.StartDate,
		"end_date":            created.EndDate,
		"is_active":           created.IsActive,
		"usage_limit":         created.UsageLimit,
		"usage_count":         created.UsageCount,
	}, "")
}

type UpdatePromotionInput struct {
	Code              *string
	Name              *string
	Description       *string
	DiscountValue     *float64
	MinPurchaseAmount *float64
	MaxDiscountAmount *float64
	StartDate         *time.Time
	EndDate           *time.Time
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
		if input.StartDate != nil {
			updates["start_date"] = *input.StartDate
		}
		if input.EndDate != nil {
			updates["end_date"] = *input.EndDate
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
		return response.NewErrorResponse(err.Error())
	}

	return s.GetPromotionByID(id)
}

func (s *PromotionService) DeletePromotion(id string) response.ApiResponse {
	pid, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Promotion not found")
	}
	res := s.db.Exec("DELETE FROM promotions WHERE id = ?", pid)
	if res.Error != nil {
		return response.NewErrorResponse("Promotion not found")
	}
	if res.RowsAffected == 0 {
		return response.NewErrorResponse("Promotion not found")
	}
	return response.NewSuccessResponse(nil, "Promotion deleted successfully")
}
