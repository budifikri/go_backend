package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/middleware"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type DokterHandler struct {
	dokterService *services.DokterService
}

func NewDokterHandler(dokterService *services.DokterService) *DokterHandler {
	return &DokterHandler{dokterService: dokterService}
}

func (h *DokterHandler) GetDokters(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	filters := make(map[string]interface{})
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if tipe := c.Query("tipe"); tipe != "" {
		filters["tipe"] = tipe
	}
	if active := c.Query("active"); active != "" {
		filters["active"] = active == "true"
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	result := h.dokterService.GetDokters(user.CompanyID, filters, limit, offset)
	return c.JSON(result)
}

func (h *DokterHandler) GetDokter(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	id := c.Params("id")

	result := h.dokterService.GetDokterByID(id, user.CompanyID)
	return c.JSON(result)
}

func (h *DokterHandler) CreateDokter(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	var req request.CreateDokterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	result := h.dokterService.CreateDokter(req, user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *DokterHandler) UpdateDokter(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	id := c.Params("id")

	var req request.UpdateDokterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	result := h.dokterService.UpdateDokter(id, req, user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

func (h *DokterHandler) DeleteDokter(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	id := c.Params("id")

	result := h.dokterService.DeleteDokter(id, user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}
