package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CompanyStatus enum
type CompanyStatus string

const (
	CompanyStatusActive    CompanyStatus = "active"
	CompanyStatusInactive  CompanyStatus = "inactive"
	CompanyStatusSuspended CompanyStatus = "suspended"
)

// Company model
type Company struct {
	ID              uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code            string        `gorm:"column:code;type:varchar(20);uniqueIndex;notNull" json:"code"`
	Nama            string        `gorm:"column:nama;type:varchar(200);notNull" json:"nama"`
	Email           string        `gorm:"column:email;type:varchar(100);uniqueIndex;notNull" json:"email"`
	Address         *string       `gorm:"column:address;type:text" json:"address,omitempty"`
	Telp            *string       `gorm:"column:telp;type:varchar(20)" json:"telp,omitempty"`
	Logo            *string       `gorm:"column:logo;type:text" json:"logo,omitempty"`
	Website         *string       `gorm:"column:website;type:varchar(255)" json:"website,omitempty"`
	TaxID           *string       `gorm:"column:tax_id;type:varchar(50)" json:"tax_id,omitempty"`
	BusinessLicense *string       `gorm:"column:business_license;type:varchar(100)" json:"business_license,omitempty"`
	Status          CompanyStatus `gorm:"column:status;type:company_status;notNull;default:'active'" json:"status"`
	CreatedAt       time.Time     `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time     `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (c *Company) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (Company) TableName() string {
	return "companies"
}
