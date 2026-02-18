package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SupplierRow struct {
	ID            uuid.UUID `json:"id" gorm:"column:id"`
	Code          string    `json:"code" gorm:"column:code"`
	Name          string    `json:"name" gorm:"column:name"`
	ContactPerson *string   `json:"contact_person" gorm:"column:contact_person"`
	Email         *string   `json:"email" gorm:"column:email"`
	Phone         *string   `json:"phone" gorm:"column:phone"`
	Address       *string   `json:"address" gorm:"column:address"`
	City          *string   `json:"city" gorm:"column:city"`
	TaxID         *string   `json:"tax_id" gorm:"column:tax_id"`
	PaymentTerms  string    `json:"payment_terms" gorm:"column:payment_terms"`
	CreditLimit   string    `json:"credit_limit" gorm:"column:credit_limit"`
	Status        string    `json:"status" gorm:"column:status"`
	Notes         *string   `json:"notes" gorm:"column:notes"`
	CompanyID     uuid.UUID `json:"company_id" gorm:"column:company_id"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"column:updated_at"`
}

type SupplierRepository struct {
	db *gorm.DB
}

func NewSupplierRepository(db *gorm.DB) *SupplierRepository {
	return &SupplierRepository{db: db}
}

func (r *SupplierRepository) FindSuppliers(filters map[string]interface{}, limit, offset int) ([]SupplierRow, int64, error) {
	var rows []SupplierRow
	var total int64

	base := r.db.Table("suppliers s")

	if v, ok := filters["status"].(string); ok && v != "" {
		base = base.Where("s.status = ?", v)
	}
	if v, ok := filters["payment_terms"].(string); ok && v != "" {
		base = base.Where("s.payment_terms = ?", v)
	}
	if v, ok := filters["search"].(string); ok && v != "" {
		like := fmt.Sprintf("%%%s%%", v)
		base = base.Where("(s.name ILIKE ? OR s.email ILIKE ? OR s.phone ILIKE ? OR s.contact_person ILIKE ?)", like, like, like, like)
	}
	if v, ok := filters["min_credit_limit"].(int); ok {
		base = base.Where("CAST(s.credit_limit AS DECIMAL) >= ?", v)
	}
	if v, ok := filters["max_credit_limit"].(int); ok {
		base = base.Where("CAST(s.credit_limit AS DECIMAL) <= ?", v)
	}
	if v, ok := filters["company_id"].(string); ok && v != "" {
		base = base.Where("s.company_id = ?", v)
	}

	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	selectClause := "s.id, s.code, s.name, s.contact_person, s.email, s.phone, s.address, s.city, s.tax_id, s.payment_terms, s.credit_limit, s.status, s.notes, s.company_id, s.created_at, s.updated_at"
	if err := base.Select(selectClause).Order("s.created_at DESC").Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (r *SupplierRepository) GetSupplierByID(id uuid.UUID, companyID *uuid.UUID) (*SupplierRow, error) {
	var row SupplierRow
	query := r.db.Table("suppliers s").
		Select("s.id, s.code, s.name, s.contact_person, s.email, s.phone, s.address, s.city, s.tax_id, s.payment_terms, s.credit_limit, s.status, s.notes, s.company_id, s.created_at, s.updated_at").
		Where("s.id = ?", id)
	if companyID != nil {
		query = query.Where("s.company_id = ?", *companyID)
	}
	if err := query.Limit(1).Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.ID == uuid.Nil {
		return nil, nil
	}
	return &row, nil
}

func (r *SupplierRepository) CreateSupplier(supplier map[string]interface{}) (*SupplierRow, error) {
	var row SupplierRow
	err := r.db.Table("suppliers").Clauses(clause.Returning{}).Create(supplier).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *SupplierRepository) UpdateSupplier(id uuid.UUID, companyID uuid.UUID, updates map[string]interface{}) (*SupplierRow, error) {
	if len(updates) == 0 {
		return nil, nil
	}
	updates["updated_at"] = time.Now()

	err := r.db.Table("suppliers").
		Where("id = ? AND company_id = ?", id, companyID).
		Updates(updates).Error
	if err != nil {
		return nil, err
	}

	return r.GetSupplierByID(id, &companyID)
}

func (r *SupplierRepository) DeactivateSupplier(id uuid.UUID, companyID uuid.UUID) error {
	return r.db.Table("suppliers").
		Where("id = ? AND company_id = ?", id, companyID).
		Updates(map[string]interface{}{"status": "inactive", "updated_at": time.Now()}).Error
}
