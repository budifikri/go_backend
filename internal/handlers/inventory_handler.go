package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/middleware"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type InventoryHandler struct {
	inventoryService *services.InventoryService
}

func NewInventoryHandler(inventoryService *services.InventoryService) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
	}
}

// GetInventory godoc
// @Summary List inventory
// @Tags Inventory
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param warehouse_id query string false "Warehouse ID"
// @Param product_id query string false "Product ID"
// @Param search query string false "Search"
// @Param stock query string false "Stock filter: all|available|minus|empty" default(available)
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/inventory [get]
func (h *InventoryHandler) GetInventory(c *fiber.Ctx) error {
	filters := make(map[string]interface{})
	user := middleware.GetUserFromContext(c)
	if user != nil && user.CompanyID != "" {
		filters["company_id"] = user.CompanyID
	}
	if warehouseID := c.Query("warehouse_id"); warehouseID != "" {
		filters["warehouse_id"] = warehouseID
	}
	if productID := c.Query("product_id"); productID != "" {
		filters["product_id"] = productID
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	stockFilter := c.Query("stock", "available")
	filters["stock"] = stockFilter
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)
	result := h.inventoryService.GetInventory(filters, limit, offset)
	return c.JSON(result)
}

// GetStockCard godoc
// @Summary Get stock card
// @Tags Inventory
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param product_id query string true "Product ID"
// @Param warehouse_id query string true "Warehouse ID"
// @Param from_date query string false "From date"
// @Param to_date query string false "To date"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/inventory/stock-card [get]
func (h *InventoryHandler) GetStockCard(c *fiber.Ctx) error {
	productID := c.Query("product_id")
	warehouseID := c.Query("warehouse_id")
	if productID == "" || warehouseID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("product_id and warehouse_id required"))
	}
	result := h.inventoryService.GetStockCard(productID, warehouseID, c.Query("from_date"), c.Query("to_date"))
	return c.JSON(result)
}

// AdjustInventory godoc
// @Summary Adjust inventory
// @Tags Inventory
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body request.InventoryAdjustmentRequest true "Adjustment payload"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/inventory/adjust [post]
func (h *InventoryHandler) AdjustInventory(c *fiber.Ctx) error {
	var req request.InventoryAdjustmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}
	user := middleware.GetUserFromContext(c)
	userID := ""
	if user != nil {
		userID = user.UserID
	}
	result := h.inventoryService.AdjustInventory(struct {
		ProductID      string
		WarehouseID    string
		AdjustmentType string
		Quantity       int
		Reason         string
		Notes          string
	}{
		ProductID:      req.ProductID,
		WarehouseID:    req.WarehouseID,
		AdjustmentType: req.AdjustmentType,
		Quantity:       req.Quantity,
		Reason:         req.Reason,
		Notes:          req.Notes,
	}, userID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

// CreateStockTransfer godoc
// @Summary Create stock transfer
// @Tags StockTransfers
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body request.StockTransferRequest true "Transfer payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/stock-transfers [post]
func (h *InventoryHandler) CreateStockTransfer(c *fiber.Ctx) error {
	var req request.StockTransferRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}
	user := middleware.GetUserFromContext(c)
	userID := ""
	if user != nil {
		userID = user.UserID
	}
	items := make([]struct {
		ProductID string
		Quantity  int
	}, len(req.Items))
	for i, item := range req.Items {
		items[i] = struct {
			ProductID string
			Quantity  int
		}{ProductID: item.ProductID, Quantity: item.Quantity}
	}
	result := h.inventoryService.CreateStockTransfer(struct {
		FromWarehouseID string
		ToWarehouseID   string
		ExpectedArrival string
		Items           []struct {
			ProductID string
			Quantity  int
		}
		Notes string
	}{
		FromWarehouseID: req.FromWarehouseID,
		ToWarehouseID:   req.ToWarehouseID,
		ExpectedArrival: req.ExpectedArrival,
		Items:           items,
		Notes:           req.Notes,
	}, userID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

// ReceiveStockTransfer godoc
// @Summary Receive stock transfer
// @Tags StockTransfers
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Transfer ID"
// @Param body body request.ReceiveTransferRequest true "Receive payload"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/stock-transfers/{id}/receive [put]
func (h *InventoryHandler) ReceiveStockTransfer(c *fiber.Ctx) error {
	id := c.Params("id")
	var req request.ReceiveTransferRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}
	user := middleware.GetUserFromContext(c)
	userID := ""
	if user != nil {
		userID = user.UserID
	}
	items := make([]struct {
		TransferItemID   string
		ReceivedQuantity int
		Notes            string
	}, len(req.Items))
	for i, item := range req.Items {
		items[i] = struct {
			TransferItemID   string
			ReceivedQuantity int
			Notes            string
		}{TransferItemID: item.TransferItemID, ReceivedQuantity: item.ReceivedQuantity, Notes: item.Notes}
	}
	result := h.inventoryService.ReceiveStockTransfer(id, items, userID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

// CreateStockOpname godoc
// @Summary Create stock opname
// @Tags StockOpname
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body request.StockOpnameRequest true "Opname payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/stock-opname [post]
func (h *InventoryHandler) CreateStockOpname(c *fiber.Ctx) error {
	var req request.StockOpnameRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}
	user := middleware.GetUserFromContext(c)
	userID := ""
	companyID := ""
	if user != nil {
		userID = user.UserID
		companyID = user.CompanyID
	}
	items := make([]struct {
		ProductID      string
		SystemQuantity int
		ActualQuantity int
		Notes          string
	}, len(req.Items))
	for i, item := range req.Items {
		items[i] = struct {
			ProductID      string
			SystemQuantity int
			ActualQuantity int
			Notes          string
		}{ProductID: item.ProductID, SystemQuantity: item.SystemQuantity, ActualQuantity: item.ActualQuantity, Notes: item.Notes}
	}
	result := h.inventoryService.CreateStockOpname(struct {
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
	}{
		WarehouseID: req.WarehouseID,
		CompanyID:   companyID,
		OpnameDate:  req.OpnameDate,
		Items:       items,
		Notes:       req.Notes,
	}, userID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

// GetStockOpnames godoc
// @Summary List stock opnames
// @Tags StockOpname
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param warehouse_id query string false "Warehouse ID"
// @Param status query string false "Status"
// @Param from_date query string false "From date"
// @Param to_date query string false "To date"
// @Param search query string false "Search"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/stock-opname [get]
func (h *InventoryHandler) GetStockOpnames(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	var companyID *string
	if user != nil && user.CompanyID != "" {
		companyID = &user.CompanyID
	}

	filters := make(map[string]interface{})
	if warehouseID := c.Query("warehouse_id"); warehouseID != "" {
		filters["warehouse_id"] = warehouseID
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if fromDate := c.Query("from_date"); fromDate != "" {
		filters["from_date"] = fromDate
	}
	if toDate := c.Query("to_date"); toDate != "" {
		filters["to_date"] = toDate
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)
	result := h.inventoryService.GetStockOpnames(companyID, filters, limit, offset)
	return c.JSON(result)
}

// GetStockOpname godoc
// @Summary Get stock opname
// @Tags StockOpname
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Opname ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/stock-opname/{id} [get]
func (h *InventoryHandler) GetStockOpname(c *fiber.Ctx) error {
	id := c.Params("id")
	result := h.inventoryService.GetStockOpnameByID(id)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// UpdateStockOpnameStatus godoc
// @Summary Update stock opname status
// @Tags StockOpname
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Opname ID"
// @Param body body request.StockOpnameStatusRequest true "Status payload"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/stock-opname/{id}/status [put]
func (h *InventoryHandler) UpdateStockOpnameStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	var req request.StockOpnameStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}
	user := middleware.GetUserFromContext(c)
	userID := ""
	if user != nil {
		userID = user.UserID
	}
	result := h.inventoryService.UpdateStockOpnameStatus(id, req.Status, userID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

// DeleteStockOpname godoc
// @Summary Delete stock opname
// @Tags StockOpname
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Opname ID"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/stock-opname/{id} [delete]
func (h *InventoryHandler) DeleteStockOpname(c *fiber.Ctx) error {
	id := c.Params("id")
	result := h.inventoryService.DeleteStockOpname(id)
	if !result.Success {
		if result.Error == "Stock opname not found" {
			return c.Status(fiber.StatusNotFound).JSON(result)
		}
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

// UpdateStockOpname godoc
// @Summary Update stock opname
// @Tags StockOpname
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Opname ID"
// @Param body body request.StockOpnameUpdateRequest true "Update payload"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/stock-opname/{id} [put]
func (h *InventoryHandler) UpdateStockOpname(c *fiber.Ctx) error {
	id := c.Params("id")
	var req request.StockOpnameUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	reqService := services.UpdateStockOpnameRequest{
		WarehouseID: req.WarehouseID,
		OpnameDate:  req.OpnameDate,
		Status:      req.Status,
		Notes:       req.Notes,
	}
	for _, item := range req.Items {
		reqService.Items = append(reqService.Items, struct {
			ID             string
			ProductID      string
			SystemQuantity int
			ActualQuantity int
			Notes          string
			Status         string
		}{
			ID:             item.ID,
			ProductID:      item.ProductID,
			SystemQuantity: item.SystemQuantity,
			ActualQuantity: item.ActualQuantity,
			Notes:          item.Notes,
			Status:         item.Status,
		})
	}

	result := h.inventoryService.UpdateStockOpname(id, reqService)
	if !result.Success {
		if result.Error == "Stock opname not found" {
			return c.Status(fiber.StatusNotFound).JSON(result)
		}
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}
