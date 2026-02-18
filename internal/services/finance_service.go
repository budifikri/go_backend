package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/repository"
	"github.com/pos-retail/go_backend/internal/types/response"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FinanceService struct {
	db   *gorm.DB
	repo *repository.FinanceRepository
}

func NewFinanceService(db *gorm.DB, repo *repository.FinanceRepository) *FinanceService {
	return &FinanceService{db: db, repo: repo}
}

type FinanceInvoiceItemInput struct {
	ProductID    *string
	Description  string
	Quantity     int
	UnitPrice    float64
	DiscountRate *float64
	TaxRate      *float64
}

type CreateIncomingInvoiceInput struct {
	SupplierID     string
	InvoiceDate    string
	DueDate        *string
	Notes          *string
	AttachmentPath *string
	Items          []FinanceInvoiceItemInput
}

type UpdateIncomingInvoiceInput struct {
	InvoiceDate *string
	DueDate     *string
	Notes       *string
	Status      *string
	Items       *[]FinanceInvoiceItemInput
}

type CreateOutgoingInvoiceInput struct {
	CustomerID     *string
	InvoiceDate    string
	DueDate        *string
	Notes          *string
	AttachmentPath *string
	Items          []FinanceInvoiceItemInput
}

type UpdateOutgoingInvoiceInput struct {
	InvoiceDate *string
	DueDate     *string
	Notes       *string
	Status      *string
	Items       *[]FinanceInvoiceItemInput
}

type CreateInvoicePaymentInput struct {
	Amount          float64
	PaymentMethod   string
	ReferenceNumber *string
	Notes           *string
}

func parseFlexibleTime(v string) (time.Time, error) {
	vs := strings.TrimSpace(v)
	if vs == "" {
		return time.Time{}, errors.New("empty time")
	}
	// Common formats from JS clients
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

type incomingInvoiceListRow struct {
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

type outgoingInvoiceListRow struct {
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

type invoiceItemRow struct {
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

type invoicePaymentRow struct {
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

func (s *FinanceService) GetIncomingInvoices(filters map[string]string, limit, offset int, companyID string) response.PaginatedResponse {
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}
	rows, total, err := s.repo.FindIncomingInvoices(filters, limit, offset, companyUUID)
	if err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}
	return response.NewPaginatedResponse(rows, total, limit, offset)
}

func (s *FinanceService) GetOutgoingInvoices(filters map[string]string, limit, offset int, companyID string) response.PaginatedResponse {
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}
	rows, total, err := s.repo.FindOutgoingInvoices(filters, limit, offset, companyUUID)
	if err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}
	return response.NewPaginatedResponse(rows, total, limit, offset)
}

func (s *FinanceService) GetIncomingInvoiceByID(id string, companyID string) response.ApiResponse {
	invoiceID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Incoming invoice not found")
	}
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Incoming invoice not found")
	}

	inv, err := s.repo.GetIncomingInvoiceByID(invoiceID, companyUUID)
	if err != nil || inv == nil {
		return response.NewErrorResponse("Incoming invoice not found")
	}
	items, _ := s.repo.GetInvoiceItems(invoiceID, models.InvoiceTypeIncoming)
	payments, _ := s.repo.GetInvoicePayments(invoiceID, models.InvoiceTypeIncoming)

	data := map[string]interface{}{}
	data["id"] = inv.ID
	data["invoice_number"] = inv.InvoiceNumber
	data["supplier_id"] = inv.SupplierID
	data["company_id"] = inv.CompanyID
	data["invoice_date"] = inv.InvoiceDate
	data["due_date"] = inv.DueDate
	data["status"] = inv.Status
	data["subtotal"] = inv.Subtotal
	data["tax_amount"] = inv.TaxAmount
	data["discount_amount"] = inv.DiscountAmount
	data["total_amount"] = inv.TotalAmount
	data["paid_amount"] = inv.PaidAmount
	data["balance_due"] = inv.BalanceDue
	data["notes"] = inv.Notes
	data["attachment_path"] = inv.AttachmentPath
	data["created_by"] = inv.CreatedBy
	data["created_at"] = inv.CreatedAt
	data["updated_at"] = inv.UpdatedAt
	data["supplier_name"] = inv.SupplierName
	data["created_by_name"] = inv.CreatedByName
	data["items"] = items
	data["payments"] = payments

	return response.NewSuccessResponse(data, "")
}

func (s *FinanceService) GetOutgoingInvoiceByID(id string, companyID string) response.ApiResponse {
	invoiceID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Outgoing invoice not found")
	}
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Outgoing invoice not found")
	}

	inv, err := s.repo.GetOutgoingInvoiceByID(invoiceID, companyUUID)
	if err != nil || inv == nil {
		return response.NewErrorResponse("Outgoing invoice not found")
	}

	items, _ := s.repo.GetInvoiceItems(invoiceID, models.InvoiceTypeOutgoing)
	payments, _ := s.repo.GetInvoicePayments(invoiceID, models.InvoiceTypeOutgoing)

	data := map[string]interface{}{}
	data["id"] = inv.ID
	data["invoice_number"] = inv.InvoiceNumber
	data["customer_id"] = inv.CustomerID
	data["company_id"] = inv.CompanyID
	data["invoice_date"] = inv.InvoiceDate
	data["due_date"] = inv.DueDate
	data["status"] = inv.Status
	data["subtotal"] = inv.Subtotal
	data["tax_amount"] = inv.TaxAmount
	data["discount_amount"] = inv.DiscountAmount
	data["total_amount"] = inv.TotalAmount
	data["paid_amount"] = inv.PaidAmount
	data["balance_due"] = inv.BalanceDue
	data["notes"] = inv.Notes
	data["attachment_path"] = inv.AttachmentPath
	data["created_by"] = inv.CreatedBy
	data["created_at"] = inv.CreatedAt
	data["updated_at"] = inv.UpdatedAt
	data["customer_name"] = inv.CustomerName
	data["created_by_name"] = inv.CreatedByName
	data["items"] = items
	data["payments"] = payments

	return response.NewSuccessResponse(data, "")
}

func (s *FinanceService) CreateIncomingInvoice(input CreateIncomingInvoiceInput, companyID string, userID string) response.ApiResponse {
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Failed to create incoming invoice")
	}
	creatorUUID, err := uuid.Parse(userID)
	if err != nil {
		return response.NewErrorResponse("Failed to create incoming invoice")
	}
	supplierUUID, err := uuid.Parse(input.SupplierID)
	if err != nil {
		return response.NewErrorResponse("Failed to create incoming invoice")
	}
	invDate, err := parseFlexibleTime(input.InvoiceDate)
	if err != nil {
		return response.NewErrorResponse("Failed to create incoming invoice")
	}
	var due *time.Time
	if input.DueDate != nil && strings.TrimSpace(*input.DueDate) != "" {
		dt, err := parseFlexibleTime(*input.DueDate)
		if err == nil {
			due = &dt
		}
	}
	if len(input.Items) == 0 {
		return response.NewErrorResponse("Failed to create incoming invoice")
	}

	invoiceNumber := fmt.Sprintf("INV-IN-%d", time.Now().UnixMilli())

	subtotal := 0.0
	taxAmount := 0.0
	discountAmount := 0.0
	for _, it := range input.Items {
		lineSubtotal := float64(it.Quantity) * it.UnitPrice
		dr := 0.0
		tr := 0.0
		if it.DiscountRate != nil {
			dr = *it.DiscountRate
		}
		if it.TaxRate != nil {
			tr = *it.TaxRate
		}
		lineDiscount := lineSubtotal * dr / 100.0
		lineTax := (lineSubtotal - lineDiscount) * tr / 100.0
		subtotal += lineSubtotal
		discountAmount += lineDiscount
		taxAmount += lineTax
	}
	totalAmount := subtotal - discountAmount + taxAmount

	var created models.IncomingInvoice
	err = s.db.Transaction(func(tx *gorm.DB) error {
		created = models.IncomingInvoice{
			ID:             uuid.New(),
			InvoiceNumber:  invoiceNumber,
			SupplierID:     supplierUUID,
			CompanyID:      companyUUID,
			InvoiceDate:    invDate,
			DueDate:        due,
			Status:         models.InvoiceStatusDraft,
			Subtotal:       subtotal,
			TaxAmount:      taxAmount,
			DiscountAmount: discountAmount,
			TotalAmount:    totalAmount,
			PaidAmount:     0,
			BalanceDue:     totalAmount,
			Notes:          input.Notes,
			AttachmentPath: input.AttachmentPath,
			CreatedBy:      creatorUUID,
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}

		for _, it := range input.Items {
			var prodUUID *uuid.UUID
			if it.ProductID != nil && strings.TrimSpace(*it.ProductID) != "" {
				pid, err := uuid.Parse(*it.ProductID)
				if err == nil {
					prodUUID = &pid
				}
			}
			dr := 0.0
			tr := 0.0
			if it.DiscountRate != nil {
				dr = *it.DiscountRate
			}
			if it.TaxRate != nil {
				tr = *it.TaxRate
			}
			lineSubtotal := float64(it.Quantity) * it.UnitPrice
			lineDiscount := lineSubtotal * dr / 100.0
			lineTax := (lineSubtotal - lineDiscount) * tr / 100.0
			lineTotal := lineSubtotal - lineDiscount + lineTax

			row := models.InvoiceItem{
				ID:             uuid.New(),
				InvoiceType:    models.InvoiceTypeIncoming,
				InvoiceID:      created.ID,
				ProductID:      prodUUID,
				Description:    it.Description,
				Quantity:       it.Quantity,
				UnitPrice:      it.UnitPrice,
				DiscountRate:   dr,
				DiscountAmount: lineDiscount,
				TaxRate:        tr,
				TaxAmount:      lineTax,
				LineTotal:      lineTotal,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return response.NewErrorResponse(err.Error())
	}

	return response.NewSuccessResponse(created, "")
}

func (s *FinanceService) CreateOutgoingInvoice(input CreateOutgoingInvoiceInput, companyID string, userID string) response.ApiResponse {
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Failed to create outgoing invoice")
	}
	creatorUUID, err := uuid.Parse(userID)
	if err != nil {
		return response.NewErrorResponse("Failed to create outgoing invoice")
	}
	invDate, err := parseFlexibleTime(input.InvoiceDate)
	if err != nil {
		return response.NewErrorResponse("Failed to create outgoing invoice")
	}
	var due *time.Time
	if input.DueDate != nil && strings.TrimSpace(*input.DueDate) != "" {
		dt, err := parseFlexibleTime(*input.DueDate)
		if err == nil {
			due = &dt
		}
	}
	if len(input.Items) == 0 {
		return response.NewErrorResponse("Failed to create outgoing invoice")
	}

	invoiceNumber := fmt.Sprintf("INV-OUT-%d", time.Now().UnixMilli())

	subtotal := 0.0
	taxAmount := 0.0
	discountAmount := 0.0
	for _, it := range input.Items {
		lineSubtotal := float64(it.Quantity) * it.UnitPrice
		dr := 0.0
		tr := 0.0
		if it.DiscountRate != nil {
			dr = *it.DiscountRate
		}
		if it.TaxRate != nil {
			tr = *it.TaxRate
		}
		lineDiscount := lineSubtotal * dr / 100.0
		lineTax := (lineSubtotal - lineDiscount) * tr / 100.0
		subtotal += lineSubtotal
		discountAmount += lineDiscount
		taxAmount += lineTax
	}
	totalAmount := subtotal - discountAmount + taxAmount

	var customerUUID *uuid.UUID
	if input.CustomerID != nil && strings.TrimSpace(*input.CustomerID) != "" {
		cid, err := uuid.Parse(*input.CustomerID)
		if err == nil {
			customerUUID = &cid
		}
	}

	var created models.OutgoingInvoice
	err = s.db.Transaction(func(tx *gorm.DB) error {
		created = models.OutgoingInvoice{
			ID:             uuid.New(),
			InvoiceNumber:  invoiceNumber,
			CustomerID:     customerUUID,
			CompanyID:      companyUUID,
			InvoiceDate:    invDate,
			DueDate:        due,
			Status:         models.InvoiceStatusDraft,
			Subtotal:       subtotal,
			TaxAmount:      taxAmount,
			DiscountAmount: discountAmount,
			TotalAmount:    totalAmount,
			PaidAmount:     0,
			BalanceDue:     totalAmount,
			Notes:          input.Notes,
			AttachmentPath: input.AttachmentPath,
			CreatedBy:      creatorUUID,
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}
		for _, it := range input.Items {
			var prodUUID *uuid.UUID
			if it.ProductID != nil && strings.TrimSpace(*it.ProductID) != "" {
				pid, err := uuid.Parse(*it.ProductID)
				if err == nil {
					prodUUID = &pid
				}
			}
			dr := 0.0
			tr := 0.0
			if it.DiscountRate != nil {
				dr = *it.DiscountRate
			}
			if it.TaxRate != nil {
				tr = *it.TaxRate
			}
			lineSubtotal := float64(it.Quantity) * it.UnitPrice
			lineDiscount := lineSubtotal * dr / 100.0
			lineTax := (lineSubtotal - lineDiscount) * tr / 100.0
			lineTotal := lineSubtotal - lineDiscount + lineTax
			row := models.InvoiceItem{
				ID:             uuid.New(),
				InvoiceType:    models.InvoiceTypeOutgoing,
				InvoiceID:      created.ID,
				ProductID:      prodUUID,
				Description:    it.Description,
				Quantity:       it.Quantity,
				UnitPrice:      it.UnitPrice,
				DiscountRate:   dr,
				DiscountAmount: lineDiscount,
				TaxRate:        tr,
				TaxAmount:      lineTax,
				LineTotal:      lineTotal,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return response.NewErrorResponse(err.Error())
	}
	return response.NewSuccessResponse(created, "")
}

func (s *FinanceService) UpdateIncomingInvoice(id string, input UpdateIncomingInvoiceInput, companyID string) response.ApiResponse {
	invoiceID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Incoming invoice not found or cannot be updated")
	}
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Incoming invoice not found or cannot be updated")
	}

	var updated models.IncomingInvoice
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var inv models.IncomingInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&inv, "id = ? AND company_id = ?", invoiceID, companyUUID).Error; err != nil {
			return err
		}
		if inv.Status != models.InvoiceStatusDraft {
			return errors.New("Incoming invoice not found or cannot be updated")
		}

		if input.InvoiceDate != nil && strings.TrimSpace(*input.InvoiceDate) != "" {
			t, err := parseFlexibleTime(*input.InvoiceDate)
			if err == nil {
				inv.InvoiceDate = t
			}
		}
		if input.DueDate != nil {
			if strings.TrimSpace(*input.DueDate) == "" {
				inv.DueDate = nil
			} else {
				t, err := parseFlexibleTime(*input.DueDate)
				if err == nil {
					inv.DueDate = &t
				}
			}
		}
		if input.Notes != nil {
			inv.Notes = input.Notes
		}
		if input.Status != nil && strings.TrimSpace(*input.Status) != "" {
			inv.Status = models.InvoiceStatus(*input.Status)
		}

		// If items provided, replace items and recalc totals.
		if input.Items != nil {
			if err := tx.Where("invoice_id = ? AND invoice_type = ?", invoiceID, models.InvoiceTypeIncoming).Delete(&models.InvoiceItem{}).Error; err != nil {
				return err
			}

			subtotal := 0.0
			taxAmount := 0.0
			discountAmount := 0.0
			for _, it := range *input.Items {
				lineSubtotal := float64(it.Quantity) * it.UnitPrice
				dr := 0.0
				tr := 0.0
				if it.DiscountRate != nil {
					dr = *it.DiscountRate
				}
				if it.TaxRate != nil {
					tr = *it.TaxRate
				}
				lineDiscount := lineSubtotal * dr / 100.0
				lineTax := (lineSubtotal - lineDiscount) * tr / 100.0
				subtotal += lineSubtotal
				discountAmount += lineDiscount
				taxAmount += lineTax
			}
			totalAmount := subtotal - discountAmount + taxAmount
			inv.Subtotal = subtotal
			inv.TaxAmount = taxAmount
			inv.DiscountAmount = discountAmount
			inv.TotalAmount = totalAmount
			inv.BalanceDue = totalAmount - inv.PaidAmount

			for _, it := range *input.Items {
				var prodUUID *uuid.UUID
				if it.ProductID != nil && strings.TrimSpace(*it.ProductID) != "" {
					pid, err := uuid.Parse(*it.ProductID)
					if err == nil {
						prodUUID = &pid
					}
				}
				dr := 0.0
				tr := 0.0
				if it.DiscountRate != nil {
					dr = *it.DiscountRate
				}
				if it.TaxRate != nil {
					tr = *it.TaxRate
				}
				lineSubtotal := float64(it.Quantity) * it.UnitPrice
				lineDiscount := lineSubtotal * dr / 100.0
				lineTax := (lineSubtotal - lineDiscount) * tr / 100.0
				lineTotal := lineSubtotal - lineDiscount + lineTax
				row := models.InvoiceItem{
					ID:             uuid.New(),
					InvoiceType:    models.InvoiceTypeIncoming,
					InvoiceID:      invoiceID,
					ProductID:      prodUUID,
					Description:    it.Description,
					Quantity:       it.Quantity,
					UnitPrice:      it.UnitPrice,
					DiscountRate:   dr,
					DiscountAmount: lineDiscount,
					TaxRate:        tr,
					TaxAmount:      lineTax,
					LineTotal:      lineTotal,
				}
				if err := tx.Create(&row).Error; err != nil {
					return err
				}
			}
		}

		if err := tx.Save(&inv).Error; err != nil {
			return err
		}
		updated = inv
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewErrorResponse("Incoming invoice not found or cannot be updated")
		}
		return response.NewErrorResponse(err.Error())
	}
	return response.NewSuccessResponse(updated, "")
}

func (s *FinanceService) UpdateOutgoingInvoice(id string, input UpdateOutgoingInvoiceInput, companyID string) response.ApiResponse {
	invoiceID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Outgoing invoice not found or cannot be updated")
	}
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Outgoing invoice not found or cannot be updated")
	}

	var updated models.OutgoingInvoice
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var inv models.OutgoingInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&inv, "id = ? AND company_id = ?", invoiceID, companyUUID).Error; err != nil {
			return err
		}
		if inv.Status != models.InvoiceStatusDraft {
			return errors.New("Outgoing invoice not found or cannot be updated")
		}

		if input.InvoiceDate != nil && strings.TrimSpace(*input.InvoiceDate) != "" {
			t, err := parseFlexibleTime(*input.InvoiceDate)
			if err == nil {
				inv.InvoiceDate = t
			}
		}
		if input.DueDate != nil {
			if strings.TrimSpace(*input.DueDate) == "" {
				inv.DueDate = nil
			} else {
				t, err := parseFlexibleTime(*input.DueDate)
				if err == nil {
					inv.DueDate = &t
				}
			}
		}
		if input.Notes != nil {
			inv.Notes = input.Notes
		}
		if input.Status != nil && strings.TrimSpace(*input.Status) != "" {
			inv.Status = models.InvoiceStatus(*input.Status)
		}

		if input.Items != nil {
			if err := tx.Where("invoice_id = ? AND invoice_type = ?", invoiceID, models.InvoiceTypeOutgoing).Delete(&models.InvoiceItem{}).Error; err != nil {
				return err
			}
			subtotal := 0.0
			taxAmount := 0.0
			discountAmount := 0.0
			for _, it := range *input.Items {
				lineSubtotal := float64(it.Quantity) * it.UnitPrice
				dr := 0.0
				tr := 0.0
				if it.DiscountRate != nil {
					dr = *it.DiscountRate
				}
				if it.TaxRate != nil {
					tr = *it.TaxRate
				}
				lineDiscount := lineSubtotal * dr / 100.0
				lineTax := (lineSubtotal - lineDiscount) * tr / 100.0
				subtotal += lineSubtotal
				discountAmount += lineDiscount
				taxAmount += lineTax
			}
			totalAmount := subtotal - discountAmount + taxAmount
			inv.Subtotal = subtotal
			inv.TaxAmount = taxAmount
			inv.DiscountAmount = discountAmount
			inv.TotalAmount = totalAmount
			inv.BalanceDue = totalAmount - inv.PaidAmount

			for _, it := range *input.Items {
				var prodUUID *uuid.UUID
				if it.ProductID != nil && strings.TrimSpace(*it.ProductID) != "" {
					pid, err := uuid.Parse(*it.ProductID)
					if err == nil {
						prodUUID = &pid
					}
				}
				dr := 0.0
				tr := 0.0
				if it.DiscountRate != nil {
					dr = *it.DiscountRate
				}
				if it.TaxRate != nil {
					tr = *it.TaxRate
				}
				lineSubtotal := float64(it.Quantity) * it.UnitPrice
				lineDiscount := lineSubtotal * dr / 100.0
				lineTax := (lineSubtotal - lineDiscount) * tr / 100.0
				lineTotal := lineSubtotal - lineDiscount + lineTax
				row := models.InvoiceItem{
					ID:             uuid.New(),
					InvoiceType:    models.InvoiceTypeOutgoing,
					InvoiceID:      invoiceID,
					ProductID:      prodUUID,
					Description:    it.Description,
					Quantity:       it.Quantity,
					UnitPrice:      it.UnitPrice,
					DiscountRate:   dr,
					DiscountAmount: lineDiscount,
					TaxRate:        tr,
					TaxAmount:      lineTax,
					LineTotal:      lineTotal,
				}
				if err := tx.Create(&row).Error; err != nil {
					return err
				}
			}
		}

		if err := tx.Save(&inv).Error; err != nil {
			return err
		}
		updated = inv
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewErrorResponse("Outgoing invoice not found or cannot be updated")
		}
		return response.NewErrorResponse(err.Error())
	}
	return response.NewSuccessResponse(updated, "")
}

func (s *FinanceService) SendIncomingInvoice(id string, companyID string) response.ApiResponse {
	invoiceID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Incoming invoice not found or cannot be sent")
	}
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Incoming invoice not found or cannot be sent")
	}

	res := s.db.Model(&models.IncomingInvoice{}).
		Where("id = ? AND company_id = ? AND status = ?", invoiceID, companyUUID, models.InvoiceStatusDraft).
		Updates(map[string]interface{}{"status": models.InvoiceStatusSent, "updated_at": time.Now()})
	if res.Error != nil || res.RowsAffected == 0 {
		return response.NewErrorResponse("Incoming invoice not found or cannot be sent")
	}

	var inv models.IncomingInvoice
	_ = s.db.First(&inv, "id = ?", invoiceID).Error
	return response.NewSuccessResponse(inv, "Invoice sent successfully")
}

func (s *FinanceService) SendOutgoingInvoice(id string, companyID string) response.ApiResponse {
	invoiceID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Outgoing invoice not found or cannot be sent")
	}
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Outgoing invoice not found or cannot be sent")
	}

	res := s.db.Model(&models.OutgoingInvoice{}).
		Where("id = ? AND company_id = ? AND status = ?", invoiceID, companyUUID, models.InvoiceStatusDraft).
		Updates(map[string]interface{}{"status": models.InvoiceStatusSent, "updated_at": time.Now()})
	if res.Error != nil || res.RowsAffected == 0 {
		return response.NewErrorResponse("Outgoing invoice not found or cannot be sent")
	}
	var inv models.OutgoingInvoice
	_ = s.db.First(&inv, "id = ?", invoiceID).Error
	return response.NewSuccessResponse(inv, "Invoice sent successfully")
}

func (s *FinanceService) CancelIncomingInvoice(id string, companyID string) response.ApiResponse {
	invoiceID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Incoming invoice not found")
	}
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Incoming invoice not found")
	}

	res := s.db.Model(&models.IncomingInvoice{}).
		Where("id = ? AND company_id = ? AND status != ?", invoiceID, companyUUID, models.InvoiceStatusPaid).
		Updates(map[string]interface{}{"status": models.InvoiceStatusCancelled, "updated_at": time.Now()})
	if res.Error != nil || res.RowsAffected == 0 {
		return response.NewErrorResponse("Incoming invoice not found")
	}
	var inv models.IncomingInvoice
	_ = s.db.First(&inv, "id = ?", invoiceID).Error
	return response.NewSuccessResponse(inv, "Invoice cancelled successfully")
}

func (s *FinanceService) CancelOutgoingInvoice(id string, companyID string) response.ApiResponse {
	invoiceID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Outgoing invoice not found")
	}
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Outgoing invoice not found")
	}

	res := s.db.Model(&models.OutgoingInvoice{}).
		Where("id = ? AND company_id = ? AND status != ?", invoiceID, companyUUID, models.InvoiceStatusPaid).
		Updates(map[string]interface{}{"status": models.InvoiceStatusCancelled, "updated_at": time.Now()})
	if res.Error != nil || res.RowsAffected == 0 {
		return response.NewErrorResponse("Outgoing invoice not found")
	}
	var inv models.OutgoingInvoice
	_ = s.db.First(&inv, "id = ?", invoiceID).Error
	return response.NewSuccessResponse(inv, "Invoice cancelled successfully")
}

func (s *FinanceService) AddIncomingInvoicePayment(invoiceID string, input CreateInvoicePaymentInput, companyID string, userID string) response.ApiResponse {
	invUUID, err := uuid.Parse(invoiceID)
	if err != nil {
		return response.NewErrorResponse("Failed to add payment")
	}
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Failed to add payment")
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return response.NewErrorResponse("Failed to add payment")
	}

	var created models.InvoicePayment
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var inv models.IncomingInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&inv, "id = ? AND company_id = ?", invUUID, companyUUID).Error; err != nil {
			return err
		}

		created = models.InvoicePayment{
			ID:              uuid.New(),
			InvoiceType:     models.InvoiceTypeIncoming,
			InvoiceID:       invUUID,
			PaymentDate:     time.Now(),
			Amount:          input.Amount,
			PaymentMethod:   models.PaymentMethodFinance(input.PaymentMethod),
			ReferenceNumber: input.ReferenceNumber,
			Notes:           input.Notes,
			CreatedBy:       userUUID,
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}

		paid := inv.PaidAmount + input.Amount
		balance := inv.TotalAmount - paid
		newStatus := models.InvoiceStatusSent
		if balance > 0 {
			newStatus = models.InvoiceStatusPartial
		} else if balance == 0 {
			newStatus = models.InvoiceStatusPaid
		}

		return tx.Model(&models.IncomingInvoice{}).
			Where("id = ?", invUUID).
			Updates(map[string]interface{}{"paid_amount": paid, "balance_due": balance, "status": newStatus, "updated_at": time.Now()}).Error
	})
	if err != nil {
		return response.NewErrorResponse(err.Error())
	}
	return response.NewSuccessResponse(created, "Payment recorded successfully")
}

func (s *FinanceService) AddOutgoingInvoicePayment(invoiceID string, input CreateInvoicePaymentInput, companyID string, userID string) response.ApiResponse {
	invUUID, err := uuid.Parse(invoiceID)
	if err != nil {
		return response.NewErrorResponse("Failed to add payment")
	}
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Failed to add payment")
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return response.NewErrorResponse("Failed to add payment")
	}

	var created models.InvoicePayment
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var inv models.OutgoingInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&inv, "id = ? AND company_id = ?", invUUID, companyUUID).Error; err != nil {
			return err
		}
		created = models.InvoicePayment{
			ID:              uuid.New(),
			InvoiceType:     models.InvoiceTypeOutgoing,
			InvoiceID:       invUUID,
			PaymentDate:     time.Now(),
			Amount:          input.Amount,
			PaymentMethod:   models.PaymentMethodFinance(input.PaymentMethod),
			ReferenceNumber: input.ReferenceNumber,
			Notes:           input.Notes,
			CreatedBy:       userUUID,
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}

		paid := inv.PaidAmount + input.Amount
		balance := inv.TotalAmount - paid
		newStatus := models.InvoiceStatusSent
		if balance > 0 {
			newStatus = models.InvoiceStatusPartial
		} else if balance == 0 {
			newStatus = models.InvoiceStatusPaid
		}
		return tx.Model(&models.OutgoingInvoice{}).
			Where("id = ?", invUUID).
			Updates(map[string]interface{}{"paid_amount": paid, "balance_due": balance, "status": newStatus, "updated_at": time.Now()}).Error
	})
	if err != nil {
		return response.NewErrorResponse(err.Error())
	}
	return response.NewSuccessResponse(created, "Payment recorded successfully")
}

type invoiceSummaryRow struct {
	Status           string  `gorm:"column:status"`
	Count            int64   `gorm:"column:count"`
	TotalAmount      float64 `gorm:"column:total_amount"`
	TotalPaid        float64 `gorm:"column:total_paid"`
	TotalOutstanding float64 `gorm:"column:total_outstanding"`
}

type invoiceStatusSummary struct {
	Count  int64   `json:"count"`
	Amount float64 `json:"amount"`
}

func (s *FinanceService) GetInvoiceSummary(filters map[string]string, companyID string) response.ApiResponse {
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Failed to get invoice summary")
	}

	typeFilter := strings.TrimSpace(filters["type"])
	fromDate := strings.TrimSpace(filters["from_date"])
	toDate := strings.TrimSpace(filters["to_date"])

	incomingWhere := "WHERE company_id = ?"
	outgoingWhere := "WHERE company_id = ?"
	params := []interface{}{companyUUID}

	if typeFilter != "" && typeFilter != "ALL" {
		if typeFilter == "INCOMING" {
			outgoingWhere = "WHERE 1=0"
		} else if typeFilter == "OUTGOING" {
			incomingWhere = "WHERE 1=0"
		}
	}

	if fromDate != "" {
		if t, err := parseFlexibleTime(fromDate); err == nil {
			incomingWhere += " AND invoice_date >= ?"
			outgoingWhere += " AND invoice_date >= ?"
			params = append(params, t)
		}
	}
	if toDate != "" {
		if t, err := parseFlexibleTime(toDate); err == nil {
			incomingWhere += " AND invoice_date <= ?"
			outgoingWhere += " AND invoice_date <= ?"
			params = append(params, t)
		}
	}

	query := fmt.Sprintf(`
		SELECT status,
			COUNT(*) as count,
			COALESCE(SUM(total_amount), 0) as total_amount,
			COALESCE(SUM(paid_amount), 0) as total_paid,
			COALESCE(SUM(balance_due), 0) as total_outstanding
		FROM (
			SELECT status, total_amount, paid_amount, balance_due FROM invoices_incoming %s
			UNION ALL
			SELECT status, total_amount, paid_amount, balance_due FROM invoices_outgoing %s
		) combined_invoices
		GROUP BY status
	`, incomingWhere, outgoingWhere)

	var rows []invoiceSummaryRow
	if err := s.db.Raw(query, params...).Scan(&rows).Error; err != nil {
		return response.NewErrorResponse("Failed to get invoice summary")
	}

	statuses := []string{"DRAFT", "SENT", "PARTIAL", "PAID", "OVERDUE", "CANCELLED"}
	byStatus := map[string]invoiceStatusSummary{}
	for _, st := range statuses {
		byStatus[st] = invoiceStatusSummary{Count: 0, Amount: 0}
		for _, r := range rows {
			if r.Status == st {
				byStatus[st] = invoiceStatusSummary{Count: r.Count, Amount: r.TotalAmount}
				break
			}
		}
	}

	totalInvoices := int64(0)
	totalAmount := 0.0
	totalPaid := 0.0
	totalOutstanding := 0.0
	for _, r := range rows {
		totalInvoices += r.Count
		totalAmount += r.TotalAmount
		totalPaid += r.TotalPaid
		totalOutstanding += r.TotalOutstanding
	}

	data := map[string]interface{}{
		"total_invoices":    totalInvoices,
		"total_amount":      totalAmount,
		"total_paid":        totalPaid,
		"total_outstanding": totalOutstanding,
		"by_status":         byStatus,
	}
	return response.NewSuccessResponse(data, "")
}
