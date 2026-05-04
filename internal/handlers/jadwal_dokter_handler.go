package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/middleware"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type JadwalDokterHandler struct {
	service *services.JadwalDokterService
}

func NewJadwalDokterHandler(service *services.JadwalDokterService) *JadwalDokterHandler {
	return &JadwalDokterHandler{service: service}
}

// CreateJadwalDokter godoc
// @Summary Create jadwal dokter
// @Tags JadwalDokter
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body request.CreateJadwalDokterRequest true "Jadwal dokter payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/jadwal-dokter [post]
func (h *JadwalDokterHandler) CreateJadwalDokter(c *fiber.Ctx) error {
	var req request.CreateJadwalDokterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	result := h.service.CreateJadwalDokter(req, user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

// GetJadwals godoc
// @Summary List jadwal dokter
// @Tags JadwalDokter
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param dokter_id query string false "Filter by dokter ID"
// @Param hari query string false "Filter by hari"
// @Param is_active query bool false "Filter by active status"
// @Param search query string false "Search by dokter name or hari"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/jadwal-dokter [get]
func (h *JadwalDokterHandler) GetJadwals(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	filters := map[string]interface{}{}

	if v := c.Query("dokter_id"); v != "" {
		filters["dokter_id"] = v
	}
	if v := c.Query("hari"); v != "" {
		filters["hari"] = v
	}
	if v := c.Query("is_active"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			filters["is_active"] = b
		}
	}
	if v := c.Query("search"); v != "" {
		filters["search"] = v
	}

	result := h.service.GetJadwals(user.CompanyID, filters, limit, offset)
	return c.JSON(result)
}

// GetJadwalDokter godoc
// @Summary Get jadwal dokter by ID
// @Tags JadwalDokter
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Jadwal Dokter ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/jadwal-dokter/{id} [get]
func (h *JadwalDokterHandler) GetJadwalDokter(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	result := h.service.GetJadwalByID(c.Params("id"), user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// UpdateJadwalDokter godoc
// @Summary Update jadwal dokter
// @Tags JadwalDokter
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Jadwal Dokter ID"
// @Param body body request.UpdateJadwalDokterRequest true "Update payload"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/jadwal-dokter/{id} [put]
func (h *JadwalDokterHandler) UpdateJadwalDokter(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	var req request.UpdateJadwalDokterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	result := h.service.UpdateJadwalDokter(c.Params("id"), req, user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// DeleteJadwalDokter godoc
// @Summary Delete jadwal dokter
// @Tags JadwalDokter
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Jadwal Dokter ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/jadwal-dokter/{id} [delete]
func (h *JadwalDokterHandler) DeleteJadwalDokter(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	result := h.service.DeleteJadwalDokter(c.Params("id"), user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}
