package middleware

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/types/response"
	"github.com/pos-retail/go_backend/internal/utils"
)

// ContextKey for user context
const ContextKeyUser = "user"

// AuthMiddleware validates JWT tokens
type AuthMiddleware struct {
	jwtUtil *utils.JWTUtil
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtUtil *utils.JWTUtil) *AuthMiddleware {
	return &AuthMiddleware{
		jwtUtil: jwtUtil,
	}
}

// Handler returns the Fiber middleware handler
func (m *AuthMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var token string

		authHeader := c.Get("Authorization")
		fmt.Printf("[DEBUG Auth] Path: %s, AuthHeader: %s\n", c.Path(), authHeader)
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				token = parts[1]
			}
		}

		if token == "" {
			token = c.Query("token")
			if token != "" {
				preview := token
				if len(preview) > 20 {
					preview = preview[:20] + "..."
				}
				fmt.Printf("[DEBUG Auth] Using token from query: %s\n", preview)
			}
		}

		if token == "" {
			fmt.Println("[DEBUG Auth] No token found!")
			return c.Status(fiber.StatusUnauthorized).JSON(response.ApiResponse{
				Success: false,
				Error:   "Authorization token required",
			})
		}

		payload, err := m.jwtUtil.VerifyToken(token)
		if err != nil {
			fmt.Printf("[DEBUG Auth] Token verification failed: %v\n", err)
			return c.Status(fiber.StatusUnauthorized).JSON(response.ApiResponse{
				Success: false,
				Error:   "Invalid or expired token",
			})
		}

		fmt.Printf("[DEBUG Auth] Token verified for user: %s\n", payload.Username)
		// Store user info in context
		c.Locals(ContextKeyUser, payload)

		return c.Next()
	}
}

// GetUserFromContext retrieves user info from context
func GetUserFromContext(c *fiber.Ctx) *utils.JWTPayload {
	if user, ok := c.Locals(ContextKeyUser).(*utils.JWTPayload); ok {
		return user
	}
	return nil
}

// RequireAuth requires authentication
func RequireAuth(jwtUtil *utils.JWTUtil) fiber.Handler {
	middleware := NewAuthMiddleware(jwtUtil)
	return middleware.Handler()
}

// RoleMiddleware creates a role-based access control middleware
func RoleMiddleware(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := GetUserFromContext(c)
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(response.ApiResponse{
				Success: false,
				Error:   "Unauthorized",
			})
		}

		for _, role := range allowedRoles {
			if user.Role == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(response.ApiResponse{
			Success: false,
			Error:   "Forbidden: Insufficient permissions",
		})
	}
}
