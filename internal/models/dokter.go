package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JenisKelamin enum
type JenisKelamin string

const (
	JenisKelaminL JenisKelamin = "L"
	JenisKelaminP JenisKelamin = "P"
)

// TipeDokter enum
type TipeDokter string

const (
	TipeDokterDokter     TipeDokter = "Dokter"
	TipeDokterBeautician TipeDokter = "Beautician"
)

// Dokter model
type Dokter struct {
	ID           uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CompanyID    uuid.UUID    `gorm:"type:uuid;notNull;index" json:"company_id"`
	Nama         string       `gorm:"column:nama;type:varchar(100);notNull" json:"nama"`
	JenisKelamin JenisKelamin `gorm:"column:jenis_kelamin;type:varchar(1);notNull" json:"jenis_kelamin"`
	TempatLahir  string       `gorm:"column:tempat_lahir;type:varchar(100);notNull" json:"tempat_lahir"`
	TanggalLahir time.Time    `gorm:"column:tanggal_lahir;type:date;notNull" json:"tanggal_lahir"`
	Alamat       string       `gorm:"column:alamat;type:text;notNull" json:"alamat"`
	NoTelp       string       `gorm:"column:no_telp;type:varchar(30);notNull" json:"no_telp"`
	Email        string       `gorm:"column:email;type:varchar(100);notNull" json:"email"`
	Tipe         TipeDokter   `gorm:"column:tipe;type:varchar(20);notNull" json:"tipe"`
	Active       bool         `gorm:"column:active;notNull;default:true" json:"active"`
	CreatedAt    time.Time    `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time    `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (d *Dokter) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

func (Dokter) TableName() string {
	return "dokters"
}
