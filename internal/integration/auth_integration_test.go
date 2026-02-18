//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/config"
	"github.com/pos-retail/go_backend/internal/database"
	"github.com/pos-retail/go_backend/internal/handlers"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/utils"
)

func requireIntegrationEnv(t *testing.T) {
	// Use separate env vars to avoid accidentally running against dev DB.
	if os.Getenv("INTEGRATION_DB_HOST") == "" {
		t.Skip("set INTEGRATION_DB_HOST/PORT/NAME/USER/PASSWORD to run integration tests")
	}
	os.Setenv("DB_HOST", os.Getenv("INTEGRATION_DB_HOST"))
	os.Setenv("DB_PORT", os.Getenv("INTEGRATION_DB_PORT"))
	os.Setenv("DB_NAME", os.Getenv("INTEGRATION_DB_NAME"))
	os.Setenv("DB_USER", os.Getenv("INTEGRATION_DB_USER"))
	os.Setenv("DB_PASSWORD", os.Getenv("INTEGRATION_DB_PASSWORD"))
	if v := os.Getenv("INTEGRATION_DB_SSL_MODE"); v != "" {
		os.Setenv("DB_SSL_MODE", v)
	}
}

func TestAuthRegisterLoginLogoutIntegration(t *testing.T) {
	requireIntegrationEnv(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}

	// Ensure core auth tables exist.
	if err := database.AutoMigrate(&models.User{}, &models.UserSession{}, &models.EmailVerification{}, &models.PasswordReset{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	jwtUtil := utils.NewJWTUtil("integration-secret", 1*time.Hour)
	authService := services.NewAuthService(db, jwtUtil)
	authHandler := handlers.NewAuthHandler(authService)

	app := fiber.New()
	app.Use(recover.New())
	api := app.Group("/api")
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/logout", authHandler.Logout)

	username := "it_" + uuid.New().String()
	password := "TestPassword123!"

	defer func() {
		// best-effort cleanup
		var user models.User
		if err := db.Where("username = ?", username).First(&user).Error; err == nil {
			_ = db.Where("user_id = ?", user.ID).Delete(&models.UserSession{}).Error
			_ = db.Delete(&models.User{}, "id = ?", user.ID).Error
		}
	}()

	// Register
	regBody := map[string]interface{}{
		"username":  username,
		"email":     username + "@example.com",
		"password":  password,
		"full_name": "Integration Test",
		"role":      "staff",
	}
	regJSON, _ := json.Marshal(regBody)
	regReq := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(regJSON))
	regReq.Header.Set("Content-Type", "application/json")
	regRes, err := app.Test(regReq, 30_000)
	if err != nil {
		t.Fatalf("register request: %v", err)
	}
	if regRes.StatusCode != fiber.StatusCreated {
		t.Fatalf("register status: got %d want %d", regRes.StatusCode, fiber.StatusCreated)
	}

	// Login
	loginBody := map[string]interface{}{"username": username, "password": password}
	loginJSON, _ := json.Marshal(loginBody)
	loginReq := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(loginJSON))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRes, err := app.Test(loginReq, 30_000)
	if err != nil {
		t.Fatalf("login request: %v", err)
	}
	if loginRes.StatusCode != fiber.StatusOK {
		t.Fatalf("login status: got %d want %d", loginRes.StatusCode, fiber.StatusOK)
	}

	var loginResp map[string]interface{}
	if err := json.NewDecoder(loginRes.Body).Decode(&loginResp); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	data, _ := loginResp["data"].(map[string]interface{})
	token, _ := data["token"].(string)
	if token == "" {
		t.Fatalf("expected token")
	}

	// Logout
	logoutReq := httptest.NewRequest("POST", "/api/auth/logout", nil)
	logoutReq.Header.Set("Authorization", "Bearer "+token)
	logoutRes, err := app.Test(logoutReq, 30_000)
	if err != nil {
		t.Fatalf("logout request: %v", err)
	}
	if logoutRes.StatusCode != fiber.StatusOK {
		t.Fatalf("logout status: got %d want %d", logoutRes.StatusCode, fiber.StatusOK)
	}
}
