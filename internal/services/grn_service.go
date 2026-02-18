package services

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/types/response"
	"gorm.io/gorm"
)

type GrnService struct {
	db *gorm.DB
}

func NewGrnService(db *gorm.DB) *GrnService {
	return &GrnService{db: db}
}

type GrnFilter struct {
	Page        int
	Limit       int
	Status      string
	PoID        string
	WarehouseID string
	StartDate   string
	EndDate     string
}

type grnHeaderRow struct {
	ID            uuid.UUID  `gorm:"column:id"`
	GrnNumber     string     `gorm:"column:grn_number"`
	PoID          uuid.UUID  `gorm:"column:po_id"`
	WarehouseID   uuid.UUID  `gorm:"column:warehouse_id"`
	ReceivedDate  time.Time  `gorm:"column:received_date"`
	Status        string     `gorm:"column:status"`
	InvoiceNumber *string    `gorm:"column:invoice_number"`
	Notes         *string    `gorm:"column:notes"`
	CreatedBy     uuid.UUID  `gorm:"column:created_by"`
	VerifiedBy    *uuid.UUID `gorm:"column:verified_by"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	VerifiedAt    *time.Time `gorm:"column:verified_at"`

	WarehouseName   *string `gorm:"column:warehouse_name"`
	WarehouseCode   *string `gorm:"column:warehouse_code"`
	WarehouseType   *string `gorm:"column:warehouse_type"`
	WarehouseStatus *string `gorm:"column:warehouse_status"`

	PoNumber             *string    `gorm:"column:po_number"`
	SupplierID           *uuid.UUID `gorm:"column:supplier_id"`
	SupplierName         *string    `gorm:"column:supplier_name"`
	SupplierCode         *string    `gorm:"column:supplier_code"`
	SupplierPaymentTerms *string    `gorm:"column:supplier_payment_terms"`
	SupplierCreditLimit  *string    `gorm:"column:supplier_credit_limit"`
	SupplierStatus       *string    `gorm:"column:supplier_status"`

	PoWarehouseID      *uuid.UUID `gorm:"column:po_warehouse_id"`
	PoOrderDate        *time.Time `gorm:"column:order_date"`
	PoExpectedDelivery *time.Time `gorm:"column:expected_delivery"`
	PoPaymentTerms     *string    `gorm:"column:payment_terms"`
	PoStatus           *string    `gorm:"column:po_status"`
	PoSubtotal         *string    `gorm:"column:subtotal"`
	PoTaxAmount        *string    `gorm:"column:tax_amount"`
	PoDiscountAmount   *string    `gorm:"column:discount_amount"`
	PoTotalAmount      *string    `gorm:"column:total_amount"`
	PoNotes            *string    `gorm:"column:po_notes"`
	PoCreatedBy        *uuid.UUID `gorm:"column:po_created_by"`
	PoApprovedBy       *uuid.UUID `gorm:"column:po_approved_by"`
	PoCreatedAt        *time.Time `gorm:"column:po_created_at"`
	PoUpdatedAt        *time.Time `gorm:"column:po_updated_at"`
}

type grnItemRow struct {
	ID               uuid.UUID  `gorm:"column:id"`
	GrnID            uuid.UUID  `gorm:"column:grn_id"`
	PoItemID         uuid.UUID  `gorm:"column:po_item_id"`
	ProductID        uuid.UUID  `gorm:"column:product_id"`
	OrderedQuantity  int        `gorm:"column:ordered_quantity"`
	ReceivedQuantity int        `gorm:"column:received_quantity"`
	RejectedQuantity int        `gorm:"column:rejected_quantity"`
	UnitPrice        string     `gorm:"column:unit_price"`
	QualityNotes     *string    `gorm:"column:quality_notes"`
	ProductName      *string    `gorm:"column:product_name"`
	SKU              *string    `gorm:"column:sku"`
	Barcode          *string    `gorm:"column:barcode"`
	UnitID           *uuid.UUID `gorm:"column:product_unit_id"`
	CostPrice        *string    `gorm:"column:product_cost_price"`
	RetailPrice      *string    `gorm:"column:product_retail_price"`
	ProductStatus    *string    `gorm:"column:product_status"`
	TaxRate          *string    `gorm:"column:product_tax_rate"`

	PoiID              *uuid.UUID `gorm:"column:poi_id"`
	PoQuantity         *int       `gorm:"column:po_quantity"`
	PoReceivedQuantity *int       `gorm:"column:po_received_quantity"`
	PoUnitPrice        *string    `gorm:"column:po_unit_price"`
	PoiTaxRate         *string    `gorm:"column:poi_tax_rate"`
	PoiDiscountRate    *string    `gorm:"column:poi_discount_rate"`
}

func (s *GrnService) GetGrns(filter GrnFilter) map[string]interface{} {
	page := filter.Page
	limit := filter.Limit
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	where := ""
	args := make([]interface{}, 0)
	add := func(cond string, val interface{}) {
		if where == "" {
			where = "WHERE " + cond
		} else {
			where = where + " AND " + cond
		}
		args = append(args, val)
	}

	if filter.Status != "" {
		add("grn.status = ?", filter.Status)
	}
	if filter.PoID != "" {
		add("grn.po_id = ?", filter.PoID)
	}
	if filter.WarehouseID != "" {
		add("grn.warehouse_id = ?", filter.WarehouseID)
	}
	if filter.StartDate != "" {
		add("grn.received_date >= ?", filter.StartDate)
	}
	if filter.EndDate != "" {
		add("grn.received_date <= ?", filter.EndDate)
	}

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM goods_received_notes grn %s", where)
	if err := s.db.Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return map[string]interface{}{"success": false, "error": "Failed to fetch GRNs", "message": "Failed to fetch GRNs"}
	}

	querySQL := fmt.Sprintf(`
		SELECT
			grn.id, grn.grn_number, grn.po_id, grn.warehouse_id,
			grn.received_date, grn.status, grn.invoice_number, grn.notes,
			grn.created_by, grn.verified_by, grn.created_at, grn.verified_at,
			w.name AS warehouse_name, w.code AS warehouse_code, w.type AS warehouse_type, w.status AS warehouse_status,
			po.po_number,
			po.supplier_id,
			s.name AS supplier_name, s.code AS supplier_code,
			s.payment_terms AS supplier_payment_terms, s.credit_limit AS supplier_credit_limit, s.status AS supplier_status
		FROM goods_received_notes grn
		LEFT JOIN warehouses w ON w.id = grn.warehouse_id
		LEFT JOIN purchase_orders po ON po.id = grn.po_id
		LEFT JOIN suppliers s ON s.id = po.supplier_id
		%s
		ORDER BY grn.created_at DESC
		LIMIT %d OFFSET %d
	`, where, limit, offset)

	var headers []grnHeaderRow
	if err := s.db.Raw(querySQL, args...).Scan(&headers).Error; err != nil {
		return map[string]interface{}{"success": false, "error": "Failed to fetch GRNs", "message": "Failed to fetch GRNs"}
	}

	data := make([]map[string]interface{}, 0, len(headers))
	for _, h := range headers {
		items, _ := s.getGrnItems(h.ID, false)
		data = append(data, map[string]interface{}{
			"id":            h.ID,
			"grnNumber":     h.GrnNumber,
			"poId":          h.PoID,
			"warehouseId":   h.WarehouseID,
			"receivedDate":  h.ReceivedDate,
			"status":        h.Status,
			"invoiceNumber": h.InvoiceNumber,
			"notes":         h.Notes,
			"createdBy":     h.CreatedBy,
			"verifiedBy":    h.VerifiedBy,
			"createdAt":     h.CreatedAt,
			"verifiedAt":    h.VerifiedAt,
			"items":         items,
			"purchaseOrder": map[string]interface{}{
				"id":       h.PoID,
				"poNumber": h.PoNumber,
				"supplier": map[string]interface{}{
					"id":           h.SupplierID,
					"code":         h.SupplierCode,
					"name":         h.SupplierName,
					"paymentTerms": h.SupplierPaymentTerms,
					"creditLimit":  h.SupplierCreditLimit,
					"status":       h.SupplierStatus,
				},
			},
			"warehouse": map[string]interface{}{
				"id":     h.WarehouseID,
				"code":   h.WarehouseCode,
				"name":   h.WarehouseName,
				"type":   h.WarehouseType,
				"status": h.WarehouseStatus,
			},
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	return map[string]interface{}{
		"success": true,
		"data":    data,
		"pagination": map[string]interface{}{
			"currentPage":  page,
			"totalPages":   totalPages,
			"totalRecords": total,
			"hasNextPage":  page < totalPages,
			"hasPrevPage":  page > 1,
		},
		"message": "GRNs retrieved successfully",
	}
}

func (s *GrnService) GetGrnByID(id string) response.ApiResponse {
	grnID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("GRN not found")
	}

	querySQL := `
		SELECT
			grn.id, grn.grn_number, grn.po_id, grn.warehouse_id,
			grn.received_date, grn.status, grn.invoice_number, grn.notes,
			grn.created_by, grn.verified_by, grn.created_at, grn.verified_at,
			w.name AS warehouse_name, w.code AS warehouse_code, w.type AS warehouse_type, w.status AS warehouse_status,
			po.po_number, po.order_date, po.expected_delivery, po.payment_terms, po.status AS po_status,
			po.subtotal, po.tax_amount, po.discount_amount, po.total_amount, po.notes AS po_notes,
			po.created_by AS po_created_by, po.approved_by AS po_approved_by, po.created_at AS po_created_at, po.updated_at AS po_updated_at,
			po.warehouse_id AS po_warehouse_id,
			s.id AS supplier_id, s.name AS supplier_name, s.code AS supplier_code,
			s.payment_terms AS supplier_payment_terms, s.credit_limit AS supplier_credit_limit, s.status AS supplier_status
		FROM goods_received_notes grn
		LEFT JOIN warehouses w ON w.id = grn.warehouse_id
		LEFT JOIN purchase_orders po ON po.id = grn.po_id
		LEFT JOIN suppliers s ON s.id = po.supplier_id
		WHERE grn.id = ?
		LIMIT 1
	`

	var h grnHeaderRow
	if err := s.db.Raw(querySQL, grnID).Scan(&h).Error; err != nil {
		return response.NewErrorResponse("GRN not found")
	}
	if h.ID == uuid.Nil {
		return response.NewErrorResponse("GRN not found")
	}

	items, _ := s.getGrnItems(h.ID, true)

	data := map[string]interface{}{
		"id":            h.ID,
		"grnNumber":     h.GrnNumber,
		"poId":          h.PoID,
		"warehouseId":   h.WarehouseID,
		"receivedDate":  h.ReceivedDate,
		"status":        h.Status,
		"invoiceNumber": h.InvoiceNumber,
		"notes":         h.Notes,
		"createdBy":     h.CreatedBy,
		"verifiedBy":    h.VerifiedBy,
		"createdAt":     h.CreatedAt,
		"verifiedAt":    h.VerifiedAt,
		"items":         items,
		"purchaseOrder": map[string]interface{}{
			"id":               h.PoID,
			"poNumber":         h.PoNumber,
			"supplierId":       h.SupplierID,
			"warehouseId":      h.PoWarehouseID,
			"orderDate":        h.PoOrderDate,
			"expectedDelivery": h.PoExpectedDelivery,
			"paymentTerms":     h.PoPaymentTerms,
			"status":           h.PoStatus,
			"subtotal":         strOrEmpty(h.PoSubtotal),
			"taxAmount":        strOrEmpty(h.PoTaxAmount),
			"discountAmount":   strOrEmpty(h.PoDiscountAmount),
			"totalAmount":      strOrEmpty(h.PoTotalAmount),
			"notes":            h.PoNotes,
			"createdBy":        h.PoCreatedBy,
			"approvedBy":       h.PoApprovedBy,
			"createdAt":        h.PoCreatedAt,
			"updatedAt":        h.PoUpdatedAt,
		},
		"warehouse": map[string]interface{}{
			"id":     h.WarehouseID,
			"code":   h.WarehouseCode,
			"name":   h.WarehouseName,
			"type":   h.WarehouseType,
			"status": h.WarehouseStatus,
		},
	}

	return response.NewSuccessResponse(data, "GRN retrieved successfully")
}

func (s *GrnService) CreateGrn(poID string, warehouseID *string, invoiceNumber *string, notes *string, userID string) response.ApiResponse {
	poUUID, err := uuid.Parse(poID)
	if err != nil {
		return response.NewErrorResponse("Purchase order not found")
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return response.NewErrorResponse("User not authenticated")
	}

	var createdID uuid.UUID
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var po struct {
			ID          uuid.UUID `gorm:"column:id"`
			Status      string    `gorm:"column:status"`
			PoNumber    string    `gorm:"column:po_number"`
			WarehouseID uuid.UUID `gorm:"column:po_warehouse_id"`
		}
		if err := tx.Raw("SELECT id, status, po_number, warehouse_id as po_warehouse_id FROM purchase_orders WHERE id = ? LIMIT 1", poUUID).Scan(&po).Error; err != nil {
			return err
		}
		if po.ID == uuid.Nil {
			return fmt.Errorf("Purchase order not found")
		}
		if po.Status != "APPROVED" && po.Status != "PARTIAL_RECEIVED" {
			return fmt.Errorf("Cannot create GRN for purchase order with status: %s. Only APPROVED or PARTIAL_RECEIVED POs are allowed.", po.Status)
		}

		wh := po.WarehouseID
		if warehouseID != nil && *warehouseID != "" {
			w, err := uuid.Parse(*warehouseID)
			if err == nil {
				wh = w
			}
		}

		grnNumber := fmt.Sprintf("GRN-%d-%d", time.Now().UnixMilli(), rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1000))
		head := map[string]interface{}{
			"grn_number":     grnNumber,
			"po_id":          poUUID,
			"warehouse_id":   wh,
			"status":         "DRAFT",
			"invoice_number": nil,
			"notes":          nil,
			"created_by":     userUUID,
			"received_date":  time.Now(),
			"created_at":     time.Now(),
		}
		if invoiceNumber != nil {
			head["invoice_number"] = *invoiceNumber
		}
		if notes != nil {
			head["notes"] = *notes
		}

		var grn models.GoodsReceivedNote
		if err := tx.Table("goods_received_notes").Create(&head).Error; err != nil {
			return err
		}
		// fetch id
		if err := tx.Raw("SELECT id FROM goods_received_notes WHERE grn_number = ? LIMIT 1", grnNumber).Scan(&grn).Error; err != nil {
			return err
		}
		createdID = grn.ID
		if createdID == uuid.Nil {
			return fmt.Errorf("Failed to create GRN")
		}

		var poItems []struct {
			PoiID            uuid.UUID `gorm:"column:poi_id"`
			ProductID        uuid.UUID `gorm:"column:product_id"`
			Quantity         int       `gorm:"column:quantity"`
			ReceivedQuantity int       `gorm:"column:received_quantity"`
			UnitPrice        string    `gorm:"column:unit_price"`
		}
		if err := tx.Raw("SELECT id as poi_id, product_id, quantity, COALESCE(received_quantity,0) as received_quantity, unit_price FROM purchase_order_items WHERE po_id = ? ORDER BY id", poUUID).Scan(&poItems).Error; err != nil {
			return err
		}
		if len(poItems) == 0 {
			return fmt.Errorf("No items found in purchase order")
		}

		for _, it := range poItems {
			remaining := it.Quantity - it.ReceivedQuantity
			if remaining <= 0 {
				continue
			}
			if err := tx.Exec(
				"INSERT INTO grn_items (grn_id, po_item_id, product_id, ordered_quantity, received_quantity, rejected_quantity, unit_price, quality_notes) VALUES (?, ?, ?, ?, ?, 0, ?, '')",
				createdID, it.PoiID, it.ProductID, it.Quantity, remaining, it.UnitPrice,
			).Error; err != nil {
				return err
			}
			if err := tx.Exec("UPDATE purchase_order_items SET received_quantity = received_quantity + ? WHERE id = ?", remaining, it.PoiID).Error; err != nil {
				return err
			}
		}

		var updated []struct {
			Quantity         int `gorm:"column:quantity"`
			ReceivedQuantity int `gorm:"column:received_quantity"`
		}
		_ = tx.Raw("SELECT quantity, COALESCE(received_quantity,0) as received_quantity FROM purchase_order_items WHERE po_id = ?", poUUID).Scan(&updated).Error
		allReceived := true
		partialReceived := false
		for _, it := range updated {
			if it.ReceivedQuantity > 0 {
				partialReceived = true
			}
			if it.ReceivedQuantity < it.Quantity {
				allReceived = false
			}
		}
		if allReceived {
			_ = tx.Exec("UPDATE purchase_orders SET status = 'RECEIVED' WHERE id = ?", poUUID).Error
		} else if partialReceived {
			_ = tx.Exec("UPDATE purchase_orders SET status = 'PARTIAL_RECEIVED' WHERE id = ?", poUUID).Error
		}
		return nil
	})

	if err != nil {
		return response.NewErrorResponse(err.Error())
	}
	return s.GetGrnByID(createdID.String())
}

func (s *GrnService) UpdateGrn(id string, invoiceNumber *string, notes *string, items *[]map[string]interface{}, userID string) response.ApiResponse {
	grnID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("GRN not found")
	}

	var existing struct {
		ID     uuid.UUID `gorm:"column:id"`
		Status string    `gorm:"column:status"`
	}
	if err := s.db.Raw("SELECT id, status FROM goods_received_notes WHERE id = ? LIMIT 1", grnID).Scan(&existing).Error; err != nil {
		return response.NewErrorResponse("GRN not found")
	}
	if existing.ID == uuid.Nil {
		return response.NewErrorResponse("GRN not found")
	}
	if existing.Status != "DRAFT" {
		return response.NewErrorResponse("Cannot update GRN with status: " + existing.Status)
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{}
		if invoiceNumber != nil {
			updates["invoice_number"] = *invoiceNumber
		}
		if notes != nil {
			updates["notes"] = *notes
		}
		if len(updates) > 0 {
			if err := tx.Table("goods_received_notes").Where("id = ?", grnID).Updates(updates).Error; err != nil {
				return err
			}
		}

		if items != nil {
			if err := tx.Exec("DELETE FROM grn_items WHERE grn_id = ?", grnID).Error; err != nil {
				return err
			}
			for _, it := range *items {
				poItemID, _ := it["poItemId"].(string)
				productID, _ := it["productId"].(string)
				ordered, _ := it["orderedQuantity"].(float64)
				received, _ := it["receivedQuantity"].(float64)
				rejected, _ := it["rejectedQuantity"].(float64)
				unitPrice, _ := it["unitPrice"].(string)
				qualityNotes, _ := it["qualityNotes"].(string)

				var poItem struct {
					ID       uuid.UUID `gorm:"column:id"`
					Quantity int       `gorm:"column:quantity"`
				}
				if err := tx.Raw("SELECT id, quantity FROM purchase_order_items WHERE id = ? LIMIT 1", poItemID).Scan(&poItem).Error; err != nil {
					return err
				}
				if poItem.ID == uuid.Nil {
					return fmt.Errorf("Purchase order item %s not found", poItemID)
				}
				if int(received) < 0 || int(received) > poItem.Quantity {
					return fmt.Errorf("Received quantity %d exceeds ordered quantity %d for item %s", int(received), poItem.Quantity, poItemID)
				}

				if err := tx.Exec(
					"INSERT INTO grn_items (grn_id, po_item_id, product_id, ordered_quantity, received_quantity, rejected_quantity, unit_price, quality_notes) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
					grnID, poItemID, productID, int(ordered), int(received), int(rejected), unitPrice, qualityNotes,
				).Error; err != nil {
					return err
				}
			}
		}
		_ = userID
		return nil
	})
	if err != nil {
		return response.NewErrorResponse(err.Error())
	}
	return s.GetGrnByID(id)
}

func (s *GrnService) CancelGrn(id string) response.ApiResponse {
	grnID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("GRN not found")
	}
	var existing struct {
		ID     uuid.UUID `gorm:"column:id"`
		Status string    `gorm:"column:status"`
	}
	_ = s.db.Raw("SELECT id, status FROM goods_received_notes WHERE id = ? LIMIT 1", grnID).Scan(&existing).Error
	if existing.ID == uuid.Nil {
		return response.NewErrorResponse("GRN not found")
	}
	if existing.Status != "DRAFT" {
		return response.NewErrorResponse("Cannot cancel GRN with status: " + existing.Status)
	}
	_ = s.db.Exec("UPDATE goods_received_notes SET status = 'CANCELLED' WHERE id = ?", grnID).Error
	return response.NewSuccessResponse(nil, "GRN cancelled successfully")
}

func (s *GrnService) VerifyGrn(id string, userID string) response.ApiResponse {
	grnID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("GRN not found")
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return response.NewErrorResponse("Unauthorized")
	}

	var existing struct {
		ID     uuid.UUID `gorm:"column:id"`
		Status string    `gorm:"column:status"`
	}
	_ = s.db.Raw("SELECT id, status FROM goods_received_notes WHERE id = ? LIMIT 1", grnID).Scan(&existing).Error
	if existing.ID == uuid.Nil {
		return response.NewErrorResponse("GRN not found")
	}
	if existing.Status != "DRAFT" {
		return response.NewErrorResponse("Cannot verify GRN with status: " + existing.Status + ". GRN must be in DRAFT status.")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		var items []struct {
			ProductID        uuid.UUID `gorm:"column:product_id"`
			ReceivedQuantity int       `gorm:"column:received_quantity"`
			PoNumber         string    `gorm:"column:po_number"`
			WarehouseID      uuid.UUID `gorm:"column:warehouse_id"`
		}
		if err := tx.Raw(`
			SELECT gri.product_id, gri.received_quantity, po.po_number, po.warehouse_id
			FROM grn_items gri
			LEFT JOIN goods_received_notes grn ON grn.id = gri.grn_id
			LEFT JOIN purchase_orders po ON po.id = grn.po_id
			WHERE gri.grn_id = ?
		`, grnID).Scan(&items).Error; err != nil {
			return err
		}

		for _, it := range items {
			var inv struct {
				ID       uuid.UUID `gorm:"column:id"`
				Quantity int       `gorm:"column:quantity"`
			}
			_ = tx.Raw("SELECT id, quantity FROM inventory WHERE product_id = ? AND warehouse_id = ? LIMIT 1", it.ProductID, it.WarehouseID).Scan(&inv).Error
			if inv.ID != uuid.Nil {
				if err := tx.Exec("UPDATE inventory SET quantity = quantity + ?, available_quantity = available_quantity + ? WHERE id = ?", it.ReceivedQuantity, it.ReceivedQuantity, inv.ID).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Exec("INSERT INTO inventory (product_id, warehouse_id, quantity, reserved_quantity, available_quantity, min_stock_level, max_stock_level, created_at, updated_at) VALUES (?, ?, ?, 0, ?, 5, 100, NOW(), NOW())", it.ProductID, it.WarehouseID, it.ReceivedQuantity, it.ReceivedQuantity).Error; err != nil {
					return err
				}
			}

			note := fmt.Sprintf("Goods received from GRN %s", it.PoNumber)
			refID := grnID
			movement := models.StockMovement{
				ID:            uuid.New(),
				ProductID:     it.ProductID,
				WarehouseID:   it.WarehouseID,
				MovementType:  models.MovementTypePurchase,
				Quantity:      it.ReceivedQuantity,
				ReferenceType: "GRN",
				ReferenceID:   &refID,
				Notes:         note,
				CreatedBy:     &userUUID,
				CreatedAt:     time.Now(),
			}
			if err := tx.Create(&movement).Error; err != nil {
				return err
			}
		}

		now := time.Now()
		if err := tx.Exec("UPDATE goods_received_notes SET status = 'VERIFIED', verified_by = ?, verified_at = ? WHERE id = ?", userUUID, now, grnID).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return response.NewErrorResponse(err.Error())
	}
	return s.GetGrnByID(id)
}

func (s *GrnService) getGrnItems(grnID uuid.UUID, includePOItem bool) ([]map[string]interface{}, error) {
	itemsSQL := `
		SELECT
			gri.id, gri.grn_id, gri.po_item_id, gri.product_id,
			gri.ordered_quantity, gri.received_quantity, gri.rejected_quantity,
			gri.unit_price, gri.quality_notes,
			p.name AS product_name, p.sku, p.barcode,
			p.unit_id AS product_unit_id, p.cost_price AS product_cost_price,
			p.retail_price AS product_retail_price, p.status AS product_status,
			p.tax_rate AS product_tax_rate
		FROM grn_items gri
		LEFT JOIN products p ON p.id = gri.product_id
		WHERE gri.grn_id = ?
		ORDER BY gri.id
	`
	if includePOItem {
		itemsSQL = `
			SELECT
				gri.id, gri.grn_id, gri.po_item_id, gri.product_id,
				gri.ordered_quantity, gri.received_quantity, gri.rejected_quantity,
				gri.unit_price, gri.quality_notes,
				p.name AS product_name, p.sku, p.barcode,
				p.unit_id AS product_unit_id, p.cost_price AS product_cost_price,
				p.retail_price AS product_retail_price, p.status AS product_status,
				p.tax_rate AS product_tax_rate,
				poi.id AS poi_id, poi.quantity AS po_quantity,
				poi.unit_price AS po_unit_price, poi.received_quantity AS po_received_quantity,
				poi.tax_rate AS poi_tax_rate, poi.discount_rate AS poi_discount_rate
			FROM grn_items gri
			LEFT JOIN products p ON p.id = gri.product_id
			LEFT JOIN purchase_order_items poi ON poi.id = gri.po_item_id
			WHERE gri.grn_id = ?
			ORDER BY gri.id
		`
	}

	var rows []grnItemRow
	if err := s.db.Raw(itemsSQL, grnID).Scan(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		item := map[string]interface{}{
			"id":               r.ID,
			"grnId":            r.GrnID,
			"poItemId":         r.PoItemID,
			"productId":        r.ProductID,
			"orderedQuantity":  r.OrderedQuantity,
			"receivedQuantity": r.ReceivedQuantity,
			"rejectedQuantity": r.RejectedQuantity,
			"unitPrice":        r.UnitPrice,
			"qualityNotes":     r.QualityNotes,
			"product": map[string]interface{}{
				"id":          r.ProductID,
				"name":        r.ProductName,
				"sku":         r.SKU,
				"barcode":     r.Barcode,
				"unitId":      r.UnitID,
				"costPrice":   strOrEmpty(r.CostPrice),
				"retailPrice": strOrEmpty(r.RetailPrice),
				"status":      r.ProductStatus,
				"taxRate":     strOrEmpty(r.TaxRate),
			},
		}
		if includePOItem {
			item["purchaseOrderItem"] = map[string]interface{}{
				"id":               r.PoiID,
				"poId":             nil,
				"productId":        r.ProductID,
				"quantity":         r.PoQuantity,
				"receivedQuantity": r.PoReceivedQuantity,
				"unitPrice":        r.PoUnitPrice,
				"taxRate":          strOrEmpty(r.PoiTaxRate),
				"discountRate":     strOrEmpty(r.PoiDiscountRate),
			}
		}
		out = append(out, item)
	}
	return out, nil
}

func strOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
