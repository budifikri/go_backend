package request

type CreateCustomerRequest struct {
	Name              string  `json:"name"`
	Email             string  `json:"email"`
	Phone             string  `json:"phone"`
	Address           string  `json:"address"`
	City              string  `json:"city"`
	Tier              string  `json:"tier"`
	CreditLimit       float64 `json:"credit_limit"`
	BankName          string  `json:"bank_name"`
	BankAccountNumber string  `json:"bank_account_number"`
	BankAccountName   string  `json:"bank_account_name"`
	BankBranch        string  `json:"bank_branch"`
}

type UpdateCustomerRequest struct {
	Name              *string  `json:"name"`
	Email             *string  `json:"email"`
	Phone             *string  `json:"phone"`
	Address           *string  `json:"address"`
	City              *string  `json:"city"`
	Tier              *string  `json:"tier"`
	Status            *string  `json:"status"`
	CreditLimit       *float64 `json:"credit_limit"`
	BankName          *string  `json:"bank_name"`
	BankAccountNumber *string  `json:"bank_account_number"`
	BankAccountName   *string  `json:"bank_account_name"`
	BankBranch        *string  `json:"bank_branch"`
}
