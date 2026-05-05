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

var appointmentLocation = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.FixedZone("GMT+7", 7*60*60)
	}
	return loc
}()

type AppointmentService struct {
	appointmentRepo *repository.AppointmentRepository
	customerRepo    *repository.CustomerRepository
	dokterRepo      *repository.DokterRepository
	treatmentRepo   *repository.TreatmentRepository
}

type AppointmentResponse struct {
	ID          uuid.UUID                `json:"id"`
	CompanyID   uuid.UUID                `json:"company_id"`
	PatientID   uuid.UUID                `json:"patient_id"`
	TreatmentID uuid.UUID                `json:"treatment_id"`
	TherapistID uuid.UUID                `json:"therapist_id"`
	BookingDate string                   `json:"booking_date"`
	StartTime   string                   `json:"start_time"`
	EndTime     string                   `json:"end_time"`
	Status      models.AppointmentStatus `json:"status"`
	Notes       string                   `json:"notes"`
	CreatedAt   string                   `json:"created_at"`
	UpdatedAt   string                   `json:"updated_at"`
	Patient     *models.Customer         `json:"patient,omitempty"`
	Treatment   *models.Treatment        `json:"treatment,omitempty"`
	Therapist   *models.Dokter           `json:"therapist,omitempty"`
}

func toAppointmentResponse(appointment *models.Appointment) *AppointmentResponse {
	if appointment == nil {
		return nil
	}

	return &AppointmentResponse{
		ID:          appointment.ID,
		CompanyID:   appointment.CompanyID,
		PatientID:   appointment.PatientID,
		TreatmentID: appointment.TreatmentID,
		TherapistID: appointment.TherapistID,
		BookingDate: appointment.BookingDate.In(appointmentLocation).Format("2006-01-02"),
		StartTime:   formatAppointmentTime(appointment.StartTime),
		EndTime:     formatAppointmentTime(appointment.EndTime),
		Status:      appointment.Status,
		Notes:       appointment.Notes,
		CreatedAt:   appointment.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   appointment.UpdatedAt.Format(time.RFC3339),
		Patient:     appointment.Patient,
		Treatment:   appointment.Treatment,
		Therapist:   appointment.Therapist,
	}
}

func toAppointmentResponseList(appointments []models.Appointment) []AppointmentResponse {
	items := make([]AppointmentResponse, 0, len(appointments))
	for i := range appointments {
		mapped := toAppointmentResponse(&appointments[i])
		if mapped != nil {
			items = append(items, *mapped)
		}
	}
	return items
}

func parseAppointmentDate(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	if parsed, err := time.ParseInLocation("2006-01-02", trimmed, appointmentLocation); err == nil {
		return parsed, nil
	}
	if parsed, err := time.Parse(time.RFC3339, trimmed); err == nil {
		return parsed.In(appointmentLocation), nil
	}
	if parsed, err := time.ParseInLocation("2006-01-02T15:04:05", trimmed, appointmentLocation); err == nil {
		return parsed, nil
	}
	if parsed, err := time.ParseInLocation("2006-01-02 15:04:05", trimmed, appointmentLocation); err == nil {
		return parsed, nil
	}

	return time.Time{}, fmt.Errorf("invalid date format")
}

func parseAppointmentTime(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("empty time")
	}
	if parsed, err := time.ParseInLocation("15:04", trimmed, appointmentLocation); err == nil {
		return parsed, nil
	}
	if parsed, err := time.ParseInLocation("15:04:05", trimmed, appointmentLocation); err == nil {
		return parsed, nil
	}
	if parsed, err := time.ParseInLocation("2006-01-02T15:04:05", trimmed, appointmentLocation); err == nil {
		return parsed, nil
	}
	if parsed, err := time.Parse(time.RFC3339, trimmed); err == nil {
		return parsed.In(appointmentLocation), nil
	}

	return time.Time{}, fmt.Errorf("invalid time format")
}

func formatAppointmentTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	localized := time.Date(2000, time.January, 1, value.Hour(), value.Minute(), 0, 0, appointmentLocation)
	return localized.Format("15:04")
}

func NewAppointmentService(
	appointmentRepo *repository.AppointmentRepository,
	customerRepo *repository.CustomerRepository,
	dokterRepo *repository.DokterRepository,
	treatmentRepo *repository.TreatmentRepository,
) *AppointmentService {
	return &AppointmentService{
		appointmentRepo: appointmentRepo,
		customerRepo:    customerRepo,
		dokterRepo:      dokterRepo,
		treatmentRepo:   treatmentRepo,
	}
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
	patient, err := s.customerRepo.GetCustomerByID(parsedPatientID, parsedCompanyID)
	if err != nil {
		return response.NewErrorResponse("Failed to validate patient: " + err.Error())
	}
	if patient == nil {
		return response.NewErrorResponse("Patient tidak ditemukan untuk company ini")
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

	parsedTreatmentID, err := uuid.Parse(input.TreatmentID)
	if err != nil {
		return response.NewErrorResponse("Invalid treatment id")
	}
	treatment, err := s.treatmentRepo.FindByIDForCompany(parsedTreatmentID, companyID)
	if err != nil {
		return response.NewErrorResponse("Failed to validate treatment: " + err.Error())
	}
	if treatment == nil {
		return response.NewErrorResponse("Treatment tidak ditemukan untuk company ini")
	}
	appointment.TreatmentID = parsedTreatmentID

	parsedTherapistID, err := uuid.Parse(input.TherapistID)
	if err != nil {
		return response.NewErrorResponse("Invalid therapist id")
	}
	therapist, err := s.dokterRepo.GetByID(input.TherapistID, companyID)
	if err != nil {
		return response.NewErrorResponse("Therapist tidak ditemukan untuk company ini")
	}
	if therapist == nil {
		return response.NewErrorResponse("Therapist tidak ditemukan untuk company ini")
	}
	appointment.TherapistID = parsedTherapistID

	// Check conflict
	hasConflict, err := s.appointmentRepo.CheckConflict(input.TherapistID, input.BookingDate, startTime, endTime, "")
	if err != nil {
		return response.NewErrorResponse("Failed to check conflict: " + err.Error())
	}
	if hasConflict {
		return response.NewErrorResponse("Therapist sudah memiliki appointment pada waktu tersebut")
	}

	if err := s.appointmentRepo.Create(appointment); err != nil {
		return response.NewErrorResponse("Failed to create appointment: " + err.Error())
	}

	created, fetchErr := s.appointmentRepo.GetByID(appointment.ID.String(), companyID)
	if fetchErr == nil && created != nil {
		return response.NewSuccessResponse(toAppointmentResponse(created), "Appointment created successfully")
	}

	return response.NewSuccessResponse(toAppointmentResponse(appointment), "Appointment created successfully")
}

func (s *AppointmentService) GetAppointments(companyID string, filters map[string]interface{}, limit, offset int) response.PaginatedResponse {
	appointments, total, err := s.appointmentRepo.GetAll(companyID, filters, limit, offset)
	if err != nil {
		return response.PaginatedResponse{Success: false, Data: nil, Pagination: response.Pagination{}}
	}

	hasMore := int64(offset+limit) < total
	return response.PaginatedResponse{
		Success: true,
		Data:    toAppointmentResponseList(appointments),
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

	return response.NewSuccessResponse(toAppointmentResponse(appointment), "Appointment fetched successfully")
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
		patient, err := s.customerRepo.GetCustomerByID(parsedPatientID, appointment.CompanyID)
		if err != nil {
			return response.NewErrorResponse("Failed to validate patient: " + err.Error())
		}
		if patient == nil {
			return response.NewErrorResponse("Patient tidak ditemukan untuk company ini")
		}
		appointment.PatientID = parsedPatientID
	}

	if input.TreatmentID != nil {
		parsedTreatmentID, err := uuid.Parse(*input.TreatmentID)
		if err != nil {
			return response.NewErrorResponse("Invalid treatment id")
		}
		treatment, err := s.treatmentRepo.FindByIDForCompany(parsedTreatmentID, companyID)
		if err != nil {
			return response.NewErrorResponse("Failed to validate treatment: " + err.Error())
		}
		if treatment == nil {
			return response.NewErrorResponse("Treatment tidak ditemukan untuk company ini")
		}
		appointment.TreatmentID = parsedTreatmentID
	}

	if input.TherapistID != nil {
		parsedTherapistID, err := uuid.Parse(*input.TherapistID)
		if err != nil {
			return response.NewErrorResponse("Invalid therapist id")
		}
		therapist, err := s.dokterRepo.GetByID(*input.TherapistID, companyID)
		if err != nil {
			return response.NewErrorResponse("Therapist tidak ditemukan untuk company ini")
		}
		if therapist == nil {
			return response.NewErrorResponse("Therapist tidak ditemukan untuk company ini")
		}
		appointment.TherapistID = parsedTherapistID
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

	if appointment.TherapistID != uuid.Nil {
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

	updated, fetchErr := s.appointmentRepo.GetByID(appointment.ID.String(), companyID)
	if fetchErr == nil && updated != nil {
		return response.NewSuccessResponse(toAppointmentResponse(updated), "Appointment updated successfully")
	}

	return response.NewSuccessResponse(toAppointmentResponse(appointment), "Appointment updated successfully")
}

func (s *AppointmentService) DeleteAppointment(id, companyID string) response.ApiResponse {
	if err := s.appointmentRepo.Delete(id, companyID); err != nil {
		return response.NewErrorResponse("Failed to delete appointment: " + err.Error())
	}

	return response.NewSuccessResponse(nil, "Appointment deleted successfully")
}
