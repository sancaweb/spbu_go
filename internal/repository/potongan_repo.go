package repository

import (
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

type PotonganRepository interface {
	FindAll() ([]entity.Potongan, error)
	FindActive() ([]entity.Potongan, error)
	FindInactive() ([]entity.Potongan, error)
	FindByID(id uint) (entity.Potongan, error)
	Create(p *entity.Potongan) error
	Update(p *entity.Potongan) error
	Delete(id uint) error
}

type potonganRepository struct {
	db *gorm.DB
}

func NewPotonganRepository(db *gorm.DB) PotonganRepository {
	return &potonganRepository{db}
}

func (r *potonganRepository) FindAll() ([]entity.Potongan, error) {
	var data []entity.Potongan
	err := r.db.Preload("Updater").Order("kode_potongan ASC").Find(&data).Error
	return data, err
}

func (r *potonganRepository) FindActive() ([]entity.Potongan, error) {
	var data []entity.Potongan
	err := r.db.Where("is_active = ?", true).Preload("Updater").Order("kode_potongan ASC").Find(&data).Error
	return data, err
}

func (r *potonganRepository) FindInactive() ([]entity.Potongan, error) {
	var data []entity.Potongan
	err := r.db.Where("is_active = ?", false).Preload("Updater").Order("kode_potongan ASC").Find(&data).Error
	return data, err
}

func (r *potonganRepository) FindByID(id uint) (entity.Potongan, error) {
	var data entity.Potongan
	err := r.db.First(&data, id).Error
	return data, err
}

func (r *potonganRepository) Create(p *entity.Potongan) error {
	return r.db.Omit("Updater").Create(p).Error
}

func (r *potonganRepository) Update(p *entity.Potongan) error {
	return r.db.Omit("Updater").Save(p).Error
}

func (r *potonganRepository) Delete(id uint) error {
	return r.db.Delete(&entity.Potongan{}, id).Error
}
