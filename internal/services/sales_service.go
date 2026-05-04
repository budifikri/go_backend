package services

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/repository"
	"github.com/pos-retail/go_backend/internal/types/response"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SalesService struct {
	db             *gorm.DB
	salesRepo      *repository.SalesRepository
	cashDrawerRepo *repository.CashDrawerRepository
	telegramRepo   *repository.TelegramRepository
}

func NewSalesService(db *gorm.DB, salesRepo *repository.SalesRepository, cashDrawerRepo *repository.CashDrawerRepository) *SalesService {
	return &SalesService{db: db, salesRepo: salesRepo, cashDrawerRepo: cashDrawerRepo}
}

func NewSalesServiceWithTelegram(db *gorm.DB, salesRepo *repository.SalesRepository, cashDrawerRepo *repository.CashDrawerRepository, telegramRepo *repository.TelegramRepository) *SalesService {
	return &SalesService{db: db, salesRepo: salesRepo, cashDrawerRepo: cashDrawerRepo, telegramRepo: telegramRepo}
}

type CreateSaleItemInput struct {
	ProductID     string
	Quantity      int
	PromotionCode string
}

type CreateSalePaymentInput struct {
	Method          string
	Amount          float64
	ReferenceNumber string
	CardLast4       string
}

type CreateSaleInput struct {
	WarehouseID         string
	CustomerID          string
	CashDrawerID        string
	Status              string
	Items               []CreateSaleItemInput
	Payments            []CreateSalePaymentInput
	LoyaltyPointsRedeem int
	Notes               string
	CompanyID           string
}

type ProcessedSaleItem struct {
	ProductID      uuid.UUID  `json:"product_id"`
	Quantity       int        `json:"quantity"`
	UnitPrice      float64    `json:"unit_price"`
	OriginalPrice  float64    `json:"original_price"`
	CostPrice      float64    `json:"cost_price"`
	DiscountAmount float64    `json:"discount_amount"`
	TaxRate        float64    `json:"tax_rate"`
	LineTotal      float64    `json:"line_total"`
	PromotionID    *uuid.UUID `json:"promotion_id,omitempty"`
	PriceTierID    *uuid.UUID `json:"price_tier_id,omitempty"`
	Notes          string     `json:"notes,omitempty"`
}

type promotionRow struct {
	ID            uuid.UUID `gorm:"column:id"`
	Code          string    `gorm:"column:code"`
	Name          string    `gorm:"column:name"`
	PromotionType string    `gorm:"column:promotion_type"`
	DiscountValue float64   `gorm:"column:discount_value"`
	BuyQuantity   int       `gorm:"column:buy_quantity"`
	GetQuantity   int       `gorm:"column:get_quantity"`
	StartDate     time.Time `gorm:"column:start_date"`
	StartTime     string    `gorm:"column:start_time"`
	EndDate       time.Time `gorm:"column:end_date"`
	EndTime       string    `gorm:"column:end_time"`
	IsActive      bool      `gorm:"column:is_active"`
}

func (s *SalesService) CreateSale(input CreateSaleInput, cashierID string) response.ApiResponse {
	warehouseID, err := uuid.Parse(input.WarehouseID)
	if err != nil {
		return response.NewErrorResponse("Invalid warehouse ID")
	}

	var customerID *uuid.UUID
	if input.CustomerID != "" {
		cid, err := uuid.Parse(input.CustomerID)
		if err != nil {
			return response.NewErrorResponse("Invalid customer ID")
		}
		customerID = &cid
	}

	companyID, err := uuid.Parse(input.CompanyID)
	if err != nil {
		return response.NewErrorResponse("Invalid company ID")
	}

	cashierUUID, err := uuid.Parse(cashierID)
	if err != nil {
		return response.NewErrorResponse("Invalid cashier ID")
	}

	if len(input.Items) == 0 {
		return response.NewErrorResponse("Items are required")
	}
	if len(input.Payments) == 0 {
		return response.NewErrorResponse("Payments are required")
	}

	var createdSale models.Sale
	processed := make([]ProcessedSaleItem, 0, len(input.Items))

	err = s.db.Transaction(func(tx *gorm.DB) error {
		var warehouse models.Warehouse
		if err := tx.First(&warehouse, "id = ?", warehouseID).Error; err != nil {
			return fmt.Errorf("Invalid warehouse")
		}
		if warehouse.Status != models.WarehouseStatusActive {
			return fmt.Errorf("Invalid warehouse")
		}

		subtotal := 0.0
		discountTotal := 0.0
		taxTotal := 0.0

		for _, item := range input.Items {
			pid, err := uuid.Parse(item.ProductID)
			if err != nil {
				return fmt.Errorf("Invalid product ID")
			}
			if item.Quantity <= 0 {
				return fmt.Errorf("Invalid quantity")
			}

			var product models.Product
			if err := tx.First(&product, "id = ?", pid).Error; err != nil {
				return fmt.Errorf("Product %s not found or inactive", item.ProductID)
			}
			if product.Status != models.ProductStatusActive {
				return fmt.Errorf("Product %s not found or inactive", item.ProductID)
			}

			// Lock inventory row to avoid oversell
			var inv models.Inventory
			invErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&inv, "product_id = ? AND warehouse_id = ?", pid, warehouseID).Error
			if invErr != nil {
				return fmt.Errorf("Insufficient stock for product %s", product.Name)
			}
			available := inv.Quantity - inv.ReservedQuantity
			if available < item.Quantity {
				return fmt.Errorf("Insufficient stock for product %s", product.Name)
			}

			retailPrice := product.RetailPrice
			tierPrice := retailPrice
			tierNotes := ""
			var priceTierID *uuid.UUID

			var tier models.PriceTier
			hasTier := false
			_ = tx.
				Where("product_id = ? AND is_active = true AND min_quantity <= ? AND (max_quantity IS NULL OR max_quantity >= ?)", pid, item.Quantity, item.Quantity).
				Order("min_quantity DESC").
				Limit(1).
				Find(&tier).Error
			if tier.ID != uuid.Nil {
				tierPrice = tier.UnitPrice
				tierNotes = fmt.Sprintf("Grosir %d", tier.MinQuantity)
				tid := tier.ID
				priceTierID = &tid
				hasTier = true
			}

			var discountPerUnit float64
			promoNotes := ""
			var promotionID *uuid.UUID
			if item.PromotionCode != "" {
				fmt.Printf("[PROMO] Looking for promo: %s\n", item.PromotionCode)
				var promo promotionRow
				pErr := tx.Raw(
					"SELECT id, code, name, promotion_type, discount_value, buy_quantity, get_quantity FROM promotions WHERE LOWER(code) = LOWER(?) AND is_active = true LIMIT 1",
					item.PromotionCode,
				).Scan(&promo).Error
				if pErr == nil && promo.ID != uuid.Nil {
					fmt.Printf("[PROMO] Found: %s - %s, discount: %v\n", promo.Code, promo.Name, promo.DiscountValue)
					switch promo.PromotionType {
					case "PERCENTAGE":
						discountPerUnit = retailPrice * (promo.DiscountValue / 100.0)
					case "FIXED_AMOUNT":
						discountPerUnit = promo.DiscountValue
					case "BUY_X_GET_Y":
						if item.Quantity >= promo.BuyQuantity {
							freeItems := (item.Quantity / promo.BuyQuantity) * promo.GetQuantity
							if freeItems > 0 {
								discountPerUnit = retailPrice
							}
						}
					case "FLASH_SALE":
						discountPerUnit = retailPrice * (promo.DiscountValue / 100.0)
					default:
						discountPerUnit = retailPrice * (promo.DiscountValue / 100.0)
					}
					fmt.Printf("[PROMO] Discount: %v\n", discountPerUnit)
					pid := promo.ID
					promotionID = &pid
					promoNotes = fmt.Sprintf("%s - %s", promo.Code, promo.Name)
				}
			}

			if discountPerUnit < 0 {
				discountPerUnit = 0
			}

			finalUnitPrice := retailPrice
			note := ""

			if discountPerUnit > 0 {
				finalUnitPrice = retailPrice - discountPerUnit
				note = promoNotes
				priceTierID = nil
			} else if hasTier {
				finalUnitPrice = tierPrice
				note = tierNotes
			}

			if finalUnitPrice < 0 {
				finalUnitPrice = 0
			}

			lineNet := finalUnitPrice * float64(item.Quantity)
			lineTax := lineNet * (product.TaxRate / 100.0)
			lineTotal := lineNet + lineTax

			subtotal += lineNet
			discountTotal += (retailPrice - finalUnitPrice) * float64(item.Quantity)
			taxTotal += lineTax

			processed = append(processed, ProcessedSaleItem{
				ProductID:      pid,
				Quantity:       item.Quantity,
				UnitPrice:      finalUnitPrice,
				OriginalPrice:  retailPrice,
				CostPrice:      product.CostPrice,
				DiscountAmount: retailPrice - finalUnitPrice,
				TaxRate:        product.TaxRate,
				LineTotal:      lineTotal,
				PromotionID:    promotionID,
				PriceTierID:    priceTierID,
				Notes:          note,
			})
		}

		totalAmount := subtotal + taxTotal
		paidAmount := 0.0
		for _, p := range input.Payments {
			paidAmount += p.Amount
		}

		changeAmount := paidAmount - totalAmount
		saleNumber := fmt.Sprintf("SLS-%s-%03d", time.Now().Format("20060102"), rand.Intn(1000))
		loyaltyEarned := int(math.Floor(totalAmount / 10000.0))

		saleStatus := models.SaleStatus(input.Status)
		if saleStatus == "" {
			saleStatus = models.SaleStatusDone
		}

		var cashDrawerID *uuid.UUID
		if input.CashDrawerID != "" {
			cid, err := uuid.Parse(input.CashDrawerID)
			if err == nil {
				cashDrawerID = &cid
			}
		}

		createdSale = models.Sale{
			ID:                    uuid.New(),
			SaleNumber:            saleNumber,
			WarehouseID:           warehouseID,
			CustomerID:            customerID,
			CashierID:             cashierUUID,
			CompanyID:             companyID,
			CashDrawerID:          cashDrawerID,
			SaleDate:              time.Now(),
			Status:                saleStatus,
			Subtotal:              subtotal,
			DiscountAmount:        discountTotal,
			TaxAmount:             taxTotal,
			TotalAmount:           totalAmount,
			PaidAmount:            paidAmount,
			ChangeAmount:          changeAmount,
			LoyaltyPointsEarned:   loyaltyEarned,
			LoyaltyPointsRedeemed: input.LoyaltyPointsRedeem,
			Notes:                 input.Notes,
		}

		if err := tx.Create(&createdSale).Error; err != nil {
			return err
		}

		for _, item := range processed {
			si := models.SaleItem{
				ID:             uuid.New(),
				SaleID:         createdSale.ID,
				ProductID:      item.ProductID,
				Quantity:       item.Quantity,
				UnitPrice:      item.UnitPrice,
				OriginalPrice:  item.OriginalPrice,
				CostPrice:      item.CostPrice,
				DiscountAmount: item.DiscountAmount,
				TaxRate:        item.TaxRate,
				PriceTierID:    item.PriceTierID,
				PromotionID:    item.PromotionID,
				Notes:          item.Notes,
			}
			if err := tx.Create(&si).Error; err != nil {
				return err
			}
		}

		for _, p := range input.Payments {
			sp := models.SalePayment{
				ID:              uuid.New(),
				SaleID:          createdSale.ID,
				PaymentMethod:   models.PaymentMethod(p.Method),
				Amount:          p.Amount,
				ReferenceNumber: p.ReferenceNumber,
				CardLast4:       p.CardLast4,
				CreatedAt:       time.Now(),
			}
			if err := tx.Create(&sp).Error; err != nil {
				return err
			}

			// Jika pembayaran CASH, catat ke cash drawer
			if p.Method == "CASH" {
				// Cek atau buat cash drawer aktif
				var drawer models.CashDrawer
				wid := warehouseID
				cid := companyID
				usrID := cashierUUID

				result := tx.Raw(`
					SELECT * FROM cash_drawers 
					WHERE warehouse_id = ? AND company_id = ? AND status = 'OPEN' 
					ORDER BY opened_at DESC LIMIT 1
				`, wid, cid).Scan(&drawer)

				if result.Error != nil || drawer.ID == uuid.Nil {
					// Buat cash drawer baru jika tidak ada yang aktif
					drawer = models.CashDrawer{
						ID:              uuid.New(),
						WarehouseID:     wid,
						CashierID:       usrID,
						CompanyID:       cid,
						Status:          models.DrawerStatusOpen,
						OpeningBalance:  p.Amount,
						ExpectedBalance: p.Amount,
						OpenedAt:        time.Now(),
						CreatedAt:       time.Now(),
					}
					if err := tx.Create(&drawer).Error; err != nil {
						return err
					}
				}

				// Update expected balance
				newBalance := drawer.ExpectedBalance + p.Amount
				if err := tx.Model(&models.CashDrawer{}).Where("id = ?", drawer.ID).
					Updates(map[string]interface{}{"expected_balance": newBalance, "updated_at": time.Now()}).Error; err != nil {
					return err
				}

				// Catat transaksi cash-in
				saleID := createdSale.ID
				txn := models.CashDrawerTransaction{
					ID:           uuid.New(),
					CashDrawerID: drawer.ID,
					Type:         models.TransactionTypeCashIn,
					Amount:       p.Amount,
					BalanceAfter: newBalance,
					SaleID:       &saleID,
					Reason:       fmt.Sprintf("Penjualan #%s", createdSale.SaleNumber),
					CreatedAt:    time.Now(),
					CreatedBy:    usrID,
				}
				if err := tx.Create(&txn).Error; err != nil {
					return err
				}
			}
		}

		// Decrease inventory and create stock movement records
		for _, item := range processed {
			if err := tx.Model(&models.Inventory{}).
				Where("product_id = ? AND warehouse_id = ?", item.ProductID, warehouseID).
				Updates(map[string]interface{}{
					"quantity":           gorm.Expr("quantity - ?", item.Quantity),
					"available_quantity": gorm.Expr("available_quantity - ?", item.Quantity),
				}).Error; err != nil {
				return err
			}

			movement := models.StockMovement{
				ID:            uuid.New(),
				ProductID:     item.ProductID,
				WarehouseID:   warehouseID,
				MovementType:  models.MovementTypeSale,
				Quantity:      item.Quantity,
				ReferenceType: "SALE",
				ReferenceID:   &createdSale.ID,
				Notes:         fmt.Sprintf("Penjualan produk %d pcs", item.Quantity),
				CreatedAt:     time.Now(),
			}
			if err := tx.Create(&movement).Error; err != nil {
				return err
			}
		}

		if customerID != nil {
			if err := tx.Model(&models.Customer{}).
				Where("id = ?", *customerID).
				Update("loyalty_points", gorm.Expr("loyalty_points + ? - ?", createdSale.LoyaltyPointsEarned, input.LoyaltyPointsRedeem)).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return response.NewErrorResponse(err.Error())
	}

	// Send Telegram notification for Penjualan
	if s.telegramRepo != nil {
		go s.sendTelegramPenjualanNotification(companyID, &createdSale, processed)
	}

	return response.NewSuccessResponse(map[string]interface{}{
		"sale_id":                 createdSale.ID,
		"sale_number":             createdSale.SaleNumber,
		"sale_date":               createdSale.SaleDate,
		"subtotal":                createdSale.Subtotal,
		"discount_amount":         createdSale.DiscountAmount,
		"tax_amount":              createdSale.TaxAmount,
		"total_amount":            createdSale.TotalAmount,
		"paid_amount":             createdSale.PaidAmount,
		"change_amount":           createdSale.ChangeAmount,
		"loyalty_points_earned":   createdSale.LoyaltyPointsEarned,
		"loyalty_points_redeemed": createdSale.LoyaltyPointsRedeemed,
		"items":                   processed,
	}, "Sale created successfully")
}

func (s *SalesService) GetSales(filters map[string]string, limit, offset int) response.PaginatedResponse {
	rows, total, err := s.salesRepo.FindSales(filters, limit, offset)
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

func (s *SalesService) GetSalesSummary(filters map[string]string) response.ApiResponse {
	summary, err := s.salesRepo.GetSalesSummary(filters)
	if err != nil {
		return response.NewErrorResponse("Failed to get sales summary")
	}

	return response.NewSuccessResponse(map[string]interface{}{
		"total_rows":      summary.TotalRows,
		"total_penjualan": summary.TotalPenjualan,
		"total_profit":    summary.TotalProfit,
		"done_rows":       summary.DoneRows,
		"cancelled_rows":  summary.CancelledRows,
		"refunded_rows":   summary.RefundedRows,
		"pending_rows":    summary.PendingRows,
	}, "")
}

func (s *SalesService) GetSaleByID(id string) response.ApiResponse {
	saleID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Invalid sale ID")
	}

	sale, err := s.salesRepo.GetSaleByID(saleID)
	if err != nil {
		return response.NewErrorResponse("Failed to get sale")
	}
	if sale == nil {
		return response.NewErrorResponse("Sale not found")
	}

	items, _ := s.salesRepo.GetSaleItems(saleID)
	payments, _ := s.salesRepo.GetSalePayments(saleID)

	// hitung total profit dari items
	totalProfit := 0.0
	for _, item := range items {
		totalProfit += item.Profit
	}

	data := map[string]interface{}{}
	data["id"] = sale.ID
	data["sale_number"] = sale.SaleNumber
	data["warehouse_id"] = sale.WarehouseID
	data["customer_id"] = sale.CustomerID
	data["cashier_id"] = sale.CashierID
	data["company_id"] = sale.CompanyID
	data["cash_drawer_id"] = sale.CashDrawerID
	data["sale_date"] = sale.SaleDate
	data["status"] = sale.Status
	data["subtotal"] = sale.Subtotal
	data["discount_amount"] = sale.DiscountAmount
	data["tax_amount"] = sale.TaxAmount
	data["total_amount"] = sale.TotalAmount
	data["paid_amount"] = sale.PaidAmount
	data["change_amount"] = sale.ChangeAmount
	data["loyalty_points_earned"] = sale.LoyaltyPointsEarned
	data["loyalty_points_redeemed"] = sale.LoyaltyPointsRedeemed
	data["notes"] = sale.Notes
	data["created_at"] = sale.CreatedAt
	data["updated_at"] = sale.UpdatedAt
	data["warehouse_name"] = sale.WarehouseName
	data["cashier_name"] = sale.CashierName
	data["customer_name"] = sale.CustomerName
	data["customer_loyalty_points"] = sale.CustomerLoyaltyPoints
	data["items"] = items
	data["payments"] = payments
	data["total_profit"] = totalProfit

	return response.NewSuccessResponse(data, "")
}

func (s *SalesService) sendTelegramPenjualanNotification(companyID uuid.UUID, sale *models.Sale, items []ProcessedSaleItem) {
	config, err := s.telegramRepo.GetConfigByCompany(companyID)
	if err != nil || config == nil || !config.IsActive || !config.NotifyPenjualan || config.TelegramIDPenjualan == "" {
		return
	}

	telegramSvc := NewTelegramService(s.db, s.telegramRepo)

	saleModel, _ := s.salesRepo.GetSaleByID(sale.ID)
	saleItems, _ := s.salesRepo.GetSaleItems(sale.ID)

	message := telegramSvc.FormatPenjualanMessageRow(saleModel, saleItems)
	_ = telegramSvc.SendNotification(config.TelegramIDPenjualan, message)
}
