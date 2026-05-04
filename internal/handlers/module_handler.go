package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/middleware"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type ModuleHandler struct {
	moduleService *services.ModuleService
}

func NewModuleHandler(moduleService *services.ModuleService) *ModuleHandler {
	return &ModuleHandler{moduleService: moduleService}
}

func (h *ModuleHandler) GetBusinessTypes(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	var isActive *bool
	if value := c.Query("is_active"); value != "" {
		parsed := value == "true"
		isActive = &parsed
	}
	result := h.moduleService.GetBusinessTypes(c.Query("search"), isActive, limit, offset)
	return c.JSON(result)
}

func (h *ModuleHandler) CreateBusinessType(c *fiber.Ctx) error {
	var req request.CreateBusinessTypeRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.CreateBusinessTypeRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}
	result := h.moduleService.CreateBusinessType(services.CreateBusinessTypeInput{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
		IsDefault:   req.IsDefault,
		IsSystem:    req.IsSystem,
		SortOrder:   req.SortOrder,
	})
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *ModuleHandler) UpdateBusinessType(c *fiber.Ctx) error {
	var req request.UpdateBusinessTypeRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.UpdateBusinessTypeRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}
	result := h.moduleService.UpdateBusinessType(c.Params("id"), services.UpdateBusinessTypeInput{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
		IsDefault:   req.IsDefault,
		IsSystem:    req.IsSystem,
		SortOrder:   req.SortOrder,
	})
	if !result.Success {
		status := fiber.StatusBadRequest
		if result.Error == "Business type not found" {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(result)
	}
	return c.JSON(result)
}

func (h *ModuleHandler) GetModulePackages(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	var isActive *bool
	if value := c.Query("is_active"); value != "" {
		parsed := value == "true"
		isActive = &parsed
	}
	result := h.moduleService.GetModulePackages(c.Query("business_type"), c.Query("search"), isActive, limit, offset)
	return c.JSON(result)
}

func (h *ModuleHandler) CreateModulePackage(c *fiber.Ctx) error {
	var req request.CreateModulePackageRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.CreateModulePackageRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}
	result := h.moduleService.CreateModulePackage(services.CreateModulePackageInput{
		BusinessType: req.BusinessType,
		Code:         req.Code,
		Name:         req.Name,
		Description:  req.Description,
		IsActive:     req.IsActive,
		IsDefault:    req.IsDefault,
		IsSystem:     req.IsSystem,
		SortOrder:    req.SortOrder,
	})
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *ModuleHandler) UpdateModulePackage(c *fiber.Ctx) error {
	var req request.UpdateModulePackageRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.UpdateModulePackageRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}
	result := h.moduleService.UpdateModulePackage(c.Params("id"), services.UpdateModulePackageInput{
		BusinessType: req.BusinessType,
		Name:         req.Name,
		Description:  req.Description,
		IsActive:     req.IsActive,
		IsDefault:    req.IsDefault,
		IsSystem:     req.IsSystem,
		SortOrder:    req.SortOrder,
	})
	if !result.Success {
		status := fiber.StatusBadRequest
		if result.Error == "Module package not found" {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(result)
	}
	return c.JSON(result)
}

func (h *ModuleHandler) GetMyModules(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil || user.CompanyID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	result := h.moduleService.GetCompanyModules(user.CompanyID, false)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

func (h *ModuleHandler) GetCompanyModules(c *fiber.Ctx) error {
	result := h.moduleService.GetCompanyModules(c.Params("id"), true)
	if !result.Success {
		status := fiber.StatusBadRequest
		if result.Error == "Company not found" {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(result)
	}
	return c.JSON(result)
}

func (h *ModuleHandler) ToggleCompanyModule(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}
	var req request.ToggleCompanyModuleRequest
	if v := c.Locals(middleware.ContextKeyValidatedBody); v != nil {
		if parsed, ok := v.(*request.ToggleCompanyModuleRequest); ok {
			req = *parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
		}
	} else if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}
	result := h.moduleService.ToggleCompanyModule(c.Params("id"), c.Params("code"), user.UserID, req.IsActive)
	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}
