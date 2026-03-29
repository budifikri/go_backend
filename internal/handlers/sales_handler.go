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

// CreateSale godoc
// @Summary Create sale
// @Tags Sales
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body request.CreateSaleRequest true "Sale payload (warehouse_id, cash_drawer_id, status, items)"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/sales [post]
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
		CashDrawerID:        req.CashDrawerID,
		Status:              req.Status,
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

// GetSale godoc
// @Summary Get sale
// @Tags Sales
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Sale ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/sales/{id} [get]
func (h *SalesHandler) GetSale(c *fiber.Ctx) error {
	result := h.salesService.GetSaleByID(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// GetSales godoc
// @Summary List sales
// @Tags Sales
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param warehouse_id query string false "Warehouse ID"
// @Param customer_id query string false "Customer ID"
// @Param cashier_id query string false "Cashier ID"
// @Param status query string false "Sale status"
// @Param date_from query string false "From date (YYYY-MM-DD)"
// @Param date_to query string false "To date (YYYY-MM-DD)"
// @Param sale_number query string false "Sale number search"
// @Param search query string false "Search"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/sales [get]
func (h *SalesHandler) GetSales(c *fiber.Ctx) error {
	filters := map[string]string{}

	filters["warehouse_id"] = c.Query("warehouse_id")
	filters["customer_id"] = c.Query("customer_id")
	filters["cashier_id"] = c.Query("cashier_id")
	filters["status"] = c.Query("status")
	filters["date_from"] = c.Query("date_from")
	filters["date_to"] = c.Query("date_to")
	filters["sale_number"] = c.Query("sale_number")
	filters["search"] = c.Query("search")

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
