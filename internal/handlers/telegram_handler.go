package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/utils"
)

type TelegramHandler struct {
	telegramService *services.TelegramService
}

func NewTelegramHandler(telegramService *services.TelegramService) *TelegramHandler {
	return &TelegramHandler{telegramService: telegramService}
}

func GetUserFromContextTelegram(c *fiber.Ctx) *UserPayload {
	user, ok := c.Locals("user").(*utils.JWTPayload)
	if !ok {
		return nil
	}

	companyID, err := uuid.Parse(user.CompanyID)
	if err != nil {
		return nil
	}

	return &UserPayload{
		UserID:    user.UserID,
		CompanyID: companyID,
		Role:      user.Role,
	}
}

func (h *TelegramHandler) GetConfig(c *fiber.Ctx) error {
	user := GetUserFromContextTelegram(c)
	if user == nil {
		return c.Status(401).JSON(fiber.Map{"success": false, "error": "Unauthorized"})
	}

	config, err := h.telegramService.GetConfigByCompany(user.CompanyID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	if config == nil {
		return c.JSON(fiber.Map{"success": true, "data": nil})
	}

	return c.JSON(fiber.Map{"success": true, "data": config})
}

func (h *TelegramHandler) SaveConfig(c *fiber.Ctx) error {
	user := GetUserFromContextTelegram(c)
	if user == nil {
		return c.Status(401).JSON(fiber.Map{"success": false, "error": "Unauthorized"})
	}

	var input request.CreateTelegramRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	result := h.telegramService.SaveConfig(user.CompanyID, input)
	if !result.Success {
		return c.Status(400).JSON(result)
	}
	return c.JSON(result)
}

func (h *TelegramHandler) TestConnection(c *fiber.Ctx) error {
	var input request.TestTelegramRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	if input.TelegramID == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Telegram ID is required"})
	}

	if input.Type == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Type is required"})
	}

	validTypes := map[string]bool{
		"penjualan":      true,
		"pembelian":      true,
		"stock_opname":   true,
		"closing_drawer": true,
	}
	if !validTypes[input.Type] {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Invalid type"})
	}

	result := h.telegramService.TestConnection(input)
	if !result.Success {
		return c.Status(400).JSON(result)
	}
	return c.JSON(result)
}
