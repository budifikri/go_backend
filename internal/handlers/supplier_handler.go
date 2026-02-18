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

func (h *SupplierHandler) GetSuppliers(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	filters := map[string]interface{}{}
	if v := c.Query("status"); v != "" {
		filters["status"] = v
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
	if req.Status != nil {
		updates["status"] = *req.Status
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

func (h *SupplierHandler) DeleteSupplier(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	result := h.supplierService.DeactivateSupplier(c.Params("id"), user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

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

func (h *SupplierHandler) GetSuppliersByStatus(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	filters := map[string]interface{}{"status": c.Params("status")}
	result := h.supplierService.GetSuppliers(filters, limit, offset, user.CompanyID)
	return c.JSON(result)
}

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
