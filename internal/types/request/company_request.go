package request

type CreateCompanyRequest struct {
	Code            string   `json:"code" validate:"required"`
	Nama            string   `json:"nama" validate:"required"`
	Email           string   `json:"email" validate:"required,email"`
	BusinessType    string   `json:"business_type" validate:"required"`
	ModuleCodes     []string `json:"module_codes" validate:"omitempty"`
	Address         *string  `json:"address" validate:"omitempty"`
	Telp            *string  `json:"telp" validate:"omitempty"`
	Website         *string  `json:"website" validate:"omitempty"`
	TaxID           *string  `json:"tax_id" validate:"omitempty"`
	BusinessLicense *string  `json:"business_license" validate:"omitempty"`
	IsActive        *bool    `json:"is_active" validate:"omitempty"`
}

type UpdateCompanyRequest struct {
	Nama            *string   `json:"nama" validate:"omitempty"`
	Email           *string   `json:"email" validate:"omitempty,email"`
	BusinessType    *string   `json:"business_type" validate:"omitempty"`
	ModuleCodes     *[]string `json:"module_codes" validate:"omitempty"`
	Address         *string   `json:"address" validate:"omitempty"`
	Telp            *string   `json:"telp" validate:"omitempty"`
	Logo            *string   `json:"logo" validate:"omitempty"`
	Website         *string   `json:"website" validate:"omitempty"`
	TaxID           *string   `json:"tax_id" validate:"omitempty"`
	BusinessLicense *string   `json:"business_license" validate:"omitempty"`
	IsActive        *bool     `json:"is_active" validate:"omitempty"`
}

type UploadCompanyLogoRequest struct {
	Logo string `json:"logo" validate:"required"`
}

type CreateBusinessTypeRequest struct {
	Code        string `json:"code" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"omitempty"`
	IsActive    *bool  `json:"is_active" validate:"omitempty"`
	IsDefault   *bool  `json:"is_default" validate:"omitempty"`
	IsSystem    *bool  `json:"is_system" validate:"omitempty"`
	SortOrder   *int   `json:"sort_order" validate:"omitempty"`
}

type UpdateBusinessTypeRequest struct {
	Name        *string `json:"name" validate:"omitempty"`
	Description *string `json:"description" validate:"omitempty"`
	IsActive    *bool   `json:"is_active" validate:"omitempty"`
	IsDefault   *bool   `json:"is_default" validate:"omitempty"`
	IsSystem    *bool   `json:"is_system" validate:"omitempty"`
	SortOrder   *int    `json:"sort_order" validate:"omitempty"`
}

type CreateModulePackageRequest struct {
	BusinessType string `json:"business_type" validate:"required"`
	Code         string `json:"code" validate:"required"`
	Name         string `json:"name" validate:"required"`
	Description  string `json:"description" validate:"omitempty"`
	IsActive     *bool  `json:"is_active" validate:"omitempty"`
	IsDefault    *bool  `json:"is_default" validate:"omitempty"`
	IsSystem     *bool  `json:"is_system" validate:"omitempty"`
	SortOrder    *int   `json:"sort_order" validate:"omitempty"`
}

type UpdateModulePackageRequest struct {
	BusinessType *string `json:"business_type" validate:"omitempty"`
	Name         *string `json:"name" validate:"omitempty"`
	Description  *string `json:"description" validate:"omitempty"`
	IsActive     *bool   `json:"is_active" validate:"omitempty"`
	IsDefault    *bool   `json:"is_default" validate:"omitempty"`
	IsSystem     *bool   `json:"is_system" validate:"omitempty"`
	SortOrder    *int    `json:"sort_order" validate:"omitempty"`
}

type ToggleCompanyModuleRequest struct {
	IsActive bool `json:"is_active" validate:"required"`
}
