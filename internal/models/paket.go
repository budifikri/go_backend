package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Paket model (Package/ Bundle)
type Paket struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	KodePaket  string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"kodepaket"`
	NmPaket    string     `gorm:"type:varchar(150);not null" json:"nm_paket"`
	Deskripsi  string     `gorm:"type:text" json:"deskripsi,omitempty"`
	HargaPaket float64    `gorm:"type:decimal(12,2);default:0;not null" json:"harga_paket"`
	IsActive   bool       `gorm:"default:true;not null" json:"is_active"`
	CompanyID  *uuid.UUID `gorm:"type:uuid" json:"company_id,omitempty"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	Details []DetailPaket `gorm:"foreignKey:IDPaket;constraint:OnDelete:CASCADE" json:"details,omitempty"`
}

func (p *Paket) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (Paket) TableName() string {
	return "paket"
}

// DetailPaket model (Package Detail)
type DetailPaket struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	IDPaket   uuid.UUID `gorm:"type:uuid;not null;index" json:"id_paket"`
	IDProduk  uuid.UUID `gorm:"type:uuid;not null;index" json:"id_produk"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Produk *Product `gorm:"foreignKey:IDProduk" json:"produk,omitempty"`
}

func (d *DetailPaket) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

func (DetailPaket) TableName() string {
	return "detail_paket"
}
