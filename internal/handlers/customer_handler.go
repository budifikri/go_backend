package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/middleware"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type CustomerHandler struct {
	customerService *services.CustomerService
}

func NewCustomerHandler(customerService *services.CustomerService) *CustomerHandler {
	return &CustomerHandler{customerService: customerService}
}

// CreateCustomer godoc
// @Summary Create customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body request.CreateCustomerRequest true "Customer payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/customers [post]
func (h *CustomerHandler) CreateCustomer(c *fiber.Ctx) error {
	var req request.CreateCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	data := map[string]interface{}{
		"name":                req.Name,
		"email":               nil,
		"phone":               nil,
		"address":             nil,
		"city":                nil,
		"tier":                nil,
		"credit_limit":        req.CreditLimit,
		"bank_name":           nil,
		"bank_account_number": nil,
		"bank_account_name":   nil,
		"bank_branch":         nil,
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
	if req.Tier != "" {
		data["tier"] = req.Tier
	}
	if req.BankName != "" {
		data["bank_name"] = req.BankName
	}
	if req.BankAccountNumber != "" {
		data["bank_account_number"] = req.BankAccountNumber
	}
	if req.BankAccountName != "" {
		data["bank_account_name"] = req.BankAccountName
	}
	if req.BankBranch != "" {
		data["bank_branch"] = req.BankBranch
	}

	result := h.customerService.CreateCustomer(data, user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

// GetCustomers godoc
// @Summary List customers
// @Tags Customers
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param tier query string false "Tier"
// @Param status query string false "Status"
// @Param search query string false "Search"
// @Param min_loyalty_points query int false "Min points"
// @Param max_loyalty_points query int false "Max points"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/customers [get]
func (h *CustomerHandler) GetCustomers(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	filters := map[string]interface{}{}
	if v := c.Query("tier"); v != "" {
		filters["tier"] = v
	}
	if v := c.Query("status"); v != "" {
		filters["status"] = v
	}
	if v := c.Query("search"); v != "" {
		filters["search"] = v
	}
	if v := c.Query("min_loyalty_points"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filters["min_loyalty_points"] = n
		}
	}
	if v := c.Query("max_loyalty_points"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filters["max_loyalty_points"] = n
		}
	}

	result := h.customerService.GetCustomers(filters, limit, offset, user.CompanyID)
	return c.JSON(result)
}

// GetCustomer godoc
// @Summary Get customer
// @Tags Customers
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Customer ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/customers/{id} [get]
func (h *CustomerHandler) GetCustomer(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	result := h.customerService.GetCustomerByID(c.Params("id"), user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// UpdateCustomer godoc
// @Summary Update customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Customer ID"
// @Param body body request.UpdateCustomerRequest true "Update payload"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/customers/{id} [put]
func (h *CustomerHandler) UpdateCustomer(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	var req request.UpdateCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
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
	if req.Tier != nil {
		updates["tier"] = *req.Tier
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.CreditLimit != nil {
		updates["credit_limit"] = *req.CreditLimit
	}
	if req.BankName != nil {
		updates["bank_name"] = *req.BankName
	}
	if req.BankAccountNumber != nil {
		updates["bank_account_number"] = *req.BankAccountNumber
	}
	if req.BankAccountName != nil {
		updates["bank_account_name"] = *req.BankAccountName
	}
	if req.BankBranch != nil {
		updates["bank_branch"] = *req.BankBranch
	}

	result := h.customerService.UpdateCustomer(c.Params("id"), user.CompanyID, updates)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// DeleteCustomer godoc
// @Summary Delete customer
// @Tags Customers
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Customer ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/customers/{id} [delete]
func (h *CustomerHandler) DeleteCustomer(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	result := h.customerService.DeleteCustomer(c.Params("id"), user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// SearchCustomers godoc
// @Summary Search customers
// @Tags Customers
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param term path string true "Search term"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/customers/search/{term} [get]
func (h *CustomerHandler) SearchCustomers(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	filters := map[string]interface{}{"search": c.Params("term")}
	result := h.customerService.GetCustomers(filters, limit, offset, user.CompanyID)
	return c.JSON(result)
}

// GetCustomersByTier godoc
// @Summary List customers by tier
// @Tags Customers
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param tier path string true "Tier"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/customers/tier/{tier} [get]
func (h *CustomerHandler) GetCustomersByTier(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	filters := map[string]interface{}{"tier": c.Params("tier")}
	result := h.customerService.GetCustomers(filters, limit, offset, user.CompanyID)
	return c.JSON(result)
}

// GetCustomersByStatus godoc
// @Summary List customers by status
// @Tags Customers
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param status path string true "Status"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/customers/status/{status} [get]
func (h *CustomerHandler) GetCustomersByStatus(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	filters := map[string]interface{}{"status": c.Params("status")}
	result := h.customerService.GetCustomers(filters, limit, offset, user.CompanyID)
	return c.JSON(result)
}
