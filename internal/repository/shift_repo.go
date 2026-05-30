package repository

import (
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

type ShiftRepository interface {
	FindAll() ([]entity.Shift, error)
	FindByID(id uint) (entity.Shift, error)
	Create(shift *entity.Shift) error
	Update(shift *entity.Shift) error
	Delete(id uint) error
	// IsUsed returns true jika shift masih dipakai di tabel trx_kedatangan_bbm.
	IsUsed(id uint) (bool, error)
}

type shiftRepository struct {
	db *gorm.DB
}

func NewShiftRepository(db *gorm.DB) ShiftRepository {
	return &shiftRepository{db}
}

func (r *shiftRepository) FindAll() ([]entity.Shift, error) {
	var data []entity.Shift
	err := r.db.Order("id ASC").Find(&data).Error
	return data, err
}

func (r *shiftRepository) FindByID(id uint) (entity.Shift, error) {
	var data entity.Shift
	err := r.db.First(&data, id).Error
	return data, err
}

func (r *shiftRepository) Create(shift *entity.Shift) error {
	return r.db.Create(shift).Error
}

func (r *shiftRepository) Update(shift *entity.Shift) error {
	return r.db.Save(shift).Error
}

func (r *shiftRepository) Delete(id uint) error {
	return r.db.Delete(&entity.Shift{}, id).Error
}

func (r *shiftRepository) IsUsed(id uint) (bool, error) {
	var count int64
	err := r.db.Table("trx_kedatangan_bbm").Where("shift_id = ?", id).Count(&count).Error
	return count > 0, err
}
