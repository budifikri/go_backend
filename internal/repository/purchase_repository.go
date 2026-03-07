package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PurchaseOrderRow struct {
	ID               uuid.UUID `json:"id" gorm:"column:id"`
	PoNumber         string    `json:"po_number" gorm:"column:po_number"`
	SupplierID       uuid.UUID `json:"supplier_id" gorm:"column:supplier_id"`
	WarehouseID      uuid.UUID `json:"warehouse_id" gorm:"column:warehouse_id"`
	OrderDate        string    `json:"order_date" gorm:"column:order_date"`
	ExpectedDelivery *string   `json:"expected_delivery" gorm:"column:expected_delivery"`
	PaymentTerms     string    `json:"payment_terms" gorm:"column:payment_terms"`
	StatusPo         string    `json:"status_po" gorm:"column:status_po"`
	StatusReceive    string    `json:"status_receive" gorm:"column:status_receive"`
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

func (r *PurchaseRepository) FindPurchaseOrders(filters map[string]string, limit, offset int) ([]PurchaseOrderRow, int64, error) {
	var rows []PurchaseOrderRow
	var total int64

	query := r.db.Table("purchase_orders po").
		Select("po.*, s.name AS supplier_name, w.name AS warehouse_name").
		Joins("LEFT JOIN suppliers s ON s.id = po.supplier_id").
		Joins("LEFT JOIN warehouses w ON w.id = po.warehouse_id")

	if v := filters["status"]; v != "" {
		query = query.Where("po.status_po = ?", v)
	}
	if v := filters["status_receive"]; v != "" {
		query = query.Where("po.status_receive = ?", v)
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

	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("po.order_date DESC").Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
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
		return nil, err
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
