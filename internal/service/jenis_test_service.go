package service

import (
	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

// JenisTestService — business logic master data jenis test.
type JenisTestService interface {
	GetAll() ([]entity.JenisTest, error)
	GetActive() ([]entity.JenisTest, error)
	GetByID(id uint) (entity.JenisTest, error)
	Create(j *entity.JenisTest) error
	Update(j *entity.JenisTest) error
	Delete(id uint) error
}

type jenisTestService struct {
	repo repository.JenisTestRepository
}

func NewJenisTestService(repo repository.JenisTestRepository) JenisTestService {
	return &jenisTestService{repo}
}

func (s *jenisTestService) GetAll() ([]entity.JenisTest, error) {
	return s.repo.FindAll()
}

func (s *jenisTestService) GetActive() ([]entity.JenisTest, error) {
	return s.repo.FindActive()
}

func (s *jenisTestService) GetByID(id uint) (entity.JenisTest, error) {
	return s.repo.FindByID(id)
}

func (s *jenisTestService) Create(j *entity.JenisTest) error {
	return s.repo.Create(j)
}

func (s *jenisTestService) Update(j *entity.JenisTest) error {
	return s.repo.Update(j)
}

func (s *jenisTestService) Delete(id uint) error {
	return s.repo.Delete(id)
}
