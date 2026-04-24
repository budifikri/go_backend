package repository

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
)

type PurchaseOrderRow struct {
	ID               uuid.UUID `json:"id" gorm:"column:id"`
	PoNumber         string    `json:"po_number" gorm:"column:po_number"`
	SupplierID       uuid.UUID `json:"supplier_id" gorm:"column:supplier_id"`
	WarehouseID      uuid.UUID `json:"warehouse_id" gorm:"column:warehouse_id"`
	OrderDate        string    `json:"order_date" gorm:"column:order_date"`
	ExpectedDelivery *string   `json:"expected_delivery" gorm:"column:expected_delivery"`
	ReceiveDate      *string   `json:"receive_date" gorm:"column:receive_date"`
	PaymentTerms     string    `json:"payment_terms" gorm:"column:payment_terms"`
	StatusPo         string    `json:"status_po" gorm:"column:status_po"`
	StatusReceive    string    `json:"status_receive" gorm:"column:status_receive"`
	ReceiveNumber    *string   `json:"receive_number" gorm:"column:receive_number"`
	NoteReceive      *string   `json:"note_receive" gorm:"column:note_receive"`
	Subtotal         string    `json:"subtotal" gorm:"column:subtotal"`
	TaxAmount        string    `json:"tax_amount" gorm:"column:tax_amount"`
	DiscountAmount   string    `json:"discount_amount" gorm:"column:discount_amount"`
	TotalAmount      string    `json:"total_amount" gorm:"column:total_amount"`
	Notes            *string   `json:"notes" gorm:"column:notes"`
	CreatedBy        uuid.UUID `json:"created_by" gorm:"column:created_by"`
	CompanyID        uuid.UUID `json:"company_id" gorm:"column:company_id"`
	CreatedAt        string    `json:"created_at" gorm:"column:created_at"`
	UpdatedAt        string    `json:"updated_at" gorm:"column:updated_at"`
	SupplierName     *string   `json:"supplier_name" gorm:"column:supplier_name"`
	WarehouseName    *string   `json:"warehouse_name" gorm:"column:warehouse_name"`
}

type PurchaseOrderItemRow struct {
	ID          uuid.UUID `json:"id" gorm:"column:id"`
	PoID        uuid.UUID `json:"po_id" gorm:"column:po_id"`
	ProductID   uuid.UUID `json:"product_id" gorm:"column:product_id"`
	QtyPo       int       `json:"qty_po" gorm:"column:qty_po"`
	QtyReceive  int       `json:"qty_receive" gorm:"column:qty_receive"`
	UnitPrice   string    `json:"unit_price" gorm:"column:unit_price"`
	Discount    string    `json:"discount" gorm:"column:discount"`
	TaxRate     string    `json:"tax_rate" gorm:"column:tax_rate"`
	LineTotal   string    `json:"line_total" gorm:"column:line_total"`
	ProductName *string   `json:"product_name" gorm:"column:product_name"`
	SKU         *string   `json:"sku" gorm:"column:sku"`
}

type PurchaseRepository struct {
	db *gorm.DB
}

func NewPurchaseRepository(db *gorm.DB) *PurchaseRepository {
	return &PurchaseRepository{db: db}
}

func (r *PurchaseRepository) FindPurchaseOrders(companyID *string, filters map[string]string, limit, offset int) ([]PurchaseOrderRow, int64, error) {
	var rows []PurchaseOrderRow
	var total int64

	query := r.db.Table("purchase_orders po").
		Select("po.*, s.name AS supplier_name, w.name AS warehouse_name").
		Joins("LEFT JOIN suppliers s ON s.id = po.supplier_id").
		Joins("LEFT JOIN warehouses w ON w.id = po.warehouse_id")

	// Filter by company_id (required for multi-tenant isolation)
	if companyID != nil && *companyID != "" {
		query = query.Where("po.company_id = ?", *companyID)
	}

	// Support case-insensitive filtering for statuses
	if v := filters["status_po"]; v != "" {
		query = query.Where("LOWER(po.status_po) = ?", strings.ToLower(v))
	}
	if v := filters["status_receive"]; v != "" {
		query = query.Where("LOWER(po.status_receive) = ?", strings.ToLower(v))
	}
	if v := filters["supplier_id"]; v != "" {
		query = query.Where("po.supplier_id = ?", v)
	}
	if v := filters["warehouse_id"]; v != "" {
		query = query.Where("po.warehouse_id = ?", v)
	}
	if v := filters["search"]; v != "" {
		like := "%" + v + "%"
		query = query.Where("po.po_number ILIKE ? OR s.name ILIKE ? OR w.name ILIKE ?", like, like, like)
	}
	if v := filters["date_from"]; v != "" {
		// Parse date and create timestamp in Asia/Jakarta timezone (UTC+7)
		// to match the database timezone
		if t, err := time.Parse("2006-01-02", v); err == nil {
			jakartaLoc, _ := time.LoadLocation("Asia/Jakarta")
			startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, jakartaLoc)
			if filters["date_to"] == "" {
				query = query.Where("po.order_date >= ?", startOfDay)
			}
		}
	}
	if v := filters["date_to"]; v != "" {
		// Parse date and create timestamp in Asia/Jakarta timezone (UTC+7)
		if t, err := time.Parse("2006-01-02", v); err == nil {
			jakartaLoc, _ := time.LoadLocation("Asia/Jakarta")
			endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, jakartaLoc)
			if filters["date_from"] == "" {
				query = query.Where("po.order_date <= ?", endOfDay)
			} else {
				// Both date_from and date_to are provided - need to get date_from value
				if vFrom := filters["date_from"]; vFrom != "" {
					if tFrom, err := time.Parse("2006-01-02", vFrom); err == nil {
						jakartaLoc, _ := time.LoadLocation("Asia/Jakarta")
						startOfDay := time.Date(tFrom.Year(), tFrom.Month(), tFrom.Day(), 0, 0, 0, 0, jakartaLoc)
						endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, jakartaLoc)
						query = query.Where("po.order_date >= ? AND po.order_date <= ?", startOfDay, endOfDay)
					}
				}
			}
		}
	}

	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("po.po_number ASC").Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (r *PurchaseRepository) GetPurchaseOrderByID(id uuid.UUID) (*PurchaseOrderRow, error) {
	var row PurchaseOrderRow
	err := r.db.Table("purchase_orders po").
		Select("po.*, s.name AS supplier_name, w.name AS warehouse_name").
		Joins("LEFT JOIN suppliers s ON s.id = po.supplier_id").
		Joins("LEFT JOIN warehouses w ON w.id = po.warehouse_id").
		Where("po.id = ?", id).
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

func (r *PurchaseRepository) GetPurchaseOrderItems(poID uuid.UUID) ([]PurchaseOrderItemRow, error) {
	var items []PurchaseOrderItemRow

	log.Printf("[DEBUG] GetPurchaseOrderItems called with poID: %s", poID.String())

	selectClause := `
		poi.id, poi.po_id, poi.product_id, poi.qty_po, poi.qty_receive,
		poi.unit_price,
		poi.discount_rate AS discount,
		poi.tax_rate,
		(
			(poi.qty_po * poi.unit_price * (1 - (poi.discount_rate / 100))) +
			((poi.qty_po * poi.unit_price * (1 - (poi.discount_rate / 100))) * (poi.tax_rate / 100))
		) AS line_total,
		p.name AS product_name,
		p.sku
	`

	err := r.db.Table("purchase_order_items poi").
		Select(selectClause).
		Joins("LEFT JOIN products p ON p.id = poi.product_id").
		Where("poi.po_id = ?", poID).
		Order("poi.id").
		Scan(&items).Error
	if err != nil {
		log.Printf("[DEBUG] GetPurchaseOrderItems error: %v", err)
		return nil, err
	}
	log.Printf("[DEBUG] GetPurchaseOrderItems returned %d items for poID: %s", len(items), poID.String())
	for i, item := range items {
		log.Printf("[DEBUG] Item %d: id=%s, po_id=%s, product_id=%s, qty=%d, unit_price=%s",
			i, item.ID.String(), item.PoID.String(), item.ProductID.String(), item.QtyPo, item.UnitPrice)
	}
	return items, nil
}

func (r *PurchaseRepository) GetPurchaseOrderItemByID(id uuid.UUID) (*PurchaseOrderItemRow, error) {
	var item PurchaseOrderItemRow
	err := r.db.Table("purchase_order_items poi").
		Select("poi.*, p.name AS product_name, p.sku").
		Joins("LEFT JOIN products p ON p.id = poi.product_id").
		Where("poi.id = ?", id).
		Scan(&item).Error
	if err != nil {
		return nil, err
	}
	if item.ID == uuid.Nil {
		return nil, nil
	}
	return &item, nil
}

func (r *PurchaseRepository) UpdatePurchaseOrderItem(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Table("purchase_order_items").Where("id = ?", id).Updates(updates).Error
}

func (r *PurchaseRepository) UpdatePurchaseOrder(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Table("purchase_orders").Where("id = ?", id).Updates(updates).Error
}

func (r *PurchaseRepository) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("purchase_order_items").Where("po_id = ?", id).Delete(nil).Error; err != nil {
			return err
		}
		if err := tx.Table("purchase_orders").Where("id = ?", id).Delete(nil).Error; err != nil {
			return err
		}
		return nil
	})
}

type PurchaseReturnRow struct {
	models.PurchaseReturn
	SupplierName  *string `gorm:"column:supplier_name"`
	WarehouseName *string `gorm:"column:warehouse_name"`
	PoNumber      *string `gorm:"column:po_number"`
}

func (r *PurchaseRepository) GetNextReturnNumber(companyID string) (string, error) {
	year := time.Now().Format("06")

	companyPrefix := ""
	if companyID != "" {
		if len(companyID) >= 4 {
			companyPrefix = strings.ToLower(companyID[len(companyID)-4:])
		}
	}

	var lastReturn struct {
		ReturnNumber string `gorm:"column:return_number"`
	}
	r.db.Table("purchase_returns").
		Where("return_number LIKE ?", fmt.Sprintf("PR-%s-%s-%%", year, companyPrefix)).
		Order("return_number DESC").
		Limit(1).Scan(&lastReturn)

	sequence := 1
	if lastReturn.ReturnNumber != "" {
		parts := strings.Split(lastReturn.ReturnNumber, "-")
		if len(parts) >= 3 {
			seqStr := parts[len(parts)-1]
			var digits string
			for _, c := range seqStr {
				if c >= '0' && c <= '9' {
					digits += string(c)
				}
			}
			if digits != "" {
				fmt.Sscanf(digits, "%d", &sequence)
				sequence++
			}
		}
	}

	return fmt.Sprintf("PR-%s-%s-%s", year, companyPrefix, formatInt(int64(sequence), 6)), nil
}

func (r *PurchaseRepository) CreatePurchaseReturn(pr *models.PurchaseReturn) error {
	return r.db.Create(pr).Error
}

func (r *PurchaseRepository) GetPurchaseReturnByID(id uuid.UUID) (*PurchaseReturnRow, error) {
	var row PurchaseReturnRow
	err := r.db.Table("purchase_returns pr").
		Select("pr.*, s.name AS supplier_name, w.name AS warehouse_name, po.po_number").
		Joins("LEFT JOIN suppliers s ON s.id = pr.supplier_id").
		Joins("LEFT JOIN warehouses w ON w.id = pr.warehouse_id").
		Joins("LEFT JOIN purchase_orders po ON po.id = pr.po_id").
		Where("pr.id = ?", id).
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

func (r *PurchaseRepository) FindPurchaseReturns(companyID *string, filters map[string]interface{}, limit, offset int) ([]PurchaseReturnRow, int64, error) {
	var rows []PurchaseReturnRow
	var total int64

	query := r.db.Table("purchase_returns pr").
		Select("pr.*, s.name AS supplier_name, w.name AS warehouse_name, po.po_number").
		Joins("LEFT JOIN suppliers s ON s.id = pr.supplier_id").
		Joins("LEFT JOIN warehouses w ON w.id = pr.warehouse_id").
		Joins("LEFT JOIN purchase_orders po ON po.id = pr.po_id")

	if companyID != nil && *companyID != "" {
		companyUUID, err := uuid.Parse(*companyID)
		if err == nil {
			query = query.Where("pr.company_id = ?", companyUUID)
		}
	}

	if warehouseID, ok := filters["warehouse_id"].(string); ok && warehouseID != "" {
		wid, err := uuid.Parse(warehouseID)
		if err == nil {
			query = query.Where("pr.warehouse_id = ?", wid)
		}
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("pr.status = ?", status)
	}

	if fromDate, ok := filters["from_date"].(string); ok && fromDate != "" {
		query = query.Where("pr.return_date >= ?", fromDate)
	}

	if toDate, ok := filters["to_date"].(string); ok && toDate != "" {
		query = query.Where("pr.return_date <= ?", toDate)
	}

	if search, ok := filters["search"].(string); ok && search != "" {
		like := "%" + search + "%"
		query = query.Where("pr.return_number ILIKE ? OR po.po_number ILIKE ? OR s.name ILIKE ?", like, like, like)
	}

	query.Count(&total)

	if err := query.Order("pr.created_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (r *PurchaseRepository) UpdatePurchaseReturn(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Table("purchase_returns").Where("id = ?", id).Updates(updates).Error
}

func (r *PurchaseRepository) DeletePurchaseReturn(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("purchase_return_items").Where("return_id = ?", id).Delete(nil).Error; err != nil {
			return err
		}
		if err := tx.Table("purchase_returns").Where("id = ?", id).Delete(nil).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *PurchaseRepository) GetPurchaseReturnItems(returnID uuid.UUID) ([]map[string]interface{}, error) {
	var items []map[string]interface{}

	query := r.db.Table("purchase_return_items pri").
		Select("pri.*, p.name AS product_name, p.sku").
		Joins("LEFT JOIN products p ON p.id = pri.product_id").
		Where("pri.return_id = ?", returnID)

	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
