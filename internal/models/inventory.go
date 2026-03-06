package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MovementType enum
type MovementType string

const (
	MovementTypeSale          MovementType = "SALE"
	MovementTypePurchase      MovementType = "PURCHASE"
	MovementTypeTransferOut   MovementType = "TRANSFER_OUT"
	MovementTypeTransferIn    MovementType = "TRANSFER_IN"
	MovementTypeAdjustmentIn  MovementType = "ADJUSTMENT_IN"
	MovementTypeAdjustmentOut MovementType = "ADJUSTMENT_OUT"
	MovementTypeReturn        MovementType = "RETURN"
	MovementTypeDamage        MovementType = "DAMAGE"
	MovementTypeOpname        MovementType = "OPNAME"
	MovementTypeExchangeIn    MovementType = "EXCHANGE_IN"
	MovementTypeExchangeOut   MovementType = "EXCHANGE_OUT"
)

// Inventory model
type Inventory struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProductID         uuid.UUID  `gorm:"type:uuid;notNull;index:idx_product_warehouse,priority:1" json:"product_id"`
	WarehouseID       uuid.UUID  `gorm:"type:uuid;notNull;index:idx_product_warehouse,priority:2" json:"warehouse_id"`
	Quantity          int        `gorm:"default:0;notNull" json:"quantity"`
	ReservedQuantity  int        `gorm:"default:0;notNull" json:"reserved_quantity"`
	AvailableQuantity int        `gorm:"default:0;notNull" json:"available_quantity"`
	MinStockLevel     int        `gorm:"default:0;notNull" json:"min_stock_level"`
	MaxStockLevel     int        `gorm:"default:0;notNull" json:"max_stock_level"`
	LastStockTake     *time.Time `json:"last_stock_take,omitempty"`
	CreatedAt         time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	Product   *Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Warehouse *Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
}

func (i *Inventory) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

func (Inventory) TableName() string {
	return "inventory"
}

// StockMovement model
type StockMovement struct {
	ID            uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProductID     uuid.UUID    `gorm:"type:uuid;notNull;index" json:"product_id"`
	WarehouseID   uuid.UUID    `gorm:"type:uuid;notNull;index" json:"warehouse_id"`
	MovementType  MovementType `gorm:"type:varchar(30);notNull" json:"movement_type"`
	Quantity      int          `gorm:"notNull" json:"quantity"`
	ReferenceType string       `gorm:"type:varchar(50)" json:"reference_type,omitempty"`
	ReferenceID   *uuid.UUID   `gorm:"type:uuid" json:"reference_id,omitempty"`
	Notes         string       `gorm:"type:text" json:"notes,omitempty"`
	CreatedBy     *uuid.UUID   `gorm:"column:created_by;type:uuid" json:"created_by,omitempty"`
	CreatedAt     time.Time    `gorm:"autoCreateTime" json:"created_at"`

	Product   *Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Warehouse *Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
}

func (sm *StockMovement) BeforeCreate(tx *gorm.DB) error {
	if sm.ID == uuid.Nil {
		sm.ID = uuid.New()
	}
	return nil
}

func (StockMovement) TableName() string {
	return "stock_movements"
}

// StockTransferStatus enum
type StockTransferStatus string

const (
	StockTransferStatusPending   StockTransferStatus = "pending"
	StockTransferStatusInTransit StockTransferStatus = "in_transit"
	StockTransferStatusReceived  StockTransferStatus = "received"
	StockTransferStatusCancelled StockTransferStatus = "cancelled"
)

// StockTransfer model
type StockTransfer struct {
	ID              uuid.UUID           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TransferNumber  string              `gorm:"type:varchar(50);uniqueIndex;notNull" json:"transfer_number"`
	FromWarehouseID uuid.UUID           `gorm:"type:uuid;notNull" json:"from_warehouse_id"`
	ToWarehouseID   uuid.UUID           `gorm:"type:uuid;notNull" json:"to_warehouse_id"`
	UserID          uuid.UUID           `gorm:"type:uuid;notNull" json:"user_id"`
	ExpectedArrival *time.Time          `json:"expected_arrival,omitempty"`
	ActualArrival   *time.Time          `json:"actual_arrival,omitempty"`
	Status          StockTransferStatus `gorm:"type:varchar(20);notNull;default:'pending'" json:"status"`
	Notes           string              `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt       time.Time           `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time           `gorm:"autoUpdateTime" json:"updated_at"`

	FromWarehouse *Warehouse          `gorm:"foreignKey:FromWarehouseID" json:"from_warehouse,omitempty"`
	ToWarehouse   *Warehouse          `gorm:"foreignKey:ToWarehouseID" json:"to_warehouse,omitempty"`
	Items         []StockTransferItem `gorm:"foreignKey:TransferID" json:"items,omitempty"`
}

func (st *StockTransfer) BeforeCreate(tx *gorm.DB) error {
	if st.ID == uuid.Nil {
		st.ID = uuid.New()
	}
	return nil
}

func (StockTransfer) TableName() string {
	return "stock_transfers"
}

// StockTransferItem model
type StockTransferItem struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TransferID       uuid.UUID `gorm:"type:uuid;notNull;index" json:"transfer_id"`
	ProductID        uuid.UUID `gorm:"type:uuid;notNull" json:"product_id"`
	Quantity         int       `gorm:"notNull" json:"quantity"`
	ReceivedQuantity *int      `json:"received_quantity,omitempty"`

	Transfer *StockTransfer `gorm:"foreignKey:TransferID" json:"transfer,omitempty"`
	Product  *Product       `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (sti *StockTransferItem) BeforeCreate(tx *gorm.DB) error {
	if sti.ID == uuid.Nil {
		sti.ID = uuid.New()
	}
	return nil
}

func (StockTransferItem) TableName() string {
	return "stock_transfer_items"
}

// StockOpnameStatus enum
type StockOpnameStatus string

const (
	StockOpnameStatusDraft      StockOpnameStatus = "draft"
	StockOpnameStatusInProgress StockOpnameStatus = "in_progress"
	StockOpnameStatusCompleted  StockOpnameStatus = "completed"
	StockOpnameStatusApproved   StockOpnameStatus = "approved"
)

// StockOpname model
type StockOpname struct {
	ID           uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OpnameNumber string            `gorm:"type:varchar(50);uniqueIndex;notNull" json:"opname_number"`
	WarehouseID  uuid.UUID         `gorm:"type:uuid;notNull" json:"warehouse_id"`
	UserID       uuid.UUID         `gorm:"type:uuid;notNull" json:"user_id"`
	OpnameDate   time.Time         `gorm:"notNull" json:"opname_date"`
	Status       StockOpnameStatus `gorm:"type:varchar(20);notNull;default:'draft'" json:"status"`
	Notes        string            `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt    time.Time         `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time         `gorm:"autoUpdateTime" json:"updated_at"`

	Warehouse *Warehouse        `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	User      *User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Items     []StockOpnameItem `gorm:"foreignKey:OpnameID" json:"items,omitempty"`
}

func (so *StockOpname) BeforeCreate(tx *gorm.DB) error {
	if so.ID == uuid.Nil {
		so.ID = uuid.New()
	}
	return nil
}

func (StockOpname) TableName() string {
	return "stock_opnames"
}

// StockOpnameItem model
type StockOpnameItem struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OpnameID       uuid.UUID `gorm:"type:uuid;notNull;index" json:"opname_id"`
	ProductID      uuid.UUID `gorm:"type:uuid;notNull" json:"product_id"`
	SystemQuantity int       `gorm:"notNull" json:"system_quantity"`
	ActualQuantity int       `gorm:"notNull" json:"actual_quantity"`
	Difference     int       `gorm:"notNull" json:"difference"`
	Status         string    `gorm:"type:varchar(20);default:'pending'" json:"status"`
	Notes          string    `gorm:"type:text" json:"notes,omitempty"`

	Opname  *StockOpname `gorm:"foreignKey:OpnameID" json:"opname,omitempty"`
	Product *Product     `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (soi *StockOpnameItem) BeforeCreate(tx *gorm.DB) error {
	if soi.ID == uuid.Nil {
		soi.ID = uuid.New()
	}
	return nil
}

func (StockOpnameItem) TableName() string {
	return "stock_opname_items"
}
