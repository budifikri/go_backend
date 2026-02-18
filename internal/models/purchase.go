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
	Status        SupplierStatus `gorm:"type:varchar(20);notNull;default:'active'" json:"status"`
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

// PoStatus enum
type PoStatus string

const (
	PoStatusDraft           PoStatus = "DRAFT"
	PoStatusPending         PoStatus = "PENDING"
	PoStatusApproved        PoStatus = "APPROVED"
	PoStatusOrdered         PoStatus = "ORDERED"
	PoStatusPartialReceived PoStatus = "PARTIAL_RECEIVED"
	PoStatusReceived        PoStatus = "RECEIVED"
	PoStatusClosed          PoStatus = "CLOSED"
	PoStatusCancelled       PoStatus = "CANCELLED"
)

// PurchaseOrder model
type PurchaseOrder struct {
	ID               uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PoNumber         string       `gorm:"column:po_number;type:varchar(50);uniqueIndex;notNull" json:"po_number"`
	SupplierID       uuid.UUID    `gorm:"column:supplier_id;type:uuid;notNull;index" json:"supplier_id"`
	WarehouseID      uuid.UUID    `gorm:"column:warehouse_id;type:uuid;notNull;index" json:"warehouse_id"`
	OrderDate        time.Time    `gorm:"column:order_date;notNull;default:now();index" json:"order_date"`
	ExpectedDelivery *time.Time   `gorm:"column:expected_delivery" json:"expected_delivery,omitempty"`
	PaymentTerms     PaymentTerms `gorm:"column:payment_terms;type:varchar(20);notNull;default:'NET_30'" json:"payment_terms"`
	Status           PoStatus     `gorm:"type:varchar(30);notNull;default:'DRAFT'" json:"status"`
	Subtotal         float64      `gorm:"column:subtotal;type:decimal(15,2);notNull;default:0" json:"subtotal"`
	TaxAmount        float64      `gorm:"column:tax_amount;type:decimal(15,2);notNull;default:0" json:"tax_amount"`
	DiscountAmount   float64      `gorm:"column:discount_amount;type:decimal(15,2);notNull;default:0" json:"discount_amount"`
	TotalAmount      float64      `gorm:"column:total_amount;type:decimal(15,2);notNull;default:0" json:"total_amount"`
	CompanyID        uuid.UUID    `gorm:"column:company_id;type:uuid;notNull;index" json:"company_id"`
	Notes            string       `gorm:"type:text" json:"notes,omitempty"`
	CreatedBy        uuid.UUID    `gorm:"column:created_by;type:uuid;notNull" json:"created_by"`
	ApprovedBy       *uuid.UUID   `gorm:"column:approved_by;type:uuid" json:"approved_by,omitempty"`
	CreatedAt        time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time    `gorm:"autoUpdateTime" json:"updated_at"`

	Items []PurchaseOrderItem `gorm:"foreignKey:PoID" json:"items,omitempty"`
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
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PoID             uuid.UUID `gorm:"column:po_id;type:uuid;notNull;index" json:"po_id"`
	ProductID        uuid.UUID `gorm:"column:product_id;type:uuid;notNull;index" json:"product_id"`
	Quantity         int       `gorm:"notNull" json:"quantity"`
	ReceivedQuantity int       `gorm:"column:received_quantity;default:0" json:"received_quantity"`
	UnitPrice        float64   `gorm:"column:unit_price;type:decimal(15,2);notNull" json:"unit_price"`
	TaxRate          float64   `gorm:"column:tax_rate;type:decimal(5,2);default:0" json:"tax_rate"`
	DiscountRate     float64   `gorm:"column:discount_rate;type:decimal(5,2);default:0" json:"discount_rate"`
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

// GrnStatus enum
type GrnStatus string

const (
	GrnStatusDraft     GrnStatus = "DRAFT"
	GrnStatusVerified  GrnStatus = "VERIFIED"
	GrnStatusPosted    GrnStatus = "POSTED"
	GrnStatusCancelled GrnStatus = "CANCELLED"
)

// GoodsReceivedNote model
type GoodsReceivedNote struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	GrnNumber     string     `gorm:"column:grn_number;type:varchar(50);uniqueIndex;notNull" json:"grn_number"`
	PoID          uuid.UUID  `gorm:"column:po_id;type:uuid;notNull;index" json:"po_id"`
	WarehouseID   uuid.UUID  `gorm:"column:warehouse_id;type:uuid;notNull;index" json:"warehouse_id"`
	ReceivedDate  time.Time  `gorm:"column:received_date;notNull;default:now()" json:"received_date"`
	Status        GrnStatus  `gorm:"type:varchar(20);notNull;default:'DRAFT'" json:"status"`
	InvoiceNumber string     `gorm:"column:invoice_number;type:varchar(50)" json:"invoice_number,omitempty"`
	Notes         string     `gorm:"type:text" json:"notes,omitempty"`
	CreatedBy     uuid.UUID  `gorm:"column:created_by;type:uuid;notNull" json:"created_by"`
	VerifiedBy    *uuid.UUID `gorm:"column:verified_by;type:uuid" json:"verified_by,omitempty"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`
	VerifiedAt    *time.Time `gorm:"column:verified_at" json:"verified_at,omitempty"`

	Items []GrnItem `gorm:"foreignKey:GrnID" json:"items,omitempty"`
}

func (g *GoodsReceivedNote) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return nil
}

func (GoodsReceivedNote) TableName() string {
	return "goods_received_notes"
}

// GrnItem model
type GrnItem struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	GrnID            uuid.UUID `gorm:"column:grn_id;type:uuid;notNull;index" json:"grn_id"`
	PoItemID         uuid.UUID `gorm:"column:po_item_id;type:uuid;notNull;index" json:"po_item_id"`
	ProductID        uuid.UUID `gorm:"column:product_id;type:uuid;notNull;index" json:"product_id"`
	OrderedQuantity  int       `gorm:"column:ordered_quantity;notNull" json:"ordered_quantity"`
	ReceivedQuantity int       `gorm:"column:received_quantity;notNull" json:"received_quantity"`
	RejectedQuantity int       `gorm:"column:rejected_quantity;default:0" json:"rejected_quantity"`
	UnitPrice        float64   `gorm:"column:unit_price;type:decimal(15,2);notNull" json:"unit_price"`
	QualityNotes     string    `gorm:"column:quality_notes;type:text" json:"quality_notes,omitempty"`
}

func (gi *GrnItem) BeforeCreate(tx *gorm.DB) error {
	if gi.ID == uuid.Nil {
		gi.ID = uuid.New()
	}
	return nil
}

func (GrnItem) TableName() string {
	return "grn_items"
}
