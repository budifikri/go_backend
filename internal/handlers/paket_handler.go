package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/middleware"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type PaketHandler struct {
	paketService *services.PaketService
}

func NewPaketHandler(paketService *services.PaketService) *PaketHandler {
	return &PaketHandler{paketService: paketService}
}

func (h *PaketHandler) GetPakets(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	filters := make(map[string]interface{})
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if active := c.Query("active"); active != "" {
		filters["active"] = active == "true"
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	result := h.paketService.GetPakets(user.CompanyID, filters, limit, offset)
	return c.JSON(result)
}

func (h *PaketHandler) GetPaket(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	id := c.Params("id")

	result := h.paketService.GetPaket(id, user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

func (h *PaketHandler) CreatePaket(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	var req request.CreatePaketRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	result := h.paketService.CreatePaket(req, user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *PaketHandler) UpdatePaket(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	id := c.Params("id")

	var req request.UpdatePaketRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	result := h.paketService.UpdatePaket(id, req, user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

func (h *PaketHandler) DeletePaket(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	id := c.Params("id")

	result := h.paketService.DeletePaket(id, user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}
