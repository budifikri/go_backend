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

type ExchangesService struct {
	db           *gorm.DB
	exchangeRepo *repository.ExchangesRepository
}

func NewExchangesService(db *gorm.DB, exchangeRepo *repository.ExchangesRepository) *ExchangesService {
	return &ExchangesService{db: db, exchangeRepo: exchangeRepo}
}

type ExchangeReturnedItemInput struct {
	SaleItemID string
	ProductID  string
	Quantity   int
	Condition  string
}

type ExchangeReceivedItemInput struct {
	ProductID string
	Quantity  int
	UnitPrice float64
}

type CreateExchangeInput struct {
	SaleID        string
	WarehouseID   string
	Reason        string
	ReturnedItems []ExchangeReturnedItemInput
	ReceivedItems []ExchangeReceivedItemInput
}

func (s *ExchangesService) CreateExchange(input CreateExchangeInput, processedBy string) response.ApiResponse {
	warehouseID, err := uuid.Parse(input.WarehouseID)
	if err != nil {
		return response.NewErrorResponse("Invalid warehouse ID")
	}
	processedUUID, err := uuid.Parse(processedBy)
	if err != nil {
		return response.NewErrorResponse("Invalid processed_by ID")
	}

	var saleID *uuid.UUID
	if input.SaleID != "" {
		id, err := uuid.Parse(input.SaleID)
		if err != nil {
			return response.NewErrorResponse("Invalid sale ID")
		}
		saleID = &id
	}

	priceDifference := 0.0
	returnedTotal := 0.0

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	exchangeNumber := fmt.Sprintf("EXC-%s-%03d", time.Now().Format("20060102"), rng.Intn(1000))
	var createdExchangeID uuid.UUID

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if saleID != nil {
			var sale struct {
				ID uuid.UUID `gorm:"column:id"`
			}
			if err := tx.Table("sales").Select("id").Where("id = ?", *saleID).Limit(1).Scan(&sale).Error; err != nil {
				return err
			}
			if sale.ID == uuid.Nil {
				return fmt.Errorf("Sale not found")
			}
		}

		priceDifference = 0
		returnedTotal = 0

		for _, item := range input.ReturnedItems {
			if item.SaleItemID != "" {
				siID, err := uuid.Parse(item.SaleItemID)
				if err != nil {
					return fmt.Errorf("Sale item not found")
				}
				var saleItem struct {
					UnitPrice float64 `gorm:"column:unit_price"`
					Quantity  int     `gorm:"column:quantity"`
				}
				if err := tx.Table("sale_items").Select("unit_price, quantity").Where("id = ?", siID).Limit(1).Scan(&saleItem).Error; err != nil {
					return err
				}
				if saleItem.Quantity == 0 {
					return fmt.Errorf("Sale item not found")
				}

				value := saleItem.UnitPrice * float64(item.Quantity)
				returnedTotal += value
				priceDifference -= value
			}
		}

		for _, item := range input.ReceivedItems {
			pid, err := uuid.Parse(item.ProductID)
			if err != nil {
				return fmt.Errorf("Product %s not found", item.ProductID)
			}
			var product struct {
				ID uuid.UUID `gorm:"column:id"`
			}
			if err := tx.Table("products").Select("id").Where("id = ? AND status = 'active'", pid).Limit(1).Scan(&product).Error; err != nil {
				return err
			}
			if product.ID == uuid.Nil {
				return fmt.Errorf("Product %s not found", item.ProductID)
			}

			var inv struct {
				Quantity         int `gorm:"column:quantity"`
				ReservedQuantity int `gorm:"column:reserved_quantity"`
			}
			if err := tx.Table("inventory").Select("quantity, reserved_quantity").Where("product_id = ? AND warehouse_id = ?", pid, warehouseID).Limit(1).Scan(&inv).Error; err != nil {
				return err
			}
			available := inv.Quantity - inv.ReservedQuantity
			if inv.Quantity == 0 && inv.ReservedQuantity == 0 {
				return fmt.Errorf("Insufficient stock for product %s", item.ProductID)
			}
			if available < item.Quantity {
				return fmt.Errorf("Insufficient stock for product %s", item.ProductID)
			}

			value := item.UnitPrice * float64(item.Quantity)
			priceDifference += value
		}

		ex := models.ItemExchange{
			ID:              uuid.New(),
			ExchangeNumber:  exchangeNumber,
			SaleID:          saleID,
			WarehouseID:     warehouseID,
			CustomerID:      nil,
			ExchangeDate:    time.Now(),
			Status:          models.ExchangeStatusDone,
			Reason:          input.Reason,
			PriceDifference: priceDifference,
			ProcessedBy:     processedUUID,
		}
		if err := tx.Create(&ex).Error; err != nil {
			return err
		}
		createdExchangeID = ex.ID

		// returned items
		div := float64(len(input.ReturnedItems))
		perItemValue := 0.0
		if div > 0 {
			perItemValue = returnedTotal / div
		}
		for _, item := range input.ReturnedItems {
			pid, _ := uuid.Parse(item.ProductID)
			var siID *uuid.UUID
			if item.SaleItemID != "" {
				id, _ := uuid.Parse(item.SaleItemID)
				siID = &id
			}
			itemValue := float64(item.Quantity) * perItemValue
			unitPrice := 0.0
			if item.Quantity > 0 {
				unitPrice = itemValue / float64(item.Quantity)
			}

			ei := models.ExchangeItem{
				ID:          uuid.New(),
				ExchangeID:  ex.ID,
				ItemType:    models.ExchangeItemTypeReturned,
				SaleItemID:  siID,
				ProductID:   pid,
				Quantity:    item.Quantity,
				UnitPrice:   unitPrice,
				TotalAmount: itemValue,
				Condition:   models.ItemCondition(item.Condition),
			}
			if err := tx.Create(&ei).Error; err != nil {
				return err
			}
		}

		// received items
		for _, item := range input.ReceivedItems {
			pid, _ := uuid.Parse(item.ProductID)
			ei := models.ExchangeItem{
				ID:          uuid.New(),
				ExchangeID:  ex.ID,
				ItemType:    models.ExchangeItemTypeReceived,
				SaleItemID:  nil,
				ProductID:   pid,
				Quantity:    item.Quantity,
				UnitPrice:   item.UnitPrice,
				TotalAmount: float64(item.Quantity) * item.UnitPrice,
				Condition:   models.ItemConditionNew,
			}
			if err := tx.Create(&ei).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return response.NewErrorResponse(err.Error())
	}

	return response.NewSuccessResponse(map[string]interface{}{
		"exchange_id":      createdExchangeID,
		"exchange_number":  exchangeNumber,
		"exchange_date":    time.Now(),
		"status":           "DONE",
		"price_difference": priceDifference,
		"payment_required": priceDifference > 0,
		"refund_required":  priceDifference < 0,
		"returned_items":   input.ReturnedItems,
		"received_items":   input.ReceivedItems,
	}, "")
}

func (s *ExchangesService) GetExchangeByID(id string) response.ApiResponse {
	eid, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Invalid exchange ID")
	}

	ex, err := s.exchangeRepo.GetExchangeByID(eid)
	if err != nil {
		return response.NewErrorResponse("Failed to get exchange")
	}
	if ex == nil {
		return response.NewErrorResponse("Exchange not found")
	}

	returnedItems, _ := s.exchangeRepo.GetExchangeItems(eid, "RETURNED")
	receivedItems, _ := s.exchangeRepo.GetExchangeItems(eid, "RECEIVED")

	data := map[string]interface{}{}
	data["id"] = ex.ID
	data["exchange_number"] = ex.ExchangeNumber
	data["sale_id"] = ex.SaleID
	data["warehouse_id"] = ex.WarehouseID
	data["customer_id"] = ex.CustomerID
	data["exchange_date"] = ex.ExchangeDate
	data["status"] = ex.Status
	data["reason"] = ex.Reason
	data["price_difference"] = ex.PriceDifference
	data["processed_by"] = ex.ProcessedBy
	data["created_at"] = ex.CreatedAt
	data["updated_at"] = ex.UpdatedAt
	data["warehouse_name"] = ex.WarehouseName
	data["processed_by_name"] = ex.ProcessedByName
	data["returned_items"] = returnedItems
	data["received_items"] = receivedItems

	return response.NewSuccessResponse(data, "")
}

func (s *ExchangesService) GetExchanges(filters map[string]string, limit, offset int) response.PaginatedResponse {
	rows, total, err := s.exchangeRepo.FindExchanges(filters, limit, offset)
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
