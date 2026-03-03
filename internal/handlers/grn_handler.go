package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/middleware"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type GrnHandler struct {
	grnService *services.GrnService
}

func NewGrnHandler(grnService *services.GrnService) *GrnHandler {
	return &GrnHandler{grnService: grnService}
}

// CreateGrn godoc
// @Summary Create GRN
// @Tags GRN
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body request.CreateGrnRequest true "GRN payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/grn [post]
func (h *GrnHandler) CreateGrn(c *fiber.Ctx) error {
	var req request.CreateGrnRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("User not authenticated"))
	}

	result := h.grnService.CreateGrn(req.PoID, req.WarehouseID, req.InvoiceNumber, req.Notes, user.UserID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

// GetGrn godoc
// @Summary Get GRN
// @Tags GRN
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "GRN ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/grn/{id} [get]
func (h *GrnHandler) GetGrn(c *fiber.Ctx) error {
	result := h.grnService.GetGrnByID(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// ListGrns godoc
// @Summary List GRNs
// @Tags GRN
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param page query int false "Page" default(1)
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset (overrides page)" default(0)
// @Param search query string false "Search"
// @Param status query string false "Status"
// @Param poId query string false "PO ID"
// @Param warehouseId query string false "Warehouse ID"
// @Param startDate query string false "Start date"
// @Param endDate query string false "End date"
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/grn [get]
func (h *GrnHandler) ListGrns(c *fiber.Ctx) error {
	offset := -1
	if v := c.Query("offset"); v != "" {
		offset = c.QueryInt("offset", 0)
	}

	filter := services.GrnFilter{
		Page:        c.QueryInt("page", 1),
		Limit:       c.QueryInt("limit", 50),
		Offset:      offset,
		Search:      c.Query("search"),
		Status:      c.Query("status"),
		PoID:        c.Query("poId"),
		WarehouseID: c.Query("warehouseId"),
		StartDate:   c.Query("startDate"),
		EndDate:     c.Query("endDate"),
	}
	result := h.grnService.GetGrns(filter)
	return c.JSON(result)
}

// UpdateGrn godoc
// @Summary Update GRN
// @Tags GRN
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "GRN ID"
// @Param body body request.UpdateGrnRequest true "Update payload"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/grn/{id} [put]
func (h *GrnHandler) UpdateGrn(c *fiber.Ctx) error {
	var req request.UpdateGrnRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	// Convert typed items to map slice for service
	var items *[]map[string]interface{}
	if req.Items != nil {
		converted := make([]map[string]interface{}, 0, len(*req.Items))
		for _, it := range *req.Items {
			m := map[string]interface{}{
				"poItemId":         it.PoItemID,
				"productId":        it.ProductID,
				"orderedQuantity":  it.OrderedQuantity,
				"receivedQuantity": it.ReceivedQuantity,
				"rejectedQuantity": it.RejectedQuantity,
				"unitPrice":        it.UnitPrice,
				"qualityNotes":     it.QualityNotes,
			}
			converted = append(converted, m)
		}
		items = &converted
	}

	result := h.grnService.UpdateGrn(c.Params("id"), req.InvoiceNumber, req.Notes, items, user.UserID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

// CancelGrn godoc
// @Summary Cancel GRN
// @Tags GRN
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "GRN ID"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/grn/{id} [delete]
func (h *GrnHandler) CancelGrn(c *fiber.Ctx) error {
	result := h.grnService.CancelGrn(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

// VerifyGrn godoc
// @Summary Verify GRN
// @Tags GRN
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "GRN ID"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/grn/{id}/verify [put]
func (h *GrnHandler) VerifyGrn(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	result := h.grnService.VerifyGrn(c.Params("id"), user.UserID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}
