package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/middleware"
	"github.com/pos-retail/go_backend/internal/services"
)

type TreatmentHandler struct {
	treatmentService *services.TreatmentService
}

func NewTreatmentHandler(treatmentService *services.TreatmentService) *TreatmentHandler {
	return &TreatmentHandler{
		treatmentService: treatmentService,
	}
}

// GetTreatments godoc
// @Summary List treatments
// @Tags Treatments
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param is_active query bool false "Filter by active"
// @Param include_inactive query bool false "Include inactive"
// @Param tag_id query string false "Filter by tag"
// @Param search query string false "Search term"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/treatments [get]
func (h *TreatmentHandler) GetTreatments(c *fiber.Ctx) error {
	filters := make(map[string]interface{})
	includeInactive := c.QueryBool("include_inactive", false)

	if v := c.Query("is_active"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			filters["is_active"] = b
		}
	} else if !includeInactive {
		filters["is_active"] = true
	}
	if tagID := c.Query("tag_id"); tagID != "" {
		filters["tag_id"] = tagID
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	user := middleware.GetUserFromContext(c)
	if user != nil && user.CompanyID != "" {
		filters["company_id"] = user.CompanyID
	}

	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if limitVal := c.QueryInt("limit", 50); limitVal > 0 {
			limit = limitVal
		}
	}
	if o := c.Query("offset"); o != "" {
		offset = c.QueryInt("offset", 0)
	}

	result := h.treatmentService.GetTreatments(filters, limit, offset)
	return c.JSON(result)
}

// GetTreatment godoc
// @Summary Get treatment
// @Tags Treatments
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Treatment ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/treatments/{id} [get]
func (h *TreatmentHandler) GetTreatment(c *fiber.Ctx) error {
	id := c.Params("id")
	result := h.treatmentService.GetTreatmentByID(id)
	return c.JSON(result)
}

// CreateTreatment godoc
// @Summary Create treatment
// @Tags Treatments
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param treatment body map[string]interface{} true "Treatment data"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/treatments [post]
func (h *TreatmentHandler) CreateTreatment(c *fiber.Ctx) error {
	var input map[string]interface{}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	user := middleware.GetUserFromContext(c)
	companyID := ""
	if user != nil {
		companyID = user.CompanyID
	}

	result := h.treatmentService.CreateTreatment(input, companyID)
	return c.JSON(result)
}

// UpdateTreatment godoc
// @Summary Update treatment
// @Tags Treatments
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Treatment ID"
// @Param treatment body map[string]interface{} true "Treatment data"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/treatments/{id} [put]
func (h *TreatmentHandler) UpdateTreatment(c *fiber.Ctx) error {
	id := c.Params("id")
	var input map[string]interface{}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	result := h.treatmentService.UpdateTreatment(id, input)
	return c.JSON(result)
}

// DeleteTreatment godoc
// @Summary Delete treatment
// @Tags Treatments
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Treatment ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/treatments/{id} [delete]
func (h *TreatmentHandler) DeleteTreatment(c *fiber.Ctx) error {
	id := c.Params("id")
	result := h.treatmentService.DeleteTreatment(id)
	return c.JSON(result)
}

// GetTags godoc
// @Summary List treatment tags
// @Tags Treatments
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/treatment-tags [get]
func (h *TreatmentHandler) GetTags(c *fiber.Ctx) error {
	result := h.treatmentService.GetTags()
	return c.JSON(result)
}

// CreateTag godoc
// @Summary Create treatment tag
// @Tags Treatments
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param tag body map[string]string true "Tag data"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/treatment-tags [post]
func (h *TreatmentHandler) CreateTag(c *fiber.Ctx) error {
	var input map[string]string
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	name := input["name"]
	if name == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Tag name is required",
		})
	}

	result := h.treatmentService.CreateTag(name)
	return c.JSON(result)
}

// UpdateTag godoc
// @Summary Update treatment tag
// @Tags Treatments
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Tag ID"
// @Param tag body map[string]string true "Tag data"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/treatment-tags/{id} [put]
func (h *TreatmentHandler) UpdateTag(c *fiber.Ctx) error {
	id := c.Params("id")
	var input map[string]string
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	name := input["name"]
	if name == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Tag name is required",
		})
	}

	result := h.treatmentService.UpdateTag(id, name)
	return c.JSON(result)
}

// DeleteTag godoc
// @Summary Delete treatment tag
// @Tags Treatments
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Tag ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/treatment-tags/{id} [delete]
func (h *TreatmentHandler) DeleteTag(c *fiber.Ctx) error {
	id := c.Params("id")
	result := h.treatmentService.DeleteTag(id)
	return c.JSON(result)
}
