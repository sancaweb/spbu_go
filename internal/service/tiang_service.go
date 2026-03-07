package service

import (
	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

type TiangService interface {
	GetAll() ([]entity.Tiang, error)
	GetByID(id uint) (entity.Tiang, error)
	Create(tiang *entity.Tiang) error
	Update(tiang *entity.Tiang) error
	Delete(id uint) error
}

type tiangService struct {
	repo repository.TiangRepository
}

func NewTiangService(repo repository.TiangRepository) TiangService {
	return &tiangService{repo}
}

func (s *tiangService) GetAll() ([]entity.Tiang, error) {
	return s.repo.FindAll()
}

func (s *tiangService) GetByID(id uint) (entity.Tiang, error) {
	return s.repo.FindByID(id)
}

func (s *tiangService) Create(tiang *entity.Tiang) error {
	return s.repo.Create(tiang)
}

func (s *tiangService) Update(tiang *entity.Tiang) error {
	return s.repo.Update(tiang)
}

func (s *tiangService) Delete(id uint) error {
	return s.repo.Delete(id)
}
