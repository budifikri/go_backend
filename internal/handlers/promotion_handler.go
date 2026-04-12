package handlers

import (
	"fmt"
	"reflect"
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
// @Param type query string false "Promotion type (PERCENTAGE, FIXED_AMOUNT, BUY_X_GET_Y, FLASH_SALE)"
// @Param scope query string false "Scope (ALL, BY_CATEGORY, BY_PRODUCT)"
// @Param search query string false "Search by code or name"
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
	result := h.promotionService.GetPromotions(isActive, promoType, scope, c.Query("search"), limit, offset)
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
// @Param body.code string false "Promotion code (auto-generated if not provided)"
// @Param body.name string true "Promotion name"
// @Param body.promotion_type string false "PERCENTAGE, FIXED_AMOUNT, BUY_X_GET_Y, FLASH_SALE"
// @Param body.scope string false "ALL, BY_CATEGORY, BY_PRODUCT"
// @Param body.discount_value number false "Discount value (percentage or fixed amount)"
// @Param body.buy_quantity number false "Buy quantity (for BUY_X_GET_Y)"
// @Param body.get_quantity number false "Free quantity (for BUY_X_GET_Y)"
// @Param body.start_date string true "Start date (RFC3339)"
// @Param body.start_time string false "Start time HH:MM (for FLASH_SALE)"
// @Param body.end_date string true "End date (RFC3339)"
// @Param body.end_time string false "End time HH:MM (for FLASH_SALE)"
// @Param body.product_ids array false "Product IDs (if scope=BY_PRODUCT)"
// @Param body.category_ids array false "Category IDs (if scope=BY_CATEGORY)"
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

	// Handle field name variations
	pType, _ := body["promo_type"].(string)
	if pType == "" {
		pType, _ = body["promotion_type"].(string)
	}
	// Normalize promotion type to uppercase enum
	pType = normalizePromoType(pType)
	// Default to PERCENTAGE if empty
	if pType == "" {
		pType = "PERCENTAGE"
	}

	scope, _ := body["scope_type"].(string)
	if scope == "" {
		scope, _ = body["scope"].(string)
	}
	// Normalize scope to uppercase enum
	scope = normalizeScope(scope)
	// Default to ALL if empty
	if scope == "" {
		scope = "ALL"
	}

	dd, _ := body["discount_value"].(float64)
	minPurchase, _ := body["min_purchase_amount"].(float64)
	maxDiscount, _ := body["max_discount_amount"].(float64)
	buyQty, _ := body["buy_quantity"].(float64)
	getQty, _ := body["get_quantity"].(float64)
	usageLimitFloat, _ := body["usage_limit"].(float64)

	fmt.Printf("[DEBUG] body keys: %v\n", reflect.ValueOf(body).MapKeys())

	// Get start_date properly with existence check
	startStr := ""
	if v, ok := body["start_date"]; ok {
		if s, ok := v.(string); ok {
			startStr = s
		}
	}
	endStr := ""
	if v, ok := body["end_date"]; ok {
		if s, ok := v.(string); ok {
			endStr = s
		}
	}
	startTime := ""
	if v, ok := body["start_time"]; ok {
		if s, ok := v.(string); ok {
			startTime = s
		}
	}
	endTime := ""
	if v, ok := body["end_time"]; ok {
		if s, ok := v.(string); ok {
			endTime = s
		}
	}

	fmt.Printf("[DEBUG] startStr='%s', endStr='%s'\n", startStr, endStr)

	var startPtr, endPtr *time.Time
	var startTimePtr, endTimePtr *time.Time

	// Simple direct parse for format "2026-04-01"
	if startStr != "" {
		// Parse as date only (midnight UTC)
		t, err := time.Parse("2006-01-02", startStr)
		if err == nil {
			startPtr = &t
			fmt.Printf("[DEBUG] start date parsed: %v\n", t)
		}
	} else {
		// Default to now if not provided
		now := time.Now()
		startPtr = &now
		fmt.Printf("[DEBUG] start date defaulted to now: %v\n", now)
	}
	if endStr != "" {
		t, err := time.Parse("2006-01-02", endStr)
		if err == nil {
			endPtr = &t
			fmt.Printf("[DEBUG] end date parsed: %v\n", t)
		}
	} else {
		// Default to one month from now if not provided
		now := time.Now()
		endDate := now.AddDate(0, 1, 0)
		endPtr = &endDate
		fmt.Printf("[DEBUG] end date defaulted to one month from now: %v\n", *endPtr)
	}

	// Handle times - parse from "HH:MM" or "HH:MM:SS" format
	if startTime != "" {
		if t, err := time.Parse("15:04", startTime); err == nil {
			startTimePtr = &t
			fmt.Printf("[DEBUG] start time parsed: %v\n", t)
		} else if t, err := time.Parse("15:04:05", startTime); err == nil {
			startTimePtr = &t
			fmt.Printf("[DEBUG] start time parsed (seconds): %v\n", t)
		}
	}
	if endTime != "" {
		if t, err := time.Parse("15:04", endTime); err == nil {
			endTimePtr = &t
			fmt.Printf("[DEBUG] end time parsed: %v\n", t)
		} else if t, err := time.Parse("15:04:05", endTime); err == nil {
			endTimePtr = &t
			fmt.Printf("[DEBUG] end time parsed (seconds): %v\n", t)
		}
	}

	var buyQtyPtr *int
	if _, ok := body["buy_quantity"]; ok {
		n := int(buyQty)
		buyQtyPtr = &n
	}
	var getQtyPtr *int
	if _, ok := body["get_quantity"]; ok {
		n := int(getQty)
		getQtyPtr = &n
	}

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
		BuyQuantity:       buyQtyPtr,
		GetQuantity:       getQtyPtr,
		StartDate:         startPtr,
		StartTime:         startTimePtr,
		EndDate:           endPtr,
		EndTime:           endTimePtr,
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
// @Param body.promotion_type string false "PERCENTAGE, FIXED_AMOUNT, BUY_X_GET_Y, FLASH_SALE"
// @Param body.scope string false "ALL, BY_CATEGORY, BY_PRODUCT"
// @Param body.discount_value number false "Discount value"
// @Param body.buy_quantity number false "Buy quantity (for BUY_X_GET_Y)"
// @Param body.get_quantity number false "Free quantity (for BUY_X_GET_Y)"
// @Param body.start_time string false "Start time HH:MM (for FLASH_SALE)"
// @Param body.end_time string false "End time HH:MM (for FLASH_SALE)"
// @Param body.product_ids array false "Product IDs"
// @Param body.category_ids array false "Category IDs"
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
	if v, ok := body["promotion_type"].(string); ok {
		input.PromotionType = &v
	}
	if v, ok := body["scope"].(string); ok {
		input.Scope = &v
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
	if v, ok := body["buy_quantity"].(float64); ok {
		n := int(v)
		input.BuyQuantity = &n
	}
	if v, ok := body["get_quantity"].(float64); ok {
		n := int(v)
		input.GetQuantity = &n
	}
	if v, ok := body["is_active"].(bool); ok {
		input.IsActive = &v
	}
	if v, ok := body["usage_limit"].(float64); ok {
		n := int(v)
		input.UsageLimit = &n
	}
	if v, ok := body["start_date"].(string); ok {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			input.StartDate = &t
		}
	}
	if v, ok := body["start_time"].(string); ok {
		if t, err := time.Parse("15:04", v); err == nil {
			input.StartTime = &t
		} else if t, err := time.Parse("15:04:05", v); err == nil {
			input.StartTime = &t
		}
	}
	if v, ok := body["end_date"].(string); ok {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			input.EndDate = &t
		}
	}
	if v, ok := body["end_time"].(string); ok {
		if t, err := time.Parse("15:04", v); err == nil {
			input.EndTime = &t
		} else if t, err := time.Parse("15:04:05", v); err == nil {
			input.EndTime = &t
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

func normalizePromoType(t string) string {
	switch t {
	case "percentage":
		return "PERCENTAGE"
	case "fixed_amount":
		return "FIXED_AMOUNT"
	case "buy_x_get_y":
		return "BUY_X_GET_Y"
	case "flash_sale":
		return "FLASH_SALE"
	}
	return t
}

func normalizeScope(s string) string {
	switch s {
	case "all":
		return "ALL"
	case "by_category":
		return "BY_CATEGORY"
	case "by_product":
		return "BY_PRODUCT"
	}
	return s
}
