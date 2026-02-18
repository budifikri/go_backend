package request

type SaleItemRequest struct {
	ProductID     string `json:"product_id"`
	Quantity      int    `json:"quantity"`
	PromotionCode string `json:"promotion_code"`
}

type SalePaymentRequest struct {
	Method          string  `json:"method"`
	Amount          float64 `json:"amount"`
	ReferenceNumber string  `json:"reference_number"`
	CardLast4       string  `json:"card_last_4"`
}

type CreateSaleRequest struct {
	WarehouseID         string               `json:"warehouse_id"`
	CustomerID          string               `json:"customer_id"`
	Items               []SaleItemRequest    `json:"items"`
	Payments            []SalePaymentRequest `json:"payments"`
	LoyaltyPointsRedeem int                  `json:"loyalty_points_redeem"`
	Notes               string               `json:"notes"`
}
