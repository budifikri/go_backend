package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
)

type SaleWithNames struct {
	models.Sale
	WarehouseName         string  `json:"warehouse_name" gorm:"column:warehouse_name"`
	CashierName           string  `json:"cashier_name" gorm:"column:cashier_name"`
	CustomerName          string  `json:"customer_name" gorm:"column:customer_name"`
	CustomerLoyaltyPoints int     `json:"customer_loyalty_points" gorm:"column:customer_loyalty_points"`
	TotalProfit           float64 `json:"total_profit" gorm:"column:total_profit"`
}

type SaleItemWithProduct struct {
	models.SaleItem
	ProductName string `json:"product_name" gorm:"column:product_name"`
	UnitName    string `json:"unit_name" gorm:"column:unit_name"`
}

type SaleItemWithProfit struct {
	models.SaleItem
	ProductName string  `json:"product_name" gorm:"column:product_name"`
	UnitName    string  `json:"unit_name" gorm:"column:unit_name"`
	Profit      float64 `json:"profit,omitempty"`
}

type SalesRepository struct {
	db *gorm.DB
}

type SalesSummary struct {
	TotalRows      int64   `json:"total_rows" gorm:"column:total_rows"`
	TotalPenjualan float64 `json:"total_penjualan" gorm:"column:total_penjualan"`
	TotalProfit    float64 `json:"total_profit" gorm:"column:total_profit"`
	DoneRows       int64   `json:"done_rows" gorm:"column:done_rows"`
	CancelledRows  int64   `json:"cancelled_rows" gorm:"column:cancelled_rows"`
	RefundedRows   int64   `json:"refunded_rows" gorm:"column:refunded_rows"`
	PendingRows    int64   `json:"pending_rows" gorm:"column:pending_rows"`
}

func NewSalesRepository(db *gorm.DB) *SalesRepository {
	return &SalesRepository{db: db}
}

func (r *SalesRepository) buildSalesFilterQuery(filters map[string]string) *gorm.DB {
	query := r.db.Table("sales s").
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
	if v := filters["cash_drawer_id"]; v != "" {
		query = query.Where("s.cash_drawer_id = ?", v)
	}
	if v := filters["status"]; v != "" {
		query = query.Where("s.status = ?", v)
	}
	if v := filters["sale_number"]; v != "" {
		query = query.Where("s.sale_number ILIKE ?", "%"+v+"%")
	}
	if v := filters["search"]; v != "" {
		like := "%" + v + "%"
		query = query.Where("s.sale_number ILIKE ? OR u.username ILIKE ? OR c.name ILIKE ?", like, like, like)
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

	return query
}

func (r *SalesRepository) buildSalesProfitSubquery() *gorm.DB {
	return r.db.Table("sale_items si").
		Select("si.sale_id, COALESCE(SUM(((COALESCE(si.unit_price, 0) - COALESCE(si.cost_price, 0)) * si.quantity) - COALESCE(si.discount_amount, 0)), 0) as base_profit").
		Group("si.sale_id")
}

func salesProfitCaseExpression() string {
	return `CASE
		WHEN s.status = 'DONE' THEN COALESCE(sp.base_profit, 0)
		WHEN s.status = 'REFUNDED' THEN -COALESCE(sp.base_profit, 0)
		ELSE 0
	END`
}

func (r *SalesRepository) FindSales(filters map[string]string, limit, offset int) ([]SaleWithNames, int64, error) {
	var rows []SaleWithNames
	var total int64
	profitSubquery := r.buildSalesProfitSubquery()

	query := r.buildSalesFilterQuery(filters).
		Joins("LEFT JOIN (?) sp ON sp.sale_id = s.id", profitSubquery).
		Select("s.*, w.name as warehouse_name, u.username as cashier_name, c.name as customer_name, c.loyalty_points as customer_loyalty_points, " + salesProfitCaseExpression() + " as total_profit").
		Session(&gorm.Session{})

	countQuery := query.Session(&gorm.Session{})
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("s.sale_date DESC").Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (r *SalesRepository) GetSalesSummary(filters map[string]string) (SalesSummary, error) {
	var summary SalesSummary
	profitSubquery := r.buildSalesProfitSubquery()

	err := r.buildSalesFilterQuery(filters).
		Joins("LEFT JOIN (?) sp ON sp.sale_id = s.id", profitSubquery).
		Select("COUNT(*) as total_rows, COALESCE(SUM(s.total_amount), 0) as total_penjualan, COALESCE(SUM(" + salesProfitCaseExpression() + "), 0) as total_profit, SUM(CASE WHEN s.status = 'DONE' THEN 1 ELSE 0 END) as done_rows, SUM(CASE WHEN s.status = 'CANCELLED' THEN 1 ELSE 0 END) as cancelled_rows, SUM(CASE WHEN s.status = 'REFUNDED' THEN 1 ELSE 0 END) as refunded_rows, SUM(CASE WHEN s.status = 'PENDING' THEN 1 ELSE 0 END) as pending_rows").
		Scan(&summary).Error

	if err != nil {
		return SalesSummary{}, err
	}

	return summary, nil
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

func (r *SalesRepository) GetSaleItems(saleID uuid.UUID) ([]SaleItemWithProfit, error) {
	var items []SaleItemWithProfit
	err := r.db.Table("sale_items si").
		Select("si.*, p.name as product_name, u.name as unit_name, (COALESCE(si.unit_price,0) - COALESCE(si.cost_price,0)) * si.quantity - COALESCE(si.discount_amount,0) as profit").
		Joins("LEFT JOIN products p ON p.id = si.product_id").
		Joins("LEFT JOIN units_of_measure u ON u.id = p.unit_id").
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
