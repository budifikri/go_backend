package request

type SaleItemRequest struct {
	ItemType      string `json:"item_type"`
	ProductID     string `json:"product_id"`
	TreatmentID   string `json:"treatment_id"`
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
	WarehouseID         string               `json:"warehouse_id" validate:"required"`
	CustomerID          string               `json:"customer_id"`
	AppointmentID       string               `json:"appointment_id"`
	CashDrawerID        string               `json:"cash_drawer_id" validate:"required"`
	Status              string               `json:"status" validate:"required,oneof=PENDING DONE CANCELLED REFUNDED"`
	Items               []SaleItemRequest    `json:"items" validate:"required,min=1"`
	Payments            []SalePaymentRequest `json:"payments"`
	LoyaltyPointsRedeem int                  `json:"loyalty_points_redeem"`
	Notes               string               `json:"notes"`
}
