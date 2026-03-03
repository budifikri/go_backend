package handlers

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/middleware"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type ProductHandler struct {
	productService *services.ProductService
}

func NewProductHandler(productService *services.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// GetProducts godoc
// @Summary List products
// @Tags Products
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param is_active query bool false "Filter by active"
// @Param include_inactive query bool false "Include inactive (ignore default active-only)"
// @Param category_id query string false "Filter by category"
// @Param search query string false "Search term"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/products [get]
func (h *ProductHandler) GetProducts(c *fiber.Ctx) error {
	filters := make(map[string]interface{})
	includeInactive := c.QueryBool("include_inactive", false)

	if v := c.Query("is_active"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			filters["is_active"] = b
		}
	} else if status := c.Query("status"); status != "" {
		// Backward compatibility: map status=active|inactive to is_active.
		if status == "active" {
			filters["is_active"] = true
		} else if status == "inactive" {
			filters["is_active"] = false
		}
	} else if !includeInactive {
		// Default active
		filters["is_active"] = true
	}
	if categoryID := c.Query("category_id"); categoryID != "" {
		filters["category_id"] = categoryID
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

	result := h.productService.GetProducts(filters, limit, offset)
	return c.JSON(result)
}

// GetProduct godoc
// @Summary Get product
// @Tags Products
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Product ID"
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/products/{id} [get]
func (h *ProductHandler) GetProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	result := h.productService.GetProductByID(id)

	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// CreateProduct godoc
// @Summary Create product
// @Tags Products
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body request.ProductCreateRequest true "Product payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/products [post]
func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var req request.ProductCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	user := middleware.GetUserFromContext(c)
	if user != nil && user.CompanyID != "" && req.CompanyID == "" {
		req.CompanyID = user.CompanyID
	}

	result := h.productService.CreateProduct(services.CreateProductRequest{
		SKU:          req.SKU,
		Barcode:      req.Barcode,
		Name:         req.Name,
		Description:  req.Description,
		CategoryID:   req.CategoryID,
		UnitID:       req.UnitID,
		CostPrice:    req.CostPrice,
		RetailPrice:  req.RetailPrice,
		TaxRate:      req.TaxRate,
		ReorderPoint: req.ReorderPoint,
		CompanyID:    req.CompanyID,
	})

	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

// UpdateProduct godoc
// @Summary Update product
// @Tags Products
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Product ID"
// @Param body body request.ProductUpdateRequest true "Product payload"
// @Success 200 {object} response.PaginatedResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	var req request.ProductUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	log.Printf("[DEBUG] UpdateProduct request: id=%s, UnitID=%s, CategoryID=%s, Name=%s", id, req.UnitID, req.CategoryID, req.Name)

	result := h.productService.UpdateProduct(id, services.CreateProductRequest{
		SKU:          req.SKU,
		Barcode:      req.Barcode,
		Name:         req.Name,
		Description:  req.Description,
		CategoryID:   req.CategoryID,
		UnitID:       req.UnitID,
		CostPrice:    req.CostPrice,
		RetailPrice:  req.RetailPrice,
		TaxRate:      req.TaxRate,
		ReorderPoint: req.ReorderPoint,
		IsActive:     req.IsActive,
	})

	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

// DeleteProduct godoc
// @Summary Delete product
// @Tags Products
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Product ID"
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	result := h.productService.DeleteProduct(id)

	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// GetCategories godoc
// @Summary List categories
// @Tags Categories
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/categories [get]
func (h *ProductHandler) GetCategories(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	var companyID *string
	if user != nil && user.CompanyID != "" {
		companyID = &user.CompanyID
	}

	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)
	includeInactive := c.QueryBool("include_inactive", false)
	var isActive *bool
	if !includeInactive {
		if v := c.Query("is_active"); v != "" {
			if b, err := strconv.ParseBool(v); err == nil {
				isActive = &b
			}
		} else {
			b := true
			isActive = &b
		}
	} else if v := c.Query("is_active"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			isActive = &b
		}
	}

	result := h.productService.GetCategoriesPaged(companyID, isActive, limit, offset)
	return c.JSON(result)
}

// GetCategory godoc
// @Summary Get category details
// @Description Get category by ID
// @Tags Categories
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Category ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/categories/{id} [get]
func (h *ProductHandler) GetCategory(c *fiber.Ctx) error {
	result := h.productService.GetCategoryByID(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// CreateCategory godoc
// @Summary Create category
// @Description Create a new category (admin/manager only)
// @Tags Categories
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body request.CreateCategoryRequest true "Category payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 403 {object} response.ApiResponse
// @Failure 409 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/categories [post]
func (h *ProductHandler) CreateCategory(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	var req request.CreateCategoryRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.CreateCategoryRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	}

	// Handle empty string parent_id => null (TS parity)
	var parentID *string
	if req.ParentID != nil {
		p := *req.ParentID
		if p != "" {
			parentID = &p
		}
	}

	companyID := user.CompanyID
	result := h.productService.CreateCategory(services.CreateCategoryInput{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		ParentID:    parentID,
		CompanyID:   &companyID,
	})
	if !result.Success {
		if result.Error == "Category code already exists" {
			return c.Status(fiber.StatusConflict).JSON(result)
		}
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

// UpdateCategory godoc
// @Summary Update category
// @Description Update category details (admin/manager only)
// @Tags Categories
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Category ID"
// @Param body body request.UpdateCategoryRequest true "Update payload"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 403 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/categories/{id} [put]
func (h *ProductHandler) UpdateCategory(c *fiber.Ctx) error {
	var req request.UpdateCategoryRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.UpdateCategoryRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	}

	result := h.productService.UpdateCategory(c.Params("id"), services.UpdateCategoryInput{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		IsActive:    req.IsActive,
	})
	if !result.Success {
		if result.Error == "Category not found" {
			return c.Status(fiber.StatusNotFound).JSON(result)
		}
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

// DeleteCategory godoc
// @Summary Delete category
// @Description Soft delete category (set is_active to false) (admin/manager only)
// @Tags Categories
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Category ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 403 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/categories/{id} [delete]
func (h *ProductHandler) DeleteCategory(c *fiber.Ctx) error {
	result := h.productService.DeleteCategory(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// GetUnits godoc
// @Summary List units
// @Tags Units
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param search query string false "Search"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/units [get]
func (h *ProductHandler) GetUnits(c *fiber.Ctx) error {
	var search *string
	if q := c.Query("search"); q != "" {
		search = &q
	}
	includeInactive := c.QueryBool("include_inactive", false)
	var isActive *bool
	if !includeInactive {
		if v := c.Query("is_active"); v != "" {
			if b, err := strconv.ParseBool(v); err == nil {
				isActive = &b
			}
		} else {
			b := true
			isActive = &b
		}
	} else if v := c.Query("is_active"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			isActive = &b
		}
	}
	var limit *int
	if q := c.Query("limit"); q != "" {
		v := c.QueryInt("limit")
		limit = &v
	}
	var offset *int
	if q := c.Query("offset"); q != "" {
		v := c.QueryInt("offset")
		offset = &v
	}

	resp, pagination := h.productService.GetUnitsWithQuery(search, isActive, limit, offset)
	if !resp.Success {
		return c.Status(fiber.StatusBadRequest).JSON(resp)
	}
	if pagination != nil {
		return c.JSON(fiber.Map{"success": true, "data": resp.Data, "pagination": pagination})
	}
	return c.JSON(resp)
}

// GetUnit godoc
// @Summary Get unit details
// @Description Get unit of measure by ID
// @Tags Units
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Unit ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/units/{id} [get]
func (h *ProductHandler) GetUnit(c *fiber.Ctx) error {
	result := h.productService.GetUnitByID(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// CreateUnit godoc
// @Summary Create unit
// @Description Create a new unit of measure
// @Tags Units
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body request.CreateUnitRequest true "Unit payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 409 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/units [post]
func (h *ProductHandler) CreateUnit(c *fiber.Ctx) error {
	var req request.CreateUnitRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.CreateUnitRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	}

	result := h.productService.CreateUnit(services.CreateUnitInput{Code: req.Code, Name: req.Name, Description: req.Description})
	if !result.Success {
		if result.Error == "Unit code already exists" {
			return c.Status(fiber.StatusConflict).JSON(result)
		}
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

// UpdateUnit godoc
// @Summary Update unit
// @Description Update unit of measure
// @Tags Units
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Unit ID"
// @Param body body request.UpdateUnitRequest true "Update payload"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/units/{id} [put]
func (h *ProductHandler) UpdateUnit(c *fiber.Ctx) error {
	var req request.UpdateUnitRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.UpdateUnitRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	}

	result := h.productService.UpdateUnit(c.Params("id"), services.UpdateUnitInput{Code: req.Code, Name: req.Name, Description: req.Description, IsActive: req.IsActive})
	if !result.Success {
		if result.Error == "Unit not found" {
			return c.Status(fiber.StatusNotFound).JSON(result)
		}
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

// DeleteUnit godoc
// @Summary Delete unit
// @Description Delete unit of measure
// @Tags Units
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Unit ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/units/{id} [delete]
func (h *ProductHandler) DeleteUnit(c *fiber.Ctx) error {
	result := h.productService.DeleteUnit(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}
