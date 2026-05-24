package repository

import (
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

// ─── COAType Repository ──────────────────────────────────────────────────────

type COATypeRepository interface {
	FindAll() ([]entity.COAType, error)
	FindByID(id uint) (*entity.COAType, error)
	Create(ct *entity.COAType) error
	Update(ct *entity.COAType) error
	Delete(id uint) error
}

type coaTypeRepository struct{ db *gorm.DB }

func NewCOATypeRepository(db *gorm.DB) COATypeRepository {
	return &coaTypeRepository{db}
}

func (r *coaTypeRepository) FindAll() ([]entity.COAType, error) {
	var types []entity.COAType
	err := r.db.Order("code ASC").Find(&types).Error
	return types, err
}

func (r *coaTypeRepository) FindByID(id uint) (*entity.COAType, error) {
	var ct entity.COAType
	err := r.db.First(&ct, id).Error
	return &ct, err
}

func (r *coaTypeRepository) Create(ct *entity.COAType) error {
	return r.db.Create(ct).Error
}

func (r *coaTypeRepository) Update(ct *entity.COAType) error {
	return r.db.Save(ct).Error
}

func (r *coaTypeRepository) Delete(id uint) error {
	return r.db.Delete(&entity.COAType{}, id).Error
}

// ─── COA Repository ──────────────────────────────────────────────────────────

type COARepository interface {
	// FindAllGrouped returns COATypes preloaded with their COAs (ordered by code)
	FindAllGrouped() ([]entity.COAType, error)
	FindAll() ([]entity.COA, error)
	FindByID(id uint) (*entity.COA, error)
	FindByCode(code string) (*entity.COA, error)
	// FindDetailOnly returns only non-header, active COAs — for transaction dropdowns
	FindDetailOnly() ([]entity.COA, error)
	// FindMaxCodeByPrefix finds the max integer code among COAs starting with prefix (non-header only)
	FindMaxCodeByPrefix(prefix string) (int, error)
	Create(coa *entity.COA) error
	Update(coa *entity.COA) error
	Delete(id uint) error
}

type coaRepository struct{ db *gorm.DB }

func NewCOARepository(db *gorm.DB) COARepository {
	return &coaRepository{db}
}

func (r *coaRepository) FindAllGrouped() ([]entity.COAType, error) {
	var types []entity.COAType
	err := r.db.Preload("COAs", func(db *gorm.DB) *gorm.DB {
		return db.Order("code ASC")
	}).Order("code ASC").Find(&types).Error
	return types, err
}

func (r *coaRepository) FindAll() ([]entity.COA, error) {
	var coas []entity.COA
	err := r.db.Preload("COAType").Order("code ASC").Find(&coas).Error
	return coas, err
}

func (r *coaRepository) FindByID(id uint) (*entity.COA, error) {
	var coa entity.COA
	err := r.db.Preload("COAType").First(&coa, id).Error
	return &coa, err
}

func (r *coaRepository) FindByCode(code string) (*entity.COA, error) {
	var coa entity.COA
	err := r.db.Where("code = ?", code).First(&coa).Error
	return &coa, err
}

func (r *coaRepository) FindDetailOnly() ([]entity.COA, error) {
	var coas []entity.COA
	err := r.db.Preload("COAType").
		Where("is_header = false AND is_active = true").
		Order("code ASC").Find(&coas).Error
	return coas, err
}

func (r *coaRepository) FindMaxCodeByPrefix(prefix string) (int, error) {
	var codes []string
	err := r.db.Unscoped().Model(&entity.COA{}).
		Where("code LIKE ? AND is_header = false", prefix+"%").
		Pluck("code", &codes).Error
	if err != nil {
		return 0, err
	}
	maxCode := 0
	for _, c := range codes {
		n := 0
		for _, ch := range c {
			if ch >= '0' && ch <= '9' {
				n = n*10 + int(ch-'0')
			} else {
				n = 0
				break
			}
		}
		if n > maxCode {
			maxCode = n
		}
	}
	return maxCode, nil
}

func (r *coaRepository) Create(coa *entity.COA) error {
	return r.db.Omit("Updater", "COAType").Create(coa).Error
}

func (r *coaRepository) Update(coa *entity.COA) error {
	return r.db.Omit("Updater", "COAType").Save(coa).Error
}

func (r *coaRepository) Delete(id uint) error {
	return r.db.Delete(&entity.COA{}, id).Error
}

// ─── JournalEntry Repository ─────────────────────────────────────────────────

type JournalEntryRepository interface {
	FindByCOA(coaID uint, limit int) ([]entity.JournalEntry, error)
	FindByRef(refType string, refID uint) ([]entity.JournalEntry, error)
	Create(entry *entity.JournalEntry) error
	CreateBatch(entries []entity.JournalEntry) error
	// DeleteByRef removes all journal entries linked to a source transaction (for reversal / re-post).
	DeleteByRef(refType string, refID uint) error
}

type journalEntryRepository struct{ db *gorm.DB }

func NewJournalEntryRepository(db *gorm.DB) JournalEntryRepository {
	return &journalEntryRepository{db}
}

func (r *journalEntryRepository) FindByCOA(coaID uint, limit int) ([]entity.JournalEntry, error) {
	var entries []entity.JournalEntry
	q := r.db.Where("coa_id = ?", coaID).
		Preload("Wallet").
		Order("trans_date DESC, id DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&entries).Error
	return entries, err
}

func (r *journalEntryRepository) FindByRef(refType string, refID uint) ([]entity.JournalEntry, error) {
	var entries []entity.JournalEntry
	err := r.db.Where("ref_type = ? AND ref_id = ?", refType, refID).
		Preload("COA").Order("id ASC").Find(&entries).Error
	return entries, err
}

func (r *journalEntryRepository) Create(entry *entity.JournalEntry) error {
	return r.db.Omit("COA", "Wallet", "Creator").Create(entry).Error
}

func (r *journalEntryRepository) CreateBatch(entries []entity.JournalEntry) error {
	if len(entries) == 0 {
		return nil
	}
	return r.db.Omit("COA", "Wallet", "Creator").CreateInBatches(entries, 100).Error
}

func (r *journalEntryRepository) DeleteByRef(refType string, refID uint) error {
	return r.db.Where("ref_type = ? AND ref_id = ?", refType, refID).
		Delete(&entity.JournalEntry{}).Error
}
