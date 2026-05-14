package service

import (
	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

type PotonganService interface {
	GetAll() ([]entity.Potongan, error)
	GetActive() ([]entity.Potongan, error)
	GetInactive() ([]entity.Potongan, error)
	GetByID(id uint) (entity.Potongan, error)
	Create(p *entity.Potongan) error
	Update(p *entity.Potongan) error
	Delete(id uint) error
}

type potonganService struct {
	repo repository.PotonganRepository
}

func NewPotonganService(repo repository.PotonganRepository) PotonganService {
	return &potonganService{repo}
}

func (s *potonganService) GetAll() ([]entity.Potongan, error) {
	return s.repo.FindAll()
}

func (s *potonganService) GetActive() ([]entity.Potongan, error) {
	return s.repo.FindActive()
}

func (s *potonganService) GetInactive() ([]entity.Potongan, error) {
	return s.repo.FindInactive()
}

func (s *potonganService) GetByID(id uint) (entity.Potongan, error) {
	return s.repo.FindByID(id)
}

func (s *potonganService) Create(p *entity.Potongan) error {
	return s.repo.Create(p)
}

func (s *potonganService) Update(p *entity.Potongan) error {
	return s.repo.Update(p)
}

func (s *potonganService) Delete(id uint) error {
	return s.repo.Delete(id)
}
