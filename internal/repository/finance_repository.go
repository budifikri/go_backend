package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
)

func parseFlexibleTime(v string) (time.Time, error) {
	vs := strings.TrimSpace(v)
	if vs == "" {
		return time.Time{}, fmt.Errorf("empty time")
	}
	if t, err := time.Parse(time.RFC3339, vs); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", vs); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02 15:04:05", vs); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid time format")
}

type IncomingInvoiceListRow struct {
	ID             uuid.UUID            `json:"id" gorm:"column:id"`
	InvoiceNumber  string               `json:"invoice_number" gorm:"column:invoice_number"`
	SupplierID     uuid.UUID            `json:"supplier_id" gorm:"column:supplier_id"`
	CompanyID      uuid.UUID            `json:"company_id" gorm:"column:company_id"`
	InvoiceDate    time.Time            `json:"invoice_date" gorm:"column:invoice_date"`
	DueDate        *time.Time           `json:"due_date" gorm:"column:due_date"`
	Status         models.InvoiceStatus `json:"status" gorm:"column:status"`
	Subtotal       float64              `json:"subtotal" gorm:"column:subtotal"`
	TaxAmount      float64              `json:"tax_amount" gorm:"column:tax_amount"`
	DiscountAmount float64              `json:"discount_amount" gorm:"column:discount_amount"`
	TotalAmount    float64              `json:"total_amount" gorm:"column:total_amount"`
	PaidAmount     float64              `json:"paid_amount" gorm:"column:paid_amount"`
	BalanceDue     float64              `json:"balance_due" gorm:"column:balance_due"`
	Notes          *string              `json:"notes" gorm:"column:notes"`
	AttachmentPath *string              `json:"attachment_path" gorm:"column:attachment_path"`
	CreatedBy      uuid.UUID            `json:"created_by" gorm:"column:created_by"`
	CreatedAt      time.Time            `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      time.Time            `json:"updated_at" gorm:"column:updated_at"`
	SupplierName   *string              `json:"supplier_name" gorm:"column:supplier_name"`
	CreatedByName  *string              `json:"created_by_name" gorm:"column:created_by_name"`
}

type OutgoingInvoiceListRow struct {
	ID             uuid.UUID            `json:"id" gorm:"column:id"`
	InvoiceNumber  string               `json:"invoice_number" gorm:"column:invoice_number"`
	CustomerID     *uuid.UUID           `json:"customer_id" gorm:"column:customer_id"`
	CompanyID      uuid.UUID            `json:"company_id" gorm:"column:company_id"`
	InvoiceDate    time.Time            `json:"invoice_date" gorm:"column:invoice_date"`
	DueDate        *time.Time           `json:"due_date" gorm:"column:due_date"`
	Status         models.InvoiceStatus `json:"status" gorm:"column:status"`
	Subtotal       float64              `json:"subtotal" gorm:"column:subtotal"`
	TaxAmount      float64              `json:"tax_amount" gorm:"column:tax_amount"`
	DiscountAmount float64              `json:"discount_amount" gorm:"column:discount_amount"`
	TotalAmount    float64              `json:"total_amount" gorm:"column:total_amount"`
	PaidAmount     float64              `json:"paid_amount" gorm:"column:paid_amount"`
	BalanceDue     float64              `json:"balance_due" gorm:"column:balance_due"`
	Notes          *string              `json:"notes" gorm:"column:notes"`
	AttachmentPath *string              `json:"attachment_path" gorm:"column:attachment_path"`
	CreatedBy      uuid.UUID            `json:"created_by" gorm:"column:created_by"`
	CreatedAt      time.Time            `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      time.Time            `json:"updated_at" gorm:"column:updated_at"`
	CustomerName   *string              `json:"customer_name" gorm:"column:customer_name"`
	CreatedByName  *string              `json:"created_by_name" gorm:"column:created_by_name"`
}

type InvoiceItemRow struct {
	ID             uuid.UUID          `json:"id" gorm:"column:id"`
	InvoiceType    models.InvoiceType `json:"invoice_type" gorm:"column:invoice_type"`
	InvoiceID      uuid.UUID          `json:"invoice_id" gorm:"column:invoice_id"`
	ProductID      *uuid.UUID         `json:"product_id" gorm:"column:product_id"`
	Description    string             `json:"description" gorm:"column:description"`
	Quantity       int                `json:"quantity" gorm:"column:quantity"`
	UnitPrice      float64            `json:"unit_price" gorm:"column:unit_price"`
	DiscountRate   float64            `json:"discount_rate" gorm:"column:discount_rate"`
	DiscountAmount float64            `json:"discount_amount" gorm:"column:discount_amount"`
	TaxRate        float64            `json:"tax_rate" gorm:"column:tax_rate"`
	TaxAmount      float64            `json:"tax_amount" gorm:"column:tax_amount"`
	LineTotal      float64            `json:"line_total" gorm:"column:line_total"`
	CreatedAt      time.Time          `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      time.Time          `json:"updated_at" gorm:"column:updated_at"`
	ProductName    *string            `json:"product_name" gorm:"column:product_name"`
}

type InvoicePaymentRow struct {
	ID              uuid.UUID                   `json:"id" gorm:"column:id"`
	InvoiceType     models.InvoiceType          `json:"invoice_type" gorm:"column:invoice_type"`
	InvoiceID       uuid.UUID                   `json:"invoice_id" gorm:"column:invoice_id"`
	PaymentDate     time.Time                   `json:"payment_date" gorm:"column:payment_date"`
	Amount          float64                     `json:"amount" gorm:"column:amount"`
	PaymentMethod   models.PaymentMethodFinance `json:"payment_method" gorm:"column:payment_method"`
	ReferenceNumber *string                     `json:"reference_number" gorm:"column:reference_number"`
	Notes           *string                     `json:"notes" gorm:"column:notes"`
	CreatedBy       uuid.UUID                   `json:"created_by" gorm:"column:created_by"`
	CreatedAt       time.Time                   `json:"created_at" gorm:"column:created_at"`
}

type FinanceRepository struct {
	db *gorm.DB
}

func NewFinanceRepository(db *gorm.DB) *FinanceRepository {
	return &FinanceRepository{db: db}
}

func (r *FinanceRepository) WithDB(db *gorm.DB) *FinanceRepository {
	return &FinanceRepository{db: db}
}

func (r *FinanceRepository) FindIncomingInvoices(filters map[string]string, limit, offset int, companyID uuid.UUID) ([]IncomingInvoiceListRow, int64, error) {
	var rows []IncomingInvoiceListRow
	var total int64

	q := r.db.Table("invoices_incoming ii").
		Select(`ii.id, ii.invoice_number, ii.supplier_id, ii.company_id, ii.invoice_date, ii.due_date, ii.status,
			ii.subtotal, ii.tax_amount, ii.discount_amount, ii.total_amount, ii.paid_amount, ii.balance_due,
			ii.notes, ii.attachment_path, ii.created_by, ii.created_at, ii.updated_at,
			s.name AS supplier_name, u.full_name AS created_by_name`).
		Joins("LEFT JOIN suppliers s ON s.id = ii.supplier_id").
		Joins("LEFT JOIN users u ON u.id = ii.created_by").
		Where("ii.company_id = ?", companyID)

	if v := strings.TrimSpace(filters["status"]); v != "" {
		q = q.Where("ii.status = ?", v)
	}
	if v := strings.TrimSpace(filters["supplier_id"]); v != "" {
		q = q.Where("ii.supplier_id = ?", v)
	}
	if v := strings.TrimSpace(filters["from_date"]); v != "" {
		if t, err := parseFlexibleTime(v); err == nil {
			q = q.Where("ii.invoice_date >= ?", t)
		}
	}
	if v := strings.TrimSpace(filters["to_date"]); v != "" {
		if t, err := parseFlexibleTime(v); err == nil {
			q = q.Where("ii.invoice_date <= ?", t)
		}
	}
	if v := strings.TrimSpace(filters["search"]); v != "" {
		like := fmt.Sprintf("%%%s%%", v)
		q = q.Where("(ii.invoice_number ILIKE ? OR ii.notes ILIKE ?)", like, like)
	}

	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("ii.created_at DESC").Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *FinanceRepository) FindOutgoingInvoices(filters map[string]string, limit, offset int, companyID uuid.UUID) ([]OutgoingInvoiceListRow, int64, error) {
	var rows []OutgoingInvoiceListRow
	var total int64

	q := r.db.Table("invoices_outgoing oi").
		Select(`oi.id, oi.invoice_number, oi.customer_id, oi.company_id, oi.invoice_date, oi.due_date, oi.status,
			oi.subtotal, oi.tax_amount, oi.discount_amount, oi.total_amount, oi.paid_amount, oi.balance_due,
			oi.notes, oi.attachment_path, oi.created_by, oi.created_at, oi.updated_at,
			c.name AS customer_name, u.full_name AS created_by_name`).
		Joins("LEFT JOIN customers c ON c.id = oi.customer_id").
		Joins("LEFT JOIN users u ON u.id = oi.created_by").
		Where("oi.company_id = ?", companyID)

	if v := strings.TrimSpace(filters["status"]); v != "" {
		q = q.Where("oi.status = ?", v)
	}
	if v := strings.TrimSpace(filters["customer_id"]); v != "" {
		q = q.Where("oi.customer_id = ?", v)
	}
	if v := strings.TrimSpace(filters["from_date"]); v != "" {
		if t, err := parseFlexibleTime(v); err == nil {
			q = q.Where("oi.invoice_date >= ?", t)
		}
	}
	if v := strings.TrimSpace(filters["to_date"]); v != "" {
		if t, err := parseFlexibleTime(v); err == nil {
			q = q.Where("oi.invoice_date <= ?", t)
		}
	}
	if v := strings.TrimSpace(filters["search"]); v != "" {
		like := fmt.Sprintf("%%%s%%", v)
		q = q.Where("(oi.invoice_number ILIKE ? OR oi.notes ILIKE ?)", like, like)
	}

	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("oi.created_at DESC").Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *FinanceRepository) GetIncomingInvoiceByID(id uuid.UUID, companyID uuid.UUID) (*IncomingInvoiceListRow, error) {
	var row IncomingInvoiceListRow
	err := r.db.Table("invoices_incoming ii").
		Select(`ii.id, ii.invoice_number, ii.supplier_id, ii.company_id, ii.invoice_date, ii.due_date, ii.status,
			ii.subtotal, ii.tax_amount, ii.discount_amount, ii.total_amount, ii.paid_amount, ii.balance_due,
			ii.notes, ii.attachment_path, ii.created_by, ii.created_at, ii.updated_at,
			s.name AS supplier_name, u.full_name AS created_by_name`).
		Joins("LEFT JOIN suppliers s ON s.id = ii.supplier_id").
		Joins("LEFT JOIN users u ON u.id = ii.created_by").
		Where("ii.id = ? AND ii.company_id = ?", id, companyID).
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

func (r *FinanceRepository) GetOutgoingInvoiceByID(id uuid.UUID, companyID uuid.UUID) (*OutgoingInvoiceListRow, error) {
	var row OutgoingInvoiceListRow
	err := r.db.Table("invoices_outgoing oi").
		Select(`oi.id, oi.invoice_number, oi.customer_id, oi.company_id, oi.invoice_date, oi.due_date, oi.status,
			oi.subtotal, oi.tax_amount, oi.discount_amount, oi.total_amount, oi.paid_amount, oi.balance_due,
			oi.notes, oi.attachment_path, oi.created_by, oi.created_at, oi.updated_at,
			c.name AS customer_name, u.full_name AS created_by_name`).
		Joins("LEFT JOIN customers c ON c.id = oi.customer_id").
		Joins("LEFT JOIN users u ON u.id = oi.created_by").
		Where("oi.id = ? AND oi.company_id = ?", id, companyID).
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

func (r *FinanceRepository) GetInvoiceItems(invoiceID uuid.UUID, invoiceType models.InvoiceType) ([]InvoiceItemRow, error) {
	var items []InvoiceItemRow
	err := r.db.Table("invoice_items it").
		Select("it.*, p.name AS product_name").
		Joins("LEFT JOIN products p ON p.id = it.product_id").
		Where("it.invoice_id = ? AND it.invoice_type = ?", invoiceID, invoiceType).
		Order("it.created_at").
		Scan(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *FinanceRepository) GetInvoicePayments(invoiceID uuid.UUID, invoiceType models.InvoiceType) ([]InvoicePaymentRow, error) {
	var payments []InvoicePaymentRow
	err := r.db.Table("invoice_payments ip").
		Where("ip.invoice_id = ? AND ip.invoice_type = ?", invoiceID, invoiceType).
		Order("ip.payment_date DESC").
		Scan(&payments).Error
	if err != nil {
		return nil, err
	}
	return payments, nil
}
