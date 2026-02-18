package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/middleware"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type SalesHandler struct {
	salesService *services.SalesService
}

func NewSalesHandler(salesService *services.SalesService) *SalesHandler {
	return &SalesHandler{salesService: salesService}
}

func (h *SalesHandler) CreateSale(c *fiber.Ctx) error {
	var req request.CreateSaleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	items := make([]services.CreateSaleItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, services.CreateSaleItemInput{
			ProductID:     it.ProductID,
			Quantity:      it.Quantity,
			PromotionCode: it.PromotionCode,
		})
	}

	payments := make([]services.CreateSalePaymentInput, 0, len(req.Payments))
	for _, p := range req.Payments {
		payments = append(payments, services.CreateSalePaymentInput{
			Method:          p.Method,
			Amount:          p.Amount,
			ReferenceNumber: p.ReferenceNumber,
			CardLast4:       p.CardLast4,
		})
	}

	result := h.salesService.CreateSale(services.CreateSaleInput{
		WarehouseID:         req.WarehouseID,
		CustomerID:          req.CustomerID,
		Items:               items,
		Payments:            payments,
		LoyaltyPointsRedeem: req.LoyaltyPointsRedeem,
		Notes:               req.Notes,
		CompanyID:           user.CompanyID,
	}, user.UserID)

	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *SalesHandler) GetSale(c *fiber.Ctx) error {
	result := h.salesService.GetSaleByID(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

func (h *SalesHandler) GetSales(c *fiber.Ctx) error {
	filters := map[string]string{}

	filters["warehouse_id"] = c.Query("warehouse_id")
	filters["customer_id"] = c.Query("customer_id")
	filters["cashier_id"] = c.Query("cashier_id")
	filters["status"] = c.Query("status")
	filters["date_from"] = c.Query("date_from")
	filters["date_to"] = c.Query("date_to")
	filters["sale_number"] = c.Query("sale_number")

	limit := c.QueryInt("limit", 50)
	if limit <= 0 {
		limit = 50
	}
	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}

	result := h.salesService.GetSales(filters, limit, offset)
	return c.JSON(result)
}
