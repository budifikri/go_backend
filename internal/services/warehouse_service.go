package services

import (
	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/repository"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type WarehouseService struct {
	warehouseRepo *repository.WarehouseRepository
}

func NewWarehouseService(warehouseRepo *repository.WarehouseRepository) *WarehouseService {
	return &WarehouseService{warehouseRepo: warehouseRepo}
}

func (s *WarehouseService) GetWarehouses(companyID *string) response.ApiResponse {
	filters := map[string]interface{}{}
	if companyID != nil && *companyID != "" {
		filters["company_id"] = *companyID
	}
	warehouses, err := s.warehouseRepo.FindAll(filters)
	if err != nil {
		return response.NewErrorResponse("Failed to get warehouses")
	}
	return response.NewSuccessResponse(warehouses, "")
}

func (s *WarehouseService) GetWarehouseByID(id string) response.ApiResponse {
	warehouseID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Warehouse not found")
	}
	wh, err := s.warehouseRepo.FindByID(warehouseID)
	if err != nil {
		return response.NewErrorResponse("Failed to get warehouse")
	}
	if wh == nil {
		return response.NewErrorResponse("Warehouse not found")
	}
	return response.NewSuccessResponse(wh, "")
}

type CreateWarehouseInput struct {
	Code      string
	Name      string
	Type      string
	Address   string
	City      string
	Phone     *string
	CompanyID *string
}

func (s *WarehouseService) CreateWarehouse(input CreateWarehouseInput) response.ApiResponse {
	existing, err := s.warehouseRepo.FindByCode(input.Code)
	if err != nil {
		return response.NewErrorResponse("Failed to create warehouse")
	}
	if existing != nil {
		return response.NewErrorResponse("Warehouse code already exists")
	}

	var companyUUID *uuid.UUID
	if input.CompanyID != nil && *input.CompanyID != "" {
		cid, err := uuid.Parse(*input.CompanyID)
		if err == nil {
			companyUUID = &cid
		}
	}

	phone := ""
	if input.Phone != nil {
		phone = *input.Phone
	}

	wh := models.Warehouse{
		ID:        uuid.New(),
		Code:      input.Code,
		Name:      input.Name,
		Type:      models.WarehouseType(input.Type),
		Status:    models.WarehouseStatusActive,
		Address:   input.Address,
		City:      input.City,
		Phone:     phone,
		CompanyID: companyUUID,
	}

	if err := s.warehouseRepo.Create(&wh); err != nil {
		return response.NewErrorResponse("Failed to create warehouse")
	}
	return response.NewSuccessResponse(wh, "")
}

type UpdateWarehouseInput struct {
	Code    *string
	Name    *string
	Type    *string
	Address *string
	City    *string
	Phone   *string
	Status  *string
}

func (s *WarehouseService) UpdateWarehouse(id string, input UpdateWarehouseInput) response.ApiResponse {
	warehouseID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Warehouse not found")
	}

	wh, err := s.warehouseRepo.FindByID(warehouseID)
	if err != nil {
		return response.NewErrorResponse("Failed to update warehouse")
	}
	if wh == nil {
		return response.NewErrorResponse("Warehouse not found")
	}

	updated := false
	if input.Code != nil {
		if *input.Code != wh.Code {
			existing, err := s.warehouseRepo.FindByCode(*input.Code)
			if err == nil && existing != nil && existing.ID != wh.ID {
				return response.NewErrorResponse("Warehouse code already exists")
			}
		}
		wh.Code = *input.Code
		updated = true
	}
	if input.Name != nil {
		wh.Name = *input.Name
		updated = true
	}
	if input.Type != nil {
		wh.Type = models.WarehouseType(*input.Type)
		updated = true
	}
	if input.Address != nil {
		wh.Address = *input.Address
		updated = true
	}
	if input.City != nil {
		wh.City = *input.City
		updated = true
	}
	if input.Phone != nil {
		wh.Phone = *input.Phone
		updated = true
	}
	if input.Status != nil {
		wh.Status = models.WarehouseStatus(*input.Status)
		updated = true
	}

	if !updated {
		return response.NewErrorResponse("No fields to update")
	}

	if err := s.warehouseRepo.Update(wh); err != nil {
		return response.NewErrorResponse("Failed to update warehouse")
	}
	return response.NewSuccessResponse(wh, "")
}

func (s *WarehouseService) DeleteWarehouse(id string) response.ApiResponse {
	warehouseID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Warehouse not found")
	}
	wh, err := s.warehouseRepo.FindByID(warehouseID)
	if err != nil {
		return response.NewErrorResponse("Failed to delete warehouse")
	}
	if wh == nil {
		return response.NewErrorResponse("Warehouse not found")
	}

	wh.Status = models.WarehouseStatusInactive
	if err := s.warehouseRepo.Update(wh); err != nil {
		return response.NewErrorResponse("Failed to delete warehouse")
	}

	return response.NewSuccessResponse(nil, "Warehouse deleted successfully")
}
