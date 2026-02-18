package request

type ProductCreateRequest struct {
	SKU          string  `json:"sku" validate:"required,max=50"`
	Barcode      string  `json:"barcode" validate:"omitempty,max=50"`
	Name         string  `json:"name" validate:"required,max=200"`
	Description  string  `json:"description"`
	CategoryID   string  `json:"category_id"`
	UnitID       string  `json:"unit_id" validate:"required,uuid"`
	CostPrice    float64 `json:"cost_price" validate:"required,gt=0"`
	RetailPrice  float64 `json:"retail_price" validate:"required,gt=0"`
	TaxRate      float64 `json:"tax_rate" validate:"omitempty,min=0,max=100"`
	ReorderPoint int     `json:"reorder_point" validate:"omitempty,min=0"`
	CompanyID    string  `json:"company_id"`
}

type ProductUpdateRequest struct {
	SKU          string  `json:"sku" validate:"omitempty,max=50"`
	Barcode      string  `json:"barcode" validate:"omitempty,max=50"`
	Name         string  `json:"name" validate:"omitempty,max=200"`
	Description  string  `json:"description"`
	CategoryID   string  `json:"category_id"`
	UnitID       string  `json:"unit_id" validate:"omitempty,uuid"`
	CostPrice    float64 `json:"cost_price" validate:"omitempty,gt=0"`
	RetailPrice  float64 `json:"retail_price" validate:"omitempty,gt=0"`
	TaxRate      float64 `json:"tax_rate" validate:"omitempty,min=0,max=100"`
	ReorderPoint int     `json:"reorder_point" validate:"omitempty,min=0"`
}

type ProductStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active inactive discontinued"`
}

type ProductPriceRequest struct {
	CostPrice   float64 `json:"cost_price" validate:"omitempty,gt=0"`
	RetailPrice float64 `json:"retail_price" validate:"omitempty,gt=0"`
	TaxRate     float64 `json:"tax_rate" validate:"omitempty,min=0,max=100"`
}

type ProductStockRequest struct {
	Quantity     int `json:"quantity" validate:"omitempty,min=0"`
	ReorderPoint int `json:"reorder_point" validate:"omitempty,min=0"`
}
