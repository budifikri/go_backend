package repository

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
)

type InventoryRepository struct {
	db *gorm.DB
}

type InventoryListRow struct {
	InventoryID       *uuid.UUID `gorm:"column:inventory_id"`
	ProductID         uuid.UUID  `gorm:"column:product_id"`
	ProductName       string     `gorm:"column:product_name"`
	WarehouseID       *uuid.UUID `gorm:"column:warehouse_id"`
	WarehouseName     *string    `gorm:"column:warehouse_name"`
	Quantity          int        `gorm:"column:quantity"`
	ReservedQuantity  int        `gorm:"column:reserved_quantity"`
	AvailableQuantity int        `gorm:"column:available_quantity"`
	MinStockLevel     int        `gorm:"column:min_stock_level"`
	MaxStockLevel     int        `gorm:"column:max_stock_level"`
}

func NewInventoryRepository(db *gorm.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

func (r *InventoryRepository) FindAll(filters map[string]interface{}, limit, offset int) ([]InventoryListRow, int64, error) {
	var rows []InventoryListRow
	var total int64

	query := r.db.Table("products p")

	warehouseID, hasWarehouseFilter := filters["warehouse_id"].(string)
	if hasWarehouseFilter && warehouseID != "" {
		query = query.Select(`
			i.id AS inventory_id,
			p.id AS product_id,
			p.name AS product_name,
			w.id AS warehouse_id,
			w.name AS warehouse_name,
			COALESCE(i.quantity, 0) AS quantity,
			COALESCE(i.reserved_quantity, 0) AS reserved_quantity,
			COALESCE(i.available_quantity, 0) AS available_quantity,
			COALESCE(i.min_stock_level, 0) AS min_stock_level,
			COALESCE(i.max_stock_level, 0) AS max_stock_level
		`).
			Joins("LEFT JOIN inventory i ON i.product_id = p.id AND i.warehouse_id = ?", warehouseID).
			Joins("LEFT JOIN warehouses w ON w.id = ?", warehouseID)
	} else {
		query = query.Select(`
			i.id AS inventory_id,
			p.id AS product_id,
			p.name AS product_name,
			w.id AS warehouse_id,
			w.name AS warehouse_name,
			COALESCE(i.quantity, 0) AS quantity,
			COALESCE(i.reserved_quantity, 0) AS reserved_quantity,
			COALESCE(i.available_quantity, 0) AS available_quantity,
			COALESCE(i.min_stock_level, 0) AS min_stock_level,
			COALESCE(i.max_stock_level, 0) AS max_stock_level
		`).
			Joins("LEFT JOIN inventory i ON i.product_id = p.id").
			Joins("LEFT JOIN warehouses w ON w.id = i.warehouse_id")
	}

	query = query.Where("p.is_active = ?", true)

	if companyID, ok := filters["company_id"].(string); ok && companyID != "" {
		query = query.Where("p.company_id = ?", companyID)
	}
	if productID, ok := filters["product_id"].(string); ok && productID != "" {
		query = query.Where("p.id = ?", productID)
	}
	if search, ok := filters["search"].(string); ok && search != "" {
		like := "%" + search + "%"
		query = query.Where("p.name ILIKE ? OR p.sku ILIKE ? OR p.barcode ILIKE ?", like, like, like)
	}

	stockFilter := "available"
	if stock, ok := filters["stock"].(string); ok && stock != "" {
		stockFilter = strings.ToLower(strings.TrimSpace(stock))
	}
	switch stockFilter {
	case "all", "all_stock", "all stock":
		// no filter
	case "minus", "stock_minus", "stock minus":
		query = query.Where("COALESCE(i.quantity, 0) < 0")
	case "empty", "stock_empty", "stock empty":
		query = query.Where("COALESCE(i.quantity, 0) = 0")
	case "available", "stock_available", "stock available":
		fallthrough
	default:
		query = query.Where("COALESCE(i.quantity, 0) > 0")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("p.name ASC, w.name ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (r *InventoryRepository) FindByProductAndWarehouse(productID, warehouseID uuid.UUID) (*models.Inventory, error) {
	var inventory models.Inventory
	if err := r.db.Preload("Product").Preload("Warehouse").First(&inventory, "product_id = ? AND warehouse_id = ?", productID, warehouseID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &inventory, nil
}

func (r *InventoryRepository) Create(inventory *models.Inventory) error {
	return r.db.Create(inventory).Error
}

func (r *InventoryRepository) Update(inventory *models.Inventory) error {
	return r.db.Save(inventory).Error
}

func (r *InventoryRepository) UpdateQuantity(productID, warehouseID uuid.UUID, quantity int) error {
	return r.db.Model(&models.Inventory{}).
		Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
		Updates(map[string]interface{}{
			"quantity":           quantity,
			"available_quantity": quantity,
		}).Error
}

func (r *InventoryRepository) CreateStockMovement(movement *models.StockMovement) error {
	return r.db.Create(movement).Error
}

func (r *InventoryRepository) GetOpeningBalance(productID, warehouseID uuid.UUID, beforeDate time.Time) (int, error) {
	var balance int

	// Match TS behavior:
	// - Sum IN movement types as +quantity
	// - Sum OUT movement types as -quantity
	// - OPNAME stores signed quantity (difference), so add it directly
	// - Only movements strictly before from_date
	err := r.db.Raw(`
		SELECT COALESCE(SUM(
			CASE
				WHEN movement_type IN ('ADJUSTMENT_IN','PURCHASE','TRANSFER_IN','RETURN','EXCHANGE_IN') THEN quantity
				WHEN movement_type IN ('SALE','ADJUSTMENT_OUT','TRANSFER_OUT','DAMAGE','EXCHANGE_OUT') THEN -quantity
				WHEN movement_type IN ('OPNAME') THEN quantity
				ELSE 0
			END
		), 0) AS balance
		FROM stock_movements
		WHERE product_id = ? AND warehouse_id = ? AND created_at < ?
	`, productID, warehouseID, beforeDate).Scan(&balance).Error
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func (r *InventoryRepository) GetStockCard(productID, warehouseID uuid.UUID, fromDate, toDate *time.Time) ([]models.StockMovement, error) {
	var movements []models.StockMovement
	query := r.db.Where("product_id = ? AND warehouse_id = ?", productID, warehouseID)

	if fromDate != nil {
		query = query.Where("created_at >= ?", fromDate)
	}
	if toDate != nil {
		// Treat toDate as an exclusive upper bound (matches TS: created_at < to_date + 1 day)
		query = query.Where("created_at < ?", toDate)
	}

	if err := query.Order("created_at ASC").Find(&movements).Error; err != nil {
		return nil, err
	}
	return movements, nil
}

func (r *InventoryRepository) GetInventoryByProduct(productID uuid.UUID) ([]models.Inventory, error) {
	var inventories []models.Inventory
	if err := r.db.Preload("Warehouse").Where("product_id = ?", productID).Find(&inventories).Error; err != nil {
		return nil, err
	}
	return inventories, nil
}

// Stock Transfer
func (r *InventoryRepository) CreateStockTransfer(transfer *models.StockTransfer) error {
	return r.db.Create(transfer).Error
}

func (r *InventoryRepository) GetStockTransferByID(id uuid.UUID) (*models.StockTransfer, error) {
	var transfer models.StockTransfer
	if err := r.db.Preload("FromWarehouse").Preload("ToWarehouse").Preload("Items").First(&transfer, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &transfer, nil
}

func (r *InventoryRepository) UpdateStockTransfer(transfer *models.StockTransfer) error {
	return r.db.Save(transfer).Error
}

func (r *InventoryRepository) FindStockTransfers(filters map[string]interface{}, limit, offset int) ([]models.StockTransfer, int64, error) {
	var transfers []models.StockTransfer
	var total int64

	query := r.db.Model(&models.StockTransfer{})

	if fromWarehouseID, ok := filters["from_warehouse_id"].(string); ok && fromWarehouseID != "" {
		query = query.Where("from_warehouse_id = ?", fromWarehouseID)
	}
	if toWarehouseID, ok := filters["to_warehouse_id"].(string); ok && toWarehouseID != "" {
		query = query.Where("to_warehouse_id = ?", toWarehouseID)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("FromWarehouse").Preload("ToWarehouse").Preload("Items").Limit(limit).Offset(offset).Order("created_at DESC").Find(&transfers).Error; err != nil {
		return nil, 0, err
	}

	return transfers, total, nil
}

// Stock Opname
func (r *InventoryRepository) CreateStockOpname(opname *models.StockOpname) error {
	return r.db.Create(opname).Error
}

func (r *InventoryRepository) GetStockOpnameByID(id uuid.UUID) (*models.StockOpname, error) {
	var opname models.StockOpname
	if err := r.db.Preload("Warehouse").Preload("User").Preload("Items.Product").Preload("Items.Product.Unit").First(&opname, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &opname, nil
}

func (r *InventoryRepository) UpdateStockOpname(opname *models.StockOpname) error {
	return r.db.Save(opname).Error
}

func (r *InventoryRepository) UpdateStockOpnameWithItems(opname *models.StockOpname, items []models.StockOpnameItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(opname).Error; err != nil {
			return err
		}
		if err := tx.Where("opname_id = ?", opname.ID).Delete(&models.StockOpnameItem{}).Error; err != nil {
			return err
		}
		for i := range items {
			if items[i].ID == uuid.Nil {
				items[i].ID = uuid.New()
			}
			items[i].OpnameID = opname.ID
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *InventoryRepository) FindStockOpnames(filters map[string]interface{}, limit, offset int) ([]models.StockOpname, int64, error) {
	var opnames []models.StockOpname
	var total int64

	query := r.db.Model(&models.StockOpname{})

	if warehouseID, ok := filters["warehouse_id"].(string); ok && warehouseID != "" {
		query = query.Where("warehouse_id = ?", warehouseID)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if fromDate, ok := filters["from_date"].(string); ok && fromDate != "" {
		query = query.Where("opname_date >= ?", fromDate)
	}
	if toDate, ok := filters["to_date"].(string); ok && toDate != "" {
		query = query.Where("opname_date <= ?", toDate)
	}
	if search, ok := filters["search"].(string); ok && search != "" {
		like := "%" + search + "%"
		query = query.Where("opname_number ILIKE ?", like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Warehouse").Preload("User").Limit(limit).Offset(offset).Order("created_at DESC").Find(&opnames).Error; err != nil {
		return nil, 0, err
	}

	return opnames, total, nil
}

func (r *InventoryRepository) GetNextTransferNumber() (string, error) {
	var count int64
	r.db.Model(&models.StockTransfer{}).Count(&count)
	return "TRF-" + time.Now().Format("20060102") + "-" + formatInt(count+1, 3), nil
}

func (r *InventoryRepository) GetNextOpnameNumber() (string, error) {
	var count int64
	r.db.Model(&models.StockOpname{}).Count(&count)
	return "OPN-" + time.Now().Format("20060102") + "-" + formatInt(count+1, 3), nil
}

func (r *InventoryRepository) CreateStockTransferItem(item *models.StockTransferItem) error {
	return r.db.Create(item).Error
}

func (r *InventoryRepository) CreateStockOpnameItem(item *models.StockOpnameItem) error {
	return r.db.Create(item).Error
}

func (r *InventoryRepository) DeleteStockOpname(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("opname_id = ?", id).Delete(&models.StockOpnameItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.StockOpname{}, "id = ?", id).Error
	})
}

func formatInt(n int64, width int) string {
	result := ""
	for n > 0 || width > 0 {
		if n > 0 || width > 0 {
			result = string(rune('0'+n%10)) + result
			n /= 10
		}
		width--
	}
	if result == "" {
		result = "0"
	}
	return result
}
