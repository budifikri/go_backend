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
	model            interface{}
	tableName        string
	hasCompanyID     bool
	companyIDCol     string
	deleteByParent   bool
	parentTable      string
	parentIDCol      string
	parentKey        string
	parentCompanyCol string
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
		"sales",
		"sale_items",
		"sale_payments",
		"sales_returns",
		"sales_return_items",
		"item_exchanges",
		"exchange_items",
		"purchase_orders",
		"purchase_order_items",
		"purchase_returns",
		"purchase_return_items",
		"invoices_incoming",
		"invoice_items",
		"invoice_payments",
		"invoices_outgoing",
		"cash_drawers",
		"cash_drawer_transactions",
		"stock_opnames",
		"stock_opname_items",
		"stock_transfers",
		"stock_transfer_items",
		"inventories",
		"stock_movements",
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
			tableName:    "inventory",
			hasCompanyID: false,
		},
		"stock_movements": {
			model:        &models.StockMovement{},
			tableName:    "stock_movements",
			hasCompanyID: false,
		},
		"stock_transfers": {
			model:        &models.StockTransfer{},
			tableName:    "stock_transfers",
			hasCompanyID: false,
		},
		"stock_transfer_items": {
			model:            &models.StockTransferItem{},
			tableName:        "stock_transfer_items",
			hasCompanyID:     false,
			deleteByParent:   true,
			parentTable:      "stock_transfers",
			parentIDCol:      "id",
			parentKey:        "transfer_id",
			parentCompanyCol: "company_id",
		},
		"stock_opnames": {
			model:        &models.StockOpname{},
			tableName:    "stock_opnames",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"stock_opname_items": {
			model:            &models.StockOpnameItem{},
			tableName:        "stock_opname_items",
			hasCompanyID:     false,
			deleteByParent:   true,
			parentTable:      "stock_opnames",
			parentIDCol:      "id",
			parentKey:        "opname_id",
			parentCompanyCol: "company_id",
		},
		"sales": {
			model:        &models.Sale{},
			tableName:    "sales",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"sale_items": {
			model:            &models.SaleItem{},
			tableName:        "sale_items",
			hasCompanyID:     false,
			deleteByParent:   true,
			parentTable:      "sales",
			parentIDCol:      "id",
			parentKey:        "sale_id",
			parentCompanyCol: "company_id",
		},
		"sale_payments": {
			model:            &models.SalePayment{},
			tableName:        "sale_payments",
			hasCompanyID:     false,
			deleteByParent:   true,
			parentTable:      "sales",
			parentIDCol:      "id",
			parentKey:        "sale_id",
			parentCompanyCol: "company_id",
		},
		"sales_returns": {
			model:        &models.SalesReturn{},
			tableName:    "sales_returns",
			hasCompanyID: false,
		},
		"sales_return_items": {
			model:        &models.SalesReturnItem{},
			tableName:    "sales_return_items",
			hasCompanyID: false,
		},
		"item_exchanges": {
			model:        &models.ItemExchange{},
			tableName:    "item_exchanges",
			hasCompanyID: false,
		},
		"exchange_items": {
			model:        &models.ExchangeItem{},
			tableName:    "exchange_items",
			hasCompanyID: false,
		},
		"purchase_orders": {
			model:        &models.PurchaseOrder{},
			tableName:    "purchase_orders",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"purchase_order_items": {
			model:            &models.PurchaseOrderItem{},
			tableName:        "purchase_order_items",
			hasCompanyID:     false,
			deleteByParent:   true,
			parentTable:      "purchase_orders",
			parentIDCol:      "id",
			parentKey:        "po_id",
			parentCompanyCol: "company_id",
		},
		"purchase_returns": {
			model:        &models.PurchaseReturn{},
			tableName:    "purchase_returns",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"purchase_return_items": {
			model:            &models.PurchaseReturnItem{},
			tableName:        "purchase_return_items",
			hasCompanyID:     false,
			deleteByParent:   true,
			parentTable:      "purchase_returns",
			parentIDCol:      "id",
			parentKey:        "return_id",
			parentCompanyCol: "company_id",
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
			hasCompanyID: false,
		},
		"invoice_payments": {
			model:        &models.InvoicePayment{},
			tableName:    "invoice_payments",
			hasCompanyID: false,
		},
		"cash_drawers": {
			model:        &models.CashDrawer{},
			tableName:    "cash_drawers",
			hasCompanyID: true,
			companyIDCol: "company_id",
		},
		"cash_drawer_transactions": {
			model:            &models.CashDrawerTransaction{},
			tableName:        "cash_drawer_transactions",
			hasCompanyID:     false,
			deleteByParent:   true,
			parentTable:      "cash_drawers",
			parentIDCol:      "id",
			parentKey:        "cash_drawer_id",
			parentCompanyCol: "company_id",
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

	results := make(map[string]int64)

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Important: delete children first, then parents. Some parents also depend on other parents
		// (example: purchase_returns.po_id -> purchase_orders.id), so ordering matters.
		plan := []string{
			// Sales
			"sale_items",
			"sale_payments",
			"sales_return_items",
			"sales_returns",
			"exchange_items",
			"item_exchanges",
			"sales",
			// Cash drawer
			"cash_drawer_transactions",
			"cash_drawers",
			// Purchases
			"purchase_return_items",
			"purchase_returns",
			"purchase_order_items",
			"purchase_orders",
			// Finance
			"invoice_items",
			"invoice_payments",
			"invoices_incoming",
			"invoices_outgoing",
			// Stock
			"stock_movements",
			"stock_opname_items",
			"stock_opnames",
			"stock_transfer_items",
			"stock_transfers",
			"inventories",
		}

		for _, tableKey := range plan {
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
	return results, response.NewSuccessResponse(results, fmt.Sprintf("Transaction data cleared: %d tables deleted", len(results)))
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
		return nil, response.NewErrorResponse(fmt.Sprintf(
			"Invalid tables: %v. Allowed tables: master %v and transactions %v",
			invalidTables,
			s.getAllowedMasterTables(),
			s.getAllowedTransactionTables(),
		))
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

	return s.deleteTable(tx, config, companyIDs, actorUserID, actorCompanyID)
}

func (s *TestDataService) deleteTable(tx *gorm.DB, config tableConfig, companyIDs []uuid.UUID, actorUserID, actorCompanyID string) (int64, error) {
	var count int64

	if config.tableName == "inventory" {
		c, err := s.deleteInventory(tx, companyIDs)
		if err != nil {
			return 0, err
		}
		count = c
	} else if config.tableName == "stock_movements" {
		c, err := s.deleteStockMovements(tx, companyIDs)
		if err != nil {
			return 0, err
		}
		count = c
	} else if config.tableName == "sales_return_items" {
		c, err := s.deleteSalesReturnItems(tx, companyIDs)
		if err != nil {
			return 0, err
		}
		count = c
	} else if config.tableName == "sales_returns" {
		c, err := s.deleteSalesReturns(tx, companyIDs)
		if err != nil {
			return 0, err
		}
		count = c
	} else if config.tableName == "exchange_items" {
		c, err := s.deleteExchangeItems(tx, companyIDs)
		if err != nil {
			return 0, err
		}
		count = c
	} else if config.tableName == "item_exchanges" {
		c, err := s.deleteItemExchanges(tx, companyIDs)
		if err != nil {
			return 0, err
		}
		count = c
	} else if config.tableName == "stock_transfer_items" {
		c, err := s.deleteStockTransferItems(tx, companyIDs)
		if err != nil {
			return 0, err
		}
		count = c
	} else if config.tableName == "stock_transfers" {
		c, err := s.deleteStockTransfers(tx, companyIDs)
		if err != nil {
			return 0, err
		}
		count = c
		// invoice_items + invoice_payments are polymorphic (invoice_type + invoice_id), so filter by
		// joining to the correct invoice table.
	} else if config.tableName == "invoice_items" || config.tableName == "invoice_payments" {
		incoming, err := s.deleteInvoiceLineTable(tx, config.tableName, "INCOMING", "invoices_incoming", companyIDs)
		if err != nil {
			return 0, err
		}
		outgoing, err := s.deleteInvoiceLineTable(tx, config.tableName, "OUTGOING", "invoices_outgoing", companyIDs)
		if err != nil {
			return 0, err
		}
		count = incoming + outgoing
	} else if config.deleteByParent && config.parentTable != "" && config.parentKey != "" && config.parentCompanyCol != "" {
		parentIDCol := config.parentIDCol
		if parentIDCol == "" {
			parentIDCol = "id"
		}

		// Use DELETE ... USING join to avoid subquery edge-cases and to ensure the
		// parent company filter is always applied correctly.
		q := fmt.Sprintf(
			"DELETE FROM %s child USING %s parent WHERE child.%s = parent.%s AND parent.%s IN (?)",
			config.tableName,
			config.parentTable,
			config.parentKey,
			parentIDCol,
			config.parentCompanyCol,
		)
		res := tx.Exec(q, companyIDs)
		if res.Error != nil {
			return 0, res.Error
		}
		count = res.RowsAffected
	} else if config.hasCompanyID {
		result := tx.Table(config.tableName).Where(config.companyIDCol+" IN ?", companyIDs).Delete(config.model)
		if result.Error != nil {
			return 0, result.Error
		}
		count = result.RowsAffected
	} else {
		result := tx.Table(config.tableName).Delete(config.model)
		if result.Error != nil {
			return 0, result.Error
		}
		count = result.RowsAffected
	}

	if l := applogger.Default(); l != nil {
		l.Log(applogger.ActionDelete, config.tableName, actorUserID, actorCompanyID, config.tableName, nil, count)
	}

	return count, nil
}

func (s *TestDataService) deleteInvoiceLineTable(tx *gorm.DB, lineTable, invoiceType, invoiceTable string, companyIDs []uuid.UUID) (int64, error) {
	// Note: invoiceType is stored as varchar in DB (INCOMING/OUTGOING)
	q := fmt.Sprintf(
		"DELETE FROM %s it USING %s inv WHERE it.invoice_type = ? AND it.invoice_id = inv.id AND inv.company_id IN (?)",
		lineTable,
		invoiceTable,
	)
	res := tx.Exec(q, invoiceType, companyIDs)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

func (s *TestDataService) deleteStockTransferItems(tx *gorm.DB, companyIDs []uuid.UUID) (int64, error) {
	q := "DELETE FROM stock_transfer_items sti USING stock_transfers st, warehouses w WHERE sti.transfer_id = st.id AND (st.from_warehouse_id = w.id OR st.to_warehouse_id = w.id) AND w.company_id IN (?)"
	res := tx.Exec(q, companyIDs)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

func (s *TestDataService) deleteStockTransfers(tx *gorm.DB, companyIDs []uuid.UUID) (int64, error) {
	q := "DELETE FROM stock_transfers st USING warehouses w WHERE (st.from_warehouse_id = w.id OR st.to_warehouse_id = w.id) AND w.company_id IN (?)"
	res := tx.Exec(q, companyIDs)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

func (s *TestDataService) deleteInventory(tx *gorm.DB, companyIDs []uuid.UUID) (int64, error) {
	q := "DELETE FROM inventory i USING warehouses w WHERE i.warehouse_id = w.id AND w.company_id IN (?)"
	res := tx.Exec(q, companyIDs)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

func (s *TestDataService) deleteStockMovements(tx *gorm.DB, companyIDs []uuid.UUID) (int64, error) {
	q := "DELETE FROM stock_movements sm USING warehouses w WHERE sm.warehouse_id = w.id AND w.company_id IN (?)"
	res := tx.Exec(q, companyIDs)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

func (s *TestDataService) deleteSalesReturnItems(tx *gorm.DB, companyIDs []uuid.UUID) (int64, error) {
	q := "DELETE FROM sales_return_items sri USING sales_returns sr, warehouses w WHERE sri.return_id = sr.id AND sr.warehouse_id = w.id AND w.company_id IN (?)"
	res := tx.Exec(q, companyIDs)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

func (s *TestDataService) deleteSalesReturns(tx *gorm.DB, companyIDs []uuid.UUID) (int64, error) {
	q := "DELETE FROM sales_returns sr USING warehouses w WHERE sr.warehouse_id = w.id AND w.company_id IN (?)"
	res := tx.Exec(q, companyIDs)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

func (s *TestDataService) deleteExchangeItems(tx *gorm.DB, companyIDs []uuid.UUID) (int64, error) {
	q := "DELETE FROM exchange_items ei USING item_exchanges ex, warehouses w WHERE ei.exchange_id = ex.id AND ex.warehouse_id = w.id AND w.company_id IN (?)"
	res := tx.Exec(q, companyIDs)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

func (s *TestDataService) deleteItemExchanges(tx *gorm.DB, companyIDs []uuid.UUID) (int64, error) {
	q := "DELETE FROM item_exchanges ex USING warehouses w WHERE ex.warehouse_id = w.id AND w.company_id IN (?)"
	res := tx.Exec(q, companyIDs)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}
