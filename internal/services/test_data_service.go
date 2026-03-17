package services

import (
	"fmt"

	"github.com/google/uuid"
	applogger "github.com/pos-retail/go_backend/internal/logger"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/types/response"
	"github.com/pos-retail/go_backend/internal/utils"
	"gorm.io/gorm"
)

type TestDataService struct {
	db *gorm.DB
}

func NewTestDataService(db *gorm.DB) *TestDataService {
	return &TestDataService{
		db: db,
	}
}

type tableConfig struct {
	model        interface{}
	tableName    string
	hasCompanyID bool
	companyIDCol string
}

func (s *TestDataService) getAllowedMasterTables() []string {
	return []string{
		"units",
		"categories",
		"warehouses",
		"products",
		"price_tiers",
		"customers",
		"suppliers",
		"promotions",
		"promotion_products",
		"promotion_categories",
		"promotion_customers",
	}
}

func (s *TestDataService) getAllowedTransactionTables() []string {
	return []string{
		"sale_payments",
		"sale_items",
		"sales",
		"sales_return_items",
		"sales_returns",
		"exchange_items",
		"item_exchanges",
		"invoice_payments",
		"invoice_items",
		"invoices_incoming",
		"invoices_outgoing",
		"cash_drawer_transactions",
		"cash_drawers",
		"purchase_order_items",
		"purchase_orders",
		"purchase_return_items",
		"purchase_returns",
		"stock_opname_items",
		"stock_opnames",
		"stock_transfer_items",
		"stock_transfers",
		"stock_movements",
		"inventories",
	}
}

func (s *TestDataService) getTableConfig(tableKey string) (tableConfig, bool) {
	tables := map[string]tableConfig{
		"units": {
			model:        &models.Unit{},
			tableName:    "units",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"categories": {
			model:        &models.Category{},
			tableName:    "categories",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"warehouses": {
			model:        &models.Warehouse{},
			tableName:    "warehouses",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"products": {
			model:        &models.Product{},
			tableName:    "products",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"price_tiers": {
			model:        &models.PriceTier{},
			tableName:    "price_tiers",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"customers": {
			model:        &models.Customer{},
			tableName:    "customers",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"suppliers": {
			model:        &models.Supplier{},
			tableName:    "suppliers",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"promotions": {
			model:        &models.Promotion{},
			tableName:    "promotions",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"promotion_products": {
			model:        &models.PromotionProduct{},
			tableName:    "promotion_products",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"promotion_categories": {
			model:        &models.PromotionCategory{},
			tableName:    "promotion_categories",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"promotion_customers": {
			model:        &models.PromotionCustomer{},
			tableName:    "promotion_customers",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"inventories": {
			model:        &models.Inventory{},
			tableName:    "inventories",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"stock_movements": {
			model:        &models.StockMovement{},
			tableName:    "stock_movements",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"stock_transfers": {
			model:        &models.StockTransfer{},
			tableName:    "stock_transfers",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"stock_transfer_items": {
			model:        &models.StockTransferItem{},
			tableName:    "stock_transfer_items",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"stock_opnames": {
			model:        &models.StockOpname{},
			tableName:    "stock_opnames",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"stock_opname_items": {
			model:        &models.StockOpnameItem{},
			tableName:    "stock_opname_items",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"sales": {
			model:        &models.Sale{},
			tableName:    "sales",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"sale_items": {
			model:        &models.SaleItem{},
			tableName:    "sale_items",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"sale_payments": {
			model:        &models.SalePayment{},
			tableName:    "sale_payments",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"sales_returns": {
			model:        &models.SalesReturn{},
			tableName:    "sales_returns",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"sales_return_items": {
			model:        &models.SalesReturnItem{},
			tableName:    "sales_return_items",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"item_exchanges": {
			model:        &models.ItemExchange{},
			tableName:    "item_exchanges",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"exchange_items": {
			model:        &models.ExchangeItem{},
			tableName:    "exchange_items",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"purchase_orders": {
			model:        &models.PurchaseOrder{},
			tableName:    "purchase_orders",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"purchase_order_items": {
			model:        &models.PurchaseOrderItem{},
			tableName:    "purchase_order_items",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"purchase_returns": {
			model:        &models.PurchaseReturn{},
			tableName:    "purchase_returns",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"purchase_return_items": {
			model:        &models.PurchaseReturnItem{},
			tableName:    "purchase_return_items",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"invoices_incoming": {
			model:        &models.IncomingInvoice{},
			tableName:    "invoices_incoming",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"invoices_outgoing": {
			model:        &models.OutgoingInvoice{},
			tableName:    "invoices_outgoing",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"invoice_items": {
			model:        &models.InvoiceItem{},
			tableName:    "invoice_items",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"invoice_payments": {
			model:        &models.InvoicePayment{},
			tableName:    "invoice_payments",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"cash_drawers": {
			model:        &models.CashDrawer{},
			tableName:    "cash_drawers",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"cash_drawer_transactions": {
			model:        &models.CashDrawerTransaction{},
			tableName:    "cash_drawer_transactions",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
	}

	config, ok := tables[tableKey]
	return config, ok
}

func (s *TestDataService) GetCompanyIDs(user *utils.JWTPayload) []uuid.UUID {
	companyIDs := make([]uuid.UUID, 0)

	if user.CompanyID != "" {
		if cid, err := uuid.Parse(user.CompanyID); err == nil {
			companyIDs = append(companyIDs, cid)
		}
	}

	for _, access := range user.CompanyAccess {
		if cid, err := uuid.Parse(access); err == nil {
			exists := false
			for _, id := range companyIDs {
				if id == cid {
					exists = true
					break
				}
			}
			if !exists {
				companyIDs = append(companyIDs, cid)
			}
		}
	}

	return companyIDs
}

func (s *TestDataService) DeleteMasterData(companyIDs []uuid.UUID, actorUserID, actorCompanyID string) (map[string]int64, response.ApiResponse) {
	if len(companyIDs) == 0 {
		return nil, response.NewErrorResponse("No company found in token")
	}

	masterTables := []string{
		"promotion_customers",
		"promotion_categories",
		"promotion_products",
		"promotions",
		"price_tiers",
		"products",
		"customers",
		"suppliers",
		"warehouses",
		"categories",
		"units",
	}

	results := make(map[string]int64)

	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, tableKey := range masterTables {
			count, err := s.deleteTableByKey(tx, tableKey, companyIDs, actorUserID, actorCompanyID)
			if err != nil {
				return fmt.Errorf("failed to delete %s: %w", tableKey, err)
			}
			results[tableKey] = count
		}
		return nil
	})

	if err != nil {
		if l := applogger.Default(); l != nil {
			l.LogError(applogger.ActionDelete, "test_data_master", actorUserID, actorCompanyID, "all", err)
		}
		return results, response.NewErrorResponse(fmt.Sprintf("Failed to delete master data: %v", err))
	}

	if l := applogger.Default(); l != nil {
		l.Log(applogger.ActionDelete, "test_data_master", actorUserID, actorCompanyID, "all", nil, results)
	}
	return results, response.NewSuccessResponse(results, fmt.Sprintf("Master data cleared: %d tables deleted", len(masterTables)))
}

func (s *TestDataService) DeleteTransactionData(companyIDs []uuid.UUID, actorUserID, actorCompanyID string) (map[string]int64, response.ApiResponse) {
	if len(companyIDs) == 0 {
		return nil, response.NewErrorResponse("No company found in token")
	}

	transactionTables := []string{
		"sale_payments",
		"sale_items",
		"sales",
		"sales_return_items",
		"sales_returns",
		"exchange_items",
		"item_exchanges",
		"invoice_payments",
		"invoice_items",
		"invoices_incoming",
		"invoices_outgoing",
		"cash_drawer_transactions",
		"cash_drawers",
		"purchase_order_items",
		"purchase_orders",
		"purchase_return_items",
		"purchase_returns",
		"stock_opname_items",
		"stock_opnames",
		"stock_transfer_items",
		"stock_transfers",
		"stock_movements",
		"inventories",
	}

	results := make(map[string]int64)

	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, tableKey := range transactionTables {
			count, err := s.deleteTableByKey(tx, tableKey, companyIDs, actorUserID, actorCompanyID)
			if err != nil {
				return fmt.Errorf("failed to delete %s: %w", tableKey, err)
			}
			results[tableKey] = count
		}
		return nil
	})

	if err != nil {
		if l := applogger.Default(); l != nil {
			l.LogError(applogger.ActionDelete, "test_data_transaction", actorUserID, actorCompanyID, "all", err)
		}
		return results, response.NewErrorResponse(fmt.Sprintf("Failed to delete transaction data: %v", err))
	}

	if l := applogger.Default(); l != nil {
		l.Log(applogger.ActionDelete, "test_data_transaction", actorUserID, actorCompanyID, "all", nil, results)
	}
	return results, response.NewSuccessResponse(results, fmt.Sprintf("Transaction data cleared: %d tables deleted", len(transactionTables)))
}

func (s *TestDataService) DeleteTableData(tables []string, companyIDs []uuid.UUID, actorUserID, actorCompanyID string) (map[string]int64, response.ApiResponse) {
	if len(companyIDs) == 0 {
		return nil, response.NewErrorResponse("No company found in token")
	}

	allowedTables := make(map[string]bool)
	for _, t := range s.getAllowedMasterTables() {
		allowedTables[t] = true
	}
	for _, t := range s.getAllowedTransactionTables() {
		allowedTables[t] = true
	}

	invalidTables := make([]string, 0)
	for _, t := range tables {
		if !allowedTables[t] {
			invalidTables = append(invalidTables, t)
		}
	}

	if len(invalidTables) > 0 {
		return nil, response.NewErrorResponse(fmt.Sprintf("Invalid tables: %v. Allowed tables: master (units, categories, warehouses, products, price_tiers, customers, suppliers, promotions, promotion_products, promotion_categories, promotion_customers) and transactions (sale_payments, sale_items, sales, sales_return_items, sales_returns, exchange_items, item_exchanges, invoice_payments, invoice_items, invoices_incoming, invoices_outgoing, cash_drawer_transactions, cash_drawers, purchase_order_items, purchase_orders, purchase_return_items, purchase_returns, stock_opname_items, stock_opnames, stock_transfer_items, stock_transfers, stock_movements, inventories)", invalidTables))
	}

	results := make(map[string]int64)

	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, tableKey := range tables {
			count, err := s.deleteTableByKey(tx, tableKey, companyIDs, actorUserID, actorCompanyID)
			if err != nil {
				return fmt.Errorf("failed to delete %s: %w", tableKey, err)
			}
			results[tableKey] = count
		}
		return nil
	})

	if err != nil {
		if l := applogger.Default(); l != nil {
			l.LogError(applogger.ActionDelete, "test_data_custom", actorUserID, actorCompanyID, "custom", err)
		}
		return results, response.NewErrorResponse(fmt.Sprintf("Failed to delete tables: %v", err))
	}

	if l := applogger.Default(); l != nil {
		l.Log(applogger.ActionDelete, "test_data_custom", actorUserID, actorCompanyID, "custom", nil, results)
	}
	return results, response.NewSuccessResponse(results, fmt.Sprintf("%d tables cleared", len(tables)))
}

func (s *TestDataService) deleteTableByKey(tx *gorm.DB, tableKey string, companyIDs []uuid.UUID, actorUserID, actorCompanyID string) (int64, error) {
	config, ok := s.getTableConfig(tableKey)
	if !ok {
		return 0, fmt.Errorf("table %s not found in config", tableKey)
	}

	return s.deleteTable(tx, config.model, config.tableName, config.hasCompanyID, config.companyIDCol, companyIDs, actorUserID, actorCompanyID)
}

func (s *TestDataService) deleteTable(tx *gorm.DB, model interface{}, tableName string, hasCompanyID bool, companyIDCol string, companyIDs []uuid.UUID, actorUserID, actorCompanyID string) (int64, error) {
	var count int64

	if hasCompanyID {
		result := tx.Table(tableName).Where(companyIDCol+" IN ?", companyIDs).Delete(model)
		if result.Error != nil {
			return 0, result.Error
		}
		count = result.RowsAffected
	} else {
		result := tx.Table(tableName).Delete(model)
		if result.Error != nil {
			return 0, result.Error
		}
		count = result.RowsAffected
	}

	if l := applogger.Default(); l != nil {
		l.Log(applogger.ActionDelete, tableName, actorUserID, actorCompanyID, tableName, nil, count)
	}

	return count, nil
}
