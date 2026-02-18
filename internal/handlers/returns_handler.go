package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/middleware"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type ReturnsHandler struct {
	returnsService *services.ReturnsService
}

func NewReturnsHandler(returnsService *services.ReturnsService) *ReturnsHandler {
	return &ReturnsHandler{returnsService: returnsService}
}

// CreateReturn godoc
// @Summary Create return
// @Tags Returns
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body request.CreateReturnRequest true "Return payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/returns [post]
func (h *ReturnsHandler) CreateReturn(c *fiber.Ctx) error {
	var req request.CreateReturnRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	items := make([]services.CreateReturnItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, services.CreateReturnItemInput{
			SaleItemID: it.SaleItemID,
			ProductID:  it.ProductID,
			Quantity:   it.Quantity,
			Condition:  it.Condition,
			Notes:      it.Notes,
		})
	}

	result := h.returnsService.CreateReturn(services.CreateReturnInput{
		SaleID:       req.SaleID,
		WarehouseID:  req.WarehouseID,
		Reason:       req.Reason,
		Items:        items,
		RefundMethod: req.RefundMethod,
	}, user.UserID)

	if !result.Success {
		// TS returns service uses 400 for most errors
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

// GetReturn godoc
// @Summary Get return
// @Tags Returns
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Return ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/returns/{id} [get]
func (h *ReturnsHandler) GetReturn(c *fiber.Ctx) error {
	result := h.returnsService.GetReturnByID(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// GetReturns godoc
// @Summary List returns
// @Tags Returns
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param warehouse_id query string false "Warehouse ID"
// @Param sale_id query string false "Sale ID"
// @Param status query string false "Return status"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/returns [get]
func (h *ReturnsHandler) GetReturns(c *fiber.Ctx) error {
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

	result := h.returnsService.GetReturns(filters, limit, offset)
	return c.JSON(result)
}
