package handlers

import (
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

func (h *PurchaseHandler) GetPurchaseOrders(c *fiber.Ctx) error {
	filters := map[string]string{}
	filters["status"] = c.Query("status")
	filters["supplier_id"] = c.Query("supplier_id")
	filters["warehouse_id"] = c.Query("warehouse_id")

	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)
	result := h.purchaseService.GetPurchaseOrders(filters, limit, offset)
	return c.JSON(result)
}

func (h *PurchaseHandler) GetPurchaseOrder(c *fiber.Ctx) error {
	result := h.purchaseService.GetPurchaseOrderByID(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

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

func (h *PurchaseHandler) CancelPurchaseOrder(c *fiber.Ctx) error {
	result := h.purchaseService.CancelPurchaseOrder(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}
