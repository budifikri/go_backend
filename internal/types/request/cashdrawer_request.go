package request

type OpenCashDrawerRequest struct {
	DrawerNumber   *string `json:"drawer_number" validate:"omitempty"`
	WarehouseID    string  `json:"warehouse_id" validate:"required,uuid"`
	OpeningBalance float64 `json:"opening_balance" validate:"required,gte=0"`
	Notes          *string `json:"notes" validate:"omitempty"`
}

type CashInOutRequest struct {
	Amount float64 `json:"amount" validate:"required,gte=0"`
	Reason string  `json:"reason" validate:"required"`
}

type CloseCashDrawerRequest struct {
	ClosingBalance float64 `json:"closing_balance" validate:"required,gte=0"`
	Notes          *string `json:"notes" validate:"omitempty"`
	PaymentMethod  *string `json:"payment_method" validate:"omitempty,oneof=CASH BANK_TRANSFER CHECK"`
}
