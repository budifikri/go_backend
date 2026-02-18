package services

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/repository"
	"github.com/pos-retail/go_backend/internal/types/response"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PurchaseService struct {
	db           *gorm.DB
	purchaseRepo *repository.PurchaseRepository
}

func NewPurchaseService(db *gorm.DB, purchaseRepo *repository.PurchaseRepository) *PurchaseService {
	return &PurchaseService{db: db, purchaseRepo: purchaseRepo}
}

type CreatePurchaseOrderItemInput struct {
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

func (s *PurchaseService) GetPurchaseOrders(filters map[string]string, limit, offset int) response.PaginatedResponse {
	rows, total, err := s.purchaseRepo.FindPurchaseOrders(filters, limit, offset)
	if err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}

	// TS converts subtotal/tax_amount/total_amount to Number for list.
	data := make([]map[string]interface{}, 0, len(rows))
	for _, po := range rows {
		m := map[string]interface{}{
			"id":                po.ID,
			"po_number":         po.PoNumber,
			"supplier_id":       po.SupplierID,
			"warehouse_id":      po.WarehouseID,
			"order_date":        po.OrderDate,
			"expected_delivery": po.ExpectedDelivery,
			"payment_terms":     po.PaymentTerms,
			"status":            po.Status,
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
			"id":                it.ID,
			"po_id":             it.PoID,
			"product_id":        it.ProductID,
			"quantity":          it.Quantity,
			"received_quantity": it.ReceivedQuantity,
			"unit_price":        toFloat(it.UnitPrice),
			"discount":          toFloat(it.Discount),
			"tax_rate":          toFloat(it.TaxRate),
			"line_total":        toFloat(it.LineTotal),
			"product_name":      it.ProductName,
			"sku":               it.SKU,
		})
	}

	data := map[string]interface{}{
		"id":                po.ID,
		"po_number":         po.PoNumber,
		"supplier_id":       po.SupplierID,
		"warehouse_id":      po.WarehouseID,
		"order_date":        po.OrderDate,
		"expected_delivery": po.ExpectedDelivery,
		"payment_terms":     po.PaymentTerms,
		"status":            po.Status,
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

	var created map[string]interface{}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		po := map[string]interface{}{
			"po_number":         poNumber,
			"supplier_id":       supplierID,
			"warehouse_id":      warehouseID,
			"expected_delivery": input.ExpectedDate,
			"status":            "DRAFT",
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
		poIDVal, ok := poRow["id"].(uuid.UUID)
		if !ok {
			// fallback fetch
			var idStr string
			if v, ok := poRow["id"].(string); ok {
				idStr = v
			}
			if idStr != "" {
				if p, err := uuid.Parse(idStr); err == nil {
					poIDVal = p
				}
			}
		}
		if poIDVal == uuid.Nil {
			// query by po_number
			var tmp struct {
				ID uuid.UUID `gorm:"column:id"`
			}
			_ = tx.Table("purchase_orders").Select("id").Where("po_number = ?", poNumber).Limit(1).Scan(&tmp).Error
			poIDVal = tmp.ID
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
				"po_id":             poIDVal,
				"product_id":        pid,
				"quantity":          item.Quantity,
				"unit_price":        item.UnitPrice,
				"discount_rate":     item.Discount,
				"tax_rate":          item.TaxRate,
				"received_quantity": 0,
			}
			if err := tx.Table("purchase_order_items").Create(poi).Error; err != nil {
				return err
			}
		}

		totalAmount := subtotal + taxAmount
		if err := tx.Table("purchase_orders").
			Where("id = ?", poIDVal).
			Updates(map[string]interface{}{"subtotal": subtotal, "tax_amount": taxAmount, "total_amount": totalAmount, "updated_at": time.Now()}).Error; err != nil {
			return err
		}

		var updated map[string]interface{}
		if err := tx.Table("purchase_orders").Where("id = ?", poIDVal).Limit(1).Find(&updated).Error; err != nil {
			// ignore
		}
		created = updated
		return nil
	})

	if err != nil {
		return response.NewErrorResponse(err.Error())
	}

	return response.NewSuccessResponse(created, "")
}

func (s *PurchaseService) UpdatePurchaseOrderStatus(id string, status string) response.ApiResponse {
	poID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Purchase order not found")
	}

	var result map[string]interface{}
	res := s.db.Table("purchase_orders").Clauses(clause.Returning{}).Where("id = ?", poID).Update("status", status).Scan(&result)
	if res.Error != nil {
		return response.NewErrorResponse("Purchase order not found")
	}
	if res.RowsAffected == 0 {
		return response.NewErrorResponse("Purchase order not found")
	}
	return response.NewSuccessResponse(result, "")
}

func (s *PurchaseService) CancelPurchaseOrder(id string) response.ApiResponse {
	poID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Purchase order not found")
	}

	res := s.db.Table("purchase_orders").Where("id = ?", poID).Update("status", "CANCELLED")
	if res.Error != nil {
		return response.NewErrorResponse("Purchase order not found")
	}
	if res.RowsAffected == 0 {
		return response.NewErrorResponse("Purchase order not found")
	}
	return response.NewSuccessResponse(nil, "Purchase order cancelled successfully")
}

func toFloat(v string) float64 {
	f, _ := strconv.ParseFloat(v, 64)
	return f
}
