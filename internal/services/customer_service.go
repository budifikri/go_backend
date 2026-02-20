package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/repository"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type CustomerService struct {
	customerRepo *repository.CustomerRepository
}

func NewCustomerService(customerRepo *repository.CustomerRepository) *CustomerService {
	return &CustomerService{customerRepo: customerRepo}
}

func (s *CustomerService) GetCustomers(filters map[string]interface{}, limit, offset int, companyID string) response.PaginatedResponse {
	cid, err := uuid.Parse(companyID)
	if err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}

	rows, total, err := s.customerRepo.FindCustomers(filters, limit, offset, cid)
	if err != nil {
		return response.PaginatedResponse{Success: false, Data: []interface{}{}, Pagination: response.Pagination{Total: 0, Limit: limit, Offset: offset, HasMore: false}}
	}

	return response.NewPaginatedResponse(rows, total, limit, offset)
}

func (s *CustomerService) GetCustomerByID(id string, companyID string) response.ApiResponse {
	cid, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Customer not found")
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Customer not found")
	}

	row, err := s.customerRepo.GetCustomerByID(uid, cid)
	if err != nil {
		return response.NewErrorResponse("Customer not found")
	}
	if row == nil {
		return response.NewErrorResponse("Customer not found")
	}

	return response.NewSuccessResponse(row, "")
}

func (s *CustomerService) CreateCustomer(input map[string]interface{}, companyID string) response.ApiResponse {
	cid, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Failed to create customer")
	}

	// Match TS: customer_code = `CUST-${new Date().toISOString().replace(/[-:]/g, '').slice(0, 14)}`
	ts := time.Now().UTC().Format("20060102T150405")
	if len(ts) > 14 {
		ts = ts[:14]
	}
	code := "CUST-" + ts

	data := map[string]interface{}{}
	for k, v := range input {
		if v != nil {
			data[k] = v
		}
	}
	data["customer_code"] = code
	data["company_id"] = cid
	if _, ok := data["tier"]; !ok {
		data["tier"] = "BRONZE"
	}
	data["is_active"] = true
	data["status"] = "active"
	data["loyalty_points"] = 0
	if _, ok := data["credit_limit"]; !ok {
		data["credit_limit"] = 0
	}
	data["credit_balance"] = 0
	data["total_purchases"] = 0
	data["created_at"] = time.Now()
	data["updated_at"] = time.Now()

	created, err := s.customerRepo.CreateCustomer(data)
	if err != nil {
		return response.NewErrorResponse(err.Error())
	}
	return response.NewSuccessResponse(created, "Customer created successfully")
}

func (s *CustomerService) UpdateCustomer(id string, companyID string, updates map[string]interface{}) response.ApiResponse {
	cid, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Customer not found")
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Customer not found")
	}

	row, err := s.customerRepo.UpdateCustomer(uid, cid, updates)
	if err != nil {
		return response.NewErrorResponse(err.Error())
	}
	if row == nil {
		return response.NewErrorResponse("Customer not found")
	}
	return response.NewSuccessResponse(row, "Customer updated successfully")
}

func (s *CustomerService) DeleteCustomer(id string, companyID string) response.ApiResponse {
	cid, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Customer not found")
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Customer not found")
	}

	affected, err := s.customerRepo.DeleteCustomer(uid, cid)
	if err != nil {
		return response.NewErrorResponse(err.Error())
	}
	if affected == 0 {
		return response.NewErrorResponse("Customer not found")
	}
	return response.NewSuccessResponse(nil, "Customer deleted successfully")
}
