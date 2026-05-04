package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/repository"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type DokterService struct {
	dokterRepo *repository.DokterRepository
}

func parseDokterDate(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}

	layouts := []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, trimmed)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date format")
}

func NewDokterService(dokterRepo *repository.DokterRepository) *DokterService {
	return &DokterService{dokterRepo: dokterRepo}
}

func (s *DokterService) CreateDokter(input request.CreateDokterRequest, companyID string) response.ApiResponse {
	parsedCompanyID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Invalid company id")
	}

	tanggalLahir, err := parseDokterDate(input.TanggalLahir)
	if err != nil {
		return response.NewErrorResponse("Format tanggal lahir tidak valid. Gunakan YYYY-MM-DD")
	}

	active := true
	if input.Active != nil {
		active = *input.Active
	}

	dokter := &models.Dokter{
		CompanyID:    parsedCompanyID,
		Nama:         input.Nama,
		JenisKelamin: input.JenisKelamin,
		TempatLahir:  input.TempatLahir,
		TanggalLahir: tanggalLahir,
		Alamat:       input.Alamat,
		NoTelp:       input.NoTelp,
		Email:        input.Email,
		Tipe:         input.Tipe,
		Active:       active,
	}

	if err := s.dokterRepo.Create(dokter); err != nil {
		return response.NewErrorResponse("Failed to create dokter: " + err.Error())
	}

	return response.NewSuccessResponse(dokter, "Dokter created successfully")
}

func (s *DokterService) GetDokters(companyID string, filters map[string]interface{}, limit, offset int) response.PaginatedResponse {
	dokters, total, err := s.dokterRepo.GetAll(companyID, filters, limit, offset)
	if err != nil {
		return response.PaginatedResponse{Success: false, Data: nil, Pagination: response.Pagination{}}
	}

	hasMore := int64(offset+limit) < total
	return response.PaginatedResponse{
		Success: true,
		Data:    dokters,
		Pagination: response.Pagination{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: hasMore,
		},
	}
}

func (s *DokterService) GetDokterByID(id, companyID string) response.ApiResponse {
	dokter, err := s.dokterRepo.GetByID(id, companyID)
	if err != nil {
		return response.NewErrorResponse(err.Error())
	}

	return response.NewSuccessResponse(dokter, "Dokter fetched successfully")
}

func (s *DokterService) UpdateDokter(id string, input request.UpdateDokterRequest, companyID string) response.ApiResponse {
	dokter, err := s.dokterRepo.GetByID(id, companyID)
	if err != nil {
		return response.NewErrorResponse(err.Error())
	}

	if input.Nama != nil {
		dokter.Nama = *input.Nama
	}
	if input.JenisKelamin != nil {
		dokter.JenisKelamin = *input.JenisKelamin
	}
	if input.TempatLahir != nil {
		dokter.TempatLahir = *input.TempatLahir
	}
	if input.TanggalLahir != nil {
		tanggalLahir, err := parseDokterDate(*input.TanggalLahir)
		if err != nil {
			return response.NewErrorResponse("Format tanggal lahir tidak valid. Gunakan YYYY-MM-DD")
		}
		dokter.TanggalLahir = tanggalLahir
	}
	if input.Alamat != nil {
		dokter.Alamat = *input.Alamat
	}
	if input.NoTelp != nil {
		dokter.NoTelp = *input.NoTelp
	}
	if input.Email != nil {
		dokter.Email = *input.Email
	}
	if input.Tipe != nil {
		dokter.Tipe = *input.Tipe
	}
	if input.Active != nil {
		dokter.Active = *input.Active
	}

	if err := s.dokterRepo.Update(dokter); err != nil {
		return response.NewErrorResponse("Failed to update dokter: " + err.Error())
	}

	return response.NewSuccessResponse(dokter, "Dokter updated successfully")
}

func (s *DokterService) DeleteDokter(id, companyID string) response.ApiResponse {
	hasDeps, err := s.dokterRepo.CheckDependencies(id, companyID)
	if err != nil {
		return response.NewErrorResponse("Failed to check dependencies: " + err.Error())
	}
	if !hasDeps {
		return response.NewErrorResponse("Dokter not found")
	}

	if err := s.dokterRepo.Delete(id, companyID); err != nil {
		return response.NewErrorResponse("Failed to delete dokter: " + err.Error())
	}

	return response.NewSuccessResponse(nil, "Dokter deleted successfully")
}
