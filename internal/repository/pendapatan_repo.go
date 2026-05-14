package repository

import (
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

type PendapatanRepository interface {
	FindAll() ([]entity.Pendapatan, error)
	FindActive() ([]entity.Pendapatan, error)
	FindInactive() ([]entity.Pendapatan, error)
	FindByID(id uint) (entity.Pendapatan, error)
	Create(p *entity.Pendapatan) error
	Update(p *entity.Pendapatan) error
	Delete(id uint) error
}

type pendapatanRepository struct {
	db *gorm.DB
}

func NewPendapatanRepository(db *gorm.DB) PendapatanRepository {
	return &pendapatanRepository{db}
}

func (r *pendapatanRepository) FindAll() ([]entity.Pendapatan, error) {
	var data []entity.Pendapatan
	err := r.db.Preload("Updater").Order("nama_pendapatan ASC").Find(&data).Error
	return data, err
}

func (r *pendapatanRepository) FindActive() ([]entity.Pendapatan, error) {
	var data []entity.Pendapatan
	err := r.db.Where("is_active = ?", true).Preload("Updater").Order("nama_pendapatan ASC").Find(&data).Error
	return data, err
}

func (r *pendapatanRepository) FindInactive() ([]entity.Pendapatan, error) {
	var data []entity.Pendapatan
	err := r.db.Where("is_active = ?", false).Preload("Updater").Order("nama_pendapatan ASC").Find(&data).Error
	return data, err
}

func (r *pendapatanRepository) FindByID(id uint) (entity.Pendapatan, error) {
	var data entity.Pendapatan
	err := r.db.First(&data, id).Error
	return data, err
}

func (r *pendapatanRepository) Create(p *entity.Pendapatan) error {
	return r.db.Omit("Updater").Create(p).Error
}

func (r *pendapatanRepository) Update(p *entity.Pendapatan) error {
	return r.db.Omit("Updater").Save(p).Error
}

func (r *pendapatanRepository) Delete(id uint) error {
	return r.db.Delete(&entity.Pendapatan{}, id).Error
}
