package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CustomerTier enum
type CustomerTier string

const (
	CustomerTierBronze   CustomerTier = "BRONZE"
	CustomerTierSilver   CustomerTier = "SILVER"
	CustomerTierGold     CustomerTier = "GOLD"
	CustomerTierPlatinum CustomerTier = "PLATINUM"
)

// CustomerStatus enum
type CustomerStatus string

const (
	CustomerStatusActive   CustomerStatus = "active"
	CustomerStatusInactive CustomerStatus = "inactive"
	CustomerStatusBlocked  CustomerStatus = "blocked"
)

// Customer model
type Customer struct {
	ID           uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CustomerCode string       `gorm:"column:customer_code;type:varchar(20);uniqueIndex;notNull" json:"customer_code"`
	Name         string       `gorm:"type:varchar(100);notNull" json:"name"`
	Email        string       `gorm:"type:varchar(100)" json:"email,omitempty"`
	Phone        string       `gorm:"type:varchar(20)" json:"phone,omitempty"`
	Address      string       `gorm:"type:text" json:"address,omitempty"`
	City         string       `gorm:"type:varchar(50)" json:"city,omitempty"`
	Tier         CustomerTier `gorm:"type:varchar(20);default:'BRONZE'" json:"tier"`
	// Legacy multi-status column kept for DB compatibility.
	Status         CustomerStatus `gorm:"column:status;type:varchar(20);notNull;default:'active'" json:"-"`
	IsActive       bool           `gorm:"column:is_active;notNull;default:true" json:"is_active"`
	LoyaltyPoints  int            `gorm:"column:loyalty_points;notNull;default:0" json:"loyalty_points"`
	CreditLimit    float64        `gorm:"column:credit_limit;type:decimal(15,2);default:0" json:"credit_limit"`
	CreditBalance  float64        `gorm:"column:credit_balance;type:decimal(15,2);notNull;default:0" json:"credit_balance"`
	TotalPurchases float64        `gorm:"column:total_purchases;type:decimal(15,2);notNull;default:0" json:"total_purchases"`
	LastPurchaseAt *time.Time     `gorm:"column:lastPurchaseDate" json:"last_purchase_date,omitempty"`
	CompanyID      uuid.UUID      `gorm:"type:uuid;notNull;index" json:"company_id"`

	BankName          string `gorm:"column:bank_name;type:varchar(50)" json:"bank_name,omitempty"`
	BankAccountNumber string `gorm:"column:bank_account_number;type:varchar(30)" json:"bank_account_number,omitempty"`
	BankAccountName   string `gorm:"column:bank_account_name;type:varchar(100)" json:"bank_account_name,omitempty"`
	BankBranch        string `gorm:"column:bank_branch;type:varchar(50)" json:"bank_branch,omitempty"`
	Allergies         string `gorm:"column:allergies;type:text" json:"allergies,omitempty"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (c *Customer) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (Customer) TableName() string {
	return "customers"
}

// SaleStatus enum
type SaleStatus string

const (
	SaleStatusPending   SaleStatus = "PENDING"
	SaleStatusDone      SaleStatus = "DONE"
	SaleStatusCancelled SaleStatus = "CANCELLED"
	SaleStatusRefunded  SaleStatus = "REFUNDED"
)

// PaymentMethod enum
type PaymentMethod string

const (
	PaymentMethodCash           PaymentMethod = "CASH"
	PaymentMethodCreditCard     PaymentMethod = "CREDIT_CARD"
	PaymentMethodDebitCard      PaymentMethod = "DEBIT_CARD"
	PaymentMethodEWallet        PaymentMethod = "EWALLET"
	PaymentMethodBankTransfer   PaymentMethod = "BANK_TRANSFER"
	PaymentMethodCustomerCredit PaymentMethod = "CUSTOMER_CREDIT"
)

// Sale model
type Sale struct {
	ID                    uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SaleNumber            string     `gorm:"column:sale_number;type:varchar(50);uniqueIndex;notNull" json:"sale_number"`
	WarehouseID           uuid.UUID  `gorm:"type:uuid;notNull;index" json:"warehouse_id"`
	CustomerID            *uuid.UUID `gorm:"type:uuid;index" json:"customer_id,omitempty"`
	CashierID             uuid.UUID  `gorm:"type:uuid;notNull;index" json:"cashier_id"`
	CompanyID             uuid.UUID  `gorm:"type:uuid;notNull;index" json:"company_id"`
	CashDrawerID          *uuid.UUID `gorm:"type:uuid;index" json:"cash_drawer_id,omitempty"`
	SaleDate              time.Time  `gorm:"column:sale_date;notNull;default:now();index" json:"sale_date"`
	Status                SaleStatus `gorm:"type:varchar(20);notNull;default:'DONE';index" json:"status"`
	Subtotal              float64    `gorm:"type:decimal(15,2);notNull;default:0" json:"subtotal"`
	TaxAmount             float64    `gorm:"column:tax_amount;type:decimal(15,2);notNull;default:0" json:"tax_amount"`
	DiscountAmount        float64    `gorm:"column:discount_amount;type:decimal(15,2);notNull;default:0" json:"discount_amount"`
	TotalAmount           float64    `gorm:"column:total_amount;type:decimal(15,2);notNull;default:0" json:"total_amount"`
	PaidAmount            float64    `gorm:"column:paid_amount;type:decimal(15,2);notNull;default:0" json:"paid_amount"`
	ChangeAmount          float64    `gorm:"column:change_amount;type:decimal(15,2);notNull;default:0" json:"change_amount"`
	LoyaltyPointsEarned   int        `gorm:"column:loyalty_points_earned;default:0" json:"loyalty_points_earned"`
	LoyaltyPointsRedeemed int        `gorm:"column:loyalty_points_redeemed;default:0" json:"loyalty_points_redeemed"`
	Notes                 string     `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt             time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	Warehouse *Warehouse    `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	Customer  *Customer     `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Items     []SaleItem    `gorm:"foreignKey:SaleID" json:"items,omitempty"`
	Payments  []SalePayment `gorm:"foreignKey:SaleID" json:"payments,omitempty"`
}

func (s *Sale) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (Sale) TableName() string {
	return "sales"
}

// SaleItem model
type SaleItem struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SaleID         uuid.UUID  `gorm:"type:uuid;notNull;index" json:"sale_id"`
	ProductID      uuid.UUID  `gorm:"type:uuid;notNull;index" json:"product_id"`
	Quantity       int        `gorm:"notNull" json:"quantity"`
	UnitPrice      float64    `gorm:"column:unit_price;type:decimal(15,2);notNull" json:"unit_price"`
	OriginalPrice  float64    `gorm:"column:original_price;type:decimal(15,2);notNull" json:"original_price"`
	CostPrice      float64    `gorm:"column:cost_price;type:decimal(15,2);notNull;default:0" json:"cost_price"`
	DiscountAmount float64    `gorm:"column:discount_amount;type:decimal(15,2);notNull;default:0" json:"discount_amount"`
	TaxRate        float64    `gorm:"column:tax_rate;type:decimal(5,2);default:0" json:"tax_rate"`
	PriceTierID    *uuid.UUID `gorm:"column:price_tier_id;type:uuid" json:"price_tier_id,omitempty"`
	PromotionID    *uuid.UUID `gorm:"column:promotion_id;type:uuid" json:"promotion_id,omitempty"`
	Notes          string     `gorm:"type:text" json:"notes,omitempty"`
}

func (si *SaleItem) BeforeCreate(tx *gorm.DB) error {
	if si.ID == uuid.Nil {
		si.ID = uuid.New()
	}
	return nil
}

func (SaleItem) TableName() string {
	return "sale_items"
}

// SalePayment model
type SalePayment struct {
	ID              uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SaleID          uuid.UUID     `gorm:"type:uuid;notNull;index" json:"sale_id"`
	PaymentMethod   PaymentMethod `gorm:"column:payment_method;type:varchar(30);notNull" json:"payment_method"`
	Amount          float64       `gorm:"type:decimal(15,2);notNull" json:"amount"`
	ReferenceNumber string        `gorm:"column:reference_number;type:varchar(100)" json:"reference_number,omitempty"`
	CardLast4       string        `gorm:"column:card_last_4;type:varchar(4)" json:"card_last_4,omitempty"`
	Notes           string        `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt       time.Time     `gorm:"autoCreateTime" json:"created_at"`
}

func (sp *SalePayment) BeforeCreate(tx *gorm.DB) error {
	if sp.ID == uuid.Nil {
		sp.ID = uuid.New()
	}
	return nil
}

func (SalePayment) TableName() string {
	return "sale_payments"
}
