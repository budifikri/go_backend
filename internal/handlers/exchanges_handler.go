package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/middleware"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type ExchangesHandler struct {
	exchangesService *services.ExchangesService
}

func NewExchangesHandler(exchangesService *services.ExchangesService) *ExchangesHandler {
	return &ExchangesHandler{exchangesService: exchangesService}
}

func (h *ExchangesHandler) CreateExchange(c *fiber.Ctx) error {
	var req request.CreateExchangeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	returnedItems := make([]services.ExchangeReturnedItemInput, 0, len(req.ReturnedItems))
	for _, it := range req.ReturnedItems {
		returnedItems = append(returnedItems, services.ExchangeReturnedItemInput{
			SaleItemID: it.SaleItemID,
			ProductID:  it.ProductID,
			Quantity:   it.Quantity,
			Condition:  it.Condition,
		})
	}

	receivedItems := make([]services.ExchangeReceivedItemInput, 0, len(req.ReceivedItems))
	for _, it := range req.ReceivedItems {
		receivedItems = append(receivedItems, services.ExchangeReceivedItemInput{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			UnitPrice: it.UnitPrice,
		})
	}

	result := h.exchangesService.CreateExchange(services.CreateExchangeInput{
		SaleID:        req.SaleID,
		WarehouseID:   req.WarehouseID,
		Reason:        req.Reason,
		ReturnedItems: returnedItems,
		ReceivedItems: receivedItems,
	}, user.UserID)

	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *ExchangesHandler) GetExchange(c *fiber.Ctx) error {
	result := h.exchangesService.GetExchangeByID(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

func (h *ExchangesHandler) GetExchanges(c *fiber.Ctx) error {
	filters := map[string]string{}
	filters["warehouse_id"] = c.Query("warehouse_id")
	filters["sale_id"] = c.Query("sale_id")
	filters["status"] = c.Query("status")

	limit := c.QueryInt("limit", 50)
	if limit <= 0 {
		limit = 50
	}
	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}

	result := h.exchangesService.GetExchanges(filters, limit, offset)
	return c.JSON(result)
}
