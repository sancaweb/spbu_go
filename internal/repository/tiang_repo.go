package repository

import (
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

type TiangRepository interface {
	FindAll() ([]entity.Tiang, error)
	FindByID(id uint) (entity.Tiang, error)
	Create(tiang *entity.Tiang) error
	Update(tiang *entity.Tiang) error
	Delete(id uint) error
}

type tiangRepository struct {
	db *gorm.DB
}

func NewTiangRepository(db *gorm.DB) TiangRepository {
	return &tiangRepository{db}
}

func (r *tiangRepository) FindAll() ([]entity.Tiang, error) {
	var tiangs []entity.Tiang
	err := r.db.Preload("Nozzles", func(db *gorm.DB) *gorm.DB {
		return db.Order("description ASC")
	}).Preload("Nozzles.BBM").Preload("Updater").Preload("Nozzles.Updater").Find(&tiangs).Error
	return tiangs, err
}

func (r *tiangRepository) FindByID(id uint) (entity.Tiang, error) {
	var tiang entity.Tiang
	err := r.db.Preload("Nozzles.BBM").Preload("Updater").First(&tiang, id).Error
	return tiang, err
}

func (r *tiangRepository) Create(tiang *entity.Tiang) error {
	return r.db.Create(tiang).Error
}

func (r *tiangRepository) Update(tiang *entity.Tiang) error {
	return r.db.Save(tiang).Error
}

func (r *tiangRepository) Delete(id uint) error {
	return r.db.Delete(&entity.Tiang{}, id).Error
}
