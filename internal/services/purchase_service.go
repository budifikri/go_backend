package services

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/repository"
	"github.com/pos-retail/go_backend/internal/types/response"
	"gorm.io/gorm"
)

type PurchaseService struct {
	db            *gorm.DB
	purchaseRepo  *repository.PurchaseRepository
	inventoryRepo *repository.InventoryRepository
}

func NewPurchaseService(db *gorm.DB, purchaseRepo *repository.PurchaseRepository) *PurchaseService {
	return &PurchaseService{db: db, purchaseRepo: purchaseRepo}
}

type CreatePurchaseOrderItemInput struct {
	ID        string
	ProductID string
	Quantity  int
	UnitPrice float64
	Discount  float64
	TaxRate   float64
}

type CreatePurchaseOrderInput struct {
	SupplierID   string
	WarehouseID  string
	ExpectedDate time.Time
	Items        []CreatePurchaseOrderItemInput
	Notes        *string
	CreatedBy    string
	CompanyID    string
}

type UpdatePurchaseOrderInput struct {
	SupplierID    string
	WarehouseID   string
	OrderDate     time.Time
	ExpectedDate  time.Time
	Items         []CreatePurchaseOrderItemInput
	Notes         *string
	StatusPo      string
	StatusReceive string
}

type ReceivePurchaseOrderItemInput struct {
	ID         string
	QtyReceive int
}

type ReceivePurchaseOrderInput struct {
	Items         []ReceivePurchaseOrderItemInput
	StatusReceive string
	ReceiveDate   time.Time
}

func normalizeStatusPo(value string) string {
	v := strings.ToUpper(strings.TrimSpace(value))
	// Frontend may send "approved"; some flows may send "approve".
	if v == "APPROVED" {
		return "APPROVE"
	}
	return v
}

func normalizeStatusReceive(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func (s *PurchaseService) GetPurchaseOrders(companyID *string, filters map[string]string, limit, offset int) response.PaginatedResponse {
	rows, total, err := s.purchaseRepo.FindPurchaseOrders(companyID, filters, limit, offset)
	if err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}

	data := make([]map[string]interface{}, 0, len(rows))
	for _, po := range rows {
		m := map[string]interface{}{
			"id":                po.ID,
			"po_number":         po.PoNumber,
			"supplier_id":       po.SupplierID,
			"warehouse_id":      po.WarehouseID,
			"order_date":        po.OrderDate,
			"expected_delivery": po.ExpectedDelivery,
			"receive_date":      po.ReceiveDate,
			"payment_terms":     po.PaymentTerms,
			"status_po":         strings.ToLower(po.StatusPo),
			"status_receive":    strings.ToLower(po.StatusReceive),
			"receive_number":    po.ReceiveNumber,
			"note_receive":      po.NoteReceive,
			"subtotal":          toFloat(po.Subtotal),
			"tax_amount":        toFloat(po.TaxAmount),
			"discount_amount":   toFloat(po.DiscountAmount),
			"total_amount":      toFloat(po.TotalAmount),
			"notes":             po.Notes,
			"created_by":        po.CreatedBy,
			"company_id":        po.CompanyID,
			"created_at":        po.CreatedAt,
			"updated_at":        po.UpdatedAt,
			"supplier_name":     po.SupplierName,
			"warehouse_name":    po.WarehouseName,
		}
		data = append(data, m)
	}

	return response.NewPaginatedResponse(data, total, limit, offset)
}

func (s *PurchaseService) GetPurchaseOrderByID(id string) response.ApiResponse {
	poID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Purchase order not found")
	}

	po, err := s.purchaseRepo.GetPurchaseOrderByID(poID)
	if err != nil {
		return response.NewErrorResponse("Purchase order not found")
	}
	if po == nil {
		return response.NewErrorResponse("Purchase order not found")
	}

	items, _ := s.purchaseRepo.GetPurchaseOrderItems(poID)
	itemsOut := make([]map[string]interface{}, 0, len(items))
	for _, it := range items {
		itemsOut = append(itemsOut, map[string]interface{}{
			"id":           it.ID,
			"po_id":        it.PoID,
			"product_id":   it.ProductID,
			"qty_po":       it.QtyPo,
			"qty_receive":  it.QtyReceive,
			"unit_price":   toFloat(it.UnitPrice),
			"discount":     toFloat(it.Discount),
			"tax_rate":     toFloat(it.TaxRate),
			"line_total":   toFloat(it.LineTotal),
			"product_name": it.ProductName,
			"sku":          it.SKU,
		})
	}

	data := map[string]interface{}{
		"id":                po.ID,
		"po_number":         po.PoNumber,
		"supplier_id":       po.SupplierID,
		"warehouse_id":      po.WarehouseID,
		"order_date":        po.OrderDate,
		"expected_delivery": po.ExpectedDelivery,
		"receive_date":      po.ReceiveDate,
		"payment_terms":     po.PaymentTerms,
		"status_po":         strings.ToLower(po.StatusPo),
		"status_receive":    strings.ToLower(po.StatusReceive),
		"receive_number":    po.ReceiveNumber,
		"note_receive":      po.NoteReceive,
		"subtotal":          toFloat(po.Subtotal),
		"tax_amount":        toFloat(po.TaxAmount),
		"discount_amount":   toFloat(po.DiscountAmount),
		"total_amount":      toFloat(po.TotalAmount),
		"notes":             po.Notes,
		"created_by":        po.CreatedBy,
		"company_id":        po.CompanyID,
		"created_at":        po.CreatedAt,
		"updated_at":        po.UpdatedAt,
		"supplier_name":     po.SupplierName,
		"warehouse_name":    po.WarehouseName,
		"items":             itemsOut,
	}

	return response.NewSuccessResponse(data, "")
}

func (s *PurchaseService) CreatePurchaseOrder(input CreatePurchaseOrderInput) response.ApiResponse {
	log.Printf("[DEBUG] CreatePurchaseOrder called with %d items", len(input.Items))
	for i, item := range input.Items {
		log.Printf("[DEBUG] Input item %d: product_id=%s, quantity=%d, unit_price=%f",
			i, item.ProductID, item.Quantity, item.UnitPrice)
	}

	supplierID, err := uuid.Parse(input.SupplierID)
	if err != nil {
		return response.NewErrorResponse("Invalid request data")
	}
	warehouseID, err := uuid.Parse(input.WarehouseID)
	if err != nil {
		return response.NewErrorResponse("Invalid request data")
	}
	createdBy, err := uuid.Parse(input.CreatedBy)
	if err != nil {
		return response.NewErrorResponse("Invalid request data")
	}
	companyID, err := uuid.Parse(input.CompanyID)
	if err != nil {
		return response.NewErrorResponse("Invalid request data")
	}

	// Validate company existence
	var company models.Company
	err = s.db.Table("companies").Where("id = ?", companyID).Limit(1).Scan(&company).Error
	if err != nil {
		return response.NewErrorResponse("Failed to validate company existence")
	}
	if company.ID == uuid.Nil {
		return response.NewErrorResponse("Company dengan ID tersebut tidak ditemukan")
	}

	// Generate PO Number: PO-{YY}-{4digit company_id}-{6digit sequence}
	// Example: PO-26-cabf-000001, PO-26-cabf-000002, ...
	companyIDStr := companyID.String()
	companyPrefix := strings.ToUpper(companyIDStr[len(companyIDStr)-4:])
	year := time.Now().Format("06")

	var lastPO struct {
		PoNumber string `gorm:"column:po_number"`
	}
	s.db.Table("purchase_orders").
		Where("po_number LIKE ?", fmt.Sprintf("PO-%s-%s-%%", year, companyPrefix)).
		Order("po_number DESC").
		Limit(1).Scan(&lastPO)

	sequence := 1
	if lastPO.PoNumber != "" {
		parts := strings.Split(lastPO.PoNumber, "-")
		if len(parts) >= 3 {
			seqStr := parts[len(parts)-1]
			var digits string
			for _, c := range seqStr {
				if c >= '0' && c <= '9' {
					digits += string(c)
				}
			}
			if digits != "" {
				if seq, err := strconv.Atoi(digits); err == nil {
					sequence = seq + 1
				}
			}
		}
	}

	poNumber := fmt.Sprintf("PO-%s-%s-%06d", year, companyPrefix, sequence)

	var createdPOID uuid.UUID
	err = s.db.Transaction(func(tx *gorm.DB) error {
		po := map[string]interface{}{
			"po_number":         poNumber,
			"supplier_id":       supplierID,
			"warehouse_id":      warehouseID,
			"expected_delivery": input.ExpectedDate,
			"status_po":         "DRAFT",
			"status_receive":    "DRAFT",
			"subtotal":          0,
			"tax_amount":        0,
			"total_amount":      0,
			"notes":             nil,
			"created_by":        createdBy,
			"company_id":        companyID,
			"created_at":        time.Now(),
			"updated_at":        time.Now(),
		}
		if input.Notes != nil {
			po["notes"] = *input.Notes
		}

		if err := tx.Table("purchase_orders").Create(po).Error; err != nil {
			return err
		}

		// Always resolve created PO id from generated po_number.
		// Scanning RETURNING into map is driver-dependent (uuid may come back as []byte/string).
		var tmp struct {
			ID uuid.UUID `gorm:"column:id"`
		}
		if err := tx.Table("purchase_orders").Select("id").Where("po_number = ?", poNumber).Limit(1).Scan(&tmp).Error; err != nil {
			return err
		}
		createdPOID = tmp.ID
		if createdPOID == uuid.Nil {
			return fmt.Errorf("failed to resolve created purchase order id")
		}

		subtotal := 0.0
		taxAmount := 0.0
		for _, item := range input.Items {
			pid, err := uuid.Parse(item.ProductID)
			if err != nil {
				return err
			}
			lineSubtotal := float64(item.Quantity) * item.UnitPrice * (1 - item.Discount/100)
			lineTax := lineSubtotal * (item.TaxRate / 100)
			subtotal += lineSubtotal
			taxAmount += lineTax

			poi := map[string]interface{}{
				"id":            uuid.New(),
				"po_id":         createdPOID,
				"product_id":    pid,
				"qty_po":        item.Quantity,
				"unit_price":    item.UnitPrice,
				"discount_rate": item.Discount,
				"tax_rate":      item.TaxRate,
				"qty_receive":   0,
			}
			log.Printf("[DEBUG] Creating item for po_id=%s, product_id=%s, qty=%d, unit_price=%f",
				createdPOID.String(), pid.String(), item.Quantity, item.UnitPrice)
			if err := tx.Table("purchase_order_items").Create(poi).Error; err != nil {
				log.Printf("[ERROR] Failed to create item: %v", err)
				return err
			}
		}

		totalAmount := subtotal + taxAmount
		log.Printf("[DEBUG] Calculated totals: subtotal=%f, taxAmount=%f, totalAmount=%f", subtotal, taxAmount, totalAmount)
		if err := tx.Table("purchase_orders").
			Where("id = ?", createdPOID).
			Updates(map[string]interface{}{"subtotal": subtotal, "tax_amount": taxAmount, "total_amount": totalAmount, "updated_at": time.Now()}).Error; err != nil {
			log.Printf("[ERROR] Failed to update PO totals: %v", err)
			return err
		}

		return nil
	})

	if err != nil {
		return response.NewErrorResponse(err.Error())
	}

	po, err := s.purchaseRepo.GetPurchaseOrderByID(createdPOID)
	if err != nil || po == nil {
		return response.NewErrorResponse("Failed to fetch created purchase order")
	}

	items, _ := s.purchaseRepo.GetPurchaseOrderItems(createdPOID)
	itemsOut := make([]map[string]interface{}, 0, len(items))
	for _, it := range items {
		itemsOut = append(itemsOut, map[string]interface{}{
			"id":           it.ID,
			"po_id":        it.PoID,
			"product_id":   it.ProductID,
			"qty_po":       it.QtyPo,
			"qty_receive":  it.QtyReceive,
			"unit_price":   toFloat(it.UnitPrice),
			"discount":     toFloat(it.Discount),
			"tax_rate":     toFloat(it.TaxRate),
			"line_total":   toFloat(it.LineTotal),
			"product_name": it.ProductName,
			"sku":          it.SKU,
		})
	}

	data := map[string]interface{}{
		"id":                po.ID,
		"po_number":         po.PoNumber,
		"supplier_id":       po.SupplierID,
		"warehouse_id":      po.WarehouseID,
		"order_date":        po.OrderDate,
		"expected_delivery": po.ExpectedDelivery,
		"receive_date":      po.ReceiveDate,
		"payment_terms":     po.PaymentTerms,
		"status_po":         strings.ToLower(po.StatusPo),
		"status_receive":    strings.ToLower(po.StatusReceive),
		"receive_number":    po.ReceiveNumber,
		"note_receive":      po.NoteReceive,
		"subtotal":          toFloat(po.Subtotal),
		"tax_amount":        toFloat(po.TaxAmount),
		"discount_amount":   toFloat(po.DiscountAmount),
		"total_amount":      toFloat(po.TotalAmount),
		"notes":             po.Notes,
		"created_by":        po.CreatedBy,
		"company_id":        po.CompanyID,
		"created_at":        po.CreatedAt,
		"updated_at":        po.UpdatedAt,
		"supplier_name":     po.SupplierName,
		"warehouse_name":    po.WarehouseName,
		"items":             itemsOut,
	}

	return response.NewSuccessResponse(data, "")
}

func (s *PurchaseService) ApprovePurchaseOrder(id string) response.ApiResponse {
	poID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Purchase order not found")
	}

	po, err := s.purchaseRepo.GetPurchaseOrderByID(poID)
	if err != nil || po == nil {
		return response.NewErrorResponse("Purchase order not found")
	}

	if normalizeStatusPo(po.StatusPo) != "DRAFT" {
		return response.NewErrorResponse("Only DRAFT purchase orders can be approved")
	}

	err = s.purchaseRepo.UpdatePurchaseOrder(poID, map[string]interface{}{
		"status_po":  "APPROVE",
		"updated_at": time.Now(),
	})
	if err != nil {
		return response.NewErrorResponse("Failed to approve purchase order")
	}

	return s.GetPurchaseOrderByID(id)
}

func (s *PurchaseService) SetPendingPurchaseOrder(id string) response.ApiResponse {
	poID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Purchase order not found")
	}

	po, err := s.purchaseRepo.GetPurchaseOrderByID(poID)
	if err != nil || po == nil {
		return response.NewErrorResponse("Purchase order not found")
	}

	currentStatus := normalizeStatusPo(po.StatusPo)
	if currentStatus != "DRAFT" && currentStatus != "APPROVE" {
		return response.NewErrorResponse("Only DRAFT or APPROVE purchase orders can be set to PENDING")
	}

	err = s.purchaseRepo.UpdatePurchaseOrder(poID, map[string]interface{}{
		"status_po":  "PENDING",
		"updated_at": time.Now(),
	})
	if err != nil {
		return response.NewErrorResponse("Failed to set purchase order to PENDING")
	}

	return s.GetPurchaseOrderByID(id)
}

func (s *PurchaseService) ReceivePurchaseOrder(id string, input ReceivePurchaseOrderInput) response.ApiResponse {
	poID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Purchase order not found")
	}

	po, err := s.purchaseRepo.GetPurchaseOrderByID(poID)
	if err != nil || po == nil {
		return response.NewErrorResponse("Purchase order not found")
	}

	if normalizeStatusPo(po.StatusPo) != "APPROVE" {
		return response.NewErrorResponse("Only APPROVED purchase orders can receive items")
	}

	if normalizeStatusReceive(po.StatusReceive) == "RECEIVE" {
		return response.NewErrorResponse("Purchase order already fully received")
	}

	items, err := s.purchaseRepo.GetPurchaseOrderItems(poID)
	if err != nil {
		return response.NewErrorResponse("Failed to get purchase order items")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		hasPartial := false

		for _, reqItem := range input.Items {
			itemID, err := uuid.Parse(reqItem.ID)
			if err != nil {
				continue
			}

			for i := range items {
				if items[i].ID == itemID {
					oldReceive := items[i].QtyReceive
					items[i].QtyReceive = reqItem.QtyReceive

					err := tx.Table("purchase_order_items").
						Where("id = ?", itemID).
						Updates(map[string]interface{}{
							"qty_receive": reqItem.QtyReceive,
						}).Error
					if err != nil {
						return err
					}

					if reqItem.QtyReceive > oldReceive && normalizeStatusReceive(input.StatusReceive) == "RECEIVE" {
						qtyToAdd := reqItem.QtyReceive - oldReceive
						if qtyToAdd > 0 {
							err := tx.Model(&models.Inventory{}).
								Where("product_id = ? AND warehouse_id = ?", items[i].ProductID, po.WarehouseID).
								UpdateColumn("quantity", gorm.Expr("quantity + ?", qtyToAdd)).Error
							if err != nil {
								return err
							}

							movement := models.StockMovement{
								ID:            uuid.New(),
								ProductID:     items[i].ProductID,
								WarehouseID:   po.WarehouseID,
								MovementType:  models.MovementTypePurchase,
								Quantity:      qtyToAdd,
								ReferenceType: "PO",
								ReferenceID:   &poID,
								Notes:         "Receive from PO",
								CreatedBy:     &po.CreatedBy,
							}
							if err := tx.Create(&movement).Error; err != nil {
								return err
							}
						}
					}

					if items[i].QtyPo > reqItem.QtyReceive {
						hasPartial = true
					}
					break
				}
			}
		}

		noteReceive := "COMPLETE"
		if hasPartial {
			noteReceive = "PARTIAL"
		}

		updates := map[string]interface{}{
			"status_receive": input.StatusReceive,
			"note_receive":   noteReceive,
			"updated_at":     time.Now(),
		}

		if normalizeStatusReceive(input.StatusReceive) == "RECEIVE" && !input.ReceiveDate.IsZero() {
			updates["receive_date"] = input.ReceiveDate
		}

		if err := tx.Table("purchase_orders").Where("id = ?", poID).Updates(updates).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return response.NewErrorResponse("Failed to receive purchase order: " + err.Error())
	}

	return s.GetPurchaseOrderByID(id)
}

func (s *PurchaseService) UpdatePurchaseOrder(id string, input UpdatePurchaseOrderInput) response.ApiResponse {
	poID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Purchase order not found")
	}

	supplierID, err := uuid.Parse(input.SupplierID)
	if err != nil {
		return response.NewErrorResponse("Invalid request data")
	}
	warehouseID, err := uuid.Parse(input.WarehouseID)
	if err != nil {
		return response.NewErrorResponse("Invalid request data")
	}
	if len(input.Items) == 0 {
		return response.NewErrorResponse("Invalid request data")
	}

	po, err := s.purchaseRepo.GetPurchaseOrderByID(poID)
	if err != nil || po == nil {
		return response.NewErrorResponse("Purchase order not found")
	}
	log.Printf("[DEBUG] UpdatePurchaseOrder: poID=%s, status_po=%s, receive_number=%v", poID, po.StatusPo, po.ReceiveNumber)
	// Allow update if status is DRAFT, PENDING, or APPROVE (for receive_number generation)
	currentStatus := normalizeStatusPo(po.StatusPo)
	if currentStatus != "DRAFT" && currentStatus != "PENDING" && currentStatus != "APPROVE" {
		log.Printf("[DEBUG] UpdatePurchaseOrder: Invalid status_po=%s (normalized=%s)", po.StatusPo, currentStatus)
		return response.NewErrorResponse("Purchase order cannot be updated")
	}

	// Normalize incoming statuses (frontend may send lowercase/synonyms)
	input.StatusPo = normalizeStatusPo(input.StatusPo)
	input.StatusReceive = normalizeStatusReceive(input.StatusReceive)

	// Generate receive number if status_po=approved and receive_number is null
	var receiveNumber string
	// Check if receive_number is empty (null in DB is nil pointer in Go)
	hasReceiveNumber := po.ReceiveNumber != nil && *po.ReceiveNumber != ""
	if input.StatusPo == "APPROVE" && !hasReceiveNumber {
		// Get company_id from the existing PO
		companyID := po.CompanyID
		companyIDStr := companyID.String()
		companyPrefix := strings.ToUpper(companyIDStr[len(companyIDStr)-4:])
		year := time.Now().Format("06")

		// Get the last receive number for this company
		var lastRN struct {
			ReceiveNumber string `gorm:"column:receive_number"`
		}
		s.db.Table("purchase_orders").
			Where("receive_number LIKE ?", fmt.Sprintf("RN-%s-%s-%%", year, companyPrefix)).
			Order("receive_number DESC").
			Limit(1).Scan(&lastRN)

		sequence := 1
		if lastRN.ReceiveNumber != "" {
			parts := strings.Split(lastRN.ReceiveNumber, "-")
			if len(parts) >= 3 {
				seqStr := parts[len(parts)-1]
				var digits string
				for _, c := range seqStr {
					if c >= '0' && c <= '9' {
						digits += string(c)
					}
				}
				if digits != "" {
					if seq, err := strconv.Atoi(digits); err == nil {
						sequence = seq + 1
					}
				}
			}
		}

		receiveNumber = fmt.Sprintf("RN-%s-%s-%06d", year, companyPrefix, sequence)
	}

	itemsExisting, _ := s.purchaseRepo.GetPurchaseOrderItems(poID)
	for _, it := range itemsExisting {
		if it.QtyReceive > 0 {
			log.Printf("[DEBUG] UpdatePurchaseOrder: Item %s has qty_receive=%d", it.ID, it.QtyReceive)
			return response.NewErrorResponse("Purchase order cannot be updated")
		}
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"supplier_id":       supplierID,
			"warehouse_id":      warehouseID,
			"order_date":        input.OrderDate,
			"expected_delivery": input.ExpectedDate,
			"updated_at":        time.Now(),
		}
		if input.Notes != nil {
			updates["notes"] = *input.Notes
		} else {
			updates["notes"] = nil
		}
		if input.StatusPo != "" {
			updates["status_po"] = input.StatusPo
		}
		if input.StatusReceive != "" {
			updates["status_receive"] = input.StatusReceive
		}
		// Add receive_number to updates if generated
		if receiveNumber != "" {
			updates["receive_number"] = receiveNumber
		}

		if err := tx.Table("purchase_orders").Where("id = ?", poID).Updates(updates).Error; err != nil {
			return err
		}

		itemIDsToKeep := make([]string, 0, len(input.Items))
		for _, item := range input.Items {
			if item.ID != "" {
				itemIDsToKeep = append(itemIDsToKeep, item.ID)
			}
		}

		// Only delete items that are NOT in the new items list
		// This preserves existing items that are still in the update
		if len(itemIDsToKeep) > 0 {
			notInClause := ""
			for i, itemID := range itemIDsToKeep {
				if i > 0 {
					notInClause += ","
				}
				notInClause += "'" + itemID + "'"
			}
			tx.Exec("DELETE FROM purchase_order_items WHERE po_id = ? AND id NOT IN ("+notInClause+")", poID)
		} else {
			// No items to keep means delete all
			tx.Exec("DELETE FROM purchase_order_items WHERE po_id = ?", poID)
		}

		subtotal := 0.0
		taxAmount := 0.0
		for _, item := range input.Items {
			pid, err := uuid.Parse(item.ProductID)
			if err != nil {
				continue
			}
			lineSubtotal := float64(item.Quantity) * item.UnitPrice * (1 - item.Discount/100)
			lineTax := lineSubtotal * (item.TaxRate / 100)
			subtotal += lineSubtotal
			taxAmount += lineTax

			if item.ID != "" {
				itemID, err := uuid.Parse(item.ID)
				if err == nil {
					tx.Table("purchase_order_items").Where("id = ?", itemID).Updates(map[string]interface{}{
						"product_id":    pid,
						"qty_po":        item.Quantity,
						"unit_price":    item.UnitPrice,
						"discount_rate": item.Discount,
						"tax_rate":      item.TaxRate,
					})
					continue
				}
			}

			poi := map[string]interface{}{
				"id":            uuid.New(),
				"po_id":         poID,
				"product_id":    pid,
				"qty_po":        item.Quantity,
				"unit_price":    item.UnitPrice,
				"discount_rate": item.Discount,
				"tax_rate":      item.TaxRate,
				"qty_receive":   0,
			}
			if err := tx.Table("purchase_order_items").Create(poi).Error; err != nil {
				return err
			}
		}

		totalAmount := subtotal + taxAmount
		if err := tx.Table("purchase_orders").Where("id = ?", poID).
			Updates(map[string]interface{}{
				"subtotal":     subtotal,
				"tax_amount":   taxAmount,
				"total_amount": totalAmount,
				"updated_at":   time.Now(),
			}).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return response.NewErrorResponse(err.Error())
	}

	return s.GetPurchaseOrderByID(id)
}

func (s *PurchaseService) UpdatePurchaseOrderStatus(id string, status string) response.ApiResponse {
	poID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Purchase order not found")
	}

	var result map[string]interface{}
	normalized := normalizeStatusPo(status)
	res := s.db.Raw("UPDATE purchase_orders SET status_po = ? WHERE id = ? RETURNING *", normalized, poID).Scan(&result)
	if res.Error != nil {
		return response.NewErrorResponse("Purchase order not found")
	}
	if len(result) == 0 {
		return response.NewErrorResponse("Purchase order not found")
	}
	return response.NewSuccessResponse(result, "")
}

func (s *PurchaseService) CancelPurchaseOrder(id string) response.ApiResponse {
	poID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Purchase order not found")
	}

	res := s.db.Table("purchase_orders").Where("id = ?", poID).Update("status_po", "CANCELLED")
	if res.Error != nil {
		return response.NewErrorResponse("Purchase order not found")
	}
	if res.RowsAffected == 0 {
		return response.NewErrorResponse("Purchase order not found")
	}
	return response.NewSuccessResponse(nil, "Purchase order cancelled successfully")
}

func (s *PurchaseService) DeletePurchaseOrder(id string) response.ApiResponse {
	poID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Invalid purchase order ID")
	}

	if err := s.purchaseRepo.Delete(poID); err != nil {
		return response.NewErrorResponse("Failed to delete purchase order: " + err.Error())
	}
	return response.NewSuccessResponse(nil, "Purchase order deleted successfully")
}

func toFloat(v string) float64 {
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

type CreatePurchaseReturnItemInput struct {
	PoItemID  string
	ProductID string
	Quantity  int
	UnitPrice float64
	Amount    float64
	Notes     string
}

type CreatePurchaseReturnInput struct {
	PoID        string
	SupplierID  string
	WarehouseID string
	CompanyID   string
	ReturnDate  string
	Reason      string
	Items       []CreatePurchaseReturnItemInput
}

func (s *PurchaseService) CreatePurchaseReturn(input CreatePurchaseReturnInput, userID string) response.ApiResponse {
	var poID uuid.UUID
	var po *repository.PurchaseOrderRow
	var err error
	var companyID uuid.UUID

	// Get companyID from input (JWT user)
	if input.CompanyID != "" {
		companyID, err = uuid.Parse(input.CompanyID)
		if err != nil {
			return response.NewErrorResponse("Invalid company ID")
		}
	}

	// If po_id provided, get company from PO if not set
	if input.PoID != "" {
		poID, err = uuid.Parse(input.PoID)
		if err != nil {
			return response.NewErrorResponse("Invalid PO ID")
		}
		po, err = s.purchaseRepo.GetPurchaseOrderByID(poID)
		if err != nil || po == nil {
			return response.NewErrorResponse("Purchase order not found")
		}
		if companyID == uuid.Nil {
			companyID = po.CompanyID
		}
	}

	supplierID, err := uuid.Parse(input.SupplierID)
	if err != nil {
		return response.NewErrorResponse("Invalid supplier ID")
	}

	warehouseID, err := uuid.Parse(input.WarehouseID)
	if err != nil {
		return response.NewErrorResponse("Invalid warehouse ID")
	}

	if companyID == uuid.Nil {
		return response.NewErrorResponse("Company ID is required")
	}

	returnNumber, _ := s.purchaseRepo.GetNextReturnNumber(companyID.String())
	uid, _ := uuid.Parse(userID)

	pr := models.PurchaseReturn{
		ID:           uuid.New(),
		ReturnNumber: returnNumber,
		PoID:         poID,
		SupplierID:   supplierID,
		WarehouseID:  warehouseID,
		CompanyID:    companyID,
		Status:       models.PurchaseReturnStatusDraft,
		Reason:       input.Reason,
		CreatedBy:    uid,
	}

	if input.ReturnDate != "" {
		t, err := time.Parse("2006-01-02", input.ReturnDate)
		if err == nil {
			pr.ReturnDate = t
		}
	}

	if err := s.purchaseRepo.CreatePurchaseReturn(&pr); err != nil {
		return response.NewErrorResponse("Failed to create purchase return: " + err.Error())
	}

	totalAmount := 0.0
	for _, item := range input.Items {
		poItemID, _ := uuid.Parse(item.PoItemID)
		productID, _ := uuid.Parse(item.ProductID)

		returnItem := models.PurchaseReturnItem{
			ID:        uuid.New(),
			ReturnID:  pr.ID,
			PoItemID:  poItemID,
			ProductID: productID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Amount:    item.Amount,
			Notes:     item.Notes,
		}
		s.db.Create(&returnItem)
		totalAmount += item.Amount
	}

	s.purchaseRepo.UpdatePurchaseReturn(pr.ID, map[string]interface{}{
		"total_amount": totalAmount,
	})

	return s.GetPurchaseReturnByID(pr.ID.String())
}

func (s *PurchaseService) GetPurchaseReturns(companyID *string, filters map[string]string, limit, offset int) response.PaginatedResponse {
	filtersIfc := make(map[string]interface{})
	for k, v := range filters {
		filtersIfc[k] = v
	}
	rows, total, err := s.purchaseRepo.FindPurchaseReturns(companyID, filtersIfc, limit, offset)
	if err != nil {
		return response.PaginatedResponse{
			Success: false,
			Data:    []interface{}{},
			Pagination: response.Pagination{
				Total:   0,
				Limit:   limit,
				Offset:  offset,
				HasMore: false,
			},
		}
	}

	data := make([]interface{}, len(rows))
	for i, row := range rows {
		items, _ := s.purchaseRepo.GetPurchaseReturnItems(row.ID)
		data[i] = map[string]interface{}{
			"id":             row.ID,
			"return_number":  row.ReturnNumber,
			"po_id":          row.PoID,
			"po_number":      row.PoNumber,
			"supplier_id":    row.SupplierID,
			"supplier_name":  row.SupplierName,
			"warehouse_id":   row.WarehouseID,
			"warehouse_name": row.WarehouseName,
			"company_id":     row.CompanyID,
			"return_date":    row.ReturnDate.Format("2006-01-02"),
			"status":         row.Status,
			"reason":         row.Reason,
			"total_amount":   row.TotalAmount,
			"created_by":     row.CreatedBy,
			"approved_by":    row.ApprovedBy,
			"created_at":     row.CreatedAt,
			"updated_at":     row.UpdatedAt,
			"items":          items,
		}
	}

	return response.NewPaginatedResponse(data, total, limit, offset)
}

func (s *PurchaseService) GetPurchaseReturnByID(id string) response.ApiResponse {
	prID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Invalid purchase return ID")
	}

	pr, err := s.purchaseRepo.GetPurchaseReturnByID(prID)
	if err != nil || pr == nil {
		return response.NewErrorResponse("Purchase return not found")
	}

	items, _ := s.purchaseRepo.GetPurchaseReturnItems(pr.ID)

	return response.NewSuccessResponse(map[string]interface{}{
		"id":             pr.ID,
		"return_number":  pr.ReturnNumber,
		"po_id":          pr.PoID,
		"po_number":      pr.PoNumber,
		"supplier_id":    pr.SupplierID,
		"supplier_name":  pr.SupplierName,
		"warehouse_id":   pr.WarehouseID,
		"warehouse_name": pr.WarehouseName,
		"company_id":     pr.CompanyID,
		"return_date":    pr.ReturnDate.Format("2006-01-02"),
		"status":         pr.Status,
		"reason":         pr.Reason,
		"total_amount":   pr.TotalAmount,
		"created_by":     pr.CreatedBy,
		"approved_by":    pr.ApprovedBy,
		"created_at":     pr.CreatedAt,
		"updated_at":     pr.UpdatedAt,
		"items":          items,
	}, "Purchase return retrieved successfully")
}

func (s *PurchaseService) UpdatePurchaseReturn(id string, input CreatePurchaseReturnInput) response.ApiResponse {
	prID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Invalid purchase return ID")
	}

	pr, err := s.purchaseRepo.GetPurchaseReturnByID(prID)
	if err != nil || pr == nil {
		return response.NewErrorResponse("Purchase return not found")
	}

	if pr.Status != models.PurchaseReturnStatusDraft {
		return response.NewErrorResponse("Only draft purchase returns can be updated")
	}

	updates := map[string]interface{}{
		"reason": input.Reason,
	}

	if input.ReturnDate != "" {
		t, err := time.Parse("2006-01-02", input.ReturnDate)
		if err == nil {
			updates["return_date"] = t
		}
	}

	if err := s.purchaseRepo.UpdatePurchaseReturn(prID, updates); err != nil {
		return response.NewErrorResponse("Failed to update purchase return")
	}

	if len(input.Items) > 0 {
		s.db.Where("return_id = ?", prID).Delete(&models.PurchaseReturnItem{})

		totalAmount := 0.0
		for _, item := range input.Items {
			poItemID, _ := uuid.Parse(item.PoItemID)
			productID, _ := uuid.Parse(item.ProductID)

			returnItem := models.PurchaseReturnItem{
				ID:        uuid.New(),
				ReturnID:  prID,
				PoItemID:  poItemID,
				ProductID: productID,
				Quantity:  item.Quantity,
				UnitPrice: item.UnitPrice,
				Amount:    item.Amount,
				Notes:     item.Notes,
			}
			s.db.Create(&returnItem)
			totalAmount += item.Amount
		}

		s.purchaseRepo.UpdatePurchaseReturn(prID, map[string]interface{}{
			"total_amount": totalAmount,
		})
	}

	return s.GetPurchaseReturnByID(id)
}

func (s *PurchaseService) UpdatePurchaseReturnStatus(id, status, userID string) response.ApiResponse {
	prID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Invalid purchase return ID")
	}

	pr, err := s.purchaseRepo.GetPurchaseReturnByID(prID)
	if err != nil || pr == nil {
		return response.NewErrorResponse("Purchase return not found")
	}

	currentStatus := pr.Status

	if status == "APPROVED" && currentStatus == models.PurchaseReturnStatusDraft {
		uid, _ := uuid.Parse(userID)
		s.purchaseRepo.UpdatePurchaseReturn(prID, map[string]interface{}{
			"status":      models.PurchaseReturnStatusApproved,
			"approved_by": uid,
		})
	} else if status == "DONE" && currentStatus == models.PurchaseReturnStatusApproved {
		items, _ := s.purchaseRepo.GetPurchaseReturnItems(prID)
		for _, item := range items {
			inventory, _ := s.inventoryRepo.FindByProductAndWarehouse(item.ProductID, pr.WarehouseID)
			if inventory != nil {
				newQty := inventory.Quantity - item.Quantity
				s.inventoryRepo.UpdateQuantity(item.ProductID, pr.WarehouseID, newQty)

				movement := models.StockMovement{
					ID:            uuid.New(),
					ProductID:     item.ProductID,
					WarehouseID:   pr.WarehouseID,
					MovementType:  models.MovementTypeReturn,
					Quantity:      -item.Quantity,
					ReferenceType: "PURCHASE_RETURN",
					ReferenceID:   &prID,
					Notes:         "Purchase Return",
				}
				s.inventoryRepo.CreateStockMovement(&movement)
			}
		}
		s.purchaseRepo.UpdatePurchaseReturn(prID, map[string]interface{}{
			"status": models.PurchaseReturnStatusDone,
		})
	} else if status == "CANCELLED" && currentStatus == models.PurchaseReturnStatusDraft {
		s.purchaseRepo.UpdatePurchaseReturn(prID, map[string]interface{}{
			"status": models.PurchaseReturnStatusCancelled,
		})
	} else {
		return response.NewErrorResponse("Invalid status transition")
	}

	return s.GetPurchaseReturnByID(id)
}

func (s *PurchaseService) DeletePurchaseReturn(id string) response.ApiResponse {
	prID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Invalid purchase return ID")
	}

	pr, err := s.purchaseRepo.GetPurchaseReturnByID(prID)
	if err != nil || pr == nil {
		return response.NewErrorResponse("Purchase return not found")
	}

	if pr.Status != models.PurchaseReturnStatusDraft {
		return response.NewErrorResponse("Only draft purchase returns can be deleted")
	}

	if err := s.purchaseRepo.DeletePurchaseReturn(prID); err != nil {
		return response.NewErrorResponse("Failed to delete purchase return: " + err.Error())
	}

	return response.NewSuccessResponse(nil, "Purchase return deleted successfully")
}
