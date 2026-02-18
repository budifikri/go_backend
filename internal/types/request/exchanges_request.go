package request

type ExchangeReturnedItemRequest struct {
	SaleItemID string `json:"sale_item_id"`
	ProductID  string `json:"product_id"`
	Quantity   int    `json:"quantity"`
	Condition  string `json:"condition"`
}

type ExchangeReceivedItemRequest struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type CreateExchangeRequest struct {
	SaleID        string                        `json:"sale_id"`
	WarehouseID   string                        `json:"warehouse_id"`
	Reason        string                        `json:"reason"`
	ReturnedItems []ExchangeReturnedItemRequest `json:"returned_items"`
	ReceivedItems []ExchangeReceivedItemRequest `json:"received_items"`
}
