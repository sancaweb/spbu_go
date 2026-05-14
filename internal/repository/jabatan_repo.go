package repository

import (
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

type JabatanRepository interface {
	FindAll() ([]entity.Jabatan, error)
	FindActive() ([]entity.Jabatan, error)
	FindInactive() ([]entity.Jabatan, error)
	FindByID(id uint) (entity.Jabatan, error)
	Create(jabatan *entity.Jabatan) error
	Update(jabatan *entity.Jabatan) error
	Delete(id uint) error
}

type jabatanRepository struct {
	db *gorm.DB
}

func NewJabatanRepository(db *gorm.DB) JabatanRepository {
	return &jabatanRepository{db}
}

func (r *jabatanRepository) FindAll() ([]entity.Jabatan, error) {
	var data []entity.Jabatan
	err := r.db.Order("kode_jabatan ASC").Find(&data).Error
	return data, err
}

func (r *jabatanRepository) FindActive() ([]entity.Jabatan, error) {
	var data []entity.Jabatan
	err := r.db.Where("is_active = ?", true).Order("kode_jabatan ASC").Find(&data).Error
	return data, err
}

func (r *jabatanRepository) FindInactive() ([]entity.Jabatan, error) {
	var data []entity.Jabatan
	err := r.db.Where("is_active = ?", false).Order("kode_jabatan ASC").Find(&data).Error
	return data, err
}

func (r *jabatanRepository) FindByID(id uint) (entity.Jabatan, error) {
	var data entity.Jabatan
	err := r.db.First(&data, id).Error
	return data, err
}

func (r *jabatanRepository) Create(jabatan *entity.Jabatan) error {
	return r.db.Create(jabatan).Error
}

func (r *jabatanRepository) Update(jabatan *entity.Jabatan) error {
	return r.db.Save(jabatan).Error
}

func (r *jabatanRepository) Delete(id uint) error {
	return r.db.Delete(&entity.Jabatan{}, id).Error
}
