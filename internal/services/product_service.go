package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/repository"
	"github.com/pos-retail/go_backend/internal/types/response"
)

var (
	ErrProductNotFound  = errors.New("product not found")
	ErrProductExists    = errors.New("product with this SKU or barcode already exists")
	ErrCategoryNotFound = errors.New("category not found")
	ErrUnitNotFound     = errors.New("unit not found")
)

type ProductService struct {
	productRepo  *repository.ProductRepository
	categoryRepo *repository.CategoryRepository
	unitRepo     *repository.UnitRepository
}

func NewProductService(
	productRepo *repository.ProductRepository,
	categoryRepo *repository.CategoryRepository,
	unitRepo *repository.UnitRepository,
) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		unitRepo:     unitRepo,
	}
}

type ProductListResponse struct {
	ID           uuid.UUID  `json:"id"`
	SKU          string     `json:"sku"`
	Barcode      string     `json:"barcode"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	CategoryID   *uuid.UUID `json:"category_id"`
	CategoryName string     `json:"category_name,omitempty"`
	UnitID       uuid.UUID  `json:"unit_id"`
	UnitName     string     `json:"unit_name,omitempty"`
	CostPrice    float64    `json:"cost_price"`
	RetailPrice  float64    `json:"retail_price"`
	Status       string     `json:"status"`
	TaxRate      float64    `json:"tax_rate"`
	ReorderPoint int        `json:"reorder_point"`
	CompanyID    *uuid.UUID `json:"company_id,omitempty"`
	CreatedAt    string     `json:"created_at"`
	UpdatedAt    string     `json:"updated_at"`
}

type ProductDetailResponse struct {
	ID           uuid.UUID           `json:"id"`
	SKU          string              `json:"sku"`
	Barcode      string              `json:"barcode"`
	Name         string              `json:"name"`
	Description  string              `json:"description"`
	CategoryID   *uuid.UUID          `json:"category_id"`
	CategoryName string              `json:"category_name,omitempty"`
	UnitID       uuid.UUID           `json:"unit_id"`
	UnitName     string              `json:"unit_name,omitempty"`
	CostPrice    float64             `json:"cost_price"`
	RetailPrice  float64             `json:"retail_price"`
	Status       string              `json:"status"`
	TaxRate      float64             `json:"tax_rate"`
	ReorderPoint int                 `json:"reorder_point"`
	CompanyID    *uuid.UUID          `json:"company_id,omitempty"`
	CreatedAt    string              `json:"created_at"`
	UpdatedAt    string              `json:"updated_at"`
	PriceTiers   []PriceTierResponse `json:"price_tiers,omitempty"`
}

type PriceTierResponse struct {
	ID          uuid.UUID `json:"id"`
	ProductID   uuid.UUID `json:"product_id"`
	TierName    string    `json:"tier_name"`
	MinQuantity int       `json:"min_quantity"`
	MaxQuantity *int      `json:"max_quantity"`
	UnitPrice   float64   `json:"unit_price"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   string    `json:"created_at"`
}

type CreateProductRequest struct {
	SKU          string  `json:"sku"`
	Barcode      string  `json:"barcode"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	CategoryID   string  `json:"category_id"`
	UnitID       string  `json:"unit_id"`
	CostPrice    float64 `json:"cost_price"`
	RetailPrice  float64 `json:"retail_price"`
	TaxRate      float64 `json:"tax_rate"`
	ReorderPoint int     `json:"reorder_point"`
	CompanyID    string  `json:"company_id"`
}

func (s *ProductService) GetProducts(filters map[string]interface{}, limit, offset int) response.PaginatedResponse {
	products, total, err := s.productRepo.FindAll(filters, limit, offset)
	if err != nil {
		return response.PaginatedResponse{
			Success: false,
			Data:    []interface{}{},
			Pagination: response.Pagination{
				Total:   0,
				Limit:   limit,
				Offset:  offset,
				HasMore: false,
			},
		}
	}

	data := make([]ProductListResponse, len(products))
	for i, p := range products {
		categoryName := ""
		if p.Category != nil {
			categoryName = p.Category.Name
		}
		unitName := ""
		if p.Unit != nil {
			unitName = p.Unit.Name
		}
		data[i] = ProductListResponse{
			ID:           p.ID,
			SKU:          p.SKU,
			Barcode:      p.Barcode,
			Name:         p.Name,
			Description:  p.Description,
			CategoryID:   p.CategoryID,
			CategoryName: categoryName,
			UnitID:       p.UnitID,
			UnitName:     unitName,
			CostPrice:    p.CostPrice,
			RetailPrice:  p.RetailPrice,
			Status:       string(p.Status),
			TaxRate:      p.TaxRate,
			ReorderPoint: p.ReorderPoint,
			CompanyID:    p.CompanyID,
			CreatedAt:    p.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:    p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return response.NewPaginatedResponse(data, total, limit, offset)
}

func (s *ProductService) GetProductByID(id string) response.ApiResponse {
	productID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Invalid product ID")
	}

	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return response.NewErrorResponse("Failed to get product")
	}
	if product == nil {
		return response.NewErrorResponse(ErrProductNotFound.Error())
	}

	tiers, _ := s.productRepo.FindPriceTiers(productID)

	tierResponses := make([]PriceTierResponse, len(tiers))
	for i, t := range tiers {
		tierResponses[i] = PriceTierResponse{
			ID:          t.ID,
			ProductID:   t.ProductID,
			TierName:    t.TierName,
			MinQuantity: t.MinQuantity,
			MaxQuantity: t.MaxQuantity,
			UnitPrice:   t.UnitPrice,
			IsActive:    t.IsActive,
			CreatedAt:   t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	categoryName := ""
	if product.Category != nil {
		categoryName = product.Category.Name
	}
	unitName := ""
	if product.Unit != nil {
		unitName = product.Unit.Name
	}

	return response.NewSuccessResponse(ProductDetailResponse{
		ID:           product.ID,
		SKU:          product.SKU,
		Barcode:      product.Barcode,
		Name:         product.Name,
		Description:  product.Description,
		CategoryID:   product.CategoryID,
		CategoryName: categoryName,
		UnitID:       product.UnitID,
		UnitName:     unitName,
		CostPrice:    product.CostPrice,
		RetailPrice:  product.RetailPrice,
		Status:       string(product.Status),
		TaxRate:      product.TaxRate,
		ReorderPoint: product.ReorderPoint,
		CompanyID:    product.CompanyID,
		CreatedAt:    product.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    product.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		PriceTiers:   tierResponses,
	}, "Product retrieved successfully")
}

func (s *ProductService) CreateProduct(req CreateProductRequest) response.ApiResponse {
	unitID, err := uuid.Parse(req.UnitID)
	if err != nil {
		return response.NewErrorResponse("Invalid unit ID")
	}

	unit, err := s.unitRepo.FindByID(unitID)
	if err != nil || unit == nil {
		return response.NewErrorResponse(ErrUnitNotFound.Error())
	}

	existingBySKU, _ := s.productRepo.FindBySKU(req.SKU)
	if existingBySKU != nil {
		return response.NewErrorResponse(ErrProductExists.Error())
	}

	if req.Barcode != "" {
		existingByBarcode, _ := s.productRepo.FindByBarcode(req.Barcode)
		if existingByBarcode != nil {
			return response.NewErrorResponse(ErrProductExists.Error())
		}
	}

	var categoryID *uuid.UUID
	if req.CategoryID != "" {
		catID, err := uuid.Parse(req.CategoryID)
		if err == nil {
			categoryID = &catID
		}
	}

	var companyID *uuid.UUID
	if req.CompanyID != "" {
		compID, err := uuid.Parse(req.CompanyID)
		if err == nil {
			companyID = &compID
		}
	}

	product := models.Product{
		ID:           uuid.New(),
		SKU:          req.SKU,
		Barcode:      req.Barcode,
		Name:         req.Name,
		Description:  req.Description,
		CategoryID:   categoryID,
		UnitID:       unitID,
		CostPrice:    req.CostPrice,
		RetailPrice:  req.RetailPrice,
		Status:       models.ProductStatusActive,
		TaxRate:      req.TaxRate,
		ReorderPoint: req.ReorderPoint,
		CompanyID:    companyID,
	}

	if err := s.productRepo.Create(&product); err != nil {
		return response.NewErrorResponse("Failed to create product")
	}

	return response.NewSuccessResponse(ProductDetailResponse{
		ID:          product.ID,
		SKU:         product.SKU,
		Barcode:     product.Barcode,
		Name:        product.Name,
		UnitID:      product.UnitID,
		CostPrice:   product.CostPrice,
		RetailPrice: product.RetailPrice,
		Status:      string(product.Status),
		TaxRate:     product.TaxRate,
		CreatedAt:   product.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   product.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}, "Product created successfully")
}

func (s *ProductService) UpdateProduct(id string, req CreateProductRequest) response.ApiResponse {
	productID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Invalid product ID")
	}

	product, err := s.productRepo.FindByID(productID)
	if err != nil || product == nil {
		return response.NewErrorResponse(ErrProductNotFound.Error())
	}

	if req.SKU != "" && req.SKU != product.SKU {
		existing, _ := s.productRepo.FindBySKU(req.SKU)
		if existing != nil && existing.ID != productID {
			return response.NewErrorResponse(ErrProductExists.Error())
		}
		product.SKU = req.SKU
	}

	if req.Barcode != "" && req.Barcode != product.Barcode {
		existing, _ := s.productRepo.FindByBarcode(req.Barcode)
		if existing != nil && existing.ID != productID {
			return response.NewErrorResponse(ErrProductExists.Error())
		}
		product.Barcode = req.Barcode
	}

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.CostPrice > 0 {
		product.CostPrice = req.CostPrice
	}
	if req.RetailPrice > 0 {
		product.RetailPrice = req.RetailPrice
	}
	if req.TaxRate >= 0 {
		product.TaxRate = req.TaxRate
	}
	if req.ReorderPoint >= 0 {
		product.ReorderPoint = req.ReorderPoint
	}

	if req.CategoryID != "" {
		catID, err := uuid.Parse(req.CategoryID)
		if err == nil {
			product.CategoryID = &catID
		}
	}
	if req.UnitID != "" {
		unitID, err := uuid.Parse(req.UnitID)
		if err == nil {
			product.UnitID = unitID
		}
	}

	if err := s.productRepo.Update(product); err != nil {
		return response.NewErrorResponse("Failed to update product")
	}

	return response.NewSuccessResponse(ProductDetailResponse{
		ID:          product.ID,
		SKU:         product.SKU,
		Barcode:     product.Barcode,
		Name:        product.Name,
		UnitID:      product.UnitID,
		CostPrice:   product.CostPrice,
		RetailPrice: product.RetailPrice,
		Status:      string(product.Status),
		TaxRate:     product.TaxRate,
		UpdatedAt:   product.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}, "Product updated successfully")
}

func (s *ProductService) DeleteProduct(id string) response.ApiResponse {
	productID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Invalid product ID")
	}

	product, err := s.productRepo.FindByID(productID)
	if err != nil || product == nil {
		return response.NewErrorResponse(ErrProductNotFound.Error())
	}

	product.Status = models.ProductStatusDiscontinued
	if err := s.productRepo.Update(product); err != nil {
		return response.NewErrorResponse("Failed to delete product")
	}

	return response.NewSuccessResponse(nil, "Product deleted successfully")
}

func (s *ProductService) GetCategories(companyID *string) response.ApiResponse {
	var compID *uuid.UUID
	if companyID != nil && *companyID != "" {
		id, err := uuid.Parse(*companyID)
		if err == nil {
			compID = &id
		}
	}

	// TS parity default pagination for categories list
	categories, err := s.categoryRepo.FindAll(compID, 50, 0)
	if err != nil {
		return response.NewErrorResponse("Failed to get categories")
	}

	return response.NewSuccessResponse(categories, "")
}

func (s *ProductService) GetCategoriesPaged(companyID *string, limit, offset int) response.ApiResponse {
	var compID *uuid.UUID
	if companyID != nil && *companyID != "" {
		id, err := uuid.Parse(*companyID)
		if err == nil {
			compID = &id
		}
	}

	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	categories, err := s.categoryRepo.FindAll(compID, limit, offset)
	if err != nil {
		return response.NewErrorResponse("Failed to get categories")
	}
	return response.NewSuccessResponse(categories, "")
}

func (s *ProductService) GetCategoryByID(id string) response.ApiResponse {
	catID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Category not found")
	}

	cat, err := s.categoryRepo.FindByID(catID)
	if err != nil {
		return response.NewErrorResponse("Failed to get category")
	}
	if cat == nil {
		return response.NewErrorResponse("Category not found")
	}
	return response.NewSuccessResponse(cat, "")
}

type CreateCategoryInput struct {
	Code        string
	Name        string
	Description *string
	ParentID    *string
	CompanyID   *string
}

func (s *ProductService) CreateCategory(input CreateCategoryInput) response.ApiResponse {
	existing, err := s.categoryRepo.FindByCode(input.Code)
	if err != nil {
		return response.NewErrorResponse("Failed to create category")
	}
	if existing != nil {
		return response.NewErrorResponse("Category code already exists")
	}

	var parentUUID *uuid.UUID
	if input.ParentID != nil {
		pid := *input.ParentID
		if pid != "" {
			p, err := uuid.Parse(pid)
			if err == nil {
				parentUUID = &p
			}
		}
	}

	var companyUUID *uuid.UUID
	if input.CompanyID != nil && *input.CompanyID != "" {
		cid, err := uuid.Parse(*input.CompanyID)
		if err == nil {
			companyUUID = &cid
		}
	}

	cat := models.Category{
		ID:          uuid.New(),
		Code:        input.Code,
		Name:        input.Name,
		Description: "",
		ParentID:    parentUUID,
		CompanyID:   companyUUID,
		IsActive:    true,
	}
	if input.Description != nil {
		cat.Description = *input.Description
	}

	if err := s.categoryRepo.Create(&cat); err != nil {
		return response.NewErrorResponse("Failed to create category")
	}
	return response.NewSuccessResponse(cat, "")
}

type UpdateCategoryInput struct {
	Code        *string
	Name        *string
	Description *string
	ParentID    *string
	IsActive    *bool
}

func (s *ProductService) UpdateCategory(id string, input UpdateCategoryInput) response.ApiResponse {
	catID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Category not found")
	}

	cat, err := s.categoryRepo.FindByID(catID)
	if err != nil {
		return response.NewErrorResponse("Failed to update category")
	}
	if cat == nil {
		return response.NewErrorResponse("Category not found")
	}

	updated := false
	if input.Code != nil {
		cat.Code = *input.Code
		updated = true
	}
	if input.Name != nil {
		cat.Name = *input.Name
		updated = true
	}
	if input.Description != nil {
		cat.Description = *input.Description
		updated = true
	}
	if input.ParentID != nil {
		pid := *input.ParentID
		if pid == "" {
			cat.ParentID = nil
		} else {
			p, err := uuid.Parse(pid)
			if err == nil {
				cat.ParentID = &p
			}
		}
		updated = true
	}
	if input.IsActive != nil {
		cat.IsActive = *input.IsActive
		updated = true
	}

	if !updated {
		return response.NewErrorResponse("No fields to update")
	}

	if err := s.categoryRepo.Update(cat); err != nil {
		return response.NewErrorResponse("Failed to update category")
	}
	return response.NewSuccessResponse(cat, "")
}

func (s *ProductService) DeleteCategory(id string) response.ApiResponse {
	catID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Category not found")
	}

	cat, err := s.categoryRepo.FindByID(catID)
	if err != nil {
		return response.NewErrorResponse("Failed to delete category")
	}
	if cat == nil {
		return response.NewErrorResponse("Category not found")
	}

	cat.IsActive = false
	if err := s.categoryRepo.Update(cat); err != nil {
		return response.NewErrorResponse("Failed to delete category")
	}
	return response.NewSuccessResponse(nil, "Category deleted successfully")
}

func (s *ProductService) GetUnits() response.ApiResponse {
	units, err := s.unitRepo.FindAll()
	if err != nil {
		return response.NewErrorResponse("Failed to fetch units")
	}

	return response.NewSuccessResponse(units, "")
}

func (s *ProductService) GetUnitsWithQuery(search *string, limit, offset *int) (response.ApiResponse, map[string]interface{}) {
	units, err := s.unitRepo.FindAllWithQuery(search, limit, offset)
	if err != nil {
		return response.NewErrorResponse("Failed to fetch units"), nil
	}

	if limit != nil || offset != nil {
		total, err := s.unitRepo.Count(search)
		if err != nil {
			return response.NewErrorResponse("Failed to fetch units"), nil
		}
		l := len(units)
		lim := l
		if limit != nil {
			lim = *limit
		}
		off := 0
		if offset != nil {
			off = *offset
		}
		return response.NewSuccessResponse(units, ""), map[string]interface{}{
			"limit":  lim,
			"offset": off,
			"total":  total,
		}
	}

	return response.NewSuccessResponse(units, ""), nil
}

func (s *ProductService) GetUnitByID(id string) response.ApiResponse {
	uid, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Unit not found")
	}
	unit, err := s.unitRepo.FindByID(uid)
	if err != nil {
		return response.NewErrorResponse("Failed to fetch unit")
	}
	if unit == nil {
		return response.NewErrorResponse("Unit not found")
	}
	return response.NewSuccessResponse(unit, "")
}

type CreateUnitInput struct {
	Code        string
	Name        string
	Description *string
}

func (s *ProductService) CreateUnit(input CreateUnitInput) response.ApiResponse {
	existing, err := s.unitRepo.FindByCode(input.Code)
	if err != nil {
		return response.NewErrorResponse("Failed to create unit")
	}
	if existing != nil {
		return response.NewErrorResponse("Unit code already exists")
	}

	unit := models.Unit{
		ID:          uuid.New(),
		Code:        input.Code,
		Name:        input.Name,
		Description: "",
		IsActive:    true,
	}
	if input.Description != nil {
		unit.Description = *input.Description
	}
	if err := s.unitRepo.Create(&unit); err != nil {
		return response.NewErrorResponse("Failed to create unit")
	}
	return response.NewSuccessResponse(unit, "")
}

type UpdateUnitInput struct {
	Code        *string
	Name        *string
	Description *string
}

func (s *ProductService) UpdateUnit(id string, input UpdateUnitInput) response.ApiResponse {
	uid, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Unit not found")
	}
	unit, err := s.unitRepo.FindByID(uid)
	if err != nil {
		return response.NewErrorResponse("Failed to update unit")
	}
	if unit == nil {
		return response.NewErrorResponse("Unit not found")
	}

	updated := false
	if input.Code != nil {
		unit.Code = *input.Code
		updated = true
	}
	if input.Name != nil {
		unit.Name = *input.Name
		updated = true
	}
	if input.Description != nil {
		unit.Description = *input.Description
		updated = true
	}
	if !updated {
		return response.NewErrorResponse("No fields to update")
	}

	if err := s.unitRepo.Update(unit); err != nil {
		return response.NewErrorResponse("Failed to update unit")
	}
	return response.NewSuccessResponse(unit, "")
}

func (s *ProductService) DeleteUnit(id string) response.ApiResponse {
	uid, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Unit not found")
	}
	unit, err := s.unitRepo.FindByID(uid)
	if err != nil {
		return response.NewErrorResponse("Failed to delete unit")
	}
	if unit == nil {
		return response.NewErrorResponse("Unit not found")
	}
	if err := s.unitRepo.Delete(uid); err != nil {
		return response.NewErrorResponse("Failed to delete unit")
	}
	return response.NewSuccessResponse(nil, "Unit deleted successfully")
}
