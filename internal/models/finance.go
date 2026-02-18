package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InvoiceType string

const (
	InvoiceTypeIncoming InvoiceType = "INCOMING"
	InvoiceTypeOutgoing InvoiceType = "OUTGOING"
)

type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "DRAFT"
	InvoiceStatusSent      InvoiceStatus = "SENT"
	InvoiceStatusPartial   InvoiceStatus = "PARTIAL"
	InvoiceStatusPaid      InvoiceStatus = "PAID"
	InvoiceStatusOverdue   InvoiceStatus = "OVERDUE"
	InvoiceStatusCancelled InvoiceStatus = "CANCELLED"
)

type PaymentMethodFinance string

const (
	PaymentMethodFinanceCash         PaymentMethodFinance = "CASH"
	PaymentMethodFinanceBankTransfer PaymentMethodFinance = "BANK_TRANSFER"
	PaymentMethodFinanceCheck        PaymentMethodFinance = "CHECK"
	PaymentMethodFinanceCard         PaymentMethodFinance = "CARD"
	PaymentMethodFinanceEWallet      PaymentMethodFinance = "EWALLET"
)

// IncomingInvoice maps to invoices_incoming
type IncomingInvoice struct {
	ID             uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	InvoiceNumber  string        `gorm:"column:invoice_number;type:varchar(50);uniqueIndex;notNull" json:"invoice_number"`
	SupplierID     uuid.UUID     `gorm:"column:supplier_id;type:uuid;notNull;index" json:"supplier_id"`
	CompanyID      uuid.UUID     `gorm:"column:company_id;type:uuid;notNull;index" json:"company_id"`
	InvoiceDate    time.Time     `gorm:"column:invoice_date;notNull;index" json:"invoice_date"`
	DueDate        *time.Time    `gorm:"column:due_date" json:"due_date,omitempty"`
	Status         InvoiceStatus `gorm:"column:status;type:varchar(20);notNull;default:'DRAFT';index" json:"status"`
	Subtotal       float64       `gorm:"column:subtotal;type:decimal(15,2);notNull;default:0" json:"subtotal"`
	TaxAmount      float64       `gorm:"column:tax_amount;type:decimal(15,2);notNull;default:0" json:"tax_amount"`
	DiscountAmount float64       `gorm:"column:discount_amount;type:decimal(15,2);notNull;default:0" json:"discount_amount"`
	TotalAmount    float64       `gorm:"column:total_amount;type:decimal(15,2);notNull;default:0" json:"total_amount"`
	PaidAmount     float64       `gorm:"column:paid_amount;type:decimal(15,2);notNull;default:0" json:"paid_amount"`
	BalanceDue     float64       `gorm:"column:balance_due;type:decimal(15,2);notNull;default:0" json:"balance_due"`
	Notes          *string       `gorm:"column:notes;type:text" json:"notes,omitempty"`
	AttachmentPath *string       `gorm:"column:attachment_path;type:text" json:"attachment_path,omitempty"`
	CreatedBy      uuid.UUID     `gorm:"column:created_by;type:uuid;notNull;index" json:"created_by"`
	CreatedAt      time.Time     `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time     `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (i *IncomingInvoice) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

func (IncomingInvoice) TableName() string { return "invoices_incoming" }

// OutgoingInvoice maps to invoices_outgoing
type OutgoingInvoice struct {
	ID             uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	InvoiceNumber  string        `gorm:"column:invoice_number;type:varchar(50);uniqueIndex;notNull" json:"invoice_number"`
	CustomerID     *uuid.UUID    `gorm:"column:customer_id;type:uuid;index" json:"customer_id,omitempty"`
	CompanyID      uuid.UUID     `gorm:"column:company_id;type:uuid;notNull;index" json:"company_id"`
	InvoiceDate    time.Time     `gorm:"column:invoice_date;notNull;index" json:"invoice_date"`
	DueDate        *time.Time    `gorm:"column:due_date" json:"due_date,omitempty"`
	Status         InvoiceStatus `gorm:"column:status;type:varchar(20);notNull;default:'DRAFT';index" json:"status"`
	Subtotal       float64       `gorm:"column:subtotal;type:decimal(15,2);notNull;default:0" json:"subtotal"`
	TaxAmount      float64       `gorm:"column:tax_amount;type:decimal(15,2);notNull;default:0" json:"tax_amount"`
	DiscountAmount float64       `gorm:"column:discount_amount;type:decimal(15,2);notNull;default:0" json:"discount_amount"`
	TotalAmount    float64       `gorm:"column:total_amount;type:decimal(15,2);notNull;default:0" json:"total_amount"`
	PaidAmount     float64       `gorm:"column:paid_amount;type:decimal(15,2);notNull;default:0" json:"paid_amount"`
	BalanceDue     float64       `gorm:"column:balance_due;type:decimal(15,2);notNull;default:0" json:"balance_due"`
	Notes          *string       `gorm:"column:notes;type:text" json:"notes,omitempty"`
	AttachmentPath *string       `gorm:"column:attachment_path;type:text" json:"attachment_path,omitempty"`
	CreatedBy      uuid.UUID     `gorm:"column:created_by;type:uuid;notNull;index" json:"created_by"`
	CreatedAt      time.Time     `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time     `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (o *OutgoingInvoice) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

func (OutgoingInvoice) TableName() string { return "invoices_outgoing" }

type InvoiceItem struct {
	ID             uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	InvoiceType    InvoiceType `gorm:"column:invoice_type;type:varchar(20);notNull;index" json:"invoice_type"`
	InvoiceID      uuid.UUID   `gorm:"column:invoice_id;type:uuid;notNull;index" json:"invoice_id"`
	ProductID      *uuid.UUID  `gorm:"column:product_id;type:uuid;index" json:"product_id,omitempty"`
	Description    string      `gorm:"column:description;type:text;notNull" json:"description"`
	Quantity       int         `gorm:"column:quantity;notNull" json:"quantity"`
	UnitPrice      float64     `gorm:"column:unit_price;type:decimal(15,2);notNull" json:"unit_price"`
	DiscountRate   float64     `gorm:"column:discount_rate;type:decimal(5,2);notNull;default:0" json:"discount_rate"`
	DiscountAmount float64     `gorm:"column:discount_amount;type:decimal(15,2);notNull;default:0" json:"discount_amount"`
	TaxRate        float64     `gorm:"column:tax_rate;type:decimal(5,2);notNull;default:0" json:"tax_rate"`
	TaxAmount      float64     `gorm:"column:tax_amount;type:decimal(15,2);notNull;default:0" json:"tax_amount"`
	LineTotal      float64     `gorm:"column:line_total;type:decimal(15,2);notNull;default:0" json:"line_total"`
	CreatedAt      time.Time   `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time   `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (it *InvoiceItem) BeforeCreate(tx *gorm.DB) error {
	if it.ID == uuid.Nil {
		it.ID = uuid.New()
	}
	return nil
}

func (InvoiceItem) TableName() string { return "invoice_items" }

type InvoicePayment struct {
	ID              uuid.UUID            `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	InvoiceType     InvoiceType          `gorm:"column:invoice_type;type:varchar(20);notNull;index" json:"invoice_type"`
	InvoiceID       uuid.UUID            `gorm:"column:invoice_id;type:uuid;notNull;index" json:"invoice_id"`
	PaymentDate     time.Time            `gorm:"column:payment_date;notNull" json:"payment_date"`
	Amount          float64              `gorm:"column:amount;type:decimal(15,2);notNull" json:"amount"`
	PaymentMethod   PaymentMethodFinance `gorm:"column:payment_method;type:varchar(30);notNull" json:"payment_method"`
	ReferenceNumber *string              `gorm:"column:reference_number;type:varchar(100)" json:"reference_number,omitempty"`
	Notes           *string              `gorm:"column:notes;type:text" json:"notes,omitempty"`
	CreatedBy       uuid.UUID            `gorm:"column:created_by;type:uuid;notNull" json:"created_by"`
	CreatedAt       time.Time            `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (p *InvoicePayment) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (InvoicePayment) TableName() string { return "invoice_payments" }
