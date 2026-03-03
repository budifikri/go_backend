package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db     *gorm.DB
	dbHost string
	dbPort int
	dbName string
}

func NewHealthHandler(db *gorm.DB, dbHost string, dbPort int, dbName string) *HealthHandler {
	return &HealthHandler{db: db, dbHost: dbHost, dbPort: dbPort, dbName: dbName}
}

// GetHealth godoc
// @Summary Health check
// @Description Check API and database health status
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/health [get]
func (h *HealthHandler) GetHealth(c *fiber.Ctx) error {
	dbStatus := "disconnected"
	if sqlDB, err := h.db.DB(); err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err == nil {
			dbStatus = "connected"
		}
	}

	return c.JSON(fiber.Map{
		"message":       "POS Retail API is running",
		"version":       "1.0.0",
		"documentation": "/docs",
		"database": fiber.Map{
			"status":   dbStatus,
			"host":     h.dbHost,
			"port":     h.dbPort,
			"database": h.dbName,
		},
	})
}
