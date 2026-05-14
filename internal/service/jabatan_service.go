package service

import (
	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

type JabatanService interface {
	GetAll() ([]entity.Jabatan, error)
	GetActive() ([]entity.Jabatan, error)
	GetInactive() ([]entity.Jabatan, error)
	GetByID(id uint) (entity.Jabatan, error)
	Create(jabatan *entity.Jabatan) error
	Update(jabatan *entity.Jabatan) error
	Delete(id uint) error
}

type jabatanService struct {
	repo repository.JabatanRepository
}

func NewJabatanService(repo repository.JabatanRepository) JabatanService {
	return &jabatanService{repo}
}

func (s *jabatanService) GetAll() ([]entity.Jabatan, error) {
	return s.repo.FindAll()
}

func (s *jabatanService) GetActive() ([]entity.Jabatan, error) {
	return s.repo.FindActive()
}

func (s *jabatanService) GetInactive() ([]entity.Jabatan, error) {
	return s.repo.FindInactive()
}

func (s *jabatanService) GetByID(id uint) (entity.Jabatan, error) {
	return s.repo.FindByID(id)
}

func (s *jabatanService) Create(jabatan *entity.Jabatan) error {
	return s.repo.Create(jabatan)
}

func (s *jabatanService) Update(jabatan *entity.Jabatan) error {
	return s.repo.Update(jabatan)
}

func (s *jabatanService) Delete(id uint) error {
	return s.repo.Delete(id)
}
