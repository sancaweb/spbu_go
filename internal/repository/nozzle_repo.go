package repository

import (
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

type NozzleRepository interface {
	Create(nozzle *entity.Nozzle) error
	Update(nozzle *entity.Nozzle) error
	Delete(id uint) error
}

type nozzleRepository struct {
	db *gorm.DB
}

func NewNozzleRepository(db *gorm.DB) NozzleRepository {
	return &nozzleRepository{db}
}

func (r *nozzleRepository) Create(nozzle *entity.Nozzle) error {
	return r.db.Create(nozzle).Error
}

func (r *nozzleRepository) Update(nozzle *entity.Nozzle) error {
	return r.db.Save(nozzle).Error
}

func (r *nozzleRepository) Delete(id uint) error {
	return r.db.Delete(&entity.Nozzle{}, id).Error
}
