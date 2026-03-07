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
	ExpectedDate string                     `json:"expected_date"`
	Items        []PurchaseOrderItemRequest `json:"items"`
	Notes        *string                    `json:"notes"`
}

type UpdatePurchaseOrderRequest struct {
	SupplierID   string                     `json:"supplier_id"`
	WarehouseID  string                     `json:"warehouse_id"`
	ExpectedDate string                     `json:"expected_date"`
	Items        []PurchaseOrderItemRequest `json:"items"`
	Notes        *string                    `json:"notes"`
}

type UpdatePurchaseOrderStatusRequest struct {
	Status string `json:"status"`
}
