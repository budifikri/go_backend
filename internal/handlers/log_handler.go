package handlers

import (
	"regexp"

	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/logger"
	"github.com/pos-retail/go_backend/internal/types/response"
)

var yearMonthPattern = regexp.MustCompile(`^\d{4}_\d{2}$`)

type LogHandler struct {
	logger *logger.Logger
}

func NewLogHandler(logger *logger.Logger) *LogHandler {
	return &LogHandler{logger: logger}
}

// GetSummary godoc
// @Summary Get CRUD logs summary
// @Tags Logs
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/logs/summary [get]
func (h *LogHandler) GetSummary(c *fiber.Ctx) error {
	return c.JSON(response.NewSuccessResponse(h.logger.GetSummary(), ""))
}

// SaveSummary godoc
// @Summary Save CRUD summary to file
// @Tags Logs
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 500 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/logs/save [post]
func (h *LogHandler) SaveSummary(c *fiber.Ctx) error {
	if err := h.logger.SaveSummary(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.NewErrorResponse("Failed to save summary"))
	}
	return c.JSON(response.NewSuccessResponse(nil, "Summary saved successfully"))
}

// ListFiles godoc
// @Summary List available log files by month
// @Tags Logs
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 500 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/logs/files [get]
func (h *LogHandler) ListFiles(c *fiber.Ctx) error {
	files, err := h.logger.ListFiles()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.NewErrorResponse("Failed to list log files"))
	}
	return c.JSON(response.NewSuccessResponse(files, ""))
}

// GetTableLogs godoc
// @Summary Get CRUD logs by table
// @Tags Logs
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param tahun_bulan path string true "Year month format (yyyy_mm)"
// @Param table path string true "Table name"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 500 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/logs/{tahun_bulan}/{table} [get]
func (h *LogHandler) GetTableLogs(c *fiber.Ctx) error {
	yearMonth := c.Params("tahun_bulan")
	if !yearMonthPattern.MatchString(yearMonth) {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid tahun_bulan format, expected yyyy_mm"))
	}

	table := c.Params("table")
	if table == "" || table == "error" {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid table name"))
	}

	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	logs, total, err := h.logger.ReadTableLogs(yearMonth, table, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.NewErrorResponse("Failed to read table logs"))
	}

	return c.JSON(response.NewPaginatedResponse(logs, total, limit, offset))
}

// GetErrorLogs godoc
// @Summary Get error logs by month
// @Tags Logs
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param tahun_bulan path string true "Year month format (yyyy_mm)"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse
// @Failure 400 {object} response.ApiResponse
// @Failure 401 {object} response.ApiResponse
// @Failure 500 {object} response.ApiResponse
// @Security BearerAuth
// @Router /api/logs/{tahun_bulan}/error [get]
func (h *LogHandler) GetErrorLogs(c *fiber.Ctx) error {
	yearMonth := c.Params("tahun_bulan")
	if !yearMonthPattern.MatchString(yearMonth) {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid tahun_bulan format, expected yyyy_mm"))
	}

	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	logs, total, err := h.logger.ReadErrorLogs(yearMonth, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.NewErrorResponse("Failed to read error logs"))
	}

	return c.JSON(response.NewPaginatedResponse(logs, total, limit, offset))
}
