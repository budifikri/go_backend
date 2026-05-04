package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductStatus enum
type ProductStatus string

const (
	ProductStatusActive       ProductStatus = "active"
	ProductStatusInactive     ProductStatus = "inactive"
	ProductStatusDiscontinued ProductStatus = "discontinued"
)

// Unit model
type Unit struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code        string     `gorm:"type:varchar(10);uniqueIndex;notNull" json:"code"`
	Name        string     `gorm:"type:varchar(50);notNull" json:"name"`
	Description string     `gorm:"type:text" json:"description,omitempty"`
	IsActive    bool       `gorm:"default:true;notNull" json:"is_active"`
	CompanyID   *uuid.UUID `gorm:"type:uuid" json:"company_id,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (u *Unit) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (Unit) TableName() string {
	return "units_of_measure"
}

// Category model
type Category struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code        string     `gorm:"type:varchar(20);uniqueIndex;notNull" json:"code"`
	Name        string     `gorm:"type:varchar(100);notNull" json:"name"`
	ProductType string     `gorm:"type:varchar(20);notNull;default:'stockable'" json:"product_type"`
	Description string     `gorm:"type:text" json:"description,omitempty"`
	ParentID    *uuid.UUID `gorm:"type:uuid" json:"parent_id,omitempty"`
	CompanyID   *uuid.UUID `gorm:"type:uuid" json:"company_id,omitempty"`
	IsActive    bool       `gorm:"default:true;notNull" json:"is_active"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`

	Parent *Category `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	c.ProductType = NormalizeCategoryProductType(c.ProductType)
	return nil
}

func (c *Category) BeforeSave(tx *gorm.DB) error {
	c.ProductType = NormalizeCategoryProductType(c.ProductType)
	return nil
}

func (c *Category) AfterFind(tx *gorm.DB) error {
	c.ProductType = NormalizeCategoryProductType(c.ProductType)
	return nil
}

func (Category) TableName() string {
	return "categories"
}

const (
	CategoryProductTypeStockable  = "stockable"
	CategoryProductTypeService    = "service"
	CategoryProductTypeConsumable = "consumable"
)

func NormalizeCategoryProductType(value string) string {
	switch value {
	case CategoryProductTypeService, CategoryProductTypeConsumable:
		return value
	default:
		return CategoryProductTypeStockable
	}
}

// WarehouseType enum
type WarehouseType string

const (
	WarehouseTypeMain    WarehouseType = "MAIN"
	WarehouseTypeBranch  WarehouseType = "BRANCH"
	WarehouseTypeStorage WarehouseType = "STORAGE"
	WarehouseTypeOutlet  WarehouseType = "OUTLET"
)

// WarehouseStatus enum
type WarehouseStatus string

const (
	WarehouseStatusActive      WarehouseStatus = "active"
	WarehouseStatusInactive    WarehouseStatus = "inactive"
	WarehouseStatusMaintenance WarehouseStatus = "maintenance"
)

// Warehouse model
type Warehouse struct {
	ID   uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code string        `gorm:"type:varchar(20);uniqueIndex;notNull" json:"code"`
	Name string        `gorm:"type:varchar(100);notNull" json:"name"`
	Type WarehouseType `gorm:"type:varchar(20);notNull" json:"type"`
	// Legacy multi-status column kept for DB compatibility.
	Status    WarehouseStatus `gorm:"column:status;type:varchar(20);notNull;default:'active'" json:"-"`
	IsActive  bool            `gorm:"column:is_active;notNull;default:true" json:"is_active"`
	Address   string          `gorm:"type:text" json:"address,omitempty"`
	City      string          `gorm:"type:varchar(50)" json:"city,omitempty"`
	Phone     string          `gorm:"type:varchar(20)" json:"phone,omitempty"`
	ManagerID *uuid.UUID      `gorm:"type:uuid" json:"manager_id,omitempty"`
	CompanyID *uuid.UUID      `gorm:"type:uuid" json:"company_id,omitempty"`
	CreatedAt time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (w *Warehouse) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

func (Warehouse) TableName() string {
	return "warehouses"
}

// Product model
type Product struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SKU         string     `gorm:"type:varchar(50);uniqueIndex;notNull" json:"sku"`
	Barcode     string     `gorm:"type:varchar(50);uniqueIndex" json:"barcode,omitempty"`
	Name        string     `gorm:"type:varchar(200);notNull" json:"name"`
	Description string     `gorm:"type:text" json:"description,omitempty"`
	CategoryID  *uuid.UUID `gorm:"type:uuid" json:"category_id,omitempty"`
	UnitID      uuid.UUID  `gorm:"type:uuid;notNull" json:"unit_id"`
	CostPrice   float64    `gorm:"type:decimal(15,2);notNull" json:"cost_price"`
	RetailPrice float64    `gorm:"type:decimal(15,2);notNull" json:"retail_price"`
	// Legacy multi-status column kept for DB compatibility.
	Status       ProductStatus `gorm:"column:status;type:varchar(20);notNull;default:'active'" json:"-"`
	IsActive     bool          `gorm:"column:is_active;notNull;default:true" json:"is_active"`
	TaxRate      float64       `gorm:"type:decimal(5,2);default:0;notNull" json:"tax_rate"`
	IsTrackable  bool          `gorm:"default:false;notNull" json:"is_trackable"`
	ReorderPoint int           `gorm:"default:0;notNull" json:"reorder_point"`
	CompanyID    *uuid.UUID    `gorm:"type:uuid" json:"company_id,omitempty"`
	CreatedAt    time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time     `gorm:"autoUpdateTime" json:"updated_at"`

	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Unit     *Unit     `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (Product) TableName() string {
	return "products"
}

// PriceTier model
type PriceTier struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProductID   uuid.UUID `gorm:"type:uuid;notNull;index" json:"product_id"`
	TierName    string    `gorm:"type:varchar(50);notNull" json:"tier_name"`
	MinQuantity int       `gorm:"notNull" json:"min_quantity"`
	MaxQuantity *int      `json:"max_quantity,omitempty"`
	UnitPrice   float64   `gorm:"type:decimal(15,2);notNull" json:"unit_price"`
	IsActive    bool      `gorm:"default:true;notNull" json:"is_active"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (pt *PriceTier) BeforeCreate(tx *gorm.DB) error {
	if pt.ID == uuid.Nil {
		pt.ID = uuid.New()
	}
	return nil
}

func (PriceTier) TableName() string {
	return "price_tiers"
}
