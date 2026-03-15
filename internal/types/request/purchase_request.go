package request

type PurchaseOrderItemRequest struct {
	ID        string  `json:"id"`
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Discount  float64 `json:"discount"`
	TaxRate   float64 `json:"tax_rate"`
}

type CreatePurchaseOrderRequest struct {
	SupplierID   string                     `json:"supplier_id"`
	WarehouseID  string                     `json:"warehouse_id"`
	CompanyID    string                     `json:"company_id"`
	ExpectedDate string                     `json:"expected_date"`
	Items        []PurchaseOrderItemRequest `json:"items"`
	Notes        *string                    `json:"notes"`
}

type UpdatePurchaseOrderRequest struct {
	SupplierID    string                     `json:"supplier_id"`
	WarehouseID   string                     `json:"warehouse_id"`
	OrderDate     string                     `json:"order_date"`
	ExpectedDate  string                     `json:"expected_date"`
	Items         []PurchaseOrderItemRequest `json:"items"`
	Notes         *string                    `json:"notes"`
	StatusPo      string                     `json:"status_po"`
	StatusReceive string                     `json:"status_receive"`
}

type UpdatePurchaseOrderStatusRequest struct {
	Status string `json:"status"`
}

type ReceivePurchaseOrderItemRequest struct {
	ID         string `json:"id"`
	QtyReceive int    `json:"qty_receive"`
}

type ReceivePurchaseOrderRequest struct {
	Items         []ReceivePurchaseOrderItemRequest `json:"items"`
	StatusReceive string                            `json:"status_receive"`
	ReceiveDate   string                            `json:"receive_date"`
}

type PurchaseReturnItemRequest struct {
	PoItemID  string  `json:"po_item_id"`
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Amount    float64 `json:"amount"`
	Notes     string  `json:"notes"`
}

type CreatePurchaseReturnRequest struct {
	PoID       string                      `json:"po_id"`
	ReturnDate string                      `json:"return_date"`
	Reason     string                      `json:"reason"`
	Items      []PurchaseReturnItemRequest `json:"items"`
}

type UpdatePurchaseReturnStatusRequest struct {
	Status string `json:"status"`
}
