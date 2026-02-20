package services

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/repository"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type SupplierService struct {
	supplierRepo *repository.SupplierRepository
}

func NewSupplierService(supplierRepo *repository.SupplierRepository) *SupplierService {
	return &SupplierService{supplierRepo: supplierRepo}
}

func (s *SupplierService) GetSuppliers(filters map[string]interface{}, limit, offset int, companyID string) response.PaginatedResponse {
	if companyID != "" {
		filters["company_id"] = companyID
	}
	rows, total, err := s.supplierRepo.FindSuppliers(filters, limit, offset)
	if err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}
	return response.NewPaginatedResponse(rows, total, limit, offset)
}

func (s *SupplierService) GetSupplierByID(id string, companyID string) response.ApiResponse {
	uid, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Supplier not found")
	}
	var cid *uuid.UUID
	if companyID != "" {
		c, err := uuid.Parse(companyID)
		if err == nil {
			cid = &c
		}
	}

	row, err := s.supplierRepo.GetSupplierByID(uid, cid)
	if err != nil {
		return response.NewErrorResponse("Supplier not found")
	}
	if row == nil {
		return response.NewErrorResponse("Supplier not found")
	}
	return response.NewSuccessResponse(row, "")
}

func (s *SupplierService) CreateSupplier(input map[string]interface{}, companyID string) response.ApiResponse {
	cid, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Failed to create supplier")
	}

	code := "SUP-" + strconv.FormatInt(time.Now().UnixMilli(), 10)
	data := map[string]interface{}{}
	for k, v := range input {
		if v != nil {
			data[k] = v
		}
	}
	data["code"] = code
	data["company_id"] = cid
	if _, ok := data["payment_terms"]; !ok {
		data["payment_terms"] = "NET_30"
	}
	if _, ok := data["credit_limit"]; !ok {
		data["credit_limit"] = "0"
	}
	data["status"] = "active"
	data["created_at"] = time.Now()
	data["updated_at"] = time.Now()

	created, err := s.supplierRepo.CreateSupplier(data)
	if err != nil {
		return response.NewErrorResponse(err.Error())
	}
	return response.NewSuccessResponse(created, "Supplier created successfully")
}

func (s *SupplierService) UpdateSupplier(id string, companyID string, updates map[string]interface{}) response.ApiResponse {
	cid, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Supplier not found")
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Supplier not found")
	}

	row, err := s.supplierRepo.UpdateSupplier(uid, cid, updates)
	if err != nil {
		return response.NewErrorResponse(err.Error())
	}
	if row == nil {
		return response.NewErrorResponse("Supplier not found")
	}
	return response.NewSuccessResponse(row, "Supplier updated successfully")
}

func (s *SupplierService) DeactivateSupplier(id string, companyID string) response.ApiResponse {
	cid, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Supplier not found")
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Supplier not found")
	}

	if err := s.supplierRepo.DeactivateSupplier(uid, cid); err != nil {
		return response.NewErrorResponse(err.Error())
	}
	return response.NewSuccessResponse(nil, "Supplier deactivated successfully")
}

func (s *SupplierService) DeleteSupplier(id string, companyID string) response.ApiResponse {
	cid, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Supplier not found")
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Supplier not found")
	}

	affected, err := s.supplierRepo.DeleteSupplier(uid, cid)
	if err != nil {
		return response.NewErrorResponse("Failed to delete supplier")
	}
	if affected == 0 {
		return response.NewErrorResponse("Supplier not found")
	}
	return response.NewSuccessResponse(nil, "Supplier deleted successfully")
}
