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

func (h *ProductHandler) GetProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	result := h.productService.GetProductByID(id)

	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

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

func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	result := h.productService.DeleteProduct(id)

	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

func (h *ProductHandler) GetCategories(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	var companyID *string
	if user != nil && user.CompanyID != "" {
		companyID = &user.CompanyID
	}

	result := h.productService.GetCategories(companyID)
	return c.JSON(result)
}

func (h *ProductHandler) GetUnits(c *fiber.Ctx) error {
	result := h.productService.GetUnits()
	return c.JSON(result)
}
