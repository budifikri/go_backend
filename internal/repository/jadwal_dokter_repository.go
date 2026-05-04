package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type JadwalDokterRepository struct {
	db *gorm.DB
}

func NewJadwalDokterRepository(db *gorm.DB) *JadwalDokterRepository {
	return &JadwalDokterRepository{db: db}
}

type JadwalDokterRow struct {
	ID         uuid.UUID `gorm:"column:id" json:"id"`
	DokterID   uuid.UUID `gorm:"column:dokter_id" json:"dokter_id"`
	CompanyID  uuid.UUID `gorm:"column:company_id" json:"company_id"`
	Hari       string    `gorm:"column:hari" json:"hari"`
	JamMulai   string    `gorm:"column:jam_mulai" json:"jam_mulai"`
	JamSelesai string    `gorm:"column:jam_selesai" json:"jam_selesai"`
	IsActive   bool      `gorm:"column:is_active" json:"is_active"`
	DokterNama string    `gorm:"column:nama" json:"dokter_nama"`
}

func (r *JadwalDokterRepository) Create(data map[string]interface{}) (*JadwalDokterRow, error) {
	jd := &JadwalDokterRow{}
	if err := r.db.Table("jadwal_dokter").Clauses(clause.Returning{}).Create(data).Scan(&jd).Error; err != nil {
		return nil, err
	}
	return jd, nil
}

func (r *JadwalDokterRepository) FindAll(companyID uuid.UUID, filters map[string]interface{}, limit, offset int) ([]JadwalDokterRow, int64, error) {
	var rows []JadwalDokterRow
	var total int64

	base := r.db.Table("jadwal_dokter jd").
		Joins("LEFT JOIN dokters d ON d.id = jd.dokter_id").
		Where("jd.company_id = ?", companyID)

	if v, ok := filters["dokter_id"]; ok {
		base = base.Where("jd.dokter_id = ?", v)
	}
	if v, ok := filters["hari"]; ok {
		base = base.Where("jd.hari = ?", v)
	}
	if v, ok := filters["is_active"]; ok {
		base = base.Where("jd.is_active = ?", v)
	}
	if v, ok := filters["search"]; ok {
		base = base.Where("d.nama ILIKE ? OR jd.hari ILIKE ?", "%"+v.(string)+"%", "%"+v.(string)+"%")
	}

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := base.Select("jd.*, d.nama as nama").
		Limit(limit).Offset(offset).
		Order("jd.hari ASC, jd.jam_mulai ASC").
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (r *JadwalDokterRepository) FindByID(id string, companyID uuid.UUID) (*JadwalDokterRow, error) {
	var row JadwalDokterRow
	err := r.db.Table("jadwal_dokter jd").
		Joins("LEFT JOIN dokters d ON d.id = jd.dokter_id").
		Where("jd.id = ? AND jd.company_id = ?", id, companyID).
		Select("jd.*, d.nama as nama").
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("jadwal dokter not found")
		}
		return nil, err
	}
	return &row, nil
}

func (r *JadwalDokterRepository) Update(id string, updates map[string]interface{}, companyID uuid.UUID) error {
	result := r.db.Table("jadwal_dokter").
		Where("id = ? AND company_id = ?", id, companyID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("jadwal dokter not found")
	}
	return nil
}

func (r *JadwalDokterRepository) Delete(id string, companyID uuid.UUID) error {
	result := r.db.Table("jadwal_dokter").
		Where("id = ? AND company_id = ?", id, companyID).
		Delete(&JadwalDokterRow{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("jadwal dokter not found")
	}
	return nil
}

func (r *JadwalDokterRepository) CheckDependencies(id string, companyID uuid.UUID) (bool, error) {
	return false, nil
}

func (r *JadwalDokterRepository) IsEligibleDokter(dokterID, companyID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Dokter{}).
		Where("id = ? AND company_id = ? AND active = ? AND tipe = ?", dokterID, companyID, true, models.TipeDokterDokter).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
