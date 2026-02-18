package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
)

type InventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

func (r *InventoryRepository) FindAll(filters map[string]interface{}, limit, offset int) ([]models.Inventory, int64, error) {
	var inventories []models.Inventory
	var total int64

	query := r.db.Model(&models.Inventory{})

	if warehouseID, ok := filters["warehouse_id"].(string); ok && warehouseID != "" {
		query = query.Where("warehouse_id = ?", warehouseID)
	}
	if productID, ok := filters["product_id"].(string); ok && productID != "" {
		query = query.Where("product_id = ?", productID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Product").Preload("Warehouse").Limit(limit).Offset(offset).Find(&inventories).Error; err != nil {
		return nil, 0, err
	}

	return inventories, total, nil
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

func (r *InventoryRepository) GetStockCard(productID, warehouseID uuid.UUID, fromDate, toDate *time.Time) ([]models.StockMovement, error) {
	var movements []models.StockMovement
	query := r.db.Where("product_id = ? AND warehouse_id = ?", productID, warehouseID)

	if fromDate != nil {
		query = query.Where("created_at >= ?", fromDate)
	}
	if toDate != nil {
		query = query.Where("created_at <= ?", toDate)
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
	if err := r.db.Preload("Warehouse").Preload("Items").First(&opname, "id = ?", id).Error; err != nil {
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

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Warehouse").Limit(limit).Offset(offset).Order("created_at DESC").Find(&opnames).Error; err != nil {
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
