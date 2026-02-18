package repository

import (
	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
)

type ExchangeRow struct {
	models.ItemExchange
	WarehouseName   string `json:"warehouse_name" gorm:"column:warehouse_name"`
	ProcessedByName string `json:"processed_by_name" gorm:"column:processed_by_name"`
}

type ExchangeItemRow struct {
	models.ExchangeItem
	ProductName string `json:"product_name" gorm:"column:product_name"`
}

type ExchangesRepository struct {
	db *gorm.DB
}

func NewExchangesRepository(db *gorm.DB) *ExchangesRepository {
	return &ExchangesRepository{db: db}
}

func (r *ExchangesRepository) FindExchanges(filters map[string]string, limit, offset int) ([]ExchangeRow, int64, error) {
	var rows []ExchangeRow
	var total int64

	query := r.db.Table("item_exchanges ie").
		Select("ie.*, w.name as warehouse_name").
		Joins("LEFT JOIN warehouses w ON w.id = ie.warehouse_id")

	if v := filters["warehouse_id"]; v != "" {
		query = query.Where("ie.warehouse_id = ?", v)
	}
	if v := filters["sale_id"]; v != "" {
		query = query.Where("ie.sale_id = ?", v)
	}
	if v := filters["status"]; v != "" {
		query = query.Where("ie.status = ?", v)
	}

	countQuery := query.Session(&gorm.Session{})
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("ie.exchange_date DESC").Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (r *ExchangesRepository) GetExchangeByID(id uuid.UUID) (*ExchangeRow, error) {
	var row ExchangeRow
	err := r.db.Table("item_exchanges ie").
		Select("ie.*, w.name as warehouse_name, u.username as processed_by_name").
		Joins("LEFT JOIN warehouses w ON w.id = ie.warehouse_id").
		Joins("LEFT JOIN users u ON u.id = ie.processed_by").
		Where("ie.id = ?", id).
		Limit(1).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == uuid.Nil {
		return nil, nil
	}
	return &row, nil
}

func (r *ExchangesRepository) GetExchangeItems(exchangeID uuid.UUID, itemType string) ([]ExchangeItemRow, error) {
	var items []ExchangeItemRow
	err := r.db.Table("exchange_items ei").
		Select("ei.*, p.name as product_name").
		Joins("LEFT JOIN products p ON p.id = ei.product_id").
		Where("ei.exchange_id = ? AND ei.item_type = ?", exchangeID, itemType).
		Scan(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}
