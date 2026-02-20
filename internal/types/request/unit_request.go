package request

type CreateUnitRequest struct {
	Code        string  `json:"code" validate:"required"`
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description" validate:"omitempty"`
}

type UpdateUnitRequest struct {
	Code        *string `json:"code" validate:"omitempty"`
	Name        *string `json:"name" validate:"omitempty"`
	Description *string `json:"description" validate:"omitempty"`
	IsActive    *bool   `json:"is_active" validate:"omitempty"`
}
