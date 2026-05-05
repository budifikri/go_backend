package services

import (
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/repository"
	"github.com/pos-retail/go_backend/internal/types/response"
)

var (
	ErrTreatmentNotFound = errors.New("treatment not found")
	ErrTagNotFound       = errors.New("treatment tag not found")
)

type TreatmentService struct {
	treatmentRepo *repository.TreatmentRepository
}

func NewTreatmentService(treatmentRepo *repository.TreatmentRepository) *TreatmentService {
	return &TreatmentService{
		treatmentRepo: treatmentRepo,
	}
}

type TreatmentListResponse struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	CompanyID   *uuid.UUID    `json:"company_id,omitempty"`
	Duration    int           `json:"duration"`
	Price       float64       `json:"price"`
	Description string        `json:"description,omitempty"`
	IsActive    bool          `json:"is_active"`
	CreatedAt   string        `json:"created_at"`
	UpdatedAt   string        `json:"updated_at"`
	Tags        []TagResponse `json:"tags,omitempty"`
}

type TagResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type TreatmentDetailResponse struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	CompanyID   *uuid.UUID    `json:"company_id,omitempty"`
	Duration    int           `json:"duration"`
	Price       float64       `json:"price"`
	Description string        `json:"description,omitempty"`
	IsActive    bool          `json:"is_active"`
	CreatedAt   string        `json:"created_at"`
	UpdatedAt   string        `json:"updated_at"`
	Tags        []TagResponse `json:"tags,omitempty"`
}

func (s *TreatmentService) GetTreatments(filters map[string]interface{}, limit, offset int) response.PaginatedResponse {
	treatments, total, err := s.treatmentRepo.FindAll(filters, limit, offset)
	if err != nil {
		log.Println("Failed to get treatments:", err)
		return response.PaginatedResponse{
			Success: false,
			Data:    err.Error(),
		}
	}

	items := make([]TreatmentListResponse, len(treatments))
	for i, t := range treatments {
		items[i] = TreatmentListResponse{
			ID:          t.ID,
			Name:        t.Name,
			CompanyID:   t.CompanyID,
			Duration:    t.Duration,
			Price:       t.Price,
			Description: t.Description,
			IsActive:    t.IsActive,
			CreatedAt:   t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Tags:        convertTags(t.Tags),
		}
	}

	return response.PaginatedResponse{
		Success: true,
		Data:    items,
		Pagination: response.Pagination{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}
}

func (s *TreatmentService) GetTreatmentByID(id string) response.ApiResponse {
	treatmentID, err := uuid.Parse(id)
	if err != nil {
		return response.ApiResponse{
			Success: false,
			Message: "Invalid treatment ID",
		}
	}

	treatment, err := s.treatmentRepo.FindByID(treatmentID)
	if err != nil {
		log.Println("Failed to get treatment:", err)
		return response.ApiResponse{
			Success: false,
			Message: "Failed to fetch treatment",
		}
	}

	if treatment == nil {
		return response.ApiResponse{
			Success: false,
			Message: "Treatment not found",
		}
	}

	return response.ApiResponse{
		Success: true,
		Data: TreatmentDetailResponse{
			ID:          treatment.ID,
			Name:        treatment.Name,
			CompanyID:   treatment.CompanyID,
			Duration:    treatment.Duration,
			Price:       treatment.Price,
			Description: treatment.Description,
			IsActive:    treatment.IsActive,
			CreatedAt:   treatment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   treatment.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Tags:        convertTags(treatment.Tags),
		},
	}
}

func (s *TreatmentService) CreateTreatment(input map[string]interface{}, companyID string) response.ApiResponse {
	treatment := &models.Treatment{
		Name:        input["name"].(string),
		Duration:    int(input["duration"].(float64)),
		Price:       input["price"].(float64),
		Description: input["description"].(string),
		IsActive:    true,
	}

	if companyID != "" {
		if cid, err := uuid.Parse(companyID); err == nil {
			treatment.CompanyID = &cid
		}
	}

	if err := s.treatmentRepo.Create(treatment); err != nil {
		log.Println("Failed to create treatment:", err)
		return response.ApiResponse{
			Success: false,
			Message: "Failed to create treatment",
		}
	}

	// Handle tags if provided
	if tagIDs, ok := input["tag_ids"].([]interface{}); ok && len(tagIDs) > 0 {
		for _, tid := range tagIDs {
			if tidStr, ok := tid.(string); ok {
				if tagID, err := uuid.Parse(tidStr); err == nil {
					relation := &models.TreatmentTagRelation{
						TreatmentID: treatment.ID,
						TagID:       tagID,
					}
					s.treatmentRepo.CreateTagRelation(relation)
				}
			}
		}
	}

	return response.ApiResponse{
		Success: true,
		Message: "Treatment created successfully",
		Data: TreatmentDetailResponse{
			ID:          treatment.ID,
			Name:        treatment.Name,
			Duration:    treatment.Duration,
			Price:       treatment.Price,
			Description: treatment.Description,
			IsActive:    treatment.IsActive,
			CreatedAt:   treatment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   treatment.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}
}

func (s *TreatmentService) UpdateTreatment(id string, input map[string]interface{}) response.ApiResponse {
	treatmentID, err := uuid.Parse(id)
	if err != nil {
		return response.ApiResponse{
			Success: false,
			Message: "Invalid treatment ID",
			Error:   err.Error(),
		}
	}

	treatment, err := s.treatmentRepo.FindByID(treatmentID)
	if err != nil {
		log.Println("Failed to find treatment:", err)
		return response.ApiResponse{
			Success: false,
			Message: "Failed to fetch treatment",
		}
	}

	if treatment == nil {
		return response.ApiResponse{
			Success: false,
			Message: "Treatment not found",
		}
	}

	// Update fields
	if name, ok := input["name"].(string); ok {
		treatment.Name = name
	}
	if duration, ok := input["duration"].(float64); ok {
		treatment.Duration = int(duration)
	}
	if price, ok := input["price"].(float64); ok {
		treatment.Price = price
	}
	if description, ok := input["description"].(string); ok {
		treatment.Description = description
	}
	if isActive, ok := input["is_active"].(bool); ok {
		treatment.IsActive = isActive
	}

	if err := s.treatmentRepo.Update(treatment); err != nil {
		log.Println("Failed to update treatment:", err)
		return response.ApiResponse{
			Success: false,
			Message: "Failed to update treatment",
		}
	}

	// Update tags if provided
	if tagIDs, ok := input["tag_ids"].([]interface{}); ok {
		// Delete existing relations
		s.treatmentRepo.DeleteTagRelationsByTreatmentID(treatmentID)

		// Create new relations
		for _, tid := range tagIDs {
			if tidStr, ok := tid.(string); ok {
				if tagID, err := uuid.Parse(tidStr); err == nil {
					relation := &models.TreatmentTagRelation{
						TreatmentID: treatment.ID,
						TagID:       tagID,
					}
					s.treatmentRepo.CreateTagRelation(relation)
				}
			}
		}
	}

	return response.ApiResponse{
		Success: true,
		Message: "Treatment updated successfully",
	}
}

func (s *TreatmentService) DeleteTreatment(id string) response.ApiResponse {
	treatmentID, err := uuid.Parse(id)
	if err != nil {
		return response.ApiResponse{
			Success: false,
			Message: "Invalid treatment ID",
			Error:   err.Error(),
		}
	}

	if err := s.treatmentRepo.Delete(treatmentID); err != nil {
		log.Println("Failed to delete treatment:", err)
		return response.ApiResponse{
			Success: false,
			Message: "Failed to delete treatment",
		}
	}

	return response.ApiResponse{
		Success: true,
		Message: "Treatment deleted successfully",
	}
}

// Tag methods
func (s *TreatmentService) GetTags() response.ApiResponse {
	tags, err := s.treatmentRepo.FindAllTags()
	if err != nil {
		log.Println("Failed to get tags:", err)
		return response.ApiResponse{
			Success: false,
			Message: "Failed to fetch tags",
		}
	}

	items := make([]TagResponse, len(tags))
	for i, tag := range tags {
		items[i] = TagResponse{
			ID:   tag.ID,
			Name: tag.Name,
		}
	}

	return response.ApiResponse{
		Success: true,
		Data:    items,
	}
}

func (s *TreatmentService) CreateTag(name string) response.ApiResponse {
	tag := &models.TreatmentTag{
		Name: name,
	}

	if err := s.treatmentRepo.CreateTag(tag); err != nil {
		log.Println("Failed to create tag:", err)
		return response.ApiResponse{
			Success: false,
			Message: "Failed to create tag",
		}
	}

	return response.ApiResponse{
		Success: true,
		Message: "Tag created successfully",
		Data: TagResponse{
			ID:   tag.ID,
			Name: tag.Name,
		},
	}
}

func (s *TreatmentService) UpdateTag(id string, name string) response.ApiResponse {
	tagID, err := uuid.Parse(id)
	if err != nil {
		return response.ApiResponse{
			Success: false,
			Message: "Invalid tag ID",
			Error:   err.Error(),
		}
	}

	tag, err := s.treatmentRepo.FindTagByID(tagID)
	if err != nil {
		return response.ApiResponse{
			Success: false,
			Message: "Failed to fetch tag",
		}
	}

	if tag == nil {
		return response.ApiResponse{
			Success: false,
			Message: "Tag not found",
		}
	}

	tag.Name = name
	if err := s.treatmentRepo.UpdateTag(tag); err != nil {
		log.Println("Failed to update tag:", err)
		return response.ApiResponse{
			Success: false,
			Message: "Failed to update tag",
		}
	}

	return response.ApiResponse{
		Success: true,
		Message: "Tag updated successfully",
	}
}

func (s *TreatmentService) DeleteTag(id string) response.ApiResponse {
	tagID, err := uuid.Parse(id)
	if err != nil {
		return response.ApiResponse{
			Success: false,
			Message: "Invalid tag ID",
			Error:   err.Error(),
		}
	}

	if err := s.treatmentRepo.DeleteTag(tagID); err != nil {
		log.Println("Failed to delete tag:", err)
		return response.ApiResponse{
			Success: false,
			Message: "Failed to delete tag",
		}
	}

	return response.ApiResponse{
		Success: true,
		Message: "Tag deleted successfully",
	}
}

func convertTags(tags []models.TreatmentTag) []TagResponse {
	result := make([]TagResponse, len(tags))
	for i, tag := range tags {
		result[i] = TagResponse{
			ID:   tag.ID,
			Name: tag.Name,
		}
	}
	return result
}
