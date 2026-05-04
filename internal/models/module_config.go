package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BusinessType struct {
	ID          uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code        BusinessTypeCode `gorm:"column:code;type:varchar(50);uniqueIndex;notNull" json:"code"`
	Name        string           `gorm:"column:name;type:varchar(100);notNull" json:"name"`
	Description string           `gorm:"column:description;type:text" json:"description"`
	IsActive    bool             `gorm:"column:is_active;notNull;default:true" json:"is_active"`
	IsDefault   bool             `gorm:"column:is_default;notNull;default:false" json:"is_default"`
	IsSystem    bool             `gorm:"column:is_system;notNull;default:false" json:"is_system"`
	SortOrder   int              `gorm:"column:sort_order;notNull;default:0" json:"sort_order"`
	CreatedAt   time.Time        `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time        `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

type ModulePackage struct {
	ID           uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BusinessType BusinessTypeCode `gorm:"column:business_type;type:varchar(50);index;notNull" json:"business_type"`
	Code         string           `gorm:"column:code;type:varchar(100);uniqueIndex;notNull" json:"code"`
	Name         string           `gorm:"column:name;type:varchar(100);notNull" json:"name"`
	Description  string           `gorm:"column:description;type:text" json:"description"`
	IsActive     bool             `gorm:"column:is_active;notNull;default:true" json:"is_active"`
	IsDefault    bool             `gorm:"column:is_default;notNull;default:false" json:"is_default"`
	IsSystem     bool             `gorm:"column:is_system;notNull;default:false" json:"is_system"`
	SortOrder    int              `gorm:"column:sort_order;notNull;default:0" json:"sort_order"`
	CreatedAt    time.Time        `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time        `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

type CompanyModule struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CompanyID   uuid.UUID  `gorm:"column:company_id;type:uuid;notNull;index:idx_company_module,unique" json:"company_id"`
	ModuleCode  string     `gorm:"column:module_code;type:varchar(100);notNull;index:idx_company_module,unique" json:"module_code"`
	IsActive    bool       `gorm:"column:is_active;notNull;default:false" json:"is_active"`
	ActivatedAt *time.Time `gorm:"column:activated_at" json:"activated_at,omitempty"`
	ActivatedBy *uuid.UUID `gorm:"column:activated_by;type:uuid" json:"activated_by,omitempty"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (m *BusinessType) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

func (m *ModulePackage) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

func (m *CompanyModule) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

func (BusinessType) TableName() string {
	return "business_types"
}

func (ModulePackage) TableName() string {
	return "module_packages"
}

func (CompanyModule) TableName() string {
	return "company_modules"
}
