package request

type CreateWarehouseRequest struct {
	Code    string  `json:"code" validate:"required"`
	Name    string  `json:"name" validate:"required"`
	Type    string  `json:"type" validate:"required,oneof=MAIN BRANCH STORAGE OUTLET"`
	Address string  `json:"address" validate:"required"`
	City    string  `json:"city" validate:"required"`
	Phone   *string `json:"phone" validate:"omitempty"`
}

type UpdateWarehouseRequest struct {
	Code     *string `json:"code" validate:"omitempty"`
	Name     *string `json:"name" validate:"omitempty"`
	Type     *string `json:"type" validate:"omitempty,oneof=MAIN BRANCH STORAGE OUTLET"`
	Address  *string `json:"address" validate:"omitempty"`
	City     *string `json:"city" validate:"omitempty"`
	Phone    *string `json:"phone" validate:"omitempty"`
	IsActive *bool   `json:"is_active" validate:"omitempty"`
}
