package services

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/repository"
	"github.com/pos-retail/go_backend/internal/types/response"
	"gorm.io/gorm"
)

var (
	ErrInventoryNotFound     = errors.New("inventory not found")
	ErrInsufficientStock     = errors.New("insufficient stock")
	ErrProductNotInWarehouse = errors.New("product not available in warehouse")
)

type UpdateStockOpnameRequest struct {
	WarehouseID string
	OpnameDate  string
	Status      string
	Notes       string
	Items       []struct {
		ID             string
		ProductID      string
		SystemQuantity int
		ActualQuantity int
		Notes          string
		Status         string
	}
}

type InventoryService struct {
	db            *gorm.DB
	inventoryRepo *repository.InventoryRepository
	productRepo   *repository.ProductRepository
	warehouseRepo *repository.WarehouseRepository
	purchaseRepo  *repository.PurchaseRepository
	telegramRepo  *repository.TelegramRepository
}

func NewInventoryService(
	db *gorm.DB,
	inventoryRepo *repository.InventoryRepository,
	productRepo *repository.ProductRepository,
	warehouseRepo *repository.WarehouseRepository,
	purchaseRepo *repository.PurchaseRepository,
) *InventoryService {
	return &InventoryService{
		db:            db,
		inventoryRepo: inventoryRepo,
		productRepo:   productRepo,
		warehouseRepo: warehouseRepo,
		purchaseRepo:  purchaseRepo,
	}
}

func NewInventoryServiceWithTelegram(
	db *gorm.DB,
	inventoryRepo *repository.InventoryRepository,
	productRepo *repository.ProductRepository,
	warehouseRepo *repository.WarehouseRepository,
	purchaseRepo *repository.PurchaseRepository,
	telegramRepo *repository.TelegramRepository,
) *InventoryService {
	return &InventoryService{
		db:            db,
		inventoryRepo: inventoryRepo,
		productRepo:   productRepo,
		warehouseRepo: warehouseRepo,
		purchaseRepo:  purchaseRepo,
		telegramRepo:  telegramRepo,
	}
}

type InventoryResponse struct {
	ID                uuid.UUID `json:"id"`
	ProductID         uuid.UUID `json:"product_id"`
	ProductName       string    `json:"product_name,omitempty"`
	ProductCode       string    `json:"product_code,omitempty"`
	CategoryName      string    `json:"category_name,omitempty"`
	UnitName          string    `json:"unit_name,omitempty"`
	WarehouseID       uuid.UUID `json:"warehouse_id"`
	WarehouseName     string    `json:"warehouse_name,omitempty"`
	Quantity          int       `json:"quantity"`
	ReservedQuantity  int       `json:"reserved_quantity"`
	AvailableQuantity int       `json:"available_quantity"`
	MinStockLevel     int       `json:"min_stock_level"`
	MaxStockLevel     int       `json:"max_stock_level"`
	Status            string    `json:"status"`
}

func (s *InventoryService) GetInventory(filters map[string]interface{}, limit, offset int) response.PaginatedResponse {
	inventories, total, err := s.inventoryRepo.FindAll(filters, limit, offset)
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

	data := make([]InventoryResponse, len(inventories))
	for i, inv := range inventories {
		warehouseID := uuid.Nil
		if inv.WarehouseID != nil {
			warehouseID = *inv.WarehouseID
		}
		warehouseName := ""
		if inv.WarehouseName != nil {
			warehouseName = *inv.WarehouseName
		}
		inventoryID := uuid.Nil
		if inv.InventoryID != nil {
			inventoryID = *inv.InventoryID
		}

		// Query product to get category and unit info
		product, err := s.productRepo.FindByID(inv.ProductID)
		productCode := ""
		categoryName := ""
		unitName := ""

		if err == nil && product != nil {
			productCode = product.SKU
			if product.Category != nil && product.Category.Name != "" {
				categoryName = product.Category.Name
			}
			if product.Unit != nil {
				if product.Unit.Code != "" {
					unitName = product.Unit.Code
				} else if product.Unit.Name != "" {
					unitName = product.Unit.Name
				}
			}
		}

		status := "normal"
		if inv.AvailableQuantity <= inv.MinStockLevel {
			status = "low_stock"
		}
		if inv.Quantity == 0 {
			status = "out_of_stock"
		}

		data[i] = InventoryResponse{
			ID:                inventoryID,
			ProductID:         inv.ProductID,
			ProductName:       inv.ProductName,
			ProductCode:       productCode,
			CategoryName:      categoryName,
			UnitName:          unitName,
			WarehouseID:       warehouseID,
			WarehouseName:     warehouseName,
			Quantity:          inv.Quantity,
			ReservedQuantity:  inv.ReservedQuantity,
			AvailableQuantity: inv.AvailableQuantity,
			MinStockLevel:     inv.MinStockLevel,
			MaxStockLevel:     inv.MaxStockLevel,
			Status:            status,
		}
	}

	return response.NewPaginatedResponse(data, total, limit, offset)
}

func (s *InventoryService) AdjustInventory(req struct {
	ProductID      string
	WarehouseID    string
	AdjustmentType string
	Quantity       int
	Reason         string
	Notes          string
}, userID string) response.ApiResponse {
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return response.NewErrorResponse("Invalid product ID")
	}

	warehouseID, err := uuid.Parse(req.WarehouseID)
	if err != nil {
		return response.NewErrorResponse("Invalid warehouse ID")
	}

	product, err := s.productRepo.FindByID(productID)
	if err != nil || product == nil {
		return response.NewErrorResponse("Product not found")
	}

	warehouse, err := s.warehouseRepo.FindByID(warehouseID)
	if err != nil || warehouse == nil {
		return response.NewErrorResponse("Warehouse not found")
	}

	inventory, err := s.inventoryRepo.FindByProductAndWarehouse(productID, warehouseID)
	if err != nil {
		return response.NewErrorResponse("Failed to get inventory")
	}

	previousQuantity := 0
	if inventory != nil {
		previousQuantity = inventory.Quantity
	}

	var newQuantity int
	var movementType models.MovementType

	if req.AdjustmentType == "ADJUSTMENT_IN" {
		newQuantity = previousQuantity + req.Quantity
		movementType = models.MovementTypeAdjustmentIn
	} else if req.AdjustmentType == "ADJUSTMENT_OUT" {
		if previousQuantity < req.Quantity {
			return response.NewErrorResponse(ErrInsufficientStock.Error())
		}
		newQuantity = previousQuantity - req.Quantity
		movementType = models.MovementTypeAdjustmentOut
	} else {
		return response.NewErrorResponse("Invalid adjustment type")
	}

	if inventory == nil {
		inventory = &models.Inventory{
			ID:                uuid.New(),
			ProductID:         productID,
			WarehouseID:       warehouseID,
			Quantity:          newQuantity,
			AvailableQuantity: newQuantity,
		}
		s.inventoryRepo.Create(inventory)
	} else {
		s.inventoryRepo.UpdateQuantity(productID, warehouseID, newQuantity)
	}

	movement := models.StockMovement{
		ID:            uuid.New(),
		ProductID:     productID,
		WarehouseID:   warehouseID,
		MovementType:  movementType,
		Quantity:      req.Quantity,
		ReferenceType: "ADJUSTMENT",
		Notes:         req.Reason,
	}
	s.inventoryRepo.CreateStockMovement(&movement)

	return response.NewSuccessResponse(map[string]interface{}{
		"product_id":        productID,
		"warehouse_id":      warehouseID,
		"previous_quantity": previousQuantity,
		"new_quantity":      newQuantity,
		"adjustment":        req.AdjustmentType,
		"adjusted_quantity": req.Quantity,
	}, "Inventory adjusted successfully")
}

func (s *InventoryService) GetStockCard(productID, warehouseID, fromDate, toDate string, limit, offset int) response.PaginatedResponse {
	pid, err := uuid.Parse(productID)
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

	wid, err := uuid.Parse(warehouseID)
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

	product, err := s.productRepo.FindByID(pid)
	if err != nil || product == nil {
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

	warehouse, err := s.warehouseRepo.FindByID(wid)
	if err != nil || warehouse == nil {
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

	// Parse dates (expected YYYY-MM-DD). Keep behavior safe if omitted.
	var from, toExclusive *time.Time
	if fromDate != "" {
		t, err := time.Parse("2006-01-02", fromDate)
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
		from = &t
	}
	if toDate != "" {
		t, err := time.Parse("2006-01-02", toDate)
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
		// Match TS: created_at < to_date + 1 day
		t2 := t.AddDate(0, 0, 1)
		toExclusive = &t2
	}

	openingBalance := 0
	if from != nil {
		bal, err := s.inventoryRepo.GetOpeningBalance(pid, wid, *from)
		if err == nil {
			openingBalance = bal
		}
	}

	movements, total, _ := s.inventoryRepo.GetStockCard(pid, wid, from, toExclusive, limit, offset)

	// Build transactions with running balance (match TS test expectations)
	transactions := make([]map[string]interface{}, 0, len(movements))
	totalIn := 0
	totalOut := 0
	runningBalance := openingBalance

	getTransactionName := func(mt models.MovementType, qty int) string {
		names := map[models.MovementType]string{
			models.MovementTypeSale:          "Penjualan",
			models.MovementTypePurchase:      "Pembelian",
			models.MovementTypeTransferOut:   "Transfer Keluar",
			models.MovementTypeTransferIn:    "Transfer Masuk",
			models.MovementTypeAdjustmentIn:  "Penyesuaian Stok",
			models.MovementTypeAdjustmentOut: "Penyesuaian Stok",
			models.MovementTypeReturn:        "Retur Penjualan",
			models.MovementTypeDamage:        "Kerusakan",
			models.MovementTypeExchangeIn:    "Pertukaran Masuk",
			models.MovementTypeExchangeOut:   "Pertukaran Keluar",
		}
		if mt == models.MovementTypeOpname {
			if qty > 0 {
				return "Stock Opname Masuk"
			}
			if qty < 0 {
				return "Stock Opname Keluar"
			}
			return "Stock Opname"
		}
		if v, ok := names[mt]; ok {
			return v
		}
		return string(mt)
	}

	getDocumentNumber := func(id uuid.UUID, mt models.MovementType, qty int) string {
		prefixMap := map[models.MovementType]string{
			models.MovementTypeSale:          "SO",
			models.MovementTypePurchase:      "PO",
			models.MovementTypeTransferOut:   "TRF-OUT",
			models.MovementTypeTransferIn:    "TRF-IN",
			models.MovementTypeAdjustmentIn:  "ADJ",
			models.MovementTypeAdjustmentOut: "ADJ",
			models.MovementTypeReturn:        "RET",
			models.MovementTypeDamage:        "DMG",
			models.MovementTypeExchangeIn:    "EXC-IN",
			models.MovementTypeExchangeOut:   "EXC-OUT",
			models.MovementTypeOpname:        "OPN",
		}
		prefix := prefixMap[mt]
		if prefix == "" {
			prefix = "MOV"
		}
		if mt == models.MovementTypeOpname {
			if qty > 0 {
				prefix = "OPN-IN"
			} else if qty < 0 {
				prefix = "OPN-OUT"
			}
		}
		s := strings.ToUpper(id.String())
		short := "0000"
		if len(s) >= 8 {
			short = s[:8]
		}
		return prefix + "-" + short
	}

	isInType := func(mt models.MovementType) bool {
		inTypes := map[models.MovementType]bool{
			models.MovementTypeAdjustmentIn: true,
			models.MovementTypePurchase:     true,
			models.MovementTypeTransferIn:   true,
			models.MovementTypeExchangeIn:   true,
		}
		return inTypes[mt]
	}

	for _, m := range movements {
		qty := m.Quantity
		mt := m.MovementType
		transactionName := getTransactionName(mt, qty)

		qtyIn := 0
		qtyOut := 0
		typeStr := "OUT"

		if mt == models.MovementTypeOpname {
			if qty >= 0 {
				qtyIn = qty
				qtyOut = 0
				typeStr = "IN"
			} else {
				qtyIn = 0
				qtyOut = -qty
				typeStr = "OUT"
			}
			// runningBalance is updated by the signed qty
			runningBalance += qty
		} else if isInType(mt) {
			qtyIn = qty
			qtyOut = 0
			typeStr = "IN"
			runningBalance += qty
		} else {
			qtyIn = 0
			qtyOut = qty
			typeStr = "OUT"
			runningBalance -= qty
		}

		totalIn += qtyIn
		totalOut += qtyOut

		desc := m.Notes
		if desc == "" {
			desc = transactionName
		}

		var ref interface{} = nil
		if mt == models.MovementTypeOpname && m.ReferenceID != nil {
			opname, _ := s.inventoryRepo.GetStockOpnameByID(*m.ReferenceID)
			if opname != nil {
				ref = opname.OpnameNumber
			}
		} else if mt == models.MovementTypePurchase && m.ReferenceID != nil {
			poNum, receiveNum, _ := s.inventoryRepo.GetPurchaseOrderByID(*m.ReferenceID)
			if receiveNum != "" {
				ref = receiveNum
			} else if poNum != "" {
				ref = poNum
			}
		} else if mt == models.MovementTypeReturn && m.ReferenceID != nil {
			pr, _ := s.purchaseRepo.GetPurchaseReturnByID(*m.ReferenceID)
			if pr != nil {
				ref = pr.ReturnNumber
			}
		}

		transactions = append(transactions, map[string]interface{}{
			"date":            m.CreatedAt.Format("2006-01-02"),
			"documentNumber":  getDocumentNumber(m.ID, mt, qty),
			"reference":       ref,
			"type":            typeStr,
			"transactionName": transactionName,
			"description":     desc,
			"qtyIn":           qtyIn,
			"qtyOut":          qtyOut,
			"balance":         runningBalance,
		})
	}

	categoryName := "Uncategorized"
	if product.Category != nil && product.Category.Name != "" {
		categoryName = product.Category.Name
	}
	unitStr := "PCS"
	if product.Unit != nil {
		if product.Unit.Code != "" {
			unitStr = product.Unit.Code
		} else if product.Unit.Name != "" {
			unitStr = product.Unit.Name
		}
	}

	return response.NewPaginatedResponse([]interface{}{
		map[string]interface{}{
			"item": map[string]interface{}{
				"id":       product.ID,
				"code":     product.SKU,
				"name":     product.Name,
				"category": categoryName,
				"unit":     unitStr,
			},
			"warehouse": map[string]interface{}{
				"id":   warehouse.ID,
				"name": warehouse.Name,
				"rack": "-",
			},
			"period": map[string]string{
				"from": fromDate,
				"to":   toDate,
			},
			"summary": map[string]int{
				"openingBalance": openingBalance,
				"totalIn":        totalIn,
				"totalOut":       totalOut,
				"closingBalance": runningBalance,
			},
			"transactions": transactions,
		},
	}, total, limit, offset)
}

func (s *InventoryService) CreateStockTransfer(req struct {
	FromWarehouseID string
	ToWarehouseID   string
	ExpectedArrival string
	Items           []struct {
		ProductID string
		Quantity  int
	}
	Notes string
}, userID string) response.ApiResponse {
	fromWarehouseID, err := uuid.Parse(req.FromWarehouseID)
	if err != nil {
		return response.NewErrorResponse("Invalid from warehouse ID")
	}

	toWarehouseID, err := uuid.Parse(req.ToWarehouseID)
	if err != nil {
		return response.NewErrorResponse("Invalid to warehouse ID")
	}

	if fromWarehouseID == toWarehouseID {
		return response.NewErrorResponse("Source and destination warehouse must be different")
	}

	transferNumber, _ := s.inventoryRepo.GetNextTransferNumber()
	uid, _ := uuid.Parse(userID)

	transfer := models.StockTransfer{
		ID:              uuid.New(),
		TransferNumber:  transferNumber,
		FromWarehouseID: fromWarehouseID,
		ToWarehouseID:   toWarehouseID,
		UserID:          uid,
		Status:          models.StockTransferStatusPending,
		Notes:           req.Notes,
	}

	if req.ExpectedArrival != "" {
		t, err := time.Parse("2006-01-02", req.ExpectedArrival)
		if err == nil {
			transfer.ExpectedArrival = &t
		}
	}

	if err := s.inventoryRepo.CreateStockTransfer(&transfer); err != nil {
		return response.NewErrorResponse("Failed to create stock transfer")
	}

	for _, item := range req.Items {
		productID, _ := uuid.Parse(item.ProductID)

		inventory, err := s.inventoryRepo.FindByProductAndWarehouse(productID, fromWarehouseID)
		if err != nil || inventory == nil {
			return response.NewErrorResponse("Product not available in source warehouse")
		}

		if inventory.AvailableQuantity < item.Quantity {
			return response.NewErrorResponse("Insufficient stock in source warehouse")
		}

		newQty := inventory.AvailableQuantity - item.Quantity
		s.inventoryRepo.UpdateQuantity(productID, fromWarehouseID, newQty)

		transferItem := models.StockTransferItem{
			ID:         uuid.New(),
			TransferID: transfer.ID,
			ProductID:  productID,
			Quantity:   item.Quantity,
		}
		s.inventoryRepo.CreateStockTransferItem(&transferItem)

		movement := models.StockMovement{
			ID:            uuid.New(),
			ProductID:     productID,
			WarehouseID:   fromWarehouseID,
			MovementType:  models.MovementTypeTransferOut,
			Quantity:      item.Quantity,
			ReferenceType: "STOCK_TRANSFER",
			ReferenceID:   &transfer.ID,
			Notes:         "Transfer to " + toWarehouseID.String(),
		}
		s.inventoryRepo.CreateStockMovement(&movement)
	}

	return response.NewSuccessResponse(transfer, "Stock transfer created successfully")
}

func (s *InventoryService) ReceiveStockTransfer(transferID string, items []struct {
	TransferItemID   string
	ReceivedQuantity int
	Notes            string
}, userID string) response.ApiResponse {
	tid, err := uuid.Parse(transferID)
	if err != nil {
		return response.NewErrorResponse("Invalid transfer ID")
	}

	transfer, err := s.inventoryRepo.GetStockTransferByID(tid)
	if err != nil || transfer == nil {
		return response.NewErrorResponse("Transfer not found")
	}

	if transfer.Status != models.StockTransferStatusPending && transfer.Status != models.StockTransferStatusInTransit {
		return response.NewErrorResponse("Transfer already received or cancelled")
	}

	for _, item := range items {
		itemID, _ := uuid.Parse(item.TransferItemID)

		for i := range transfer.Items {
			if transfer.Items[i].ID == itemID {
				recQty := item.ReceivedQuantity
				if recQty == 0 {
					recQty = transfer.Items[i].Quantity
				}
				transfer.Items[i].ReceivedQuantity = &recQty

				inventory, err := s.inventoryRepo.FindByProductAndWarehouse(transfer.Items[i].ProductID, transfer.ToWarehouseID)
				if err == nil && inventory != nil {
					newQty := inventory.Quantity + recQty
					s.inventoryRepo.UpdateQuantity(transfer.Items[i].ProductID, transfer.ToWarehouseID, newQty)
				} else {
					newInv := models.Inventory{
						ID:                uuid.New(),
						ProductID:         transfer.Items[i].ProductID,
						WarehouseID:       transfer.ToWarehouseID,
						Quantity:          recQty,
						AvailableQuantity: recQty,
					}
					s.inventoryRepo.Create(&newInv)
				}

				movement := models.StockMovement{
					ID:            uuid.New(),
					ProductID:     transfer.Items[i].ProductID,
					WarehouseID:   transfer.ToWarehouseID,
					MovementType:  models.MovementTypeTransferIn,
					Quantity:      recQty,
					ReferenceType: "STOCK_TRANSFER",
					ReferenceID:   &transfer.ID,
					Notes:         "Received from transfer",
				}
				s.inventoryRepo.CreateStockMovement(&movement)
			}
		}
	}

	now := time.Now()
	transfer.ActualArrival = &now
	transfer.Status = models.StockTransferStatusReceived
	s.inventoryRepo.UpdateStockTransfer(transfer)

	return response.NewSuccessResponse(transfer, "Stock transfer received successfully")
}

func (s *InventoryService) CreateStockOpname(req struct {
	WarehouseID string
	CompanyID   string
	OpnameDate  string
	Items       []struct {
		ProductID      string
		SystemQuantity int
		ActualQuantity int
		Notes          string
	}
	Notes string
}, userID string) response.ApiResponse {
	warehouseID, err := uuid.Parse(req.WarehouseID)
	if err != nil {
		return response.NewErrorResponse("Invalid warehouse ID")
	}

	warehouse, err := s.warehouseRepo.FindByID(warehouseID)
	if err != nil || warehouse == nil {
		return response.NewErrorResponse("Warehouse not found")
	}

	opnameNumber, _ := s.inventoryRepo.GetNextOpnameNumber(req.CompanyID)
	uid, _ := uuid.Parse(userID)
	companyUID, _ := uuid.Parse(req.CompanyID)

	opname := models.StockOpname{
		ID:           uuid.New(),
		OpnameNumber: opnameNumber,
		WarehouseID:  warehouseID,
		UserID:       uid,
		CompanyID:    companyUID,
		Status:       models.StockOpnameStatusDraft,
		Notes:        req.Notes,
	}

	if req.OpnameDate != "" {
		t, err := time.Parse("2006-01-02", req.OpnameDate)
		if err == nil {
			opname.OpnameDate = t
		}
	}

	if err := s.inventoryRepo.CreateStockOpname(&opname); err != nil {
		return response.NewErrorResponse("Failed to create stock opname")
	}

	totalSelisih := 0.0

	for _, item := range req.Items {
		productID, _ := uuid.Parse(item.ProductID)

		inventory, _ := s.inventoryRepo.FindByProductAndWarehouse(productID, warehouseID)
		systemQty := 0
		if inventory != nil {
			systemQty = inventory.Quantity
		}

		difference := item.ActualQuantity - systemQty

		// ambil cost_price dari product
		var productCostPrice float64
		var product models.Product
		if err := s.db.First(&product, "id = ?", productID).Error; err == nil {
			productCostPrice = product.CostPrice
		}

		opnameItem := models.StockOpnameItem{
			ID:             uuid.New(),
			OpnameID:       opname.ID,
			ProductID:      productID,
			SystemQuantity: systemQty,
			ActualQuantity: item.ActualQuantity,
			CostPrice:      productCostPrice,
			Difference:     difference,
			Notes:          item.Notes,
		}
		s.inventoryRepo.CreateStockOpnameItem(&opnameItem)

		// akumulasi total_selisih
		nilaiSelisih := float64(difference) * productCostPrice
		totalSelisih += nilaiSelisih
	}

	// simpan total_selisih ke header
	opname.TotalSelisih = totalSelisih
	s.inventoryRepo.UpdateStockOpname(&opname)

	return s.GetStockOpnameByID(opname.ID.String())
}

func (s *InventoryService) GetStockOpnames(companyID *string, filters map[string]interface{}, limit, offset int) response.PaginatedResponse {
	if companyID != nil {
		filters["company_id"] = *companyID
	}
	opnames, total, err := s.inventoryRepo.FindStockOpnames(filters, limit, offset)
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

	return response.NewPaginatedResponse(opnames, total, limit, offset)
}

func (s *InventoryService) GetStockOpnameByID(id string) response.ApiResponse {
	opnameID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Invalid opname ID")
	}

	opname, err := s.inventoryRepo.GetStockOpnameByID(opnameID)
	if err != nil || opname == nil {
		return response.NewErrorResponse("Stock opname not found")
	}

	transformedItems := make([]map[string]interface{}, len(opname.Items))
	for i, item := range opname.Items {
		itemMap := map[string]interface{}{
			"id":              item.ID,
			"opname_id":       item.OpnameID,
			"product_id":      item.ProductID,
			"system_quantity": item.SystemQuantity,
			"actual_quantity": item.ActualQuantity,
			"cost_price":      item.CostPrice,
			"difference":      item.Difference,
			"status":          item.Status,
			"notes":           item.Notes,
		}
		if item.Product != nil {
			itemMap["product_name"] = item.Product.Name
			itemMap["product_sku"] = item.Product.SKU
			if item.Product.Unit != nil {
				itemMap["product_unit_name"] = item.Product.Unit.Name
			}
		}
		transformedItems[i] = itemMap
	}

	opnameMap := map[string]interface{}{
		"id":            opname.ID,
		"opname_number": opname.OpnameNumber,
		"warehouse_id":  opname.WarehouseID,
		"user_id":       opname.UserID,
		"opname_date":   opname.OpnameDate,
		"status":        opname.Status,
		"notes":         opname.Notes,
		"total_selisih": opname.TotalSelisih,
		"created_at":    opname.CreatedAt,
		"updated_at":    opname.UpdatedAt,
		"warehouse":     opname.Warehouse,
		"user":          opname.User,
		"items":         transformedItems,
	}

	return response.NewSuccessResponse(opnameMap, "Stock opname retrieved successfully")
}

func (s *InventoryService) UpdateStockOpnameStatus(id, status, userID string) response.ApiResponse {
	opnameID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Invalid opname ID")
	}

	opname, err := s.inventoryRepo.GetStockOpnameByID(opnameID)
	if err != nil || opname == nil {
		return response.NewErrorResponse("Stock opname not found")
	}

	currentStatus := opname.Status

	if status == "COMPLETED" && currentStatus == models.StockOpnameStatusDraft {
		for _, item := range opname.Items {
			inventory, _ := s.inventoryRepo.FindByProductAndWarehouse(item.ProductID, opname.WarehouseID)

			currentQty := 0
			if inventory != nil {
				currentQty = inventory.Quantity
			}

			newQty := currentQty + item.Difference
			s.inventoryRepo.UpdateQuantity(item.ProductID, opname.WarehouseID, newQty)

			movement := models.StockMovement{
				ID:            uuid.New(),
				ProductID:     item.ProductID,
				WarehouseID:   opname.WarehouseID,
				MovementType:  models.MovementTypeOpname,
				Quantity:      item.Difference,
				ReferenceType: "STOCK_OPNAME",
				ReferenceID:   &opname.ID,
				Notes:         "Stock opname adjustment",
			}
			s.inventoryRepo.CreateStockMovement(&movement)
		}
		opname.Status = models.StockOpnameStatusCompleted
	} else if status == "APPROVED" && currentStatus == models.StockOpnameStatusCompleted {
		opname.Status = models.StockOpnameStatusApproved
	} else {
		return response.NewErrorResponse("Invalid status transition")
	}

	s.inventoryRepo.UpdateStockOpname(opname)

	// Send Telegram notification for Stock Opname
	if s.telegramRepo != nil && status == "COMPLETED" {
		go s.sendTelegramStockOpnameNotification(opname.WarehouseID, opnameID)
	}

	return response.NewSuccessResponse(map[string]interface{}{
		"id":     opname.ID,
		"status": opname.Status,
	}, "Stock opname status updated successfully")
}

func (s *InventoryService) DeleteStockOpname(id string) response.ApiResponse {
	opnameID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Invalid stock opname ID")
	}

	opname, err := s.inventoryRepo.GetStockOpnameByID(opnameID)
	if err != nil || opname == nil {
		return response.NewErrorResponse("Stock opname not found")
	}

	if err := s.inventoryRepo.DeleteStockOpname(opnameID); err != nil {
		return response.NewErrorResponse("Failed to delete stock opname")
	}

	return response.NewSuccessResponse(nil, "Stock opname deleted successfully")
}

func (s *InventoryService) UpdateStockOpname(id string, req UpdateStockOpnameRequest) response.ApiResponse {
	opnameID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Invalid stock opname ID")
	}

	opname, err := s.inventoryRepo.GetStockOpnameByID(opnameID)
	if err != nil || opname == nil {
		return response.NewErrorResponse("Stock opname not found")
	}

	currentStatus := opname.Status

	if currentStatus != models.StockOpnameStatusDraft && currentStatus != models.StockOpnameStatusInProgress {
		return response.NewErrorResponse("Cannot update stock opname that is already completed or approved")
	}

	isApprovalTransition := req.Status == "approved" && currentStatus == models.StockOpnameStatusDraft

	if req.WarehouseID != "" {
		wid, err := uuid.Parse(req.WarehouseID)
		if err != nil {
			return response.NewErrorResponse("Invalid warehouse ID")
		}
		opname.WarehouseID = wid
	}

	if req.OpnameDate != "" {
		t, err := time.Parse("2006-01-02", req.OpnameDate)
		if err != nil {
			return response.NewErrorResponse("Invalid opname_date format. Expected YYYY-MM-DD")
		}
		opname.OpnameDate = t
	}

	if req.Notes != "" {
		opname.Notes = req.Notes
	}

	if isApprovalTransition {
		opname.Status = models.StockOpnameStatusApproved
	} else if req.Status != "" {
		opname.Status = models.StockOpnameStatus(req.Status)
	}

	items := make([]struct {
		ID             string
		ProductID      uuid.UUID
		SystemQuantity int
		ActualQuantity int
		CostPrice      float64
		Difference     int
		Status         string
		Notes          string
	}, 0, len(req.Items))

	totalSelisih := 0.0

	for _, item := range req.Items {
		pid, err := uuid.Parse(item.ProductID)
		if err != nil {
			continue
		}
		difference := item.ActualQuantity - item.SystemQuantity

		// ambil cost_price dari product
		var costPrice float64
		var product models.Product
		if err := s.db.First(&product, "id = ?", pid).Error; err == nil {
			costPrice = product.CostPrice
		}

		nilaiSelisih := float64(difference) * costPrice
		totalSelisih += nilaiSelisih

		status := item.Status
		if status == "" {
			if isApprovalTransition {
				status = "done"
			} else {
				status = "pending"
			}
		}
		items = append(items, struct {
			ID             string
			ProductID      uuid.UUID
			SystemQuantity int
			ActualQuantity int
			CostPrice      float64
			Difference     int
			Status         string
			Notes          string
		}{
			ID:             item.ID,
			ProductID:      pid,
			SystemQuantity: item.SystemQuantity,
			ActualQuantity: item.ActualQuantity,
			CostPrice:      costPrice,
			Difference:     difference,
			Status:         status,
			Notes:          item.Notes,
		})
	}

	// simpan total_selisih ke header
	opname.TotalSelisih = totalSelisih
	if err := s.inventoryRepo.UpdateStockOpname(&opname); err != nil {
		return response.NewErrorResponse("Failed to update stock opname total")
	}

	if err := s.inventoryRepo.UpdateStockOpnameWithItems(opname, items); err != nil {
		return response.NewErrorResponse("Failed to update stock opname")
	}

	if isApprovalTransition {
		s.inventoryRepo.UpdateStockOpnameItemsStatus(opnameID, "done")

		for _, item := range items {
			inventory, _ := s.inventoryRepo.FindByProductAndWarehouse(item.ProductID, opname.WarehouseID)

			currentQty := 0
			if inventory != nil {
				currentQty = inventory.Quantity
			}

			newQty := currentQty + item.Difference
			s.inventoryRepo.UpdateQuantity(item.ProductID, opname.WarehouseID, newQty)

			notes := item.Notes
			if notes == "" {
				notes = "Stock opname adjustment - approved"
			}

			movement := models.StockMovement{
				ID:            uuid.New(),
				ProductID:     item.ProductID,
				WarehouseID:   opname.WarehouseID,
				MovementType:  models.MovementTypeOpname,
				Quantity:      item.Difference,
				ReferenceType: "STOCK_OPNAME",
				ReferenceID:   &opnameID,
				Notes:         notes,
			}
			s.inventoryRepo.CreateStockMovement(&movement)
		}
	}

	updated, _ := s.inventoryRepo.GetStockOpnameByID(opnameID)
	return response.NewSuccessResponse(updated, "Stock opname updated successfully")
}

func (s *InventoryService) sendTelegramStockOpnameNotification(warehouseID uuid.UUID, opnameID uuid.UUID) {
	warehouse, _ := s.warehouseRepo.FindByID(warehouseID)
	warehouseName := "Unknown"
	var companyID uuid.UUID
	if warehouse != nil {
		warehouseName = warehouse.Name
		if warehouse.CompanyID != nil {
			companyID = *warehouse.CompanyID
		}
	} else {
		return
	}

	opname, _ := s.inventoryRepo.GetStockOpnameByID(opnameID)
	if opname == nil {
		return
	}

	config, err := s.telegramRepo.GetConfigByCompany(companyID)
	if err != nil || config == nil || !config.IsActive || !config.NotifyStockOpname || config.TelegramIDStockOpname == "" {
		return
	}

	telegramSvc := NewTelegramService(s.db, s.telegramRepo)

	discrepancyCount := 0
	for _, item := range opname.Items {
		if item.Difference != 0 {
			discrepancyCount++
		}
	}

	message := telegramSvc.FormatStockOpnameMessage(opname, warehouseName, discrepancyCount)
	_ = telegramSvc.SendNotification(config.TelegramIDStockOpname, message)
}
