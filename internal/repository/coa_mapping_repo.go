package repository

import (
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

// ─── COAMapping Repository ───────────────────────────────────────────────────

type COAMappingRepository interface {
	FindAll() ([]entity.COAMapping, error)
	FindByTransType(transType string) ([]entity.COAMapping, error)
	FindByTransTypeAndRole(transType, role string, bbmID *uint) (*entity.COAMapping, error)
	FindByBBM(bbmID uint) ([]entity.COAMapping, error)
	Upsert(m *entity.COAMapping) error
	Delete(id uint) error
}

type coaMappingRepository struct{ db *gorm.DB }

func NewCOAMappingRepository(db *gorm.DB) COAMappingRepository {
	return &coaMappingRepository{db}
}

func (r *coaMappingRepository) FindAll() ([]entity.COAMapping, error) {
	var ms []entity.COAMapping
	err := r.db.Preload("COA").Preload("BBM").
		Order("trans_type, role, bbm_id").Find(&ms).Error
	return ms, err
}

func (r *coaMappingRepository) FindByTransType(transType string) ([]entity.COAMapping, error) {
	var ms []entity.COAMapping
	err := r.db.Preload("COA").Preload("BBM").
		Where("trans_type = ?", transType).
		Order("role, bbm_id").Find(&ms).Error
	return ms, err
}

func (r *coaMappingRepository) FindByTransTypeAndRole(transType, role string, bbmID *uint) (*entity.COAMapping, error) {
	var m entity.COAMapping
	q := r.db.Where("trans_type = ? AND role = ?", transType, role)
	if bbmID == nil {
		q = q.Where("bbm_id IS NULL")
	} else {
		q = q.Where("bbm_id = ?", *bbmID)
	}
	err := q.First(&m).Error
	return &m, err
}

func (r *coaMappingRepository) FindByBBM(bbmID uint) ([]entity.COAMapping, error) {
	var ms []entity.COAMapping
	err := r.db.Preload("COA").Where("bbm_id = ?", bbmID).Find(&ms).Error
	return ms, err
}

func (r *coaMappingRepository) Upsert(m *entity.COAMapping) error {
	existing, err := r.FindByTransTypeAndRole(m.TransType, m.Role, m.BBMID)
	if err == nil {
		existing.COAID = m.COAID
		existing.Label = m.Label
		return r.db.Omit("COA", "BBM").Save(existing).Error
	}
	return r.db.Omit("COA", "BBM").Create(m).Error
}

func (r *coaMappingRepository) Delete(id uint) error {
	return r.db.Delete(&entity.COAMapping{}, id).Error
}
