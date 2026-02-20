package request

type CreateCompanyRequest struct {
	Code            string  `json:"code" validate:"required"`
	Nama            string  `json:"nama" validate:"required"`
	Email           string  `json:"email" validate:"required,email"`
	Address         *string `json:"address" validate:"omitempty"`
	Telp            *string `json:"telp" validate:"omitempty"`
	Website         *string `json:"website" validate:"omitempty"`
	TaxID           *string `json:"tax_id" validate:"omitempty"`
	BusinessLicense *string `json:"business_license" validate:"omitempty"`
	IsActive        *bool   `json:"is_active" validate:"omitempty"`
}

type UpdateCompanyRequest struct {
	Nama            *string `json:"nama" validate:"omitempty"`
	Email           *string `json:"email" validate:"omitempty,email"`
	Address         *string `json:"address" validate:"omitempty"`
	Telp            *string `json:"telp" validate:"omitempty"`
	Logo            *string `json:"logo" validate:"omitempty"`
	Website         *string `json:"website" validate:"omitempty"`
	TaxID           *string `json:"tax_id" validate:"omitempty"`
	BusinessLicense *string `json:"business_license" validate:"omitempty"`
	IsActive        *bool   `json:"is_active" validate:"omitempty"`
}

type UploadCompanyLogoRequest struct {
	Logo string `json:"logo" validate:"required"`
}
