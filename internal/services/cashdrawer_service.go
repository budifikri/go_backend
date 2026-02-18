package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/repository"
	"github.com/pos-retail/go_backend/internal/types/response"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CashDrawerService struct {
	db             *gorm.DB
	repo           *repository.CashDrawerRepository
	financeService *FinanceService
}

func NewCashDrawerService(db *gorm.DB, repo *repository.CashDrawerRepository, financeService *FinanceService) *CashDrawerService {
	return &CashDrawerService{db: db, repo: repo, financeService: financeService}
}

type OpenCashDrawerInput struct {
	DrawerNumber   *string
	WarehouseID    string
	OpeningBalance float64
	Notes          *string
}

type CloseCashDrawerInput struct {
	ClosingBalance float64
	Notes          *string
	PaymentMethod  *string
}

type CashInOutInput struct {
	Amount float64
	Reason string
}

type cashDrawerRow struct {
	ID               uuid.UUID           `json:"id" gorm:"column:id"`
	DrawerNumber     string              `json:"drawer_number" gorm:"column:drawer_number"`
	WarehouseID      uuid.UUID           `json:"warehouse_id" gorm:"column:warehouse_id"`
	CashierID        uuid.UUID           `json:"cashier_id" gorm:"column:cashier_id"`
	CompanyID        uuid.UUID           `json:"company_id" gorm:"column:company_id"`
	Status           models.DrawerStatus `json:"status" gorm:"column:status"`
	OpeningBalance   float64             `json:"opening_balance" gorm:"column:opening_balance"`
	ClosingBalance   *float64            `json:"closing_balance" gorm:"column:closing_balance"`
	ExpectedBalance  float64             `json:"expected_balance" gorm:"column:expected_balance"`
	Variance         *float64            `json:"variance" gorm:"column:variance"`
	OpenedAt         time.Time           `json:"opened_at" gorm:"column:opened_at"`
	ClosedAt         *time.Time          `json:"closed_at" gorm:"column:closed_at"`
	Notes            *string             `json:"notes" gorm:"column:notes"`
	DepositInvoiceID *uuid.UUID          `json:"deposit_invoice_id" gorm:"column:deposit_invoice_id"`
	CreatedAt        time.Time           `json:"created_at" gorm:"column:created_at"`
	UpdatedAt        time.Time           `json:"updated_at" gorm:"column:updated_at"`
	WarehouseName    *string             `json:"warehouse_name" gorm:"column:warehouse_name"`
	CashierName      *string             `json:"cashier_name" gorm:"column:cashier_name"`
}

type cashDrawerTransactionRow struct {
	ID            uuid.UUID              `json:"id" gorm:"column:id"`
	CashDrawerID  uuid.UUID              `json:"cash_drawer_id" gorm:"column:cash_drawer_id"`
	Type          models.TransactionType `json:"type" gorm:"column:type"`
	Amount        float64                `json:"amount" gorm:"column:amount"`
	BalanceAfter  float64                `json:"balance_after" gorm:"column:balance_after"`
	SaleID        *uuid.UUID             `json:"sale_id" gorm:"column:sale_id"`
	PaymentMethod *string                `json:"payment_method" gorm:"column:payment_method"`
	Reason        string                 `json:"reason" gorm:"column:reason"`
	CreatedAt     time.Time              `json:"created_at" gorm:"column:created_at"`
	CreatedBy     uuid.UUID              `json:"created_by" gorm:"column:created_by"`
	CreatedByName *string                `json:"created_by_name" gorm:"column:created_by_name"`
}

func (s *CashDrawerService) OpenCashDrawer(input OpenCashDrawerInput, companyID string, userID string) response.ApiResponse {
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Failed to open cash drawer")
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return response.NewErrorResponse("Failed to open cash drawer")
	}
	warehouseUUID, err := uuid.Parse(input.WarehouseID)
	if err != nil {
		return response.NewErrorResponse("Failed to open cash drawer")
	}

	drawerNumber := ""
	if input.DrawerNumber != nil && strings.TrimSpace(*input.DrawerNumber) != "" {
		drawerNumber = strings.TrimSpace(*input.DrawerNumber)
	} else {
		drawerNumber = s.repo.CreateDrawerNumber()
	}

	var created models.CashDrawer
	err = s.db.Transaction(func(tx *gorm.DB) error {
		exists, err := s.repo.WithDB(tx).DrawerExistsOpenForCashierWarehouse(companyUUID, userUUID, warehouseUUID)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("Cash drawer is already open for this cashier at this warehouse")
		}

		created = models.CashDrawer{
			ID:              uuid.New(),
			DrawerNumber:    drawerNumber,
			WarehouseID:     warehouseUUID,
			CashierID:       userUUID,
			CompanyID:       companyUUID,
			Status:          models.DrawerStatusOpen,
			OpeningBalance:  input.OpeningBalance,
			ExpectedBalance: input.OpeningBalance,
			OpenedAt:        time.Now(),
			Notes:           input.Notes,
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}

		reason := "Opening cash drawer"
		if input.Notes != nil && strings.TrimSpace(*input.Notes) != "" {
			reason = *input.Notes
		}
		txRow := models.CashDrawerTransaction{
			ID:           uuid.New(),
			CashDrawerID: created.ID,
			Type:         models.TransactionTypeOpening,
			Amount:       input.OpeningBalance,
			BalanceAfter: input.OpeningBalance,
			Reason:       reason,
			CreatedAt:    time.Now(),
			CreatedBy:    userUUID,
		}
		return tx.Create(&txRow).Error
	})
	if err != nil {
		return response.NewErrorResponse(err.Error())
	}
	return response.NewSuccessResponse(created, "")
}

func (s *CashDrawerService) GetCurrentDrawer(warehouseID *string, companyID string, cashierID string) response.ApiResponse {
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("No open cash drawer found")
	}
	cashierUUID, err := uuid.Parse(cashierID)
	if err != nil {
		return response.NewErrorResponse("No open cash drawer found")
	}
	var wid *uuid.UUID
	if warehouseID != nil && strings.TrimSpace(*warehouseID) != "" {
		w, err := uuid.Parse(*warehouseID)
		if err == nil {
			wid = &w
		}
	}
	row, err := s.repo.GetCurrentDrawer(wid, companyUUID, cashierUUID)
	if err != nil || row == nil {
		return response.NewErrorResponse("No open cash drawer found")
	}
	return response.NewSuccessResponse(row, "")
}

func (s *CashDrawerService) GetCashDrawerByID(id string, companyID string) response.ApiResponse {
	drawerUUID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Cash drawer not found")
	}
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Cash drawer not found")
	}

	drawer, err := s.repo.GetCashDrawerByID(drawerUUID, companyUUID)
	if err != nil || drawer == nil {
		return response.NewErrorResponse("Cash drawer not found")
	}

	txs, _ := s.repo.GetAllDrawerTransactions(drawerUUID)

	data := map[string]interface{}{}
	data["id"] = drawer.ID
	data["drawer_number"] = drawer.DrawerNumber
	data["warehouse_id"] = drawer.WarehouseID
	data["cashier_id"] = drawer.CashierID
	data["company_id"] = drawer.CompanyID
	data["status"] = drawer.Status
	data["opening_balance"] = drawer.OpeningBalance
	data["closing_balance"] = drawer.ClosingBalance
	data["expected_balance"] = drawer.ExpectedBalance
	data["variance"] = drawer.Variance
	data["opened_at"] = drawer.OpenedAt
	data["closed_at"] = drawer.ClosedAt
	data["notes"] = drawer.Notes
	data["deposit_invoice_id"] = drawer.DepositInvoiceID
	data["created_at"] = drawer.CreatedAt
	data["updated_at"] = drawer.UpdatedAt
	data["cashier_name"] = drawer.CashierName
	data["warehouse_name"] = drawer.WarehouseName
	data["transactions"] = txs

	return response.NewSuccessResponse(data, "")
}

func (s *CashDrawerService) ListCashDrawers(filters map[string]string, limit, offset int, companyID string) response.PaginatedResponse {
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}
	rows, total, err := s.repo.ListCashDrawers(filters, limit, offset, companyUUID)
	if err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}
	return response.NewPaginatedResponse(rows, total, limit, offset)
}

func (s *CashDrawerService) AddCashIn(drawerID string, input CashInOutInput, companyID string, userID string) response.ApiResponse {
	drawerUUID, err := uuid.Parse(drawerID)
	if err != nil {
		return response.NewErrorResponse("Cash drawer not found")
	}
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Cash drawer not found")
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return response.NewErrorResponse("Failed to add cash in")
	}

	var created models.CashDrawerTransaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var drawer models.CashDrawer
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&drawer, "id = ? AND company_id = ?", drawerUUID, companyUUID).Error; err != nil {
			return err
		}
		if drawer.Status != models.DrawerStatusOpen {
			return errors.New("Cash drawer must be OPEN to add cash in")
		}

		newBalance := drawer.ExpectedBalance + input.Amount
		if err := tx.Model(&models.CashDrawer{}).Where("id = ?", drawerUUID).
			Updates(map[string]interface{}{"expected_balance": newBalance, "updated_at": time.Now()}).Error; err != nil {
			return err
		}

		created = models.CashDrawerTransaction{
			ID:           uuid.New(),
			CashDrawerID: drawerUUID,
			Type:         models.TransactionTypeCashIn,
			Amount:       input.Amount,
			BalanceAfter: newBalance,
			Reason:       input.Reason,
			CreatedAt:    time.Now(),
			CreatedBy:    userUUID,
		}
		return tx.Create(&created).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewErrorResponse("Cash drawer not found")
		}
		return response.NewErrorResponse(err.Error())
	}
	return response.NewSuccessResponse(created, "")
}

func (s *CashDrawerService) AddCashOut(drawerID string, input CashInOutInput, companyID string, userID string) response.ApiResponse {
	drawerUUID, err := uuid.Parse(drawerID)
	if err != nil {
		return response.NewErrorResponse("Cash drawer not found")
	}
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Cash drawer not found")
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return response.NewErrorResponse("Failed to add cash out")
	}

	var created models.CashDrawerTransaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var drawer models.CashDrawer
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&drawer, "id = ? AND company_id = ?", drawerUUID, companyUUID).Error; err != nil {
			return err
		}
		if drawer.Status != models.DrawerStatusOpen {
			return errors.New("Cash drawer must be OPEN to add cash out")
		}

		newBalance := drawer.ExpectedBalance - input.Amount
		if err := tx.Model(&models.CashDrawer{}).Where("id = ?", drawerUUID).
			Updates(map[string]interface{}{"expected_balance": newBalance, "updated_at": time.Now()}).Error; err != nil {
			return err
		}

		created = models.CashDrawerTransaction{
			ID:           uuid.New(),
			CashDrawerID: drawerUUID,
			Type:         models.TransactionTypeCashOut,
			Amount:       input.Amount,
			BalanceAfter: newBalance,
			Reason:       input.Reason,
			CreatedAt:    time.Now(),
			CreatedBy:    userUUID,
		}
		return tx.Create(&created).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewErrorResponse("Cash drawer not found")
		}
		return response.NewErrorResponse(err.Error())
	}
	return response.NewSuccessResponse(created, "")
}

func (s *CashDrawerService) GetDrawerTransactions(drawerID string, txType *string, limit, offset int) response.PaginatedResponse {
	drawerUUID, err := uuid.Parse(drawerID)
	if err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}
	rows, total, err := s.repo.GetDrawerTransactions(drawerUUID, txType, limit, offset)
	if err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}
	return response.NewPaginatedResponse(rows, total, limit, offset)
}

type drawerSummary struct {
	OpeningBalance   float64  `json:"opening_balance"`
	ClosingBalance   *float64 `json:"closing_balance"`
	ExpectedBalance  float64  `json:"expected_balance"`
	Variance         float64  `json:"variance"`
	CashInTotal      float64  `json:"cash_in_total"`
	CashOutTotal     float64  `json:"cash_out_total"`
	SalesTotal       float64  `json:"sales_total"`
	RefundsTotal     float64  `json:"refunds_total"`
	TransactionCount int64    `json:"transaction_count"`
	SalesCount       int64    `json:"sales_count"`
	DurationMinutes  int      `json:"duration_minutes"`
}

type txSumRow struct {
	Type   string  `gorm:"column:type"`
	Amount float64 `gorm:"column:amount"`
}

type salesSumRow struct {
	Count int64   `gorm:"column:count"`
	Total float64 `gorm:"column:total_amount"`
}

func (s *CashDrawerService) GetDrawerSummary(id string, companyID string) response.ApiResponse {
	drawerUUID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Cash drawer not found")
	}
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Cash drawer not found")
	}

	var drawer models.CashDrawer
	if err := s.db.First(&drawer, "id = ? AND company_id = ?", drawerUUID, companyUUID).Error; err != nil {
		return response.NewErrorResponse("Cash drawer not found")
	}

	var sums []txSumRow
	_ = s.db.Raw(`SELECT type, COALESCE(SUM(amount),0) as amount FROM cash_drawer_transactions WHERE cash_drawer_id = ? GROUP BY type`, drawerUUID).Scan(&sums).Error
	mapSum := map[string]float64{}
	for _, r := range sums {
		mapSum[r.Type] = r.Amount
	}

	cashIn := mapSum[string(models.TransactionTypeCashIn)]
	cashOut := mapSum[string(models.TransactionTypeCashOut)]
	returnOut := mapSum[string(models.TransactionTypeReturnOut)]

	var salesAgg salesSumRow
	_ = s.db.Raw(`SELECT COUNT(*) as count, COALESCE(SUM(total_amount),0) as total_amount FROM sales WHERE cash_drawer_id = ?`, drawerUUID).Scan(&salesAgg).Error

	var txCount int64
	_ = s.db.Table("cash_drawer_transactions").Where("cash_drawer_id = ?", drawerUUID).Count(&txCount).Error

	openedAt := drawer.OpenedAt
	closedAt := time.Now()
	if drawer.ClosedAt != nil {
		closedAt = *drawer.ClosedAt
	}
	durationMinutes := int(closedAt.Sub(openedAt).Minutes())
	if durationMinutes < 0 {
		durationMinutes = 0
	}

	variance := 0.0
	if drawer.Variance != nil {
		variance = *drawer.Variance
	}

	data := drawerSummary{
		OpeningBalance:   drawer.OpeningBalance,
		ClosingBalance:   drawer.ClosingBalance,
		ExpectedBalance:  drawer.ExpectedBalance,
		Variance:         variance,
		CashInTotal:      cashIn,
		CashOutTotal:     cashOut,
		SalesTotal:       salesAgg.Total,
		RefundsTotal:     returnOut,
		TransactionCount: txCount,
		SalesCount:       salesAgg.Count,
		DurationMinutes:  durationMinutes,
	}
	return response.NewSuccessResponse(data, "")
}

func (s *CashDrawerService) CloseCashDrawer(id string, input CloseCashDrawerInput, companyID string, userID string) response.ApiResponse {
	drawerUUID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Cash drawer not found")
	}
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Cash drawer not found")
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return response.NewErrorResponse("Failed to close cash drawer")
	}

	var result models.CashDrawer
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var drawer models.CashDrawer
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&drawer, "id = ? AND company_id = ?", drawerUUID, companyUUID).Error; err != nil {
			return err
		}
		if drawer.Status != models.DrawerStatusOpen {
			return errors.New("Cash drawer must be OPEN to close")
		}

		variance := input.ClosingBalance - drawer.ExpectedBalance
		closing := input.ClosingBalance
		now := time.Now()
		drawer.ClosingBalance = &closing
		drawer.Variance = &variance
		drawer.Status = models.DrawerStatusClosed
		drawer.ClosedAt = &now
		drawer.Notes = input.Notes
		drawer.UpdatedAt = now

		if err := tx.Save(&drawer).Error; err != nil {
			return err
		}

		reason := "Closing cash drawer"
		if input.Notes != nil && strings.TrimSpace(*input.Notes) != "" {
			reason = *input.Notes
		}
		txRow := models.CashDrawerTransaction{
			ID:           uuid.New(),
			CashDrawerID: drawerUUID,
			Type:         models.TransactionTypeClosing,
			Amount:       input.ClosingBalance,
			BalanceAfter: input.ClosingBalance,
			Reason:       reason,
			CreatedAt:    now,
			CreatedBy:    userUUID,
		}
		if err := tx.Create(&txRow).Error; err != nil {
			return err
		}

		result = drawer
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewErrorResponse("Cash drawer not found")
		}
		return response.NewErrorResponse(err.Error())
	}

	// Create deposit invoice (best-effort; do not fail drawer closure).
	variance := 0.0
	if result.Variance != nil {
		variance = *result.Variance
	}
	notes := fmt.Sprintf("Cash drawer deposit for drawer %s. Amount: %.2f, Variance: %.2f", result.DrawerNumber, input.ClosingBalance, variance)
	invoiceInput := CreateOutgoingInvoiceInput{
		CustomerID:  nil,
		InvoiceDate: time.Now().Format(time.RFC3339),
		DueDate:     func() *string { s := time.Now().Format(time.RFC3339); return &s }(),
		Notes:       &notes,
		Items: []FinanceInvoiceItemInput{{
			Description: "Cash Drawer Deposit",
			Quantity:    1,
			UnitPrice:   input.ClosingBalance,
		}},
	}
	invRes := s.financeService.CreateOutgoingInvoice(invoiceInput, companyID, userID)
	if invRes.Success {
		if inv, ok := invRes.Data.(models.OutgoingInvoice); ok {
			_ = s.db.Model(&models.CashDrawer{}).Where("id = ?", result.ID).Update("deposit_invoice_id", inv.ID).Error
			result.DepositInvoiceID = &inv.ID
		}
	}

	return response.NewSuccessResponse(result, "")
}
