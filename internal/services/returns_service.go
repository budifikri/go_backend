package services

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/repository"
	"github.com/pos-retail/go_backend/internal/types/response"
	"gorm.io/gorm"
)

type ReturnsService struct {
	db         *gorm.DB
	returnRepo *repository.ReturnsRepository
}

func NewReturnsService(db *gorm.DB, returnRepo *repository.ReturnsRepository) *ReturnsService {
	return &ReturnsService{db: db, returnRepo: returnRepo}
}

type CreateReturnItemInput struct {
	SaleItemID string
	ProductID  string
	Quantity   int
	Condition  string
	Notes      string
}

type CreateReturnInput struct {
	SaleID       string
	WarehouseID  string
	Reason       string
	Items        []CreateReturnItemInput
	RefundMethod string
}

func (s *ReturnsService) CreateReturn(input CreateReturnInput, processedBy string) response.ApiResponse {
	saleID, err := uuid.Parse(input.SaleID)
	if err != nil {
		return response.NewErrorResponse("Invalid sale ID")
	}
	warehouseID, err := uuid.Parse(input.WarehouseID)
	if err != nil {
		return response.NewErrorResponse("Invalid warehouse ID")
	}
	processedUUID, err := uuid.Parse(processedBy)
	if err != nil {
		return response.NewErrorResponse("Invalid processed_by ID")
	}
	if len(input.Items) == 0 {
		return response.NewErrorResponse("Items are required")
	}

	processedItems := make([]map[string]interface{}, 0, len(input.Items))
	totalRefund := 0.0
	var createdReturnID uuid.UUID
	returnNumber := ""

	err = s.db.Transaction(func(tx *gorm.DB) error {
		var sale struct {
			ID         uuid.UUID         `gorm:"column:id"`
			Status     models.SaleStatus `gorm:"column:status"`
			CustomerID *uuid.UUID        `gorm:"column:customer_id"`
		}
		if err := tx.Table("sales").Select("id, status, customer_id").Where("id = ?", saleID).Limit(1).Scan(&sale).Error; err != nil {
			return err
		}
		if sale.ID == uuid.Nil {
			return fmt.Errorf("Sale not found")
		}
		if string(sale.Status) != "DONE" {
			return fmt.Errorf("Sale is not done")
		}

		processedItems = processedItems[:0]
		totalRefund = 0

		for _, item := range input.Items {
			siID, err := uuid.Parse(item.SaleItemID)
			if err != nil {
				return fmt.Errorf("Sale item not found")
			}
			pid, err := uuid.Parse(item.ProductID)
			if err != nil {
				return fmt.Errorf("Sale item not found")
			}
			if item.Quantity <= 0 {
				return fmt.Errorf("Return quantity exceeds sold quantity")
			}

			var saleItem struct {
				Quantity  int     `gorm:"column:quantity"`
				UnitPrice float64 `gorm:"column:unit_price"`
			}
			err = tx.Table("sale_items").
				Select("quantity, unit_price").
				Where("id = ? AND sale_id = ?", siID, saleID).
				Limit(1).
				Scan(&saleItem).Error
			if err != nil {
				return err
			}
			if saleItem.Quantity == 0 {
				return fmt.Errorf("Sale item not found")
			}
			if item.Quantity > saleItem.Quantity {
				return fmt.Errorf("Return quantity exceeds sold quantity")
			}

			refundAmount := saleItem.UnitPrice * float64(item.Quantity)
			totalRefund += refundAmount

			processedItems = append(processedItems, map[string]interface{}{
				"sale_item_id":  item.SaleItemID,
				"product_id":    item.ProductID,
				"quantity":      item.Quantity,
				"unit_price":    saleItem.UnitPrice,
				"refund_amount": refundAmount,
				"condition":     item.Condition,
				"notes":         item.Notes,
			})
			_ = pid
		}

		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		returnNumber = fmt.Sprintf("RTN-%s-%03d", time.Now().Format("20060102"), rng.Intn(1000))

		sr := models.SalesReturn{
			ID:           uuid.New(),
			ReturnNumber: returnNumber,
			SaleID:       saleID,
			WarehouseID:  warehouseID,
			CustomerID:   sale.CustomerID,
			ReturnDate:   time.Now(),
			Status:       models.ReturnStatusDone,
			Reason:       input.Reason,
			TotalAmount:  totalRefund,
			RefundMethod: models.RefundMethod(input.RefundMethod),
			ProcessedBy:  processedUUID,
		}
		if err := tx.Create(&sr).Error; err != nil {
			return err
		}
		createdReturnID = sr.ID

		for _, item := range processedItems {
			siID, _ := uuid.Parse(item["sale_item_id"].(string))
			pid, _ := uuid.Parse(item["product_id"].(string))
			qty := item["quantity"].(int)
			unitPrice := item["unit_price"].(float64)
			refundAmount := item["refund_amount"].(float64)
			cond := item["condition"].(string)
			notes := ""
			if v, ok := item["notes"].(string); ok {
				notes = v
			}

			sri := models.SalesReturnItem{
				ID:           uuid.New(),
				ReturnID:     sr.ID,
				SaleItemID:   siID,
				ProductID:    pid,
				Quantity:     qty,
				UnitPrice:    unitPrice,
				RefundAmount: refundAmount,
				Condition:    models.ItemCondition(cond),
				Notes:        notes,
			}
			if err := tx.Create(&sri).Error; err != nil {
				return err
			}
		}

		if sale.CustomerID != nil {
			var points struct {
				Points int `gorm:"column:points"`
			}
			_ = tx.Table("sales").
				Select("COALESCE(SUM(loyalty_points_earned), 0) as points").
				Where("customer_id = ? AND id = ?", *sale.CustomerID, saleID).
				Scan(&points).Error
			if points.Points > 0 {
				if err := tx.Model(&models.Customer{}).
					Where("id = ?", *sale.CustomerID).
					Update("loyalty_points", gorm.Expr("loyalty_points - ?", points.Points)).Error; err != nil {
					return err
				}
			}
		}

		var returned struct {
			TotalReturned int `gorm:"column:total_returned"`
		}
		_ = tx.Table("sales_return_items sri").
			Select("COALESCE(SUM(sri.quantity), 0) as total_returned").
			Joins("JOIN sales_returns sr ON sr.id = sri.return_id").
			Where("sr.sale_id = ?", saleID).
			Scan(&returned).Error

		var sold struct {
			TotalSold int `gorm:"column:total_sold"`
		}
		_ = tx.Table("sale_items").
			Select("COALESCE(SUM(quantity), 0) as total_sold").
			Where("sale_id = ?", saleID).
			Scan(&sold).Error

		if returned.TotalReturned >= sold.TotalSold && sold.TotalSold > 0 {
			if err := tx.Table("sales").Where("id = ?", saleID).Update("status", "REFUNDED").Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return response.NewErrorResponse(err.Error())
	}

	return response.NewSuccessResponse(map[string]interface{}{
		"return_id":     createdReturnID,
		"return_number": returnNumber,
		"sale_id":       input.SaleID,
		"return_date":   time.Now(),
		"status":        "DONE",
		"total_refund":  totalRefund,
		"refund_method": input.RefundMethod,
		"items":         processedItems,
	}, "")
}

func (s *ReturnsService) GetReturnByID(id string) response.ApiResponse {
	rid, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Invalid return ID")
	}

	ret, err := s.returnRepo.GetReturnByID(rid)
	if err != nil {
		return response.NewErrorResponse("Failed to get return")
	}
	if ret == nil {
		return response.NewErrorResponse("Return not found")
	}

	items, _ := s.returnRepo.GetReturnItems(rid)

	data := map[string]interface{}{}
	data["id"] = ret.ID
	data["return_number"] = ret.ReturnNumber
	data["sale_id"] = ret.SaleID
	data["warehouse_id"] = ret.WarehouseID
	data["customer_id"] = ret.CustomerID
	data["return_date"] = ret.ReturnDate
	data["status"] = ret.Status
	data["reason"] = ret.Reason
	data["total_amount"] = ret.TotalAmount
	data["refund_method"] = ret.RefundMethod
	data["processed_by"] = ret.ProcessedBy
	data["created_at"] = ret.CreatedAt
	data["updated_at"] = ret.UpdatedAt
	data["sale_number"] = ret.SaleNumber
	data["warehouse_name"] = ret.WarehouseName
	data["processed_by_name"] = ret.ProcessedByName
	data["items"] = items

	return response.NewSuccessResponse(data, "")
}

func (s *ReturnsService) GetReturns(filters map[string]string, limit, offset int) response.PaginatedResponse {
	rows, total, err := s.returnRepo.FindReturns(filters, limit, offset)
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

	return response.NewPaginatedResponse(rows, total, limit, offset)
}
