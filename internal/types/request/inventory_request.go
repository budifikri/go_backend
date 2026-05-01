package request

type InventoryAdjustmentRequest struct {
	ProductID      string `json:"product_id"`
	WarehouseID    string `json:"warehouse_id"`
	AdjustmentType string `json:"adjustment_type"`
	Quantity       int    `json:"quantity"`
	Reason         string `json:"reason"`
	Notes          string `json:"notes"`
}

type StockTransferItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type StockTransferRequest struct {
	FromWarehouseID string                     `json:"from_warehouse_id"`
	ToWarehouseID   string                     `json:"to_warehouse_id"`
	ExpectedArrival string                     `json:"expected_arrival"`
	Items           []StockTransferItemRequest `json:"items"`
	Notes           string                     `json:"notes"`
}

type ReceiveTransferItemRequest struct {
	TransferItemID   string `json:"transfer_item_id"`
	ReceivedQuantity int    `json:"received_quantity"`
	Notes            string `json:"notes"`
}

type ReceiveTransferRequest struct {
	Items []ReceiveTransferItemRequest `json:"items"`
}

type StockOpnameItemRequest struct {
	ProductID      string  `json:"product_id"`
	SystemQuantity int     `json:"system_quantity"`
	ActualQuantity int     `json:"actual_quantity"`
	CostPrice      float64 `json:"cost_price"`
	Notes          string  `json:"notes"`
}

type StockOpnameRequest struct {
	WarehouseID string                   `json:"warehouse_id"`
	OpnameDate  string                   `json:"opname_date"`
	IsOpening   bool                     `json:"is_opening"`
	Items       []StockOpnameItemRequest `json:"items"`
	Notes       string                   `json:"notes"`
}

type StockOpnameStatusRequest struct {
	Status string `json:"status"`
}

type StockOpnameUpdateItemRequest struct {
	ID             string  `json:"id"`
	ProductID      string  `json:"product_id"`
	SystemQuantity int     `json:"system_quantity"`
	ActualQuantity int     `json:"actual_quantity"`
	CostPrice      float64 `json:"cost_price"`
	Notes          string  `json:"notes"`
	Status         string  `json:"status"`
}

type StockOpnameUpdateRequest struct {
	WarehouseID string                         `json:"warehouse_id"`
	OpnameDate  string                         `json:"opname_date"`
	IsOpening   bool                           `json:"is_opening"`
	Status      string                         `json:"status"`
	Notes       string                         `json:"notes"`
	Items       []StockOpnameUpdateItemRequest `json:"items"`
}
