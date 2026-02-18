package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
)

type SaleWithNames struct {
	models.Sale
	WarehouseName         string `json:"warehouse_name" gorm:"column:warehouse_name"`
	CashierName           string `json:"cashier_name" gorm:"column:cashier_name"`
	CustomerName          string `json:"customer_name" gorm:"column:customer_name"`
	CustomerLoyaltyPoints int    `json:"customer_loyalty_points" gorm:"column:customer_loyalty_points"`
}

type SaleItemWithProduct struct {
	models.SaleItem
	ProductName string `json:"product_name" gorm:"column:product_name"`
}

type SalesRepository struct {
	db *gorm.DB
}

func NewSalesRepository(db *gorm.DB) *SalesRepository {
	return &SalesRepository{db: db}
}

func (r *SalesRepository) FindSales(filters map[string]string, limit, offset int) ([]SaleWithNames, int64, error) {
	var rows []SaleWithNames
	var total int64

	query := r.db.Table("sales s").
		Select("s.*, w.name as warehouse_name, u.username as cashier_name, c.name as customer_name, c.loyalty_points as customer_loyalty_points").
		Joins("LEFT JOIN warehouses w ON w.id = s.warehouse_id").
		Joins("LEFT JOIN users u ON u.id = s.cashier_id").
		Joins("LEFT JOIN customers c ON c.id = s.customer_id")

	if v := filters["warehouse_id"]; v != "" {
		query = query.Where("s.warehouse_id = ?", v)
	}
	if v := filters["customer_id"]; v != "" {
		query = query.Where("s.customer_id = ?", v)
	}
	if v := filters["cashier_id"]; v != "" {
		query = query.Where("s.cashier_id = ?", v)
	}
	if v := filters["status"]; v != "" {
		query = query.Where("s.status = ?", v)
	}
	if v := filters["sale_number"]; v != "" {
		query = query.Where("s.sale_number ILIKE ?", "%"+v+"%")
	}
	if v := filters["date_from"]; v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			query = query.Where("s.sale_date >= ?", t)
		}
	}
	if v := filters["date_to"]; v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			query = query.Where("s.sale_date <= ?", t)
		}
	}

	countQuery := query.Session(&gorm.Session{})
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("s.sale_date DESC").Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (r *SalesRepository) GetSaleByID(id uuid.UUID) (*SaleWithNames, error) {
	var row SaleWithNames

	err := r.db.Table("sales s").
		Select("s.*, w.name as warehouse_name, u.username as cashier_name, c.name as customer_name, c.loyalty_points as customer_loyalty_points").
		Joins("LEFT JOIN warehouses w ON w.id = s.warehouse_id").
		Joins("LEFT JOIN users u ON u.id = s.cashier_id").
		Joins("LEFT JOIN customers c ON c.id = s.customer_id").
		Where("s.id = ?", id).
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

func (r *SalesRepository) GetSaleItems(saleID uuid.UUID) ([]SaleItemWithProduct, error) {
	var items []SaleItemWithProduct
	err := r.db.Table("sale_items si").
		Select("si.*, p.name as product_name").
		Joins("LEFT JOIN products p ON p.id = si.product_id").
		Where("si.sale_id = ?", saleID).
		Scan(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *SalesRepository) GetSalePayments(saleID uuid.UUID) ([]models.SalePayment, error) {
	var payments []models.SalePayment
	err := r.db.Table("sale_payments").Where("sale_id = ?", saleID).Find(&payments).Error
	if err != nil {
		return nil, err
	}
	return payments, nil
}
