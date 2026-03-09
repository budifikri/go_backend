package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/middleware"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type PurchaseHandler struct {
	purchaseService *services.PurchaseService
}

func NewPurchaseHandler(purchaseService *services.PurchaseService) *PurchaseHandler {
	return &PurchaseHandler{purchaseService: purchaseService}
}

// GetPurchaseOrders godoc
// @Summary List purchase orders
// @Tags Purchases
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param status query string false "PO status"
// @Param supplier_id query string false "Supplier ID"
// @Param warehouse_id query string false "Warehouse ID"
// @Param search query string false "Search"
// @Param date_from query string false "Filter from date (YYYY-MM-DD)"
// @Param date_to query string false "Filter to date (YYYY-MM-DD)"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/purchases [get]
func (h *PurchaseHandler) GetPurchaseOrders(c *fiber.Ctx) error {
	filters := map[string]string{}
	// Support both status and status_po/status_receive for filtering
	// Backwards compatibility: if 'status' is provided, it filters status_po
	filters["status_po"] = c.Query("status_po")
	if filters["status_po"] == "" {
		filters["status_po"] = c.Query("status")
	}
	filters["status_receive"] = c.Query("status_receive")
	filters["supplier_id"] = c.Query("supplier_id")
	filters["warehouse_id"] = c.Query("warehouse_id")
	filters["search"] = c.Query("search")
	filters["date_from"] = c.Query("date_from")
	filters["date_to"] = c.Query("date_to")

	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)
	result := h.purchaseService.GetPurchaseOrders(filters, limit, offset)
	return c.JSON(result)
}

// GetPurchaseOrder godoc
// @Summary Get purchase order
// @Tags Purchases
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Purchase Order ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/purchases/{id} [get]
func (h *PurchaseHandler) GetPurchaseOrder(c *fiber.Ctx) error {
	result := h.purchaseService.GetPurchaseOrderByID(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// CreatePurchaseOrder godoc
// @Summary Create purchase order
// @Tags Purchases
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body request.CreatePurchaseOrderRequest true "Purchase order payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/purchases [post]
func (h *PurchaseHandler) CreatePurchaseOrder(c *fiber.Ctx) error {
	var req request.CreatePurchaseOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	expected, err := time.Parse(time.RFC3339, req.ExpectedDate)
	if err != nil {
		// TS sends date string, parsed with new Date(body.expected_date)
		expected = time.Now()
	}

	items := make([]services.CreatePurchaseOrderItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, services.CreatePurchaseOrderItemInput{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			UnitPrice: it.UnitPrice,
			Discount:  it.Discount,
			TaxRate:   it.TaxRate,
		})
	}

	result := h.purchaseService.CreatePurchaseOrder(services.CreatePurchaseOrderInput{
		SupplierID:   req.SupplierID,
		WarehouseID:  req.WarehouseID,
		ExpectedDate: expected,
		Items:        items,
		Notes:        req.Notes,
		CreatedBy:    user.UserID,
		CompanyID:    user.CompanyID,
	})
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

// UpdatePurchaseOrder godoc
// @Summary Update purchase order
// @Tags Purchases
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Purchase Order ID"
// @Param body body request.UpdatePurchaseOrderRequest true "Purchase order payload"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/purchases/{id} [put]
func (h *PurchaseHandler) UpdatePurchaseOrder(c *fiber.Ctx) error {
	var req request.UpdatePurchaseOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	log.Printf("[DEBUG] UpdatePurchaseOrder payload: supplier_id=%s warehouse_id=%s expected_date=%s items=%d", req.SupplierID, req.WarehouseID, req.ExpectedDate, len(req.Items))

	if req.SupplierID == "" || req.WarehouseID == "" || req.ExpectedDate == "" || len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse(
			"Invalid request data: supplier_id, warehouse_id, expected_date, and items are required",
		))
	}

	orderDate := time.Now()
	if req.OrderDate != "" {
		orderDate, _ = time.Parse(time.RFC3339, req.OrderDate)
		if orderDate.IsZero() {
			orderDate, _ = time.Parse("2006-01-02", req.OrderDate)
		}
		if orderDate.IsZero() {
			orderDate = time.Now()
		}
	}

	expected, err := time.Parse(time.RFC3339, req.ExpectedDate)
	if err != nil {
		if t2, err2 := time.Parse("2006-01-02", req.ExpectedDate); err2 == nil {
			expected = t2
		} else {
			expected = time.Now()
		}
	}

	items := make([]services.CreatePurchaseOrderItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, services.CreatePurchaseOrderItemInput{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			UnitPrice: it.UnitPrice,
			Discount:  it.Discount,
			TaxRate:   it.TaxRate,
		})
	}

	result := h.purchaseService.UpdatePurchaseOrder(c.Params("id"), services.UpdatePurchaseOrderInput{
		SupplierID:    req.SupplierID,
		WarehouseID:   req.WarehouseID,
		OrderDate:     orderDate,
		ExpectedDate:  expected,
		Items:         items,
		Notes:         req.Notes,
		StatusPo:      req.StatusPo,
		StatusReceive: req.StatusReceive,
	})
	if !result.Success {
		if result.Error == "Purchase order not found" {
			return c.Status(fiber.StatusNotFound).JSON(result)
		}
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

// UpdatePurchaseOrderStatus godoc
// @Summary Update purchase order status
// @Tags Purchases
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Purchase Order ID"
// @Param body body request.UpdatePurchaseOrderStatusRequest true "Status payload"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/purchases/{id}/status [put]
func (h *PurchaseHandler) UpdatePurchaseOrderStatus(c *fiber.Ctx) error {
	var req request.UpdatePurchaseOrderStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}
	result := h.purchaseService.UpdatePurchaseOrderStatus(c.Params("id"), req.Status)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// CancelPurchaseOrder godoc
// @Summary Cancel purchase order
// @Tags Purchases
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Purchase Order ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/purchases/{id}/cancel [post]
func (h *PurchaseHandler) CancelPurchaseOrder(c *fiber.Ctx) error {
	result := h.purchaseService.CancelPurchaseOrder(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// DeletePurchaseOrder godoc
// @Summary Delete purchase order permanently
// @Tags Purchases
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Purchase Order ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/purchases/{id} [delete]
func (h *PurchaseHandler) DeletePurchaseOrder(c *fiber.Ctx) error {
	result := h.purchaseService.DeletePurchaseOrder(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// ApprovePurchaseOrder godoc
// @Summary Approve purchase order
// @Tags Purchases
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Purchase Order ID"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/purchases/{id}/approve [put]
func (h *PurchaseHandler) ApprovePurchaseOrder(c *fiber.Ctx) error {
	result := h.purchaseService.ApprovePurchaseOrder(c.Params("id"))
	if !result.Success {
		if result.Error == "Purchase order not found" {
			return c.Status(fiber.StatusNotFound).JSON(result)
		}
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

// SetPendingPurchaseOrder godoc
// @Summary Set purchase order to PENDING
// @Tags Purchases
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Purchase Order ID"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/purchases/{id}/pending [put]
func (h *PurchaseHandler) SetPendingPurchaseOrder(c *fiber.Ctx) error {
	result := h.purchaseService.SetPendingPurchaseOrder(c.Params("id"))
	if !result.Success {
		if result.Error == "Purchase order not found" {
			return c.Status(fiber.StatusNotFound).JSON(result)
		}
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

// ReceivePurchaseOrder godoc
// @Summary Receive purchase order items
// @Tags Purchases
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Purchase Order ID"
// @Param body body request.ReceivePurchaseOrderRequest true "Receive payload"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/purchases/{id}/receive [put]
func (h *PurchaseHandler) ReceivePurchaseOrder(c *fiber.Ctx) error {
	var req request.ReceivePurchaseOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	if len(req.Items) == 0 || req.StatusReceive == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request data"))
	}

	items := make([]services.ReceivePurchaseOrderItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, services.ReceivePurchaseOrderItemInput{
			ID:         it.ID,
			QtyReceive: it.QtyReceive,
		})
	}

	receiveDate := time.Now()
	if req.ReceiveDate != "" {
		receiveDate, _ = time.Parse(time.RFC3339, req.ReceiveDate)
		if receiveDate.IsZero() {
			receiveDate, _ = time.Parse("2006-01-02", req.ReceiveDate)
		}
		if receiveDate.IsZero() {
			receiveDate = time.Now()
		}
	}

	result := h.purchaseService.ReceivePurchaseOrder(c.Params("id"), services.ReceivePurchaseOrderInput{
		Items:         items,
		StatusReceive: req.StatusReceive,
		ReceiveDate:   receiveDate,
	})
	if !result.Success {
		if result.Error == "Purchase order not found" {
			return c.Status(fiber.StatusNotFound).JSON(result)
		}
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}
