package request

type ReturnItemRequest struct {
	SaleItemID string `json:"sale_item_id"`
	ProductID  string `json:"product_id"`
	Quantity   int    `json:"quantity"`
	Condition  string `json:"condition"`
	Notes      string `json:"notes"`
}

type CreateReturnRequest struct {
	SaleID       string              `json:"sale_id"`
	WarehouseID  string              `json:"warehouse_id"`
	Reason       string              `json:"reason"`
	Items        []ReturnItemRequest `json:"items"`
	RefundMethod string              `json:"refund_method"`
}
