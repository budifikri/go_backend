package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/middleware"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type SupplierHandler struct {
	supplierService *services.SupplierService
}

func NewSupplierHandler(supplierService *services.SupplierService) *SupplierHandler {
	return &SupplierHandler{supplierService: supplierService}
}

// CreateSupplier godoc
// @Summary Create supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body request.CreateSupplierRequest true "Supplier payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/suppliers [post]
func (h *SupplierHandler) CreateSupplier(c *fiber.Ctx) error {
	var req request.CreateSupplierRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	data := map[string]interface{}{
		"name":           req.Name,
		"contact_person": nil,
		"email":          nil,
		"phone":          nil,
		"address":        nil,
		"city":           nil,
		"tax_id":         nil,
		"payment_terms":  nil,
		"credit_limit":   nil,
		"notes":          nil,
	}
	if req.ContactPerson != "" {
		data["contact_person"] = req.ContactPerson
	}
	if req.Email != "" {
		data["email"] = req.Email
	}
	if req.Phone != "" {
		data["phone"] = req.Phone
	}
	if req.Address != "" {
		data["address"] = req.Address
	}
	if req.City != "" {
		data["city"] = req.City
	}
	if req.TaxID != "" {
		data["tax_id"] = req.TaxID
	}
	if req.PaymentTerms != "" {
		data["payment_terms"] = req.PaymentTerms
	}
	if req.CreditLimit != 0 {
		data["credit_limit"] = req.CreditLimit
	}
	if req.Notes != "" {
		data["notes"] = req.Notes
	}

	result := h.supplierService.CreateSupplier(data, user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

// GetSuppliers godoc
// @Summary List suppliers
// @Tags Suppliers
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param is_active query bool false "Filter by active"
// @Param payment_terms query string false "Payment terms"
// @Param search query string false "Search"
// @Param min_credit_limit query int false "Min credit limit"
// @Param max_credit_limit query int false "Max credit limit"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/suppliers [get]
func (h *SupplierHandler) GetSuppliers(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	filters := map[string]interface{}{}
	if v := c.Query("is_active"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			filters["is_active"] = b
		}
	} else if v := c.Query("status"); v != "" {
		// Backward compatibility: map status=active|inactive to is_active.
		if v == "active" {
			filters["is_active"] = true
		} else if v == "inactive" {
			filters["is_active"] = false
		}
	}
	if v := c.Query("payment_terms"); v != "" {
		filters["payment_terms"] = v
	}
	if v := c.Query("search"); v != "" {
		filters["search"] = v
	}
	if v := c.Query("min_credit_limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filters["min_credit_limit"] = n
		}
	}
	if v := c.Query("max_credit_limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filters["max_credit_limit"] = n
		}
	}

	result := h.supplierService.GetSuppliers(filters, limit, offset, user.CompanyID)
	return c.JSON(result)
}

// GetSupplier godoc
// @Summary Get supplier
// @Tags Suppliers
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Supplier ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/suppliers/{id} [get]
func (h *SupplierHandler) GetSupplier(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	result := h.supplierService.GetSupplierByID(c.Params("id"), user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// UpdateSupplier godoc
// @Summary Update supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Supplier ID"
// @Param body body request.UpdateSupplierRequest true "Update payload"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/suppliers/{id} [put]
func (h *SupplierHandler) UpdateSupplier(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	var req request.UpdateSupplierRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.ContactPerson != nil {
		updates["contact_person"] = *req.ContactPerson
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.City != nil {
		updates["city"] = *req.City
	}
	if req.TaxID != nil {
		updates["tax_id"] = *req.TaxID
	}
	if req.PaymentTerms != nil {
		updates["payment_terms"] = *req.PaymentTerms
	}
	if req.CreditLimit != nil {
		updates["credit_limit"] = *req.CreditLimit
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
		if *req.IsActive {
			updates["status"] = "active"
		} else {
			updates["status"] = "inactive"
		}
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	result := h.supplierService.UpdateSupplier(c.Params("id"), user.CompanyID, updates)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

// DeleteSupplier godoc
// @Summary Delete supplier
// @Tags Suppliers
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Supplier ID"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/suppliers/{id} [delete]
func (h *SupplierHandler) DeleteSupplier(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	result := h.supplierService.DeleteSupplier(c.Params("id"), user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

// SearchSuppliers godoc
// @Summary Search suppliers
// @Tags Suppliers
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param term path string true "Search term"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/suppliers/search/{term} [get]
func (h *SupplierHandler) SearchSuppliers(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	filters := map[string]interface{}{"search": c.Params("term")}
	result := h.supplierService.GetSuppliers(filters, limit, offset, user.CompanyID)
	return c.JSON(result)
}

// GetSuppliersByStatus godoc
// @Summary List suppliers by active (deprecated)
// @Tags Suppliers
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param status path string true "Status (active|inactive)"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/suppliers/status/{status} [get]
func (h *SupplierHandler) GetSuppliersByStatus(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	filters := map[string]interface{}{}
	s := c.Params("status")
	if s == "active" {
		filters["is_active"] = true
	} else if s == "inactive" {
		filters["is_active"] = false
	}
	result := h.supplierService.GetSuppliers(filters, limit, offset, user.CompanyID)
	return c.JSON(result)
}

// GetSuppliersByPaymentTerms godoc
// @Summary List suppliers by payment terms
// @Tags Suppliers
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param terms path string true "Payment terms"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/suppliers/payment-terms/{terms} [get]
func (h *SupplierHandler) GetSuppliersByPaymentTerms(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	filters := map[string]interface{}{"payment_terms": c.Params("terms")}
	result := h.supplierService.GetSuppliers(filters, limit, offset, user.CompanyID)
	return c.JSON(result)
}
