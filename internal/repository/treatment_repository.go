package repository

import (
	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
)

type TreatmentRepository struct {
	db *gorm.DB
}

func NewTreatmentRepository(db *gorm.DB) *TreatmentRepository {
	return &TreatmentRepository{db: db}
}

func (r *TreatmentRepository) FindAll(filters map[string]interface{}, limit, offset int) ([]models.Treatment, int64, error) {
	var treatments []models.Treatment
	var total int64

	query := r.db.Model(&models.Treatment{})

	if v, ok := filters["is_active"].(bool); ok {
		query = query.Where("is_active = ?", v)
	}
	if companyID, ok := filters["company_id"].(string); ok && companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}
	if search, ok := filters["search"].(string); ok && search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
	}
	if tagID, ok := filters["tag_id"].(string); ok && tagID != "" {
		query = query.Joins("JOIN treatment_tag_relations ON treatments.id = treatment_tag_relations.treatment_id").
			Where("treatment_tag_relations.tag_id = ?", tagID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Tags").Limit(limit).Offset(offset).Order("created_at DESC").Find(&treatments).Error; err != nil {
		return nil, 0, err
	}

	return treatments, total, nil
}

func (r *TreatmentRepository) FindByID(id uuid.UUID) (*models.Treatment, error) {
	var treatment models.Treatment
	if err := r.db.Preload("Tags").First(&treatment, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &treatment, nil
}

func (r *TreatmentRepository) Create(treatment *models.Treatment) error {
	return r.db.Create(treatment).Error
}

func (r *TreatmentRepository) Update(treatment *models.Treatment) error {
	return r.db.Save(treatment).Error
}

func (r *TreatmentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Treatment{}, "id = ?", id).Error
}

func (r *TreatmentRepository) FindAllTags() ([]models.TreatmentTag, error) {
	var tags []models.TreatmentTag
	if err := r.db.Order("name ASC").Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *TreatmentRepository) FindTagByID(id uuid.UUID) (*models.TreatmentTag, error) {
	var tag models.TreatmentTag
	if err := r.db.First(&tag, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

func (r *TreatmentRepository) CreateTag(tag *models.TreatmentTag) error {
	return r.db.Create(tag).Error
}

func (r *TreatmentRepository) UpdateTag(tag *models.TreatmentTag) error {
	return r.db.Save(tag).Error
}

func (r *TreatmentRepository) DeleteTag(id uuid.UUID) error {
	return r.db.Delete(&models.TreatmentTag{}, "id = ?", id).Error
}

func (r *TreatmentRepository) DeleteTagRelationsByTreatmentID(treatmentID uuid.UUID) error {
	return r.db.Delete(&models.TreatmentTagRelation{}, "treatment_id = ?", treatmentID).Error
}

func (r *TreatmentRepository) DeleteTagRelationsByTagID(tagID uuid.UUID) error {
	return r.db.Delete(&models.TreatmentTagRelation{}, "tag_id = ?", tagID).Error
}

func (r *TreatmentRepository) CreateTagRelation(relation *models.TreatmentTagRelation) error {
	return r.db.Create(relation).Error
}
