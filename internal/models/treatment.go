package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Treatment model
type Treatment struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string     `gorm:"type:varchar(150);not null" json:"name"`
	CompanyID   *uuid.UUID `gorm:"type:uuid;index" json:"company_id,omitempty"`
	Duration    int        `gorm:"default:60;not null" json:"duration"` // in minutes
	Price       float64    `gorm:"type:decimal(12,2);default:0;not null" json:"price"`
	Description string     `gorm:"type:text" json:"description,omitempty"`
	IsActive    bool       `gorm:"default:true;not null" json:"is_active"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	Tags []TreatmentTag `gorm:"many2many:treatment_tag_relations;" json:"tags,omitempty"`
}

func (t *Treatment) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (Treatment) TableName() string {
	return "treatments"
}

// TreatmentTag model
type TreatmentTag struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
}

func (TreatmentTag) TableName() string {
	return "treatment_tags"
}

// TreatmentTagRelation model
type TreatmentTagRelation struct {
	TreatmentID uuid.UUID `gorm:"type:uuid;primaryKey" json:"treatment_id"`
	TagID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"tag_id"`
}

func (TreatmentTagRelation) TableName() string {
	return "treatment_tag_relations"
}
