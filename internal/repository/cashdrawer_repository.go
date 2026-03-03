package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
)

type CashDrawerRow struct {
	ID               uuid.UUID           `json:"id" gorm:"column:id"`
	DrawerNumber     string              `json:"drawer_number" gorm:"column:drawer_number"`
	WarehouseID      uuid.UUID           `json:"warehouse_id" gorm:"column:warehouse_id"`
	CashierID        uuid.UUID           `json:"cashier_id" gorm:"column:cashier_id"`
	CompanyID        uuid.UUID           `json:"company_id" gorm:"column:company_id"`
	Status           models.DrawerStatus `json:"status" gorm:"column:status"`
	OpeningBalance   float64             `json:"opening_balance" gorm:"column:opening_balance"`
	ClosingBalance   *float64            `json:"closing_balance" gorm:"column:closing_balance"`
	ExpectedBalance  float64             `json:"expected_balance" gorm:"column:expected_balance"`
	Variance         *float64            `json:"variance" gorm:"column:variance"`
	OpenedAt         time.Time           `json:"opened_at" gorm:"column:opened_at"`
	ClosedAt         *time.Time          `json:"closed_at" gorm:"column:closed_at"`
	Notes            *string             `json:"notes" gorm:"column:notes"`
	DepositInvoiceID *uuid.UUID          `json:"deposit_invoice_id" gorm:"column:deposit_invoice_id"`
	CreatedAt        time.Time           `json:"created_at" gorm:"column:created_at"`
	UpdatedAt        time.Time           `json:"updated_at" gorm:"column:updated_at"`
	WarehouseName    *string             `json:"warehouse_name" gorm:"column:warehouse_name"`
	CashierName      *string             `json:"cashier_name" gorm:"column:cashier_name"`
}

type CashDrawerTransactionRow struct {
	ID            uuid.UUID              `json:"id" gorm:"column:id"`
	CashDrawerID  uuid.UUID              `json:"cash_drawer_id" gorm:"column:cash_drawer_id"`
	Type          models.TransactionType `json:"type" gorm:"column:type"`
	Amount        float64                `json:"amount" gorm:"column:amount"`
	BalanceAfter  float64                `json:"balance_after" gorm:"column:balance_after"`
	SaleID        *uuid.UUID             `json:"sale_id" gorm:"column:sale_id"`
	PaymentMethod *string                `json:"payment_method" gorm:"column:payment_method"`
	Reason        string                 `json:"reason" gorm:"column:reason"`
	CreatedAt     time.Time              `json:"created_at" gorm:"column:created_at"`
	CreatedBy     uuid.UUID              `json:"created_by" gorm:"column:created_by"`
	CreatedByName *string                `json:"created_by_name" gorm:"column:created_by_name"`
}

type CashDrawerRepository struct {
	db *gorm.DB
}

func NewCashDrawerRepository(db *gorm.DB) *CashDrawerRepository {
	return &CashDrawerRepository{db: db}
}

func (r *CashDrawerRepository) WithDB(db *gorm.DB) *CashDrawerRepository {
	return &CashDrawerRepository{db: db}
}

func (r *CashDrawerRepository) GetCurrentDrawer(warehouseID *uuid.UUID, companyID uuid.UUID, cashierID uuid.UUID) (*CashDrawerRow, error) {
	q := r.db.Table("cash_drawers cd").
		Select("cd.*, w.name AS warehouse_name, u.full_name AS cashier_name").
		Joins("LEFT JOIN warehouses w ON w.id = cd.warehouse_id").
		Joins("LEFT JOIN users u ON u.id = cd.cashier_id").
		Where("cd.company_id = ? AND cd.cashier_id = ? AND cd.status = ?", companyID, cashierID, models.DrawerStatusOpen)
	if warehouseID != nil {
		q = q.Where("cd.warehouse_id = ?", *warehouseID)
	}
	var row CashDrawerRow
	if err := q.Order("cd.opened_at DESC").Limit(1).Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.ID == uuid.Nil {
		return nil, nil
	}
	return &row, nil
}

func (r *CashDrawerRepository) ListCashDrawers(filters map[string]string, limit, offset int, companyID uuid.UUID) ([]CashDrawerRow, int64, error) {
	var rows []CashDrawerRow
	var total int64

	q := r.db.Table("cash_drawers cd").
		Select("cd.*, w.name AS warehouse_name, u.full_name AS cashier_name").
		Joins("LEFT JOIN warehouses w ON w.id = cd.warehouse_id").
		Joins("LEFT JOIN users u ON u.id = cd.cashier_id").
		Where("cd.company_id = ?", companyID)

	if v := strings.TrimSpace(filters["warehouse_id"]); v != "" {
		q = q.Where("cd.warehouse_id = ?", v)
	}
	if v := strings.TrimSpace(filters["cashier_id"]); v != "" {
		q = q.Where("cd.cashier_id = ?", v)
	}
	if v := strings.TrimSpace(filters["status"]); v != "" {
		q = q.Where("cd.status = ?", v)
	}
	if v := strings.TrimSpace(filters["search"]); v != "" {
		like := "%" + v + "%"
		q = q.Where("cd.drawer_number ILIKE ? OR w.name ILIKE ? OR u.full_name ILIKE ?", like, like, like)
	}
	if v := strings.TrimSpace(filters["from_date"]); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			q = q.Where("cd.opened_at >= ?", t)
		}
	}
	if v := strings.TrimSpace(filters["to_date"]); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			q = q.Where("cd.opened_at <= ?", t)
		}
	}

	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("cd.opened_at DESC").Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *CashDrawerRepository) GetCashDrawerByID(id uuid.UUID, companyID uuid.UUID) (*CashDrawerRow, error) {
	var row CashDrawerRow
	err := r.db.Table("cash_drawers cd").
		Select("cd.*, w.name AS warehouse_name, u.full_name AS cashier_name").
		Joins("LEFT JOIN warehouses w ON w.id = cd.warehouse_id").
		Joins("LEFT JOIN users u ON u.id = cd.cashier_id").
		Where("cd.id = ? AND cd.company_id = ?", id, companyID).
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

func (r *CashDrawerRepository) GetDrawerTransactions(drawerID uuid.UUID, txType *string, search string, limit, offset int) ([]CashDrawerTransactionRow, int64, error) {
	var rows []CashDrawerTransactionRow
	var total int64

	q := r.db.Table("cash_drawer_transactions cdt").
		Select("cdt.*, u.full_name AS created_by_name").
		Joins("LEFT JOIN users u ON u.id = cdt.created_by").
		Where("cdt.cash_drawer_id = ?", drawerID)
	if txType != nil && strings.TrimSpace(*txType) != "" {
		q = q.Where("cdt.type = ?", strings.TrimSpace(*txType))
	}
	if strings.TrimSpace(search) != "" {
		like := "%" + strings.TrimSpace(search) + "%"
		q = q.Where("cdt.reason ILIKE ? OR u.full_name ILIKE ?", like, like)
	}

	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("cdt.created_at DESC").Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *CashDrawerRepository) GetAllDrawerTransactions(drawerID uuid.UUID) ([]CashDrawerTransactionRow, error) {
	var rows []CashDrawerTransactionRow
	err := r.db.Table("cash_drawer_transactions cdt").
		Select("cdt.*, u.full_name AS created_by_name").
		Joins("LEFT JOIN users u ON u.id = cdt.created_by").
		Where("cdt.cash_drawer_id = ?", drawerID).
		Order("cdt.created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *CashDrawerRepository) DrawerExistsOpenForCashierWarehouse(companyID uuid.UUID, cashierID uuid.UUID, warehouseID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Table("cash_drawers").
		Where("company_id = ? AND cashier_id = ? AND warehouse_id = ? AND status = ?", companyID, cashierID, warehouseID, models.DrawerStatusOpen).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *CashDrawerRepository) CreateDrawerNumber() string {
	return fmt.Sprintf("CD-%d", time.Now().UnixMilli())
}
