package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/middleware"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type WarehouseHandler struct {
	warehouseService *services.WarehouseService
}

func NewWarehouseHandler(warehouseService *services.WarehouseService) *WarehouseHandler {
	return &WarehouseHandler{warehouseService: warehouseService}
}

// GetWarehouses godoc
// @Summary List warehouses
// @Description List warehouses (default active only)
// @Tags Warehouses
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param is_active query bool false "Filter by active"
// @Param include_inactive query bool false "Include inactive (ignore default active-only)"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/warehouses [get]
func (h *WarehouseHandler) GetWarehouses(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	var companyID *string
	if user != nil && user.CompanyID != "" {
		companyID = &user.CompanyID
	}

	includeInactive := c.QueryBool("include_inactive", false)
	var isActive *bool
	if v := c.Query("is_active"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			isActive = &b
		}
	} else if !includeInactive {
		b := true
		isActive = &b
	}
	result := h.warehouseService.GetWarehouses(companyID, isActive)
	return c.JSON(result)
}

// GetWarehouse godoc
// @Summary Get warehouse details
// @Description Get warehouse by ID
// @Tags Warehouses
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Warehouse ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/warehouses/{id} [get]
func (h *WarehouseHandler) GetWarehouse(c *fiber.Ctx) error {
	result := h.warehouseService.GetWarehouseByID(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// CreateWarehouse godoc
// @Summary Create warehouse
// @Description Create a new warehouse
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body request.CreateWarehouseRequest true "Warehouse payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 409 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/warehouses [post]
func (h *WarehouseHandler) CreateWarehouse(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	var req request.CreateWarehouseRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.CreateWarehouseRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	}

	companyID := user.CompanyID
	result := h.warehouseService.CreateWarehouse(services.CreateWarehouseInput{
		Code:      req.Code,
		Name:      req.Name,
		Type:      req.Type,
		Address:   req.Address,
		City:      req.City,
		Phone:     req.Phone,
		CompanyID: &companyID,
	})
	if !result.Success {
		if result.Error == "Warehouse code already exists" {
			return c.Status(fiber.StatusConflict).JSON(result)
		}
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

// UpdateWarehouse godoc
// @Summary Update warehouse
// @Description Update warehouse details
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Warehouse ID"
// @Param body body request.UpdateWarehouseRequest true "Update payload"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Failure 409 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/warehouses/{id} [put]
func (h *WarehouseHandler) UpdateWarehouse(c *fiber.Ctx) error {
	var req request.UpdateWarehouseRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.UpdateWarehouseRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	}

	result := h.warehouseService.UpdateWarehouse(c.Params("id"), services.UpdateWarehouseInput{
		Code:     req.Code,
		Name:     req.Name,
		Type:     req.Type,
		Address:  req.Address,
		City:     req.City,
		Phone:    req.Phone,
		IsActive: req.IsActive,
	})
	if !result.Success {
		if result.Error == "Warehouse not found" {
			return c.Status(fiber.StatusNotFound).JSON(result)
		}
		if result.Error == "Warehouse code already exists" {
			return c.Status(fiber.StatusConflict).JSON(result)
		}
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

// DeleteWarehouse godoc
// @Summary Delete warehouse
// @Description Delete warehouse
// @Tags Warehouses
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Warehouse ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/warehouses/{id} [delete]
func (h *WarehouseHandler) DeleteWarehouse(c *fiber.Ctx) error {
	result := h.warehouseService.DeleteWarehouse(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}
