package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CompanyStatus enum
type CompanyStatus string

const (
	CompanyStatusActive   CompanyStatus = "active"
	CompanyStatusInactive CompanyStatus = "inactive"
)

// Company model
type Company struct {
	ID        uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code      string        `gorm:"type:varchar(20);uniqueIndex;notNull" json:"code"`
	Name      string        `gorm:"type:varchar(100);notNull" json:"name"`
	Address   string        `gorm:"type:text" json:"address,omitempty"`
	City      string        `gorm:"type:varchar(50)" json:"city,omitempty"`
	Phone     string        `gorm:"type:varchar(20)" json:"phone,omitempty"`
	Email     string        `gorm:"type:varchar(100)" json:"email,omitempty"`
	TaxID     string        `gorm:"type:varchar(50)" json:"tax_id,omitempty"`
	Logo      string        `gorm:"type:text" json:"logo,omitempty"`
	Status    CompanyStatus `gorm:"type:company_status;notNull;default:'active'" json:"status"`
	CreatedAt time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
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
