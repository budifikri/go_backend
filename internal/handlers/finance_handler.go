package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/middleware"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type FinanceHandler struct {
	financeService *services.FinanceService
}

func NewFinanceHandler(financeService *services.FinanceService) *FinanceHandler {
	return &FinanceHandler{financeService: financeService}
}

// GetIncomingInvoices godoc
// @Summary List incoming invoices
// @Tags Finance
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param status query string false "Status"
// @Param supplier_id query string false "Supplier ID"
// @Param from_date query string false "From date"
// @Param to_date query string false "To date"
// @Param search query string false "Search"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/invoices/incoming [get]
func (h *FinanceHandler) GetIncomingInvoices(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	filters := map[string]string{}
	if v := c.Query("status"); v != "" {
		filters["status"] = v
	}
	if v := c.Query("supplier_id"); v != "" {
		filters["supplier_id"] = v
	}
	if v := c.Query("from_date"); v != "" {
		filters["from_date"] = v
	}
	if v := c.Query("to_date"); v != "" {
		filters["to_date"] = v
	}
	if v := c.Query("search"); v != "" {
		filters["search"] = v
	}

	limit := c.QueryInt("limit", 50)
	if limit <= 0 {
		limit = 50
	}
	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}

	result := h.financeService.GetIncomingInvoices(filters, limit, offset, user.CompanyID)
	return c.JSON(result)
}

// GetIncomingInvoice godoc
// @Summary Get incoming invoice
// @Tags Finance
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Invoice ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/invoices/incoming/{id} [get]
func (h *FinanceHandler) GetIncomingInvoice(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	result := h.financeService.GetIncomingInvoiceByID(c.Params("id"), user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   result.Error,
			"message": "Failed to get incoming invoice",
		})
	}
	return c.JSON(result)
}

// CreateIncomingInvoice godoc
// @Summary Create incoming invoice
// @Tags Finance
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body request.CreateIncomingInvoiceRequest true "Invoice payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/invoices/incoming [post]
func (h *FinanceHandler) CreateIncomingInvoice(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	var req request.CreateIncomingInvoiceRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.CreateIncomingInvoiceRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	}

	items := make([]services.FinanceInvoiceItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, services.FinanceInvoiceItemInput{
			ProductID:    it.ProductID,
			Description:  it.Description,
			Quantity:     it.Quantity,
			UnitPrice:    it.UnitPrice,
			DiscountRate: it.DiscountRate,
			TaxRate:      it.TaxRate,
		})
	}

	result := h.financeService.CreateIncomingInvoice(services.CreateIncomingInvoiceInput{
		SupplierID:     req.SupplierID,
		InvoiceDate:    req.InvoiceDate,
		DueDate:        req.DueDate,
		Notes:          req.Notes,
		AttachmentPath: req.AttachmentPath,
		Items:          items,
	}, user.CompanyID, user.UserID)

	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   result.Error,
			"message": "Failed to create incoming invoice",
		})
	}
	result.Message = "Incoming invoice created successfully"
	return c.Status(fiber.StatusCreated).JSON(result)
}

// UpdateIncomingInvoice godoc
// @Summary Update incoming invoice
// @Tags Finance
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Invoice ID"
// @Param body body request.UpdateIncomingInvoiceRequest true "Update payload"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/invoices/incoming/{id} [put]
func (h *FinanceHandler) UpdateIncomingInvoice(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	var req request.UpdateIncomingInvoiceRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.UpdateIncomingInvoiceRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	}

	var items *[]services.FinanceInvoiceItemInput
	if req.Items != nil {
		converted := make([]services.FinanceInvoiceItemInput, 0, len(*req.Items))
		for _, it := range *req.Items {
			converted = append(converted, services.FinanceInvoiceItemInput{
				ProductID:    it.ProductID,
				Description:  it.Description,
				Quantity:     it.Quantity,
				UnitPrice:    it.UnitPrice,
				DiscountRate: it.DiscountRate,
				TaxRate:      it.TaxRate,
			})
		}
		items = &converted
	}

	result := h.financeService.UpdateIncomingInvoice(c.Params("id"), services.UpdateIncomingInvoiceInput{
		InvoiceDate: req.InvoiceDate,
		DueDate:     req.DueDate,
		Notes:       req.Notes,
		Status:      req.Status,
		Items:       items,
	}, user.CompanyID)

	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   result.Error,
			"message": "Failed to update incoming invoice",
		})
	}
	result.Message = "Incoming invoice updated successfully"
	return c.JSON(result)
}

// SendIncomingInvoice godoc
// @Summary Send incoming invoice
// @Tags Finance
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Invoice ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/invoices/incoming/{id}/send [post]
func (h *FinanceHandler) SendIncomingInvoice(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	result := h.financeService.SendIncomingInvoice(c.Params("id"), user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   result.Error,
			"message": "Failed to send incoming invoice",
		})
	}
	result.Message = "Invoice sent successfully"
	return c.JSON(result)
}

// CancelIncomingInvoice godoc
// @Summary Cancel incoming invoice
// @Tags Finance
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Invoice ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/invoices/incoming/{id}/cancel [post]
func (h *FinanceHandler) CancelIncomingInvoice(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	result := h.financeService.CancelIncomingInvoice(c.Params("id"), user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   result.Error,
			"message": "Failed to cancel invoice",
		})
	}
	result.Message = "Invoice cancelled successfully"
	return c.JSON(result)
}

// AddIncomingInvoicePayment godoc
// @Summary Add incoming invoice payment
// @Tags Finance
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Invoice ID"
// @Param body body request.CreateInvoicePaymentRequest true "Payment payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/invoices/incoming/{id}/payments [post]
func (h *FinanceHandler) AddIncomingInvoicePayment(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	var req request.CreateInvoicePaymentRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.CreateInvoicePaymentRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	}

	result := h.financeService.AddIncomingInvoicePayment(c.Params("id"), services.CreateInvoicePaymentInput{
		Amount:          req.Amount,
		PaymentMethod:   req.PaymentMethod,
		ReferenceNumber: req.ReferenceNumber,
		Notes:           req.Notes,
	}, user.CompanyID, user.UserID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   result.Error,
			"message": "Failed to add payment",
		})
	}
	result.Message = "Payment recorded successfully"
	return c.Status(fiber.StatusCreated).JSON(result)
}

// GetOutgoingInvoices godoc
// @Summary List outgoing invoices
// @Tags Finance
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param status query string false "Status"
// @Param customer_id query string false "Customer ID"
// @Param from_date query string false "From date"
// @Param to_date query string false "To date"
// @Param search query string false "Search"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/invoices/outgoing [get]
func (h *FinanceHandler) GetOutgoingInvoices(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	filters := map[string]string{}
	if v := c.Query("status"); v != "" {
		filters["status"] = v
	}
	if v := c.Query("customer_id"); v != "" {
		filters["customer_id"] = v
	}
	if v := c.Query("from_date"); v != "" {
		filters["from_date"] = v
	}
	if v := c.Query("to_date"); v != "" {
		filters["to_date"] = v
	}
	if v := c.Query("search"); v != "" {
		filters["search"] = v
	}

	limit := c.QueryInt("limit", 50)
	if limit <= 0 {
		limit = 50
	}
	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}

	result := h.financeService.GetOutgoingInvoices(filters, limit, offset, user.CompanyID)
	return c.JSON(result)
}

// GetOutgoingInvoice godoc
// @Summary Get outgoing invoice
// @Tags Finance
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Invoice ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/invoices/outgoing/{id} [get]
func (h *FinanceHandler) GetOutgoingInvoice(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	result := h.financeService.GetOutgoingInvoiceByID(c.Params("id"), user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   result.Error,
			"message": "Failed to get outgoing invoice",
		})
	}
	return c.JSON(result)
}

// CreateOutgoingInvoice godoc
// @Summary Create outgoing invoice
// @Tags Finance
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body request.CreateOutgoingInvoiceRequest true "Invoice payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/invoices/outgoing [post]
func (h *FinanceHandler) CreateOutgoingInvoice(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	var req request.CreateOutgoingInvoiceRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.CreateOutgoingInvoiceRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	}

	items := make([]services.FinanceInvoiceItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, services.FinanceInvoiceItemInput{
			ProductID:    it.ProductID,
			Description:  it.Description,
			Quantity:     it.Quantity,
			UnitPrice:    it.UnitPrice,
			DiscountRate: it.DiscountRate,
			TaxRate:      it.TaxRate,
		})
	}

	result := h.financeService.CreateOutgoingInvoice(services.CreateOutgoingInvoiceInput{
		CustomerID:     req.CustomerID,
		InvoiceDate:    req.InvoiceDate,
		DueDate:        req.DueDate,
		Notes:          req.Notes,
		AttachmentPath: req.AttachmentPath,
		Items:          items,
	}, user.CompanyID, user.UserID)

	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   result.Error,
			"message": "Failed to create outgoing invoice",
		})
	}
	result.Message = "Outgoing invoice created successfully"
	return c.Status(fiber.StatusCreated).JSON(result)
}

// UpdateOutgoingInvoice godoc
// @Summary Update outgoing invoice
// @Tags Finance
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Invoice ID"
// @Param body body request.UpdateOutgoingInvoiceRequest true "Update payload"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/invoices/outgoing/{id} [put]
func (h *FinanceHandler) UpdateOutgoingInvoice(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	var req request.UpdateOutgoingInvoiceRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.UpdateOutgoingInvoiceRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	}

	var items *[]services.FinanceInvoiceItemInput
	if req.Items != nil {
		converted := make([]services.FinanceInvoiceItemInput, 0, len(*req.Items))
		for _, it := range *req.Items {
			converted = append(converted, services.FinanceInvoiceItemInput{
				ProductID:    it.ProductID,
				Description:  it.Description,
				Quantity:     it.Quantity,
				UnitPrice:    it.UnitPrice,
				DiscountRate: it.DiscountRate,
				TaxRate:      it.TaxRate,
			})
		}
		items = &converted
	}

	result := h.financeService.UpdateOutgoingInvoice(c.Params("id"), services.UpdateOutgoingInvoiceInput{
		InvoiceDate: req.InvoiceDate,
		DueDate:     req.DueDate,
		Notes:       req.Notes,
		Status:      req.Status,
		Items:       items,
	}, user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   result.Error,
			"message": "Failed to update outgoing invoice",
		})
	}
	result.Message = "Outgoing invoice updated successfully"
	return c.JSON(result)
}

// SendOutgoingInvoice godoc
// @Summary Send outgoing invoice
// @Tags Finance
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Invoice ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/invoices/outgoing/{id}/send [post]
func (h *FinanceHandler) SendOutgoingInvoice(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	result := h.financeService.SendOutgoingInvoice(c.Params("id"), user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   result.Error,
			"message": "Failed to send outgoing invoice",
		})
	}
	result.Message = "Invoice sent successfully"
	return c.JSON(result)
}

// CancelOutgoingInvoice godoc
// @Summary Cancel outgoing invoice
// @Tags Finance
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Invoice ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/invoices/outgoing/{id}/cancel [post]
func (h *FinanceHandler) CancelOutgoingInvoice(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	result := h.financeService.CancelOutgoingInvoice(c.Params("id"), user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   result.Error,
			"message": "Failed to cancel invoice",
		})
	}
	result.Message = "Invoice cancelled successfully"
	return c.JSON(result)
}

// AddOutgoingInvoicePayment godoc
// @Summary Add outgoing invoice payment
// @Tags Finance
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Invoice ID"
// @Param body body request.CreateInvoicePaymentRequest true "Payment payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/invoices/outgoing/{id}/payments [post]
func (h *FinanceHandler) AddOutgoingInvoicePayment(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	var req request.CreateInvoicePaymentRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.CreateInvoicePaymentRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	}
	result := h.financeService.AddOutgoingInvoicePayment(c.Params("id"), services.CreateInvoicePaymentInput{
		Amount:          req.Amount,
		PaymentMethod:   req.PaymentMethod,
		ReferenceNumber: req.ReferenceNumber,
		Notes:           req.Notes,
	}, user.CompanyID, user.UserID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   result.Error,
			"message": "Failed to add payment",
		})
	}
	result.Message = "Payment recorded successfully"
	return c.Status(fiber.StatusCreated).JSON(result)
}

// GetInvoiceSummary godoc
// @Summary Get invoice summary
// @Tags Finance
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param type query string false "Type (INCOMING|OUTGOING|ALL)"
// @Param from_date query string false "From date"
// @Param to_date query string false "To date"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/invoices/summary [get]
func (h *FinanceHandler) GetInvoiceSummary(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	filters := map[string]string{}
	if v := c.Query("type"); v != "" && v != "ALL" {
		filters["type"] = v
	}
	if v := c.Query("from_date"); v != "" {
		filters["from_date"] = v
	}
	if v := c.Query("to_date"); v != "" {
		filters["to_date"] = v
	}

	result := h.financeService.GetInvoiceSummary(filters, user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   result.Error,
			"message": "Failed to get invoice summary",
		})
	}
	return c.JSON(result)
}
