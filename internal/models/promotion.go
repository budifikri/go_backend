package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Promotion model
type Promotion struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code              string    `gorm:"type:varchar(50);uniqueIndex;notNull" json:"code"`
	Name              string    `gorm:"type:varchar(100);notNull" json:"name"`
	Description       string    `gorm:"type:text" json:"description,omitempty"`
	PromotionType     string    `gorm:"column:promotion_type;type:varchar(20);notNull" json:"promotion_type"`
	Scope             string    `gorm:"type:varchar(20);notNull;default:'ALL'" json:"scope"`
	DiscountValue     float64   `gorm:"column:discount_value;type:decimal(15,2);notNull" json:"discount_value"`
	MinPurchaseAmount float64   `gorm:"column:min_purchase_amount;type:decimal(15,2);default:0" json:"min_purchase_amount"`
	MaxDiscountAmount *float64  `gorm:"column:max_discount_amount;type:decimal(15,2)" json:"max_discount_amount,omitempty"`
	StartDate         time.Time `gorm:"column:start_date;notNull" json:"start_date"`
	EndDate           time.Time `gorm:"column:end_date;notNull" json:"end_date"`
	IsActive          bool      `gorm:"column:is_active;default:true;notNull" json:"is_active"`
	UsageLimit        *int      `gorm:"column:usage_limit" json:"usage_limit,omitempty"`
	UsageCount        int       `gorm:"column:usage_count;notNull;default:0" json:"usage_count"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (p *Promotion) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (Promotion) TableName() string {
	return "promotions"
}

type PromotionProduct struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PromotionID uuid.UUID `gorm:"column:promotion_id;type:uuid;notNull;index" json:"promotion_id"`
	ProductID   uuid.UUID `gorm:"column:product_id;type:uuid;notNull;index" json:"product_id"`
}

func (pp *PromotionProduct) BeforeCreate(tx *gorm.DB) error {
	if pp.ID == uuid.Nil {
		pp.ID = uuid.New()
	}
	return nil
}

func (PromotionProduct) TableName() string {
	return "promotion_products"
}

type PromotionCategory struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PromotionID uuid.UUID `gorm:"column:promotion_id;type:uuid;notNull;index" json:"promotion_id"`
	CategoryID  uuid.UUID `gorm:"column:category_id;type:uuid;notNull;index" json:"category_id"`
}

func (pc *PromotionCategory) BeforeCreate(tx *gorm.DB) error {
	if pc.ID == uuid.Nil {
		pc.ID = uuid.New()
	}
	return nil
}

func (PromotionCategory) TableName() string {
	return "promotion_categories"
}

type PromotionCustomer struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PromotionID uuid.UUID `gorm:"column:promotion_id;type:uuid;notNull;index" json:"promotion_id"`
	CustomerID  uuid.UUID `gorm:"column:customer_id;type:uuid;notNull;index" json:"customer_id"`
}

func (pc *PromotionCustomer) BeforeCreate(tx *gorm.DB) error {
	if pc.ID == uuid.Nil {
		pc.ID = uuid.New()
	}
	return nil
}

func (PromotionCustomer) TableName() string {
	return "promotion_customers"
}
