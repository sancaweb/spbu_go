package repository

import (
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

type BBMRepository interface {
	FindAll() ([]entity.BBM, error)
	FindActive() ([]entity.BBM, error)
	FindByID(id uint) (entity.BBM, error)
	Create(bbm *entity.BBM) error
	Update(bbm *entity.BBM) error
	Delete(id uint) error
}

type bbmRepository struct {
	db *gorm.DB
}

func NewBBMRepository(db *gorm.DB) BBMRepository {
	return &bbmRepository{db}
}

func (r *bbmRepository) FindAll() ([]entity.BBM, error) {
	var bbms []entity.BBM
	err := r.db.Preload("Updater").Find(&bbms).Error
	return bbms, err
}

func (r *bbmRepository) FindActive() ([]entity.BBM, error) {
	var bbms []entity.BBM
	err := r.db.Where("is_active = ?", true).Preload("Updater").Find(&bbms).Error
	return bbms, err
}

func (r *bbmRepository) FindByID(id uint) (entity.BBM, error) {
	var bbm entity.BBM
	err := r.db.Preload("Updater").First(&bbm, id).Error
	return bbm, err
}

func (r *bbmRepository) Create(bbm *entity.BBM) error {
	return r.db.Create(bbm).Error
}

func (r *bbmRepository) Update(bbm *entity.BBM) error {
	return r.db.Save(bbm).Error
}

func (r *bbmRepository) Delete(id uint) error {
	return r.db.Delete(&entity.BBM{}, id).Error
}
