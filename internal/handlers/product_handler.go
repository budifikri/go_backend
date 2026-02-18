package handlers

import (
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
// @Param status query string false "Filter by status"
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

	if status := c.Query("status"); status != "" {
		filters["status"] = status
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
// @Success 200 {object} response.ApiResponse
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
// @Success 200 {object} response.ApiResponse
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
// @Success 200 {object} response.ApiResponse
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
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/categories [get]
func (h *ProductHandler) GetCategories(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	var companyID *string
	if user != nil && user.CompanyID != "" {
		companyID = &user.CompanyID
	}

	result := h.productService.GetCategories(companyID)
	return c.JSON(result)
}

// GetUnits godoc
// @Summary List units
// @Tags Units
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/units [get]
func (h *ProductHandler) GetUnits(c *fiber.Ctx) error {
	result := h.productService.GetUnits()
	return c.JSON(result)
}
