package repository

import (
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

// JenisTestRepository — kontrak akses data master jenis test.
type JenisTestRepository interface {
	FindAll() ([]entity.JenisTest, error)
	FindActive() ([]entity.JenisTest, error)
	FindByID(id uint) (entity.JenisTest, error)
	Create(j *entity.JenisTest) error
	Update(j *entity.JenisTest) error
	Delete(id uint) error
}

type jenisTestRepository struct {
	db *gorm.DB
}

func NewJenisTestRepository(db *gorm.DB) JenisTestRepository {
	return &jenisTestRepository{db}
}

func (r *jenisTestRepository) FindAll() ([]entity.JenisTest, error) {
	var list []entity.JenisTest
	err := r.db.Order("nama_test ASC").Find(&list).Error
	return list, err
}

func (r *jenisTestRepository) FindActive() ([]entity.JenisTest, error) {
	var list []entity.JenisTest
	err := r.db.Where("is_active = ?", true).Order("nama_test ASC").Find(&list).Error
	return list, err
}

func (r *jenisTestRepository) FindByID(id uint) (entity.JenisTest, error) {
	var j entity.JenisTest
	err := r.db.First(&j, id).Error
	return j, err
}

func (r *jenisTestRepository) Create(j *entity.JenisTest) error {
	return r.db.Create(j).Error
}

func (r *jenisTestRepository) Update(j *entity.JenisTest) error {
	return r.db.Save(j).Error
}

func (r *jenisTestRepository) Delete(id uint) error {
	return r.db.Delete(&entity.JenisTest{}, id).Error
}
