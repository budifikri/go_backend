package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/repository"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
)

type JadwalDokterService struct {
	repo *repository.JadwalDokterRepository
}

func NewJadwalDokterService(repo *repository.JadwalDokterRepository) *JadwalDokterService {
	return &JadwalDokterService{repo: repo}
}

var validHari = map[string]bool{
	"Senin": true, "Selasa": true, "Rabu": true, "Kamis": true,
	"Jumat": true, "Sabtu": true, "Minggu": true,
}

func validateTimeFormat(t string) bool {
	_, err := time.Parse("15:04", t)
	return err == nil
}

func (s *JadwalDokterService) CreateJadwalDokter(input request.CreateJadwalDokterRequest, companyID string) response.ApiResponse {
	// Validate dokter_id
	dokterID, err := uuid.Parse(input.DokterID)
	if err != nil {
		return response.NewErrorResponse("Invalid dokter_id format")
	}

	// Validate company_id
	cid, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Invalid company_id")
	}

	// Validate hari
	if !validHari[input.Hari] {
		return response.NewErrorResponse("Invalid hari. Must be one of: Senin, Selasa, Rabu, Kamis, Jumat, Sabtu, Minggu")
	}

	// Validate jam format
	if !validateTimeFormat(input.JamMulai) {
		return response.NewErrorResponse("Invalid jam_mulai format. Use HH:MM")
	}
	if !validateTimeFormat(input.JamSelesai) {
		return response.NewErrorResponse("Invalid jam_selesai format. Use HH:MM")
	}

	// Validate jam_mulai < jam_selesai
	start, _ := time.Parse("15:04", input.JamMulai)
	end, _ := time.Parse("15:04", input.JamSelesai)
	if !start.Before(end) {
		return response.NewErrorResponse("jam_mulai must be before jam_selesai")
	}

	eligible, err := s.repo.IsEligibleDokter(dokterID, cid)
	if err != nil {
		return response.NewErrorResponse(fmt.Sprintf("Failed to validate dokter: %v", err))
	}
	if !eligible {
		return response.NewErrorResponse("Dokter tidak valid. Hanya dokter aktif dengan tipe Dokter yang bisa dipilih")
	}

	data := map[string]interface{}{
		"dokter_id":   dokterID,
		"company_id":  cid,
		"hari":        input.Hari,
		"jam_mulai":   input.JamMulai,
		"jam_selesai": input.JamSelesai,
		"is_active":   true,
	}

	created, err := s.repo.Create(data)
	if err != nil {
		return response.NewErrorResponse(fmt.Sprintf("Failed to create jadwal dokter: %v", err))
	}

	return response.NewSuccessResponse(created, "Jadwal dokter created successfully")
}

func (s *JadwalDokterService) GetJadwals(companyID string, filters map[string]interface{}, limit, offset int) response.PaginatedResponse {
	cid, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewPaginatedResponse(nil, 0, limit, offset)
	}

	rows, total, err := s.repo.FindAll(cid, filters, limit, offset)
	if err != nil {
		return response.NewPaginatedResponse(nil, 0, limit, offset)
	}

	return response.NewPaginatedResponse(rows, total, limit, offset)
}

func (s *JadwalDokterService) GetJadwalByID(id, companyID string) response.ApiResponse {
	cid, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Invalid company_id")
	}

	row, err := s.repo.FindByID(id, cid)
	if err != nil {
		if errors.Is(err, errors.New("jadwal dokter not found")) {
			return response.NewErrorResponse("Jadwal dokter not found")
		}
		return response.NewErrorResponse(fmt.Sprintf("Failed to get jadwal dokter: %v", err))
	}

	return response.NewSuccessResponse(row, "Jadwal dokter fetched successfully")
}

func (s *JadwalDokterService) UpdateJadwalDokter(id string, input request.UpdateJadwalDokterRequest, companyID string) response.ApiResponse {
	cid, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Invalid company_id")
	}

	// Check if exists
	_, err = s.repo.FindByID(id, cid)
	if err != nil {
		return response.NewErrorResponse("Jadwal dokter not found")
	}

	updates := map[string]interface{}{}

	if input.DokterID != nil {
		dokterID, err := uuid.Parse(*input.DokterID)
		if err != nil {
			return response.NewErrorResponse("Invalid dokter_id format")
		}
		eligible, err := s.repo.IsEligibleDokter(dokterID, cid)
		if err != nil {
			return response.NewErrorResponse(fmt.Sprintf("Failed to validate dokter: %v", err))
		}
		if !eligible {
			return response.NewErrorResponse("Dokter tidak valid. Hanya dokter aktif dengan tipe Dokter yang bisa dipilih")
		}
		updates["dokter_id"] = dokterID
	}

	if input.Hari != nil {
		if !validHari[*input.Hari] {
			return response.NewErrorResponse("Invalid hari")
		}
		updates["hari"] = *input.Hari
	}

	if input.JamMulai != nil {
		if !validateTimeFormat(*input.JamMulai) {
			return response.NewErrorResponse("Invalid jam_mulai format")
		}
		updates["jam_mulai"] = *input.JamMulai
	}

	if input.JamSelesai != nil {
		if !validateTimeFormat(*input.JamSelesai) {
			return response.NewErrorResponse("Invalid jam_selesai format")
		}
		updates["jam_selesai"] = *input.JamSelesai
	}

	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}

	// Validate start < end if both are being updated
	jamMulai := updates["jam_mulai"]
	jamSelesai := updates["jam_selesai"]
	if jamMulai != nil && jamSelesai != nil {
		start, _ := time.Parse("15:04", jamMulai.(string))
		end, _ := time.Parse("15:04", jamSelesai.(string))
		if !start.Before(end) {
			return response.NewErrorResponse("jam_mulai must be before jam_selesai")
		}
	}

	if err := s.repo.Update(id, updates, cid); err != nil {
		return response.NewErrorResponse(fmt.Sprintf("Failed to update jadwal dokter: %v", err))
	}

	updated, _ := s.repo.FindByID(id, cid)
	return response.NewSuccessResponse(updated, "Jadwal dokter updated successfully")
}

func (s *JadwalDokterService) DeleteJadwalDokter(id, companyID string) response.ApiResponse {
	cid, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Invalid company_id")
	}

	// Check dependencies
	hasDeps, err := s.repo.CheckDependencies(id, cid)
	if err != nil {
		return response.NewErrorResponse(fmt.Sprintf("Failed to check dependencies: %v", err))
	}
	if hasDeps {
		return response.NewErrorResponse("Cannot delete jadwal dokter because it has dependencies")
	}

	if err := s.repo.Delete(id, cid); err != nil {
		return response.NewErrorResponse(fmt.Sprintf("Failed to delete jadwal dokter: %v", err))
	}

	return response.NewSuccessResponse(nil, "Jadwal dokter deleted successfully")
}
