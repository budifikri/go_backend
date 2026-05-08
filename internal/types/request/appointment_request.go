package request

import "github.com/pos-retail/go_backend/internal/models"

type CreateAppointmentRequest struct {
	PatientID   string `json:"patient_id" validate:"required,uuid4"`
	TreatmentID string `json:"treatment_id" validate:"required,uuid4"`
	TherapistID string `json:"therapist_id" validate:"required,uuid4"`
	BookingDate string `json:"booking_date" validate:"required,datetime=2006-01-02"`
	StartTime   string `json:"start_time" validate:"required,datetime=15:04"`
	EndTime     string `json:"end_time" validate:"required,datetime=15:04"`
	Notes       string `json:"notes" validate:"omitempty"`
}

type UpdateAppointmentRequest struct {
	PatientID   *string                   `json:"patient_id" validate:"omitempty,uuid4"`
	TreatmentID *string                   `json:"treatment_id" validate:"omitempty,uuid4"`
	TherapistID *string                   `json:"therapist_id" validate:"omitempty,uuid4"`
	SalesID     *string                   `json:"sales_id" validate:"omitempty,uuid4"`
	BookingDate *string                   `json:"booking_date" validate:"omitempty,datetime=2006-01-02"`
	StartTime   *string                   `json:"start_time" validate:"omitempty,datetime=15:04"`
	EndTime     *string                   `json:"end_time" validate:"omitempty,datetime=15:04"`
	Status      *models.AppointmentStatus `json:"status" validate:"omitempty,oneof=scheduled confirmed completed cancelled"`
	Notes       *string                   `json:"notes" validate:"omitempty"`
}

type AppointmentFilterRequest struct {
	DateFrom    string `json:"date_from" validate:"omitempty,datetime=2006-01-02"`
	DateTo      string `json:"date_to" validate:"omitempty,datetime=2006-01-02"`
	Status      string `json:"status" validate:"omitempty,oneof=scheduled confirmed completed cancelled"`
	TherapistID string `json:"therapist_id" validate:"omitempty,uuid4"`
	PatientID   string `json:"patient_id" validate:"omitempty,uuid4"`
	Limit       int    `json:"limit" validate:"omitempty,min=1,max=100"`
	Offset      int    `json:"offset" validate:"omitempty,min=0"`
}
