package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReturnStatus enum
type ReturnStatus string

const (
	ReturnStatusPending   ReturnStatus = "PENDING"
	ReturnStatusDone      ReturnStatus = "DONE"
	ReturnStatusCancelled ReturnStatus = "CANCELLED"
)

// RefundMethod enum
type RefundMethod string

const (
	RefundMethodCash        RefundMethod = "CASH"
	RefundMethodCreditCard  RefundMethod = "CREDIT_CARD"
	RefundMethodDebitCard   RefundMethod = "DEBIT_CARD"
	RefundMethodBank        RefundMethod = "BANK_TRANSFER"
	RefundMethodStoreCredit RefundMethod = "STORE_CREDIT"
	RefundMethodEWallet     RefundMethod = "EWALLET"
)

// ItemCondition enum
type ItemCondition string

const (
	ItemConditionNew     ItemCondition = "NEW"
	ItemConditionUsed    ItemCondition = "USED"
	ItemConditionDamaged ItemCondition = "DAMAGED"
)

// SalesReturn model
type SalesReturn struct {
	ID           uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ReturnNumber string       `gorm:"column:return_number;type:varchar(50);uniqueIndex;notNull" json:"return_number"`
	SaleID       uuid.UUID    `gorm:"column:sale_id;type:uuid;notNull;index" json:"sale_id"`
	WarehouseID  uuid.UUID    `gorm:"column:warehouse_id;type:uuid;notNull;index" json:"warehouse_id"`
	CustomerID   *uuid.UUID   `gorm:"column:customer_id;type:uuid;index" json:"customer_id,omitempty"`
	ReturnDate   time.Time    `gorm:"column:return_date;notNull;default:now();index" json:"return_date"`
	Status       ReturnStatus `gorm:"type:varchar(20);notNull;default:'DONE'" json:"status"`
	Reason       string       `gorm:"type:text" json:"reason"`
	TotalAmount  float64      `gorm:"column:total_amount;type:decimal(15,2);notNull;default:0" json:"total_amount"`
	RefundMethod RefundMethod `gorm:"column:refund_method;type:varchar(20);notNull" json:"refund_method"`
	ProcessedBy  uuid.UUID    `gorm:"column:processed_by;type:uuid;notNull;index" json:"processed_by"`
	CreatedAt    time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time    `gorm:"autoUpdateTime" json:"updated_at"`

	Items []SalesReturnItem `gorm:"foreignKey:ReturnID" json:"items,omitempty"`
}

func (sr *SalesReturn) BeforeCreate(tx *gorm.DB) error {
	if sr.ID == uuid.Nil {
		sr.ID = uuid.New()
	}
	return nil
}

func (SalesReturn) TableName() string {
	return "sales_returns"
}

// SalesReturnItem model
type SalesReturnItem struct {
	ID           uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ReturnID     uuid.UUID     `gorm:"column:return_id;type:uuid;notNull;index" json:"return_id"`
	SaleItemID   uuid.UUID     `gorm:"column:sale_item_id;type:uuid;notNull;index" json:"sale_item_id"`
	ProductID    uuid.UUID     `gorm:"column:product_id;type:uuid;notNull;index" json:"product_id"`
	Quantity     int           `gorm:"notNull" json:"quantity"`
	UnitPrice    float64       `gorm:"column:unit_price;type:decimal(15,2);notNull" json:"unit_price"`
	RefundAmount float64       `gorm:"column:refund_amount;type:decimal(15,2);notNull" json:"refund_amount"`
	Condition    ItemCondition `gorm:"type:varchar(20);notNull" json:"condition"`
	Notes        string        `gorm:"type:text" json:"notes,omitempty"`
}

func (sri *SalesReturnItem) BeforeCreate(tx *gorm.DB) error {
	if sri.ID == uuid.Nil {
		sri.ID = uuid.New()
	}
	return nil
}

func (SalesReturnItem) TableName() string {
	return "sales_return_items"
}

// ExchangeStatus enum
type ExchangeStatus string

const (
	ExchangeStatusPending   ExchangeStatus = "PENDING"
	ExchangeStatusDone      ExchangeStatus = "DONE"
	ExchangeStatusCancelled ExchangeStatus = "CANCELLED"
)

// ExchangeItemType enum
type ExchangeItemType string

const (
	ExchangeItemTypeReturned ExchangeItemType = "RETURNED"
	ExchangeItemTypeReceived ExchangeItemType = "RECEIVED"
)

// ItemExchange model
type ItemExchange struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExchangeNumber  string         `gorm:"column:exchange_number;type:varchar(50);uniqueIndex;notNull" json:"exchange_number"`
	SaleID          *uuid.UUID     `gorm:"column:sale_id;type:uuid;index" json:"sale_id,omitempty"`
	WarehouseID     uuid.UUID      `gorm:"column:warehouse_id;type:uuid;notNull;index" json:"warehouse_id"`
	CustomerID      *uuid.UUID     `gorm:"column:customer_id;type:uuid;index" json:"customer_id,omitempty"`
	ExchangeDate    time.Time      `gorm:"column:exchange_date;notNull;default:now();index" json:"exchange_date"`
	Status          ExchangeStatus `gorm:"type:varchar(20);notNull;default:'DONE'" json:"status"`
	Reason          string         `gorm:"type:text" json:"reason"`
	PriceDifference float64        `gorm:"column:price_difference;type:decimal(15,2);notNull;default:0" json:"price_difference"`
	ProcessedBy     uuid.UUID      `gorm:"column:processed_by;type:uuid;notNull;index" json:"processed_by"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`

	Items []ExchangeItem `gorm:"foreignKey:ExchangeID" json:"items,omitempty"`
}

func (ie *ItemExchange) BeforeCreate(tx *gorm.DB) error {
	if ie.ID == uuid.Nil {
		ie.ID = uuid.New()
	}
	return nil
}

func (ItemExchange) TableName() string {
	return "item_exchanges"
}

// ExchangeItem model
type ExchangeItem struct {
	ID          uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExchangeID  uuid.UUID        `gorm:"column:exchange_id;type:uuid;notNull;index" json:"exchange_id"`
	ItemType    ExchangeItemType `gorm:"column:item_type;type:varchar(20);notNull" json:"item_type"`
	SaleItemID  *uuid.UUID       `gorm:"column:sale_item_id;type:uuid;index" json:"sale_item_id,omitempty"`
	ProductID   uuid.UUID        `gorm:"column:product_id;type:uuid;notNull;index" json:"product_id"`
	Quantity    int              `gorm:"notNull" json:"quantity"`
	UnitPrice   float64          `gorm:"column:unit_price;type:decimal(15,2);notNull" json:"unit_price"`
	TotalAmount float64          `gorm:"column:total_amount;type:decimal(15,2);notNull" json:"total_amount"`
	Condition   ItemCondition    `gorm:"type:varchar(20);notNull" json:"condition"`
}

func (ei *ExchangeItem) BeforeCreate(tx *gorm.DB) error {
	if ei.ID == uuid.Nil {
		ei.ID = uuid.New()
	}
	return nil
}

func (ExchangeItem) TableName() string {
	return "exchange_items"
}
