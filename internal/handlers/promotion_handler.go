package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type PromotionHandler struct {
	promotionService *services.PromotionService
}

func NewPromotionHandler(promotionService *services.PromotionService) *PromotionHandler {
	return &PromotionHandler{promotionService: promotionService}
}

// GetPromotions godoc
// @Summary List promotions
// @Tags Promotions
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param is_active query bool false "Is active"
// @Param type query string false "Promotion type"
// @Param scope query string false "Scope"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/promotions [get]
func (h *PromotionHandler) GetPromotions(c *fiber.Ctx) error {
	var isActive *bool
	if v := c.Query("is_active"); v != "" {
		b := v == "true"
		isActive = &b
	}
	var promoType *string
	if v := c.Query("type"); v != "" {
		promoType = &v
	}
	var scope *string
	if v := c.Query("scope"); v != "" {
		scope = &v
	}
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)
	result := h.promotionService.GetPromotions(isActive, promoType, scope, limit, offset)
	return c.JSON(result)
}

// GetPromotion godoc
// @Summary Get promotion
// @Tags Promotions
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Promotion ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/promotions/{id} [get]
func (h *PromotionHandler) GetPromotion(c *fiber.Ctx) error {
	result := h.promotionService.GetPromotionByID(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// CreatePromotion godoc
// @Summary Create promotion
// @Tags Promotions
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body object true "Promotion payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/promotions [post]
func (h *PromotionHandler) CreatePromotion(c *fiber.Ctx) error {
	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	code, _ := body["code"].(string)
	name, _ := body["name"].(string)
	desc, _ := body["description"].(string)
	pType, _ := body["promotion_type"].(string)
	scope, _ := body["scope"].(string)
	dd, _ := body["discount_value"].(float64)
	minPurchase, _ := body["min_purchase_amount"].(float64)
	maxDiscount, _ := body["max_discount_amount"].(float64)
	startStr, _ := body["start_date"].(string)
	endStr, _ := body["end_date"].(string)
	usageLimitFloat, _ := body["usage_limit"].(float64)

	start, _ := time.Parse(time.RFC3339, startStr)
	end, _ := time.Parse(time.RFC3339, endStr)

	var descPtr *string
	if desc != "" {
		descPtr = &desc
	}
	var minPtr *float64
	if _, ok := body["min_purchase_amount"]; ok {
		minPtr = &minPurchase
	}
	var maxPtr *float64
	if _, ok := body["max_discount_amount"]; ok {
		maxPtr = &maxDiscount
	}
	var usageLimit *int
	if _, ok := body["usage_limit"]; ok {
		n := int(usageLimitFloat)
		usageLimit = &n
	}

	getStringSlice := func(key string) []string {
		v, ok := body[key]
		if !ok {
			return nil
		}
		arr, ok := v.([]interface{})
		if !ok {
			return nil
		}
		out := make([]string, 0, len(arr))
		for _, x := range arr {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}

	result := h.promotionService.CreatePromotion(services.CreatePromotionInput{
		Code:              code,
		Name:              name,
		Description:       descPtr,
		PromotionType:     pType,
		Scope:             scope,
		DiscountValue:     dd,
		MinPurchaseAmount: minPtr,
		MaxDiscountAmount: maxPtr,
		StartDate:         start,
		EndDate:           end,
		UsageLimit:        usageLimit,
		ProductIDs:        getStringSlice("product_ids"),
		CategoryIDs:       getStringSlice("category_ids"),
		CustomerIDs:       getStringSlice("customer_ids"),
	})

	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

// UpdatePromotion godoc
// @Summary Update promotion
// @Tags Promotions
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Promotion ID"
// @Param body body object true "Update payload"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/promotions/{id} [put]
func (h *PromotionHandler) UpdatePromotion(c *fiber.Ctx) error {
	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	input := services.UpdatePromotionInput{}
	if v, ok := body["code"].(string); ok {
		input.Code = &v
	}
	if v, ok := body["name"].(string); ok {
		input.Name = &v
	}
	if v, ok := body["description"].(string); ok {
		input.Description = &v
	}
	if v, ok := body["discount_value"].(float64); ok {
		input.DiscountValue = &v
	}
	if v, ok := body["min_purchase_amount"].(float64); ok {
		input.MinPurchaseAmount = &v
	}
	if v, ok := body["max_discount_amount"].(float64); ok {
		input.MaxDiscountAmount = &v
	}
	if v, ok := body["is_active"].(bool); ok {
		input.IsActive = &v
	}
	if v, ok := body["usage_limit"].(float64); ok {
		n := int(v)
		input.UsageLimit = &n
	}
	if v, ok := body["start_date"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			input.StartDate = &t
		}
	}
	if v, ok := body["end_date"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			input.EndDate = &t
		}
	}
	// association lists
	if v, ok := body["product_ids"]; ok {
		ids := parseStringSlice(v)
		input.ProductIDs = &ids
	}
	if v, ok := body["category_ids"]; ok {
		ids := parseStringSlice(v)
		input.CategoryIDs = &ids
	}
	if v, ok := body["customer_ids"]; ok {
		ids := parseStringSlice(v)
		input.CustomerIDs = &ids
	}

	result := h.promotionService.UpdatePromotion(c.Params("id"), input)
	if !result.Success {
		status := fiber.StatusBadRequest
		if result.Error == "Promotion not found" {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(result)
	}
	return c.JSON(result)
}

// DeletePromotion godoc
// @Summary Delete promotion
// @Tags Promotions
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Promotion ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/promotions/{id} [delete]
func (h *PromotionHandler) DeletePromotion(c *fiber.Ctx) error {
	result := h.promotionService.DeletePromotion(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

func parseStringSlice(v interface{}) []string {
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, x := range arr {
		if s, ok := x.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
