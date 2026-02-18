package request

type FinanceInvoiceItemRequest struct {
	ProductID    *string  `json:"product_id" validate:"omitempty,uuid"`
	Description  string   `json:"description" validate:"required"`
	Quantity     int      `json:"quantity" validate:"required,gte=1"`
	UnitPrice    float64  `json:"unit_price" validate:"required,gte=0"`
	DiscountRate *float64 `json:"discount_rate" validate:"omitempty,gte=0"`
	TaxRate      *float64 `json:"tax_rate" validate:"omitempty,gte=0"`
}

type CreateIncomingInvoiceRequest struct {
	SupplierID     string                      `json:"supplier_id" validate:"required,uuid"`
	InvoiceDate    string                      `json:"invoice_date" validate:"required"`
	DueDate        *string                     `json:"due_date" validate:"omitempty"`
	Notes          *string                     `json:"notes" validate:"omitempty"`
	AttachmentPath *string                     `json:"attachment_path" validate:"omitempty"`
	Items          []FinanceInvoiceItemRequest `json:"items" validate:"required,dive"`
}

type UpdateIncomingInvoiceRequest struct {
	InvoiceDate *string                      `json:"invoice_date" validate:"omitempty"`
	DueDate     *string                      `json:"due_date" validate:"omitempty"`
	Notes       *string                      `json:"notes" validate:"omitempty"`
	Status      *string                      `json:"status" validate:"omitempty,oneof=DRAFT SENT PARTIAL PAID OVERDUE CANCELLED"`
	Items       *[]FinanceInvoiceItemRequest `json:"items" validate:"omitempty,dive"`
}

type CreateOutgoingInvoiceRequest struct {
	CustomerID     *string                     `json:"customer_id" validate:"omitempty,uuid"`
	InvoiceDate    string                      `json:"invoice_date" validate:"required"`
	DueDate        *string                     `json:"due_date" validate:"omitempty"`
	Notes          *string                     `json:"notes" validate:"omitempty"`
	AttachmentPath *string                     `json:"attachment_path" validate:"omitempty"`
	Items          []FinanceInvoiceItemRequest `json:"items" validate:"required,dive"`
}

type UpdateOutgoingInvoiceRequest struct {
	InvoiceDate *string                      `json:"invoice_date" validate:"omitempty"`
	DueDate     *string                      `json:"due_date" validate:"omitempty"`
	Notes       *string                      `json:"notes" validate:"omitempty"`
	Status      *string                      `json:"status" validate:"omitempty,oneof=DRAFT SENT PARTIAL PAID OVERDUE CANCELLED"`
	Items       *[]FinanceInvoiceItemRequest `json:"items" validate:"omitempty,dive"`
}

type CreateInvoicePaymentRequest struct {
	Amount          float64 `json:"amount" validate:"required,gte=0"`
	PaymentMethod   string  `json:"payment_method" validate:"required,oneof=CASH BANK_TRANSFER CHECK CARD EWALLET"`
	ReferenceNumber *string `json:"reference_number" validate:"omitempty"`
	Notes           *string `json:"notes" validate:"omitempty"`
}
