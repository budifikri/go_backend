package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CustomerRow struct {
	ID                uuid.UUID  `json:"id" gorm:"column:id"`
	CustomerCode      string     `json:"customer_code" gorm:"column:customer_code"`
	Name              string     `json:"name" gorm:"column:name"`
	Email             *string    `json:"email" gorm:"column:email"`
	Phone             *string    `json:"phone" gorm:"column:phone"`
	Address           *string    `json:"address" gorm:"column:address"`
	City              *string    `json:"city" gorm:"column:city"`
	Tier              string     `json:"tier" gorm:"column:tier"`
	Status            string     `json:"status" gorm:"column:status"`
	LoyaltyPoints     int        `json:"loyalty_points" gorm:"column:loyalty_points"`
	CreditLimit       string     `json:"credit_limit" gorm:"column:credit_limit"`
	CreditBalance     string     `json:"credit_balance" gorm:"column:credit_balance"`
	TotalPurchases    string     `json:"total_purchases" gorm:"column:total_purchases"`
	LastPurchaseDate  *time.Time `json:"lastPurchaseDate" gorm:"column:lastPurchaseDate"`
	BankName          *string    `json:"bank_name" gorm:"column:bank_name"`
	BankAccountNumber *string    `json:"bank_account_number" gorm:"column:bank_account_number"`
	BankAccountName   *string    `json:"bank_account_name" gorm:"column:bank_account_name"`
	BankBranch        *string    `json:"bank_branch" gorm:"column:bank_branch"`
	CreatedAt         time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"column:updated_at"`
	CompanyID         uuid.UUID  `json:"company_id" gorm:"column:company_id"`
}

type CustomerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) FindCustomers(filters map[string]interface{}, limit, offset int, companyID uuid.UUID) ([]CustomerRow, int64, error) {
	var rows []CustomerRow
	var total int64

	base := r.db.Table("customers c").Where("c.company_id = ?", companyID)

	if v, ok := filters["tier"].(string); ok && v != "" {
		base = base.Where("c.tier = ?", v)
	}
	if v, ok := filters["status"].(string); ok && v != "" {
		base = base.Where("c.status = ?", v)
	}
	if v, ok := filters["search"].(string); ok && v != "" {
		like := fmt.Sprintf("%%%s%%", v)
		base = base.Where("(c.name ILIKE ? OR c.email ILIKE ? OR c.phone ILIKE ?)", like, like, like)
	}
	if v, ok := filters["min_loyalty_points"].(int); ok {
		base = base.Where("c.loyalty_points >= ?", v)
	}
	if v, ok := filters["max_loyalty_points"].(int); ok {
		base = base.Where("c.loyalty_points <= ?", v)
	}

	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	selectClause := "c.id, c.customer_code, c.name, c.email, c.phone, c.address, c.city, c.tier, c.status, c.loyalty_points, c.credit_limit, c.credit_balance, c.total_purchases, c.\"lastPurchaseDate\", c.bank_name, c.bank_account_number, c.bank_account_name, c.bank_branch, c.created_at, c.updated_at, c.company_id"
	if err := base.Select(selectClause).Order("c.created_at DESC").Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (r *CustomerRepository) GetCustomerByID(id uuid.UUID, companyID uuid.UUID) (*CustomerRow, error) {
	var row CustomerRow

	selectClause := "c.id, c.customer_code, c.name, c.email, c.phone, c.address, c.city, c.tier, c.status, c.loyalty_points, c.credit_limit, c.credit_balance, c.total_purchases, c.\"lastPurchaseDate\", c.bank_name, c.bank_account_number, c.bank_account_name, c.bank_branch, c.created_at, c.updated_at, c.company_id"
	err := r.db.Table("customers c").
		Select(selectClause).
		Where("c.id = ? AND c.company_id = ?", id, companyID).
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

func (r *CustomerRepository) CreateCustomer(customer map[string]interface{}) (*CustomerRow, error) {
	var row CustomerRow
	err := r.db.Table("customers").Clauses(clause.Returning{}).Create(customer).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *CustomerRepository) UpdateCustomer(id uuid.UUID, companyID uuid.UUID, updates map[string]interface{}) (*CustomerRow, error) {
	if len(updates) == 0 {
		return nil, nil
	}
	updates["updated_at"] = time.Now()

	err := r.db.Table("customers").
		Where("id = ? AND company_id = ?", id, companyID).
		Updates(updates).Error
	if err != nil {
		return nil, err
	}

	return r.GetCustomerByID(id, companyID)
}

func (r *CustomerRepository) DeactivateCustomer(id uuid.UUID, companyID uuid.UUID) error {
	return r.db.Table("customers").Where("id = ? AND company_id = ?", id, companyID).
		Updates(map[string]interface{}{"status": "inactive", "updated_at": time.Now()}).Error
}

func (r *CustomerRepository) DeleteCustomer(id uuid.UUID, companyID uuid.UUID) (int64, error) {
	res := r.db.Exec("DELETE FROM customers WHERE id = ? AND company_id = ?", id, companyID)
	return res.RowsAffected, res.Error
}
