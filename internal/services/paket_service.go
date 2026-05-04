package services

import (
	"github.com/google/uuid"

	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/repository"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type PaketService struct {
	paketRepo *repository.PaketRepository
}

func NewPaketService(paketRepo *repository.PaketRepository) *PaketService {
	return &PaketService{paketRepo: paketRepo}
}

func (s *PaketService) CreatePaket(input request.CreatePaketRequest, companyID string) response.ApiResponse {
	parsedCompanyID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Invalid company id")
	}

	// Validate items
	if len(input.Items) == 0 {
		return response.NewErrorResponse("Paket must have at least one product")
	}

	// Check all products exist
	for _, item := range input.Items {
		produkID, err := uuid.Parse(item.IDProduk)
		if err != nil {
			return response.NewErrorResponse("Invalid product id: " + item.IDProduk)
		}
		exists, err := s.paketRepo.CheckProdukExists(produkID)
		if err != nil {
			return response.NewErrorResponse("Failed to check product: " + err.Error())
		}
		if !exists {
			return response.NewErrorResponse("Product not found: " + item.IDProduk)
		}
	}

	// Set default values
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	paket := &models.Paket{
		KodePaket: input.KodePaket,
		NmPaket:   input.NmPaket,
		Deskripsi: input.Deskripsi,
		IsActive:  isActive,
		CompanyID: &parsedCompanyID,
	}

	// Build details
	details := make([]models.DetailPaket, 0, len(input.Items))
	for _, item := range input.Items {
		produkID, _ := uuid.Parse(item.IDProduk)
		details = append(details, models.DetailPaket{
			IDProduk: produkID,
		})
	}

	// Create with transaction
	if err := s.paketRepo.Create(paket, details); err != nil {
		return response.NewErrorResponse("Failed to create paket: " + err.Error())
	}

	// Calculate and update harga_paket
	if err := s.calculateAndUpdateHarga(paket.ID); err != nil {
		return response.NewErrorResponse("Paket created but failed to calculate harga: " + err.Error())
	}

	// Fetch complete data
	result, err := s.paketRepo.GetByID(paket.ID, companyID)
	if err != nil {
		return response.NewErrorResponse("Paket created but failed to fetch: " + err.Error())
	}

	return response.NewSuccessResponse(result, "Paket created successfully")
}

func (s *PaketService) GetPaket(id, companyID string) response.ApiResponse {
	paketID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Invalid paket id")
	}

	paket, err := s.paketRepo.GetByID(paketID, companyID)
	if err != nil {
		return response.NewErrorResponse(err.Error())
	}

	return response.NewSuccessResponse(paket, "Paket fetched successfully")
}

func (s *PaketService) GetPakets(companyID string, filters map[string]interface{}, limit, offset int) response.PaginatedResponse {
	pakets, total, err := s.paketRepo.GetAll(companyID, filters, limit, offset)
	if err != nil {
		return response.PaginatedResponse{Success: false, Data: nil, Pagination: response.Pagination{}}
	}

	hasMore := int64(offset+limit) < total
	return response.PaginatedResponse{
		Success: true,
		Data:    pakets,
		Pagination: response.Pagination{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: hasMore,
		},
	}
}

func (s *PaketService) UpdatePaket(id string, input request.UpdatePaketRequest, companyID string) response.ApiResponse {
	paketID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Invalid paket id")
	}

	paket, err := s.paketRepo.GetByID(paketID, companyID)
	if err != nil {
		return response.NewErrorResponse(err.Error())
	}

	// Update fields
	if input.KodePaket != nil {
		paket.KodePaket = *input.KodePaket
	}
	if input.NmPaket != nil {
		paket.NmPaket = *input.NmPaket
	}
	if input.Deskripsi != nil {
		paket.Deskripsi = *input.Deskripsi
	}
	if input.IsActive != nil {
		paket.IsActive = *input.IsActive
	}

	// Build details if provided
	var details []models.DetailPaket
	if len(input.Items) > 0 {
		details = make([]models.DetailPaket, 0, len(input.Items))
		for _, item := range input.Items {
			produkID, err := uuid.Parse(item.IDProduk)
			if err != nil {
				return response.NewErrorResponse("Invalid product id: " + item.IDProduk)
			}
			// Check product exists
			exists, err := s.paketRepo.CheckProdukExists(produkID)
			if err != nil {
				return response.NewErrorResponse("Failed to check product: " + err.Error())
			}
			if !exists {
				return response.NewErrorResponse("Product not found: " + item.IDProduk)
			}
			details = append(details, models.DetailPaket{
				IDProduk: produkID,
			})
		}
	}

	// Update with transaction
	if err := s.paketRepo.Update(paket, details); err != nil {
		return response.NewErrorResponse("Failed to update paket: " + err.Error())
	}

	// Calculate and update harga_paket if items were updated
	if len(input.Items) > 0 {
		if err := s.calculateAndUpdateHarga(paket.ID); err != nil {
			return response.NewErrorResponse("Paket updated but failed to calculate harga: " + err.Error())
		}
	}

	// Fetch complete data
	result, err := s.paketRepo.GetByID(paket.ID, companyID)
	if err != nil {
		return response.NewErrorResponse("Paket updated but failed to fetch: " + err.Error())
	}

	return response.NewSuccessResponse(result, "Paket updated successfully")
}

func (s *PaketService) DeletePaket(id, companyID string) response.ApiResponse {
	paketID, err := uuid.Parse(id)
	if err != nil {
		return response.NewErrorResponse("Invalid paket id")
	}

	if err := s.paketRepo.Delete(paketID, companyID); err != nil {
		return response.NewErrorResponse("Failed to delete paket: " + err.Error())
	}

	return response.NewSuccessResponse(nil, "Paket deleted successfully")
}

// calculateAndUpdateHarga calculates total harga from detail products and updates paket
func (s *PaketService) calculateAndUpdateHarga(paketID uuid.UUID) error {
	total, err := s.paketRepo.CalculateTotalHarga(paketID)
	if err != nil {
		return err
	}

	// Update harga_paket directly using repo
	return s.paketRepo.Update(&models.Paket{
		ID:         paketID,
		HargaPaket: total,
	}, nil)
}
