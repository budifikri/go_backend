package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JadwalDokter model
type JadwalDokter struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DokterID   uuid.UUID `gorm:"type:uuid;notNull;index" json:"dokter_id"`
	CompanyID  uuid.UUID `gorm:"type:uuid;notNull;index" json:"company_id"`
	Hari       string    `gorm:"type:varchar(20);notNull" json:"hari"`
	JamMulai   string    `gorm:"type:varchar(5);notNull" json:"jam_mulai"`
	JamSelesai string    `gorm:"type:varchar(5);notNull" json:"jam_selesai"`
	IsActive   bool      `gorm:"column:is_active;notNull;default:true" json:"is_active"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Dokter *Dokter `gorm:"foreignKey:DokterID;references:ID" json:"dokter,omitempty"`
}

func (j *JadwalDokter) BeforeCreate(tx *gorm.DB) error {
	if j.ID == uuid.Nil {
		j.ID = uuid.New()
	}
	return nil
}

func (JadwalDokter) TableName() string {
	return "jadwal_dokter"
}
