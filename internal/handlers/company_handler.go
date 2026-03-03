package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/middleware"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type CompanyHandler struct {
	companyService *services.CompanyService
}

func NewCompanyHandler(companyService *services.CompanyService) *CompanyHandler {
	return &CompanyHandler{companyService: companyService}
}

// GetCompanies godoc
// @Summary List companies (admin only)
// @Tags Companies
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 403 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/companies [get]
func (h *CompanyHandler) GetCompanies(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	result := h.companyService.GetCompanies(limit, offset)
	return c.JSON(result)
}

// GetCurrentCompany godoc
// @Summary Get current user's company
// @Tags Companies
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/companies/current [get]
func (h *CompanyHandler) GetCurrentCompany(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil || user.CompanyID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("User not authenticated"))
	}
	result := h.companyService.GetCompanyByUserCompanyID(user.CompanyID)
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// GetCompany godoc
// @Summary Get company
// @Tags Companies
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Company ID"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/companies/{id} [get]
func (h *CompanyHandler) GetCompany(c *fiber.Ctx) error {
	result := h.companyService.GetCompanyByID(c.Params("id"))
	if !result.Success {
		return c.Status(fiber.StatusNotFound).JSON(result)
	}
	return c.JSON(result)
}

// CreateCompany godoc
// @Summary Create company (admin only)
// @Tags Companies
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body request.CreateCompanyRequest true "Company payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 403 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/companies [post]
func (h *CompanyHandler) CreateCompany(c *fiber.Ctx) error {
	var req request.CreateCompanyRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.CreateCompanyRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	}

	result := h.companyService.CreateCompany(services.CreateCompanyInput{
		Code:            req.Code,
		Nama:            req.Nama,
		Email:           req.Email,
		Address:         req.Address,
		Telp:            req.Telp,
		Website:         req.Website,
		TaxID:           req.TaxID,
		BusinessLicense: req.BusinessLicense,
		IsActive:        req.IsActive,
	})
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	result.Message = "Company created successfully"
	return c.Status(fiber.StatusCreated).JSON(result)
}

// UpdateCompany godoc
// @Summary Update company (admin only)
// @Tags Companies
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Company ID"
// @Param body body request.UpdateCompanyRequest true "Update payload"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 403 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/companies/{id} [put]
func (h *CompanyHandler) UpdateCompany(c *fiber.Ctx) error {
	var req request.UpdateCompanyRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.UpdateCompanyRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	}

	result := h.companyService.UpdateCompany(c.Params("id"), services.UpdateCompanyInput{
		Nama:            req.Nama,
		Email:           req.Email,
		Address:         req.Address,
		Telp:            req.Telp,
		Logo:            req.Logo,
		Website:         req.Website,
		TaxID:           req.TaxID,
		BusinessLicense: req.BusinessLicense,
		IsActive:        req.IsActive,
	})
	if !result.Success {
		if result.Error == "Company not found" {
			return c.Status(fiber.StatusNotFound).JSON(result)
		}
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	result.Message = "Company updated successfully"
	return c.JSON(result)
}

// DeleteCompany godoc
// @Summary Delete company (admin only)
// @Tags Companies
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Company ID"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 403 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/companies/{id} [delete]
func (h *CompanyHandler) DeleteCompany(c *fiber.Ctx) error {
	result := h.companyService.DeleteCompany(c.Params("id"))
	if !result.Success {
		if result.Error == "Company not found" {
			return c.Status(fiber.StatusNotFound).JSON(result)
		}
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

// UploadCompanyLogo godoc
// @Summary Upload company logo (admin only)
// @Tags Companies
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Company ID"
// @Param body body request.UploadCompanyLogoRequest true "Logo payload"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 403 {object} response.ApiResponse
// @Failure 404 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/companies/{id}/logo [post]
func (h *CompanyHandler) UploadCompanyLogo(c *fiber.Ctx) error {
	var req request.UploadCompanyLogoRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.UploadCompanyLogoRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	}

	result := h.companyService.UploadCompanyLogo(c.Params("id"), req.Logo)
	if !result.Success {
		if result.Error == "Company not found" {
			return c.Status(fiber.StatusNotFound).JSON(result)
		}
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	result.Message = "Logo uploaded successfully"
	return c.JSON(result)
}
