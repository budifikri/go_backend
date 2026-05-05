package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AppointmentStatus enum
type AppointmentStatus string

const (
	AppointmentStatusScheduled AppointmentStatus = "scheduled"
	AppointmentStatusConfirmed AppointmentStatus = "confirmed"
	AppointmentStatusCompleted AppointmentStatus = "completed"
	AppointmentStatusCancelled AppointmentStatus = "cancelled"
)

// Appointment model
type Appointment struct {
	ID          uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CompanyID   uuid.UUID         `gorm:"type:uuid;notNull;index" json:"company_id"`
	PatientID   uuid.UUID         `gorm:"type:uuid;notNull;index" json:"patient_id"`
	TreatmentID *uuid.UUID        `gorm:"type:uuid;index" json:"treatment_id"`
	TherapistID *uuid.UUID        `gorm:"type:uuid;index" json:"therapist_id"`
	BookingDate time.Time         `gorm:"column:booking_date;type:date;notNull;index" json:"booking_date"`
	StartTime   time.Time         `gorm:"column:start_time;type:time;notNull" json:"start_time"`
	EndTime     time.Time         `gorm:"column:end_time;type:time;notNull" json:"end_time"`
	Status      AppointmentStatus `gorm:"column:status;type:varchar(20);notNull;default:'scheduled'" json:"status"`
	Notes       string            `gorm:"column:notes;type:text" json:"notes"`
	CreatedAt   time.Time         `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time         `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relations
	Patient   *Customer  `gorm:"foreignKey:PatientID;references:ID" json:"patient,omitempty"`
	Treatment *Treatment `gorm:"foreignKey:TreatmentID;references:ID" json:"treatment,omitempty"`
	Therapist *Dokter    `gorm:"foreignKey:TherapistID;references:ID" json:"therapist,omitempty"`
}

func (a *Appointment) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

func (Appointment) TableName() string {
	return "appointments"
}
