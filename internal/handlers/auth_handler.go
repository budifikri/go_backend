package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/middleware"
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

// Login godoc
// @Summary Login
// @Description Authenticate user and return token/session
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body request.LoginRequest true "Login payload"
// @Success 200 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Router /api/auth/login [post]
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

// Register godoc
// @Summary Register
// @Description Register a new user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body request.RegisterRequest true "Register payload"
// @Success 201 {object} response.ApiResponse
// @Failure 400 {object} response.ApiResponse
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req request.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	result := h.authService.Register(struct {
		Username    string `json:"username"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		FullName    string `json:"full_name"`
		Role        string `json:"role"`
		CompanyName string `json:"company_name"`
	}{
		Username:    req.Username,
		Email:       req.Email,
		Password:    req.Password,
		FullName:    req.FullName,
		Role:        req.Role,
		CompanyName: req.CompanyName,
	})

	if !result.Success {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// GetMe godoc
// @Summary Get current user
// @Description Get authenticated user's profile
// @Tags Authentication
// @Produce json
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/auth/me [get]
func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	payload := middleware.GetUserFromContext(c)
	if payload == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	result := h.authService.GetMe(payload)
	return c.Status(fiber.StatusOK).JSON(result)
}

// Logout godoc
// @Summary Logout
// @Description Invalidate session/token
// @Tags Authentication
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Authorization header required"))
	}

	token := authHeader[7:] // Remove "Bearer " prefix
	result := h.authService.Logout(token)

	return c.Status(fiber.StatusOK).JSON(result)
}
