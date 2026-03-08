package services

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/repository"
	"github.com/pos-retail/go_backend/internal/types/response"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (s *PurchaseService) GetPurchaseOrders(filters map[string]string, limit, offset int) response.PaginatedResponse {
	rows, total, err := s.purchaseRepo.FindPurchaseOrders(filters, limit, offset)
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
			"status_po":         po.StatusPo,
			"status_receive":    po.StatusReceive,
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
		"status_po":         po.StatusPo,
		"status_receive":    po.StatusReceive,
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

	poNumber := fmt.Sprintf("PO-%d", time.Now().UnixMilli())

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

		var poRow map[string]interface{}
		if err := tx.Table("purchase_orders").Clauses(clause.Returning{}).Create(po).Scan(&poRow).Error; err != nil {
			return err
		}

		var ok bool
		createdPOID, ok = poRow["id"].(uuid.UUID)
		if !ok {
			var idStr string
			if v, ok := poRow["id"].(string); ok {
				idStr = v
			}
			if idStr != "" {
				if p, err := uuid.Parse(idStr); err == nil {
					createdPOID = p
				}
			}
		}
		if createdPOID == uuid.Nil {
			var tmp struct {
				ID uuid.UUID `gorm:"column:id"`
			}
			_ = tx.Table("purchase_orders").Select("id").Where("po_number = ?", poNumber).Limit(1).Scan(&tmp).Error
			createdPOID = tmp.ID
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
				"po_id":         createdPOID,
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
		if err := tx.Table("purchase_orders").
			Where("id = ?", createdPOID).
			Updates(map[string]interface{}{"subtotal": subtotal, "tax_amount": taxAmount, "total_amount": totalAmount, "updated_at": time.Now()}).Error; err != nil {
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
		"status_po":         po.StatusPo,
		"status_receive":    po.StatusReceive,
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

	if po.StatusPo != "DRAFT" {
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

	if po.StatusPo != "DRAFT" && po.StatusPo != "APPROVE" {
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

	if po.StatusPo != "APPROVE" {
		return response.NewErrorResponse("Only APPROVED purchase orders can receive items")
	}

	if po.StatusReceive == "RECEIVE" {
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

					if reqItem.QtyReceive > oldReceive && input.StatusReceive == "RECEIVE" {
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
								MovementType:  "IN",
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

		if input.StatusReceive == "RECEIVE" && !input.ReceiveDate.IsZero() {
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
	if po.StatusPo != "DRAFT" && po.StatusPo != "PENDING" {
		return response.NewErrorResponse("Purchase order cannot be updated")
	}

	itemsExisting, _ := s.purchaseRepo.GetPurchaseOrderItems(poID)
	for _, it := range itemsExisting {
		if it.QtyReceive > 0 {
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

		if err := tx.Table("purchase_orders").Where("id = ?", poID).Updates(updates).Error; err != nil {
			return err
		}

		itemIDsToKeep := make([]string, 0, len(input.Items))
		for _, item := range input.Items {
			if item.ID != "" {
				itemIDsToKeep = append(itemIDsToKeep, item.ID)
			}
		}

		deleteQuery := tx.Exec("DELETE FROM purchase_order_items WHERE po_id = ?", poID)
		if deleteQuery.Error != nil {
			return deleteQuery.Error
		}
		if len(itemIDsToKeep) > 0 {
			notInClause := ""
			for i, itemID := range itemIDsToKeep {
				if i > 0 {
					notInClause += ","
				}
				notInClause += "'" + itemID + "'"
			}
			tx.Exec("DELETE FROM purchase_order_items WHERE po_id = ? AND id NOT IN ("+notInClause+")", poID)
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
	res := s.db.Raw("UPDATE purchase_orders SET status_po = ? WHERE id = ? RETURNING *", status, poID).Scan(&result)
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
