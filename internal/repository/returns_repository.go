package repository

import (
	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
)

type SalesReturnRow struct {
	models.SalesReturn
	SaleNumber      string `json:"sale_number" gorm:"column:sale_number"`
	WarehouseName   string `json:"warehouse_name" gorm:"column:warehouse_name"`
	ProcessedByName string `json:"processed_by_name" gorm:"column:processed_by_name"`
}

type SalesReturnItemRow struct {
	models.SalesReturnItem
	ProductName string `json:"product_name" gorm:"column:product_name"`
}

type ReturnsRepository struct {
	db *gorm.DB
}

func NewReturnsRepository(db *gorm.DB) *ReturnsRepository {
	return &ReturnsRepository{db: db}
}

func (r *ReturnsRepository) FindReturns(filters map[string]string, limit, offset int) ([]SalesReturnRow, int64, error) {
	var rows []SalesReturnRow
	var total int64

	query := r.db.Table("sales_returns sr").
		Select("sr.*, s.sale_number").
		Joins("LEFT JOIN sales s ON s.id = sr.sale_id")

	if v := filters["warehouse_id"]; v != "" {
		query = query.Where("sr.warehouse_id = ?", v)
	}
	if v := filters["sale_id"]; v != "" {
		query = query.Where("sr.sale_id = ?", v)
	}
	if v := filters["status"]; v != "" {
		query = query.Where("sr.status = ?", v)
	}
	if v := filters["search"]; v != "" {
		like := "%" + v + "%"
		query = query.Where("sr.return_number ILIKE ? OR s.sale_number ILIKE ?", like, like)
	}

	countQuery := query.Session(&gorm.Session{})
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("sr.return_date DESC").Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (r *ReturnsRepository) GetReturnByID(id uuid.UUID) (*SalesReturnRow, error) {
	var row SalesReturnRow

	err := r.db.Table("sales_returns sr").
		Select("sr.*, s.sale_number, w.name as warehouse_name, u.username as processed_by_name").
		Joins("LEFT JOIN sales s ON s.id = sr.sale_id").
		Joins("LEFT JOIN warehouses w ON w.id = sr.warehouse_id").
		Joins("LEFT JOIN users u ON u.id = sr.processed_by").
		Where("sr.id = ?", id).
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

func (r *ReturnsRepository) GetReturnItems(returnID uuid.UUID) ([]SalesReturnItemRow, error) {
	var items []SalesReturnItemRow

	err := r.db.Table("sales_return_items sri").
		Select("sri.*, p.name as product_name").
		Joins("LEFT JOIN products p ON p.id = sri.product_id").
		Where("sri.return_id = ?", returnID).
		Scan(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}
