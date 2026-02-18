package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

// AuthHandler handles auth endpoints
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login POST /api/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req request.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	result := h.authService.Login(req.Username, req.Password)

	if !result.Success {
		return c.Status(fiber.StatusUnauthorized).JSON(result)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

// Register POST /api/auth/register
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req request.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	result := h.authService.Register(struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
		Role     string `json:"role"`
	}{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Role:     req.Role,
	})

	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// Logout POST /api/auth/logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Authorization header required"))
	}

	token := authHeader[7:] // Remove "Bearer " prefix
	result := h.authService.Logout(token)

	return c.Status(fiber.StatusOK).JSON(result)
}
