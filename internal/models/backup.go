package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BackupLog struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CompanyID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"company_id"`
	Filename     string         `gorm:"size:255;not null" json:"filename"`
	FilePath     string         `gorm:"size:500;not null" json:"file_path"`
	FileSize     int64          `gorm:"not null" json:"file_size"`
	Status       string         `gorm:"size:50;default:'completed'" json:"status"`
	ErrorMessage string         `gorm:"type:text" json:"error_message,omitempty"`
	CreatedBy    string         `gorm:"size:100" json:"created_by"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	IsAuto       bool           `gorm:"default:false" json:"is_auto"`
	TableCount   int            `gorm:"default:0" json:"table_count"`
	RowCount     int64          `gorm:"default:0" json:"row_count"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (BackupLog) TableName() string {
	return "backup_logs"
}

type BackupSchedule struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	CompanyID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"company_id"`
	Enabled       bool      `gorm:"default:false" json:"enabled"`
	Schedule      string    `gorm:"size:100;default:'0 2 * * *'" json:"schedule"`
	RetentionDays int       `gorm:"default:7" json:"retention_days"`
	LastBackupAt  time.Time `json:"last_backup_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (BackupSchedule) TableName() string {
	return "backup_schedules"
}

type CreateBackupRequest struct {
	IsAuto bool `json:"is_auto"`
}

type UpdateScheduleRequest struct {
	Enabled       bool   `json:"enabled"`
	Schedule      string `json:"schedule"`
	RetentionDays int    `json:"retention_days"`
	Frequency     string `json:"frequency,omitempty"`
	Day           string `json:"day,omitempty"`
	Hour          string `json:"hour,omitempty"`
}

type BackupResponse struct {
	ID                uint      `json:"id"`
	CompanyID         uuid.UUID `json:"company_id"`
	Filename          string    `json:"filename"`
	FilePath          string    `json:"file_path"`
	FileSize          int64     `json:"file_size"`
	FileSizeFormatted string    `json:"file_size_formatted"`
	Status            string    `json:"status"`
	CreatedBy         string    `json:"created_by"`
	CreatedAt         time.Time `json:"created_at"`
	IsAuto            bool      `json:"is_auto"`
	TableCount        int       `json:"table_count"`
	RowCount          int64     `json:"row_count"`
}

type ScheduleResponse struct {
	CompanyID     uuid.UUID  `json:"company_id"`
	Enabled       bool       `json:"enabled"`
	Schedule      string     `json:"schedule"`
	RetentionDays int        `json:"retention_days"`
	LastBackupAt  *time.Time `json:"last_backup_at,omitempty"`
	Frequency     string     `json:"frequency"`
	Day           string     `json:"day,omitempty"`
	Hour          string     `json:"hour,omitempty"`
}

type RestoreRequest struct {
	Filename string `json:"filename" validate:"required"`
	Confirm  bool   `json:"confirm"`
}

type RestoreResult struct {
	Status        string    `json:"status"`
	CompanyID     uuid.UUID `json:"company_id"`
	TablesCleared int       `json:"tables_cleared"`
	RowsRestored  int64     `json:"rows_restored"`
	Duration      string    `json:"duration"`
	SafetyBackup  string    `json:"safety_backup"`
}

type RestoreValidation struct {
	Filename          string    `json:"filename"`
	FileSize          int64     `json:"file_size"`
	FileSizeFormatted string    `json:"file_size_formatted"`
	CreatedAt         time.Time `json:"created_at"`
	TableCount        int       `json:"table_count"`
	RowCount          int64     `json:"row_count"`
	CompanyID         uuid.UUID `json:"company_id"`
	CompanyName       string    `json:"company_name"`
	IsValid           bool      `json:"is_valid"`
	ErrorMessage      string    `json:"error_message,omitempty"`
}

type DeleteDataRequest struct {
	Scope    string `json:"scope" validate:"required,oneof=all master transaction"`
	Backuped bool   `json:"backuped"`
}

type TableCount struct {
	TableName string `json:"table_name"`
	RowCount  int64  `json:"row_count"`
}

type ScopeCountResponse struct {
	Scope  string       `json:"scope"`
	Tables []TableCount `json:"tables"`
	Total  int64        `json:"total"`
}

type DeleteDataResponse struct {
	Scope          string           `json:"scope"`
	TablesCleared  []string         `json:"tables_cleared"`
	RecordsDeleted map[string]int64 `json:"records_deleted"`
	TotalRecords   int64            `json:"total_records"`
}
