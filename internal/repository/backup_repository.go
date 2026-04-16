package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
)

type BackupRepository struct {
	db *gorm.DB
}

func NewBackupRepository(db *gorm.DB) *BackupRepository {
	return &BackupRepository{db: db}
}

func (r *BackupRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&models.BackupLog{}, &models.BackupSchedule{})
}

func (r *BackupRepository) CreateBackupLog(log *models.BackupLog) error {
	return r.db.Create(log).Error
}

func (r *BackupRepository) UpdateBackupLog(log *models.BackupLog) error {
	return r.db.Save(log).Error
}

func (r *BackupRepository) GetBackupLogByFilename(companyID uuid.UUID, filename string) (*models.BackupLog, error) {
	var log models.BackupLog
	err := r.db.Where("company_id = ? AND filename = ?", companyID, filename).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *BackupRepository) GetBackupLogs(companyID uuid.UUID, limit int) ([]models.BackupLog, error) {
	var logs []models.BackupLog
	query := r.db.Where("company_id = ?", companyID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&logs).Error
	return logs, err
}

func (r *BackupRepository) DeleteBackupLog(companyID uuid.UUID, filename string) error {
	return r.db.Where("company_id = ? AND filename = ?", companyID, filename).Delete(&models.BackupLog{}).Error
}

func (r *BackupRepository) DeleteOldBackups(companyID uuid.UUID, retentionDays int) ([]string, error) {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	var oldLogs []models.BackupLog
	err := r.db.Where("company_id = ? AND created_at < ?", companyID, cutoffDate).Find(&oldLogs).Error
	if err != nil {
		return nil, err
	}

	var deletedFiles []string
	for _, log := range oldLogs {
		deletedFiles = append(deletedFiles, log.Filename)
	}

	if len(deletedFiles) > 0 {
		err = r.db.Where("company_id = ? AND created_at < ?", companyID, cutoffDate).Delete(&models.BackupLog{}).Error
		if err != nil {
			return nil, err
		}
	}

	return deletedFiles, nil
}

func (r *BackupRepository) GetSchedule(companyID uuid.UUID) (*models.BackupSchedule, error) {
	var schedule models.BackupSchedule
	err := r.db.Where("company_id = ?", companyID).First(&schedule).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &schedule, nil
}

func (r *BackupRepository) UpsertSchedule(schedule *models.BackupSchedule) error {
	var existing models.BackupSchedule
	err := r.db.Where("company_id = ?", schedule.CompanyID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(schedule).Error
	}
	if err != nil {
		return err
	}
	schedule.ID = existing.ID
	schedule.CreatedAt = existing.CreatedAt
	return r.db.Save(schedule).Error
}

func (r *BackupRepository) UpdateLastBackup(companyID uuid.UUID) error {
	return r.db.Model(&models.BackupSchedule{}).Where("company_id = ?", companyID).Update("last_backup_at", time.Now()).Error
}

func (r *BackupRepository) GetAllEnabledSchedules() ([]models.BackupSchedule, error) {
	var schedules []models.BackupSchedule
	err := r.db.Where("enabled = ?", true).Find(&schedules).Error
	return schedules, err
}

func (r *BackupRepository) GetBackupCount(companyID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.BackupLog{}).Where("company_id = ?", companyID).Count(&count).Error
	return count, err
}

func (r *BackupRepository) GetTotalBackupSize(companyID uuid.UUID) (int64, error) {
	var total int64
	err := r.db.Model(&models.BackupLog{}).Where("company_id = ?", companyID).Select("COALESCE(SUM(file_size), 0)").Scan(&total).Error
	return total, err
}
