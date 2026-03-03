package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/middleware"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type CashDrawerHandler struct {
	cashDrawerService *services.CashDrawerService
}

func NewCashDrawerHandler(cashDrawerService *services.CashDrawerService) *CashDrawerHandler {
	return &CashDrawerHandler{cashDrawerService: cashDrawerService}
}

// OpenCashDrawer godoc
// @Summary Open cash drawer
// @Tags CashDrawers
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body request.OpenCashDrawerRequest true "Open drawer payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/cash-drawers/open [post]
func (h *CashDrawerHandler) OpenCashDrawer(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	var req request.OpenCashDrawerRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.OpenCashDrawerRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	}

	result := h.cashDrawerService.OpenCashDrawer(services.OpenCashDrawerInput{
		DrawerNumber:   req.DrawerNumber,
		WarehouseID:    req.WarehouseID,
		OpeningBalance: req.OpeningBalance,
		Notes:          req.Notes,
	}, user.CompanyID, user.UserID)

	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   result.Error,
			"message": "Failed to open cash drawer",
		})
	}
	result.Message = "Cash drawer opened successfully"
	return c.Status(fiber.StatusCreated).JSON(result)
}

// GetCurrentDrawer godoc
// @Summary Get current open drawer
// @Tags CashDrawers
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param warehouse_id query string false "Warehouse ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/cash-drawers/current [get]
func (h *CashDrawerHandler) GetCurrentDrawer(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	var wid *string
	if q := c.Query("warehouse_id"); q != "" {
		wid = &q
	}
	result := h.cashDrawerService.GetCurrentDrawer(wid, user.CompanyID, user.UserID)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   result.Error,
			"message": "No open cash drawer found",
		})
	}
	return c.JSON(result)
}

// CashIn godoc
// @Summary Add cash in
// @Tags CashDrawers
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Drawer ID"
// @Param body body request.CashInOutRequest true "Cash in payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/cash-drawers/{id}/cash-in [post]
func (h *CashDrawerHandler) CashIn(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	var req request.CashInOutRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.CashInOutRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	}
	result := h.cashDrawerService.AddCashIn(c.Params("id"), services.CashInOutInput{Amount: req.Amount, Reason: req.Reason}, user.CompanyID, user.UserID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   result.Error,
			"message": "Failed to add cash in",
		})
	}
	result.Message = "Cash in added successfully"
	return c.Status(fiber.StatusCreated).JSON(result)
}

// CashOut godoc
// @Summary Add cash out
// @Tags CashDrawers
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Drawer ID"
// @Param body body request.CashInOutRequest true "Cash out payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/cash-drawers/{id}/cash-out [post]
func (h *CashDrawerHandler) CashOut(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	var req request.CashInOutRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.CashInOutRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	}
	result := h.cashDrawerService.AddCashOut(c.Params("id"), services.CashInOutInput{Amount: req.Amount, Reason: req.Reason}, user.CompanyID, user.UserID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   result.Error,
			"message": "Failed to add cash out",
		})
	}
	result.Message = "Cash out added successfully"
	return c.Status(fiber.StatusCreated).JSON(result)
}

// GetTransactions godoc
// @Summary List drawer transactions
// @Tags CashDrawers
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Drawer ID"
// @Param type query string false "Transaction type"
// @Param search query string false "Search"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/cash-drawers/{id}/transactions [get]
func (h *CashDrawerHandler) GetTransactions(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 50)
	if limit <= 0 {
		limit = 50
	}
	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}
	var typ *string
	if q := c.Query("type"); q != "" {
		typ = &q
	}
	result := h.cashDrawerService.GetDrawerTransactions(c.Params("id"), typ, c.Query("search"), limit, offset)
	return c.JSON(result)
}

// CloseCashDrawer godoc
// @Summary Close cash drawer
// @Tags CashDrawers
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Drawer ID"
// @Param body body request.CloseCashDrawerRequest true "Close drawer payload"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/cash-drawers/{id}/close [post]
func (h *CashDrawerHandler) CloseCashDrawer(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	var req request.CloseCashDrawerRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.CloseCashDrawerRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	}

	result := h.cashDrawerService.CloseCashDrawer(c.Params("id"), services.CloseCashDrawerInput{
		ClosingBalance: req.ClosingBalance,
		Notes:          req.Notes,
		PaymentMethod:  req.PaymentMethod,
	}, user.CompanyID, user.UserID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   result.Error,
			"message": "Failed to close cash drawer",
		})
	}
	result.Message = "Cash drawer closed successfully"
	return c.JSON(result)
}

// GetSummary godoc
// @Summary Get drawer summary
// @Tags CashDrawers
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Drawer ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/cash-drawers/{id}/summary [get]
func (h *CashDrawerHandler) GetSummary(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	result := h.cashDrawerService.GetDrawerSummary(c.Params("id"), user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   result.Error,
			"message": "Failed to get drawer summary",
		})
	}
	return c.JSON(result)
}

// ListCashDrawers godoc
// @Summary List cash drawers
// @Tags CashDrawers
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param warehouse_id query string false "Warehouse ID"
// @Param cashier_id query string false "Cashier ID"
// @Param status query string false "Status"
// @Param search query string false "Search"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/cash-drawers [get]
func (h *CashDrawerHandler) ListCashDrawers(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	filters := map[string]string{}
	if v := c.Query("warehouse_id"); v != "" {
		filters["warehouse_id"] = v
	}
	if v := c.Query("cashier_id"); v != "" {
		filters["cashier_id"] = v
	}
	if v := c.Query("status"); v != "" {
		filters["status"] = v
	}
	if v := c.Query("search"); v != "" {
		filters["search"] = v
	}
	limit := c.QueryInt("limit", 50)
	if limit <= 0 {
		limit = 50
	}
	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}
	result := h.cashDrawerService.ListCashDrawers(filters, limit, offset, user.CompanyID)
	return c.JSON(result)
}

// GetCashDrawer godoc
// @Summary Get cash drawer
// @Tags CashDrawers
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Drawer ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/cash-drawers/{id} [get]
func (h *CashDrawerHandler) GetCashDrawer(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	result := h.cashDrawerService.GetCashDrawerByID(c.Params("id"), user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   result.Error,
			"message": "Cash drawer not found",
		})
	}
	return c.JSON(result)
}
