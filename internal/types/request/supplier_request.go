package request

type CreateSupplierRequest struct {
	Name          string  `json:"name"`
	ContactPerson string  `json:"contact_person"`
	Email         string  `json:"email"`
	Phone         string  `json:"phone"`
	Address       string  `json:"address"`
	City          string  `json:"city"`
	TaxID         string  `json:"tax_id"`
	PaymentTerms  string  `json:"payment_terms"`
	CreditLimit   float64 `json:"credit_limit"`
	Notes         string  `json:"notes"`
}

type UpdateSupplierRequest struct {
	Name          *string  `json:"name"`
	ContactPerson *string  `json:"contact_person"`
	Email         *string  `json:"email"`
	Phone         *string  `json:"phone"`
	Address       *string  `json:"address"`
	City          *string  `json:"city"`
	TaxID         *string  `json:"tax_id"`
	PaymentTerms  *string  `json:"payment_terms"`
	CreditLimit   *float64 `json:"credit_limit"`
	Status        *string  `json:"status"`
	Notes         *string  `json:"notes"`
}
