package repository

import (
	"errors"

	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
)

type AppointmentRepository struct {
	db *gorm.DB
}

func NewAppointmentRepository(db *gorm.DB) *AppointmentRepository {
	return &AppointmentRepository{db: db}
}

func (r *AppointmentRepository) Create(appointment *models.Appointment) error {
	return r.db.Create(appointment).Error
}

func (r *AppointmentRepository) GetByID(id, companyID string) (*models.Appointment, error) {
	var appointment models.Appointment
	err := r.db.Preload("Patient").Preload("Treatment").Preload("Therapist").Preload("Sale").
		Where("id = ? AND company_id = ?", id, companyID).First(&appointment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("appointment not found")
		}
		return nil, err
	}
	return &appointment, nil
}

func (r *AppointmentRepository) GetAll(companyID string, filters map[string]interface{}, limit, offset int) ([]models.Appointment, int64, error) {
	var appointments []models.Appointment
	var total int64

	query := r.db.Model(&models.Appointment{}).Where("company_id = ?", companyID)

	// Apply filters
	if dateFrom, ok := filters["date_from"].(string); ok && dateFrom != "" {
		query = query.Where("booking_date >= ?", dateFrom)
	}
	if dateTo, ok := filters["date_to"].(string); ok && dateTo != "" {
		query = query.Where("booking_date <= ?", dateTo)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if therapistID, ok := filters["therapist_id"].(string); ok && therapistID != "" {
		query = query.Where("therapist_id = ?", therapistID)
	}
	if patientID, ok := filters["patient_id"].(string); ok && patientID != "" {
		query = query.Where("patient_id = ?", patientID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Patient").Preload("Treatment").Preload("Therapist").Preload("Sale").
		Limit(limit).Offset(offset).Order("booking_date DESC, start_time ASC").Find(&appointments).Error; err != nil {
		return nil, 0, err
	}

	return appointments, total, nil
}

func (r *AppointmentRepository) Update(appointment *models.Appointment) error {
	return r.db.Save(appointment).Error
}

func (r *AppointmentRepository) Delete(id, companyID string) error {
	result := r.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&models.Appointment{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("appointment not found")
	}
	return nil
}

func (r *AppointmentRepository) CheckConflict(therapistID, bookingDate string, startTime, endTime models.ClockTime, excludeID string) (bool, error) {
	var count int64
	startValue := startTime.Format("15:04:05")
	endValue := endTime.Format("15:04:05")
	query := r.db.Model(&models.Appointment{}).
		Where("therapist_id = ? AND booking_date = ? AND status NOT IN ('cancelled')", therapistID, bookingDate).
		Where("(CAST(start_time AS time) < CAST(? AS time) AND CAST(end_time AS time) > CAST(? AS time)) OR (CAST(start_time AS time) < CAST(? AS time) AND CAST(end_time AS time) > CAST(? AS time)) OR (CAST(start_time AS time) >= CAST(? AS time) AND CAST(end_time AS time) <= CAST(? AS time))",
			endValue, startValue, endValue, startValue, startValue, endValue)

	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
