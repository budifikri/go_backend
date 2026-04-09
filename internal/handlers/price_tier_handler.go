package handlers

import (
	"fmt"

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

// GetPriceTiers godoc
// @Summary List price tiers
// @Tags PriceTiers
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param product_id query string false "Product ID"
// @Param search query string false "Search"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/price-tiers [get]
func (h *PriceTierHandler) GetPriceTiers(c *fiber.Ctx) error {
	var pid *string
	if v := c.Query("product_id"); v != "" {
		pid = &v
	}
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	result := h.priceTierService.GetPriceTiers(pid, c.Query("search"), limit, offset)
	return c.JSON(result)
}

// GetPriceTier godoc
// @Summary Get price tier
// @Tags PriceTiers
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Price Tier ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/price-tiers/{id} [get]
func (h *PriceTierHandler) GetPriceTier(c *fiber.Ctx) error {
	result := h.priceTierService.GetPriceTierByID(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// CreatePriceTier godoc
// @Summary Create price tier
// @Tags PriceTiers
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body object true "Price tier payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/price-tiers [post]
func (h *PriceTierHandler) CreatePriceTier(c *fiber.Ctx) error {
	var rawBody map[string]interface{}
	if err := c.BodyParser(&rawBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	if tiers, ok := rawBody["tiers"].([]interface{}); ok {
		productID, ok := rawBody["product_id"].(string)
		if !ok || productID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("product_id is required"))
		}

		var tierInputs []services.PriceTierInput
		for i, t := range tiers {
			tierMap, ok := t.(map[string]interface{})
			if !ok {
				return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse(fmt.Sprintf("Invalid tier format at index %d", i)))
			}

			tierName, _ := tierMap["tier_name"].(string)
			minQty := int(0)
			if v, ok := tierMap["min_quantity"].(float64); ok {
				minQty = int(v)
			}
			unitPrice := float64(0)
			if v, ok := tierMap["unit_price"].(float64); ok {
				unitPrice = v
			}

			tierInputs = append(tierInputs, services.PriceTierInput{
				TierName:    tierName,
				MinQuantity: minQty,
				UnitPrice:   unitPrice,
			})
		}

		result := h.priceTierService.SaveProductPriceTiers(productID, tierInputs)
		if !result.Success {
			return c.Status(fiber.StatusBadRequest).JSON(result)
		}
		return c.JSON(result)
	}

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

// UpdatePriceTier godoc
// @Summary Update price tier
// @Tags PriceTiers
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Price Tier ID"
// @Param body body object true "Update payload"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/price-tiers/{id} [put]
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

// DeletePriceTier godoc
// @Summary Delete price tier
// @Tags PriceTiers
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Price Tier ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/price-tiers/{id} [delete]
func (h *PriceTierHandler) DeletePriceTier(c *fiber.Ctx) error {
	result := h.priceTierService.DeletePriceTier(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// SaveProductPriceTiers godoc
// @Summary Save price tiers for a product
// @Tags PriceTiers
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param product_id path string true "Product ID"
// @Param body body object{product_id=string,tiers=[]object{tier_name=string,min_quantity=number,unit_price=number}} true "Price tiers payload"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/price-tiers/product/{product_id} [post]
func (h *PriceTierHandler) SaveProductPriceTiers(c *fiber.Ctx) error {
	var body struct {
		ProductID string                    `json:"product_id"`
		Tiers     []services.PriceTierInput `json:"tiers"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}
	productID := c.Params("product_id")
	if productID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Product ID is required"))
	}
	result := h.priceTierService.SaveProductPriceTiers(productID, body.Tiers)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

// GetPriceTiersByProduct godoc
// @Summary Get price tiers by product ID
// @Tags PriceTiers
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param product_id path string true "Product ID"
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/price-tiers/product/{product_id} [get]
func (h *PriceTierHandler) GetPriceTiersByProduct(c *fiber.Ctx) error {
	productID := c.Params("product_id")
	if productID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Product ID is required"))
	}
	result := h.priceTierService.GetPriceTiers(&productID, "", 100, 0)
	return c.JSON(result)
}
