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

type AppointmentService struct {
	appointmentRepo *repository.AppointmentRepository
}

func parseAppointmentDate(value string) (time.Time, error) {
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

func parseAppointmentTime(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("empty time")
	}

	layouts := []string{
		"15:04",
		"15:04:05",
		"2006-01-02T15:04:05",
		time.RFC3339,
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, trimmed)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid time format")
}

func NewAppointmentService(appointmentRepo *repository.AppointmentRepository) *AppointmentService {
	return &AppointmentService{appointmentRepo: appointmentRepo}
}

func (s *AppointmentService) CreateAppointment(input request.CreateAppointmentRequest, companyID string) response.ApiResponse {
	parsedCompanyID, err := uuid.Parse(companyID)
	if err != nil {
		return response.NewErrorResponse("Invalid company id")
	}

	parsedPatientID, err := uuid.Parse(input.PatientID)
	if err != nil {
		return response.NewErrorResponse("Invalid patient id")
	}

	bookingDate, err := parseAppointmentDate(input.BookingDate)
	if err != nil {
		return response.NewErrorResponse("Format booking date tidak valid. Gunakan YYYY-MM-DD")
	}

	startTime, err := parseAppointmentTime(input.StartTime)
	if err != nil {
		return response.NewErrorResponse("Format start time tidak valid. Gunakan HH:MM")
	}

	endTime, err := parseAppointmentTime(input.EndTime)
	if err != nil {
		return response.NewErrorResponse("Format end time tidak valid. Gunakan HH:MM")
	}

	if !endTime.After(startTime) {
		return response.NewErrorResponse("End time harus lebih besar dari start time")
	}

	appointment := &models.Appointment{
		CompanyID:   parsedCompanyID,
		PatientID:   parsedPatientID,
		BookingDate: bookingDate,
		StartTime:   startTime,
		EndTime:     endTime,
		Status:      models.AppointmentStatusScheduled,
		Notes:       input.Notes,
	}

	if input.TreatmentID != "" {
		parsedTreatmentID, err := uuid.Parse(input.TreatmentID)
		if err != nil {
			return response.NewErrorResponse("Invalid treatment id")
		}
		appointment.TreatmentID = &parsedTreatmentID
	}

	if input.TherapistID != "" {
		parsedTherapistID, err := uuid.Parse(input.TherapistID)
		if err != nil {
			return response.NewErrorResponse("Invalid therapist id")
		}
		appointment.TherapistID = &parsedTherapistID

		// Check conflict
		hasConflict, err := s.appointmentRepo.CheckConflict(input.TherapistID, input.BookingDate, startTime, endTime, "")
		if err != nil {
			return response.NewErrorResponse("Failed to check conflict: " + err.Error())
		}
		if hasConflict {
			return response.NewErrorResponse("Therapist sudah memiliki appointment pada waktu tersebut")
		}
	}

	if err := s.appointmentRepo.Create(appointment); err != nil {
		return response.NewErrorResponse("Failed to create appointment: " + err.Error())
	}

	return response.NewSuccessResponse(appointment, "Appointment created successfully")
}

func (s *AppointmentService) GetAppointments(companyID string, filters map[string]interface{}, limit, offset int) response.PaginatedResponse {
	appointments, total, err := s.appointmentRepo.GetAll(companyID, filters, limit, offset)
	if err != nil {
		return response.PaginatedResponse{Success: false, Data: nil, Pagination: response.Pagination{}}
	}

	hasMore := int64(offset+limit) < total
	return response.PaginatedResponse{
		Success: true,
		Data:    appointments,
		Pagination: response.Pagination{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: hasMore,
		},
	}
}

func (s *AppointmentService) GetAppointmentByID(id, companyID string) response.ApiResponse {
	appointment, err := s.appointmentRepo.GetByID(id, companyID)
	if err != nil {
		return response.NewErrorResponse(err.Error())
	}

	return response.NewSuccessResponse(appointment, "Appointment fetched successfully")
}

func (s *AppointmentService) UpdateAppointment(id string, input request.UpdateAppointmentRequest, companyID string) response.ApiResponse {
	appointment, err := s.appointmentRepo.GetByID(id, companyID)
	if err != nil {
		return response.NewErrorResponse(err.Error())
	}

	if input.PatientID != nil {
		parsedPatientID, err := uuid.Parse(*input.PatientID)
		if err != nil {
			return response.NewErrorResponse("Invalid patient id")
		}
		appointment.PatientID = parsedPatientID
	}

	if input.TreatmentID != nil {
		if *input.TreatmentID == "" {
			appointment.TreatmentID = nil
		} else {
			parsedTreatmentID, err := uuid.Parse(*input.TreatmentID)
			if err != nil {
				return response.NewErrorResponse("Invalid treatment id")
			}
			appointment.TreatmentID = &parsedTreatmentID
		}
	}

	if input.TherapistID != nil {
		if *input.TherapistID == "" {
			appointment.TherapistID = nil
		} else {
			parsedTherapistID, err := uuid.Parse(*input.TherapistID)
			if err != nil {
				return response.NewErrorResponse("Invalid therapist id")
			}
			appointment.TherapistID = &parsedTherapistID
		}
	}

	if input.BookingDate != nil {
		bookingDate, err := parseAppointmentDate(*input.BookingDate)
		if err != nil {
			return response.NewErrorResponse("Format booking date tidak valid. Gunakan YYYY-MM-DD")
		}
		appointment.BookingDate = bookingDate
	}

	if input.StartTime != nil {
		startTime, err := parseAppointmentTime(*input.StartTime)
		if err != nil {
			return response.NewErrorResponse("Format start time tidak valid. Gunakan HH:MM")
		}
		appointment.StartTime = startTime
	}

	if input.EndTime != nil {
		endTime, err := parseAppointmentTime(*input.EndTime)
		if err != nil {
			return response.NewErrorResponse("Format end time tidak valid. Gunakan HH:MM")
		}
		appointment.EndTime = endTime
	}

	if input.Status != nil {
		appointment.Status = *input.Status
	}

	if input.Notes != nil {
		appointment.Notes = *input.Notes
	}

	if !appointment.EndTime.After(appointment.StartTime) {
		return response.NewErrorResponse("End time harus lebih besar dari start time")
	}

	if appointment.TherapistID != nil {
		therapistID := appointment.TherapistID.String()
		bookingDate := appointment.BookingDate.Format("2006-01-02")
		hasConflict, err := s.appointmentRepo.CheckConflict(therapistID, bookingDate, appointment.StartTime, appointment.EndTime, id)
		if err != nil {
			return response.NewErrorResponse("Failed to check conflict: " + err.Error())
		}
		if hasConflict {
			return response.NewErrorResponse("Therapist sudah memiliki appointment pada waktu tersebut")
		}
	}

	if err := s.appointmentRepo.Update(appointment); err != nil {
		return response.NewErrorResponse("Failed to update appointment: " + err.Error())
	}

	return response.NewSuccessResponse(appointment, "Appointment updated successfully")
}

func (s *AppointmentService) DeleteAppointment(id, companyID string) response.ApiResponse {
	if err := s.appointmentRepo.Delete(id, companyID); err != nil {
		return response.NewErrorResponse("Failed to delete appointment: " + err.Error())
	}

	return response.NewSuccessResponse(nil, "Appointment deleted successfully")
}
