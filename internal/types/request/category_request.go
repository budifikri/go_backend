package request

type CreateCategoryRequest struct {
	Code        string  `json:"code" validate:"required"`
	Name        string  `json:"name" validate:"required"`
	ProductType string  `json:"product_type" validate:"omitempty,oneof=stockable service consumable"`
	Description *string `json:"description" validate:"omitempty"`
	ParentID    *string `json:"parent_id" validate:"omitempty"`
}

type UpdateCategoryRequest struct {
	Code        *string `json:"code" validate:"omitempty"`
	Name        *string `json:"name" validate:"omitempty"`
	ProductType *string `json:"product_type" validate:"omitempty,oneof=stockable service consumable"`
	Description *string `json:"description" validate:"omitempty"`
	ParentID    *string `json:"parent_id" validate:"omitempty"`
	IsActive    *bool   `json:"is_active" validate:"omitempty"`
}
