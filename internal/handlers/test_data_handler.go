package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pos-retail/go_backend/internal/middleware"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type TestDataHandler struct {
	testDataService *services.TestDataService
}

func NewTestDataHandler(testDataService *services.TestDataService) *TestDataHandler {
	return &TestDataHandler{
		testDataService: testDataService,
	}
}

// DeleteMasterData godoc
// @Summary Clear all master data
// @Description Deletes all master data (units, categories, warehouses, products, price_tiers, customers, suppliers, promotions) for the user's company
// @Tags Remove Data
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.ApiResponse
// @Router /api/remove-data/master [delete]
func (h *TestDataHandler) DeleteMasterData(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	companyIDs := h.testDataService.GetCompanyIDs(user)
	results, apiResponse := h.testDataService.DeleteMasterData(companyIDs, user.UserID, user.CompanyID)

	if !apiResponse.Success {
		return c.Status(fiber.StatusInternalServerError).JSON(apiResponse)
	}

	return c.JSON(response.NewSuccessResponse(results, apiResponse.Message))
}

// DeleteTransactionData godoc
// @Summary Clear all transaction data
// @Description Deletes all transaction data (sales, purchases, inventory, invoices, etc) for the user's company
// @Tags Remove Data
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.ApiResponse
// @Router /api/remove-data/transactions [delete]
func (h *TestDataHandler) DeleteTransactionData(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	companyIDs := h.testDataService.GetCompanyIDs(user)
	results, apiResponse := h.testDataService.DeleteTransactionData(companyIDs, user.UserID, user.CompanyID)

	if !apiResponse.Success {
		return c.Status(fiber.StatusInternalServerError).JSON(apiResponse)
	}

	return c.JSON(response.NewSuccessResponse(results, apiResponse.Message))
}

// DeleteTableData godoc
// @Summary Clear specific tables
// @Description Deletes data from specified tables for the user's company
// @Tags Remove Data
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tables body request.DeleteTableRequest true "Tables to delete"
// @Success 200 {object} response.ApiResponse
// @Router /api/remove-data/table [delete]
func (h *TestDataHandler) DeleteTableData(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("Unauthorized"))
	}

	var req request.DeleteTableRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("Invalid request body"))
	}

	if len(req.Tables) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("At least one table must be specified"))
	}

	companyIDs := h.testDataService.GetCompanyIDs(user)
	results, apiResponse := h.testDataService.DeleteTableData(req.Tables, companyIDs, user.UserID, user.CompanyID)

	if !apiResponse.Success {
		return c.Status(fiber.StatusBadRequest).JSON(apiResponse)
	}

	return c.JSON(response.NewSuccessResponse(results, apiResponse.Message))
}
