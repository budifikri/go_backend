package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type PriceTierHandler struct {
	priceTierService *services.PriceTierService
}

func NewPriceTierHandler(priceTierService *services.PriceTierService) *PriceTierHandler {
	return &PriceTierHandler{priceTierService: priceTierService}
}

func (h *PriceTierHandler) GetPriceTiers(c *fiber.Ctx) error {
	var pid *string
	if v := c.Query("product_id"); v != "" {
		pid = &v
	}
	result := h.priceTierService.GetPriceTiers(pid)
	return c.JSON(result)
}

func (h *PriceTierHandler) GetPriceTier(c *fiber.Ctx) error {
	result := h.priceTierService.GetPriceTierByID(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

func (h *PriceTierHandler) CreatePriceTier(c *fiber.Ctx) error {
	var body struct {
		ProductID   string  `json:"product_id"`
		TierName    string  `json:"tier_name"`
		MinQuantity int     `json:"min_quantity"`
		MaxQuantity *int    `json:"max_quantity"`
		UnitPrice   float64 `json:"unit_price"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}
	result := h.priceTierService.CreatePriceTier(services.CreatePriceTierInput{
		ProductID:   body.ProductID,
		TierName:    body.TierName,
		MinQuantity: body.MinQuantity,
		MaxQuantity: body.MaxQuantity,
		UnitPrice:   body.UnitPrice,
	})
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *PriceTierHandler) UpdatePriceTier(c *fiber.Ctx) error {
	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}
	input := services.UpdatePriceTierInput{}
	if v, ok := body["tier_name"].(string); ok {
		input.TierName = &v
	}
	if v, ok := body["min_quantity"].(float64); ok {
		n := int(v)
		input.MinQuantity = &n
	}
	if _, ok := body["max_quantity"]; ok {
		if v, ok := body["max_quantity"].(float64); ok {
			n := int(v)
			ptr := &n
			input.MaxQuantity = &ptr
		} else {
			var nilPtr *int
			input.MaxQuantity = &nilPtr
		}
	}
	if v, ok := body["unit_price"].(float64); ok {
		input.UnitPrice = &v
	}
	if v, ok := body["is_active"].(bool); ok {
		input.IsActive = &v
	}
	result := h.priceTierService.UpdatePriceTier(c.Params("id"), input)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

func (h *PriceTierHandler) DeletePriceTier(c *fiber.Ctx) error {
	result := h.priceTierService.DeletePriceTier(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}
