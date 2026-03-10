package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentTerms enum
type PaymentTerms string

const (
	PaymentTermsCash  PaymentTerms = "CASH"
	PaymentTermsNet30 PaymentTerms = "NET_30"
	PaymentTermsNet60 PaymentTerms = "NET_60"
	PaymentTermsNet90 PaymentTerms = "NET_90"
	PaymentTermsCOD   PaymentTerms = "COD"
)

// SupplierStatus enum
type SupplierStatus string

const (
	SupplierStatusActive      SupplierStatus = "active"
	SupplierStatusInactive    SupplierStatus = "inactive"
	SupplierStatusBlacklisted SupplierStatus = "blacklisted"
)

// Supplier model
type Supplier struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code          string         `gorm:"type:varchar(20);uniqueIndex;notNull" json:"code"`
	Name          string         `gorm:"type:varchar(100);notNull" json:"name"`
	ContactPerson string         `gorm:"column:contact_person;type:varchar(100)" json:"contact_person,omitempty"`
	Email         string         `gorm:"type:varchar(100)" json:"email,omitempty"`
	Phone         string         `gorm:"type:varchar(20)" json:"phone,omitempty"`
	Address       string         `gorm:"type:text" json:"address,omitempty"`
	City          string         `gorm:"type:varchar(50)" json:"city,omitempty"`
	TaxID         string         `gorm:"column:tax_id;type:varchar(50)" json:"tax_id,omitempty"`
	PaymentTerms  PaymentTerms   `gorm:"column:payment_terms;type:varchar(20);notNull;default:'NET_30'" json:"payment_terms"`
	CreditLimit   float64        `gorm:"column:credit_limit;type:decimal(15,2);notNull;default:0" json:"credit_limit"`
	Status        SupplierStatus `gorm:"column:status;type:varchar(20);notNull;default:'active'" json:"-"`
	IsActive      bool           `gorm:"column:is_active;notNull;default:true" json:"is_active"`
	CompanyID     uuid.UUID      `gorm:"column:company_id;type:uuid;notNull;index" json:"company_id"`
	Notes         string         `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (s *Supplier) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (Supplier) TableName() string {
	return "suppliers"
}

// StatusPo enum
type StatusPo string

const (
	StatusPoDraft   StatusPo = "DRAFT"
	StatusPoApprove StatusPo = "APPROVE"
	StatusPoPending StatusPo = "PENDING"
)

// StatusReceive enum
type StatusReceive string

const (
	StatusReceiveDraft   StatusReceive = "DRAFT"
	StatusReceiveReject  StatusReceive = "REJECT"
	StatusReceiveReceive StatusReceive = "RECEIVE"
)

// PurchaseOrder model
type PurchaseOrder struct {
	ID               uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PoNumber         string        `gorm:"column:po_number;type:varchar(50);uniqueIndex;notNull" json:"po_number"`
	SupplierID       uuid.UUID     `gorm:"column:supplier_id;type:uuid;notNull;index" json:"supplier_id"`
	WarehouseID      uuid.UUID     `gorm:"column:warehouse_id;type:uuid;notNull;index" json:"warehouse_id"`
	OrderDate        time.Time     `gorm:"column:order_date;notNull;default:now();index" json:"order_date"`
	ExpectedDelivery *time.Time    `gorm:"column:expected_delivery" json:"expected_delivery,omitempty"`
	ReceiveDate      *time.Time    `gorm:"column:receive_date" json:"receive_date,omitempty"`
	PaymentTerms     PaymentTerms  `gorm:"column:payment_terms;type:varchar(20);notNull;default:'NET_30'" json:"payment_terms"`
	StatusPo         StatusPo      `gorm:"column:status_po;type:varchar(20);notNull;default:'DRAFT'" json:"status_po"`
	StatusReceive    StatusReceive `gorm:"column:status_receive;type:varchar(20);notNull;default:'DRAFT'" json:"status_receive"`
	ReceiveNumber    string        `gorm:"column:receive_number;type:varchar(50)" json:"receive_number"`
	NoteReceive      string        `gorm:"column:note_receive;type:text" json:"note_receive"`
	Subtotal         float64       `gorm:"column:subtotal;type:decimal(15,2);notNull;default:0" json:"subtotal"`
	TaxAmount        float64       `gorm:"column:tax_amount;type:decimal(15,2);notNull;default:0" json:"tax_amount"`
	DiscountAmount   float64       `gorm:"column:discount_amount;type:decimal(15,2);notNull;default:0" json:"discount_amount"`
	TotalAmount      float64       `gorm:"column:total_amount;type:decimal(15,2);notNull;default:0" json:"total_amount"`
	CompanyID        uuid.UUID     `gorm:"column:company_id;type:uuid;notNull;index;references:companies(id)" json:"company_id"`
	Notes            string        `gorm:"type:text" json:"notes,omitempty"`
	CreatedBy        uuid.UUID     `gorm:"column:created_by;type:uuid;notNull" json:"created_by"`
	ApprovedBy       *uuid.UUID    `gorm:"column:approved_by;type:uuid" json:"approved_by,omitempty"`
	CreatedAt        time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time     `gorm:"autoUpdateTime" json:"updated_at"`

	Items []PurchaseOrderItem `gorm:"foreignKey:PoID" json:"items,omitempty"`

	Company Company `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;" json:"-"`
}

func (po *PurchaseOrder) BeforeCreate(tx *gorm.DB) error {
	if po.ID == uuid.Nil {
		po.ID = uuid.New()
	}
	return nil
}

func (PurchaseOrder) TableName() string {
	return "purchase_orders"
}

// PurchaseOrderItem model
type PurchaseOrderItem struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PoID         uuid.UUID `gorm:"column:po_id;type:uuid;notNull;index" json:"po_id"`
	ProductID    uuid.UUID `gorm:"column:product_id;type:uuid;notNull;index" json:"product_id"`
	QtyPo        int       `gorm:"column:qty_po;notNull" json:"qty_po"`
	QtyReceive   int       `gorm:"column:qty_receive;default:0" json:"qty_receive"`
	UnitPrice    float64   `gorm:"column:unit_price;type:decimal(15,2);notNull" json:"unit_price"`
	TaxRate      float64   `gorm:"column:tax_rate;type:decimal(5,2);default:0" json:"tax_rate"`
	DiscountRate float64   `gorm:"column:discount_rate;type:decimal(5,2);default:0" json:"discount_rate"`
}

func (poi *PurchaseOrderItem) BeforeCreate(tx *gorm.DB) error {
	if poi.ID == uuid.Nil {
		poi.ID = uuid.New()
	}
	return nil
}

func (PurchaseOrderItem) TableName() string {
	return "purchase_order_items"
}