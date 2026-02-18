package request

type CreateGrnRequest struct {
	PoID          string  `json:"poId"`
	WarehouseID   *string `json:"warehouseId"`
	InvoiceNumber *string `json:"invoiceNumber"`
	Notes         *string `json:"notes"`
}

type UpdateGrnItemRequest struct {
	ID               string `json:"id"`
	PoItemID         string `json:"poItemId"`
	ProductID        string `json:"productId"`
	OrderedQuantity  int    `json:"orderedQuantity"`
	ReceivedQuantity int    `json:"receivedQuantity"`
	RejectedQuantity int    `json:"rejectedQuantity"`
	UnitPrice        string `json:"unitPrice"`
	QualityNotes     string `json:"qualityNotes"`
}

type UpdateGrnRequest struct {
	InvoiceNumber *string                 `json:"invoiceNumber"`
	Notes         *string                 `json:"notes"`
	Items         *[]UpdateGrnItemRequest `json:"items"`
}
