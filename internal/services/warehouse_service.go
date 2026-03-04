package services

import (
	"github.com/google/uuid"
	applogger "github.com/pos-retail/go_backend/internal/logger"
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

func (s *WarehouseService) GetWarehouses(companyID *string, isActive *bool, search string, limit, offset int) response.PaginatedResponse {
	filters := map[string]interface{}{}
	if companyID != nil && *companyID != "" {
		filters["company_id"] = *companyID
	}
	if isActive != nil {
		filters["is_active"] = *isActive
	}
	if search != "" {
		filters["search"] = search
	}
	warehouses, total, err := s.warehouseRepo.FindAll(filters, limit, offset)
	if err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}
	return response.NewPaginatedResponse(warehouses, total, limit, offset)
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
		IsActive:  true,
		Address:   input.Address,
		City:      input.City,
		Phone:     phone,
		CompanyID: companyUUID,
	}

	if err := s.warehouseRepo.Create(&wh); err != nil {
		if l := applogger.Default(); l != nil {
			l.LogError(applogger.ActionCreate, "warehouses", "", warehouseCompanyIDString(companyUUID), wh.ID.String(), err)
		}
		return response.NewErrorResponse("Failed to create warehouse")
	}
	if l := applogger.Default(); l != nil {
		l.Log(applogger.ActionCreate, "warehouses", "", warehouseCompanyIDString(companyUUID), wh.ID.String(), nil, wh)
	}
	return response.NewSuccessResponse(wh, "")
}

type UpdateWarehouseInput struct {
	Code     *string
	Name     *string
	Type     *string
	Address  *string
	City     *string
	Phone    *string
	IsActive *bool
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
	oldWarehouse := *wh

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
	if input.IsActive != nil {
		wh.IsActive = *input.IsActive
		if *input.IsActive {
			wh.Status = models.WarehouseStatusActive
		} else {
			wh.Status = models.WarehouseStatusInactive
		}
		updated = true
	}

	if !updated {
		return response.NewErrorResponse("No fields to update")
	}

	if err := s.warehouseRepo.Update(wh); err != nil {
		if l := applogger.Default(); l != nil {
			l.LogError(applogger.ActionUpdate, "warehouses", "", warehouseCompanyIDString(wh.CompanyID), wh.ID.String(), err)
		}
		return response.NewErrorResponse("Failed to update warehouse")
	}
	if l := applogger.Default(); l != nil {
		l.Log(applogger.ActionUpdate, "warehouses", "", warehouseCompanyIDString(wh.CompanyID), wh.ID.String(), oldWarehouse, wh)
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

	if err := s.warehouseRepo.Delete(wh.ID); err != nil {
		if l := applogger.Default(); l != nil {
			l.LogError(applogger.ActionDelete, "warehouses", "", warehouseCompanyIDString(wh.CompanyID), wh.ID.String(), err)
		}
		return response.NewErrorResponse("Failed to delete warehouse")
	}
	if l := applogger.Default(); l != nil {
		l.Log(applogger.ActionDelete, "warehouses", "", warehouseCompanyIDString(wh.CompanyID), wh.ID.String(), wh, nil)
	}

	return response.NewSuccessResponse(nil, "Warehouse deleted successfully")
}

func warehouseCompanyIDString(v *uuid.UUID) string {
	if v == nil {
		return ""
	}
	return v.String()
}
