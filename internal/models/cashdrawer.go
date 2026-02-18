package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DrawerStatus string

const (
	DrawerStatusOpen   DrawerStatus = "OPEN"
	DrawerStatusClosed DrawerStatus = "CLOSED"
)

type TransactionType string

const (
	TransactionTypeOpening   TransactionType = "OPENING"
	TransactionTypeCashIn    TransactionType = "CASH_IN"
	TransactionTypeCashOut   TransactionType = "CASH_OUT"
	TransactionTypeClosing   TransactionType = "CLOSING"
	TransactionTypeSaleIn    TransactionType = "SALE_IN"
	TransactionTypeSaleOut   TransactionType = "SALE_OUT"
	TransactionTypeReturnOut TransactionType = "RETURN_OUT"
)

type CashDrawer struct {
	ID               uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DrawerNumber     string       `gorm:"column:drawer_number;type:varchar(20);uniqueIndex;notNull" json:"drawer_number"`
	WarehouseID      uuid.UUID    `gorm:"column:warehouse_id;type:uuid;notNull;index" json:"warehouse_id"`
	CashierID        uuid.UUID    `gorm:"column:cashier_id;type:uuid;notNull;index" json:"cashier_id"`
	CompanyID        uuid.UUID    `gorm:"column:company_id;type:uuid;notNull;index" json:"company_id"`
	Status           DrawerStatus `gorm:"column:status;type:varchar(10);notNull;default:'OPEN';index" json:"status"`
	OpeningBalance   float64      `gorm:"column:opening_balance;type:decimal(15,2);notNull;default:0" json:"opening_balance"`
	ClosingBalance   *float64     `gorm:"column:closing_balance;type:decimal(15,2)" json:"closing_balance,omitempty"`
	ExpectedBalance  float64      `gorm:"column:expected_balance;type:decimal(15,2);notNull;default:0" json:"expected_balance"`
	Variance         *float64     `gorm:"column:variance;type:decimal(15,2)" json:"variance,omitempty"`
	OpenedAt         time.Time    `gorm:"column:opened_at;notNull;index" json:"opened_at"`
	ClosedAt         *time.Time   `gorm:"column:closed_at" json:"closed_at,omitempty"`
	Notes            *string      `gorm:"column:notes;type:text" json:"notes,omitempty"`
	DepositInvoiceID *uuid.UUID   `gorm:"column:deposit_invoice_id;type:uuid" json:"deposit_invoice_id,omitempty"`
	CreatedAt        time.Time    `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time    `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (d *CashDrawer) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

func (CashDrawer) TableName() string { return "cash_drawers" }

type CashDrawerTransaction struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CashDrawerID  uuid.UUID       `gorm:"column:cash_drawer_id;type:uuid;notNull;index" json:"cash_drawer_id"`
	Type          TransactionType `gorm:"column:type;type:varchar(20);notNull;index" json:"type"`
	Amount        float64         `gorm:"column:amount;type:decimal(15,2);notNull" json:"amount"`
	BalanceAfter  float64         `gorm:"column:balance_after;type:decimal(15,2);notNull" json:"balance_after"`
	SaleID        *uuid.UUID      `gorm:"column:sale_id;type:uuid;index" json:"sale_id,omitempty"`
	PaymentMethod *string         `gorm:"column:payment_method;type:varchar(30)" json:"payment_method,omitempty"`
	Reason        string          `gorm:"column:reason;type:text;notNull" json:"reason"`
	CreatedAt     time.Time       `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	CreatedBy     uuid.UUID       `gorm:"column:created_by;type:uuid;notNull;index" json:"created_by"`
}

func (t *CashDrawerTransaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (CashDrawerTransaction) TableName() string { return "cash_drawer_transactions" }
