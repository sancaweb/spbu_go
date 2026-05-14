package service

import (
	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

type PendapatanService interface {
	GetAll() ([]entity.Pendapatan, error)
	GetActive() ([]entity.Pendapatan, error)
	GetInactive() ([]entity.Pendapatan, error)
	GetByID(id uint) (entity.Pendapatan, error)
	Create(p *entity.Pendapatan) error
	Update(p *entity.Pendapatan) error
	Delete(id uint) error
}

type pendapatanService struct {
	repo repository.PendapatanRepository
}

func NewPendapatanService(repo repository.PendapatanRepository) PendapatanService {
	return &pendapatanService{repo}
}

func (s *pendapatanService) GetAll() ([]entity.Pendapatan, error) {
	return s.repo.FindAll()
}

func (s *pendapatanService) GetActive() ([]entity.Pendapatan, error) {
	return s.repo.FindActive()
}

func (s *pendapatanService) GetInactive() ([]entity.Pendapatan, error) {
	return s.repo.FindInactive()
}

func (s *pendapatanService) GetByID(id uint) (entity.Pendapatan, error) {
	return s.repo.FindByID(id)
}

func (s *pendapatanService) Create(p *entity.Pendapatan) error {
	return s.repo.Create(p)
}

func (s *pendapatanService) Update(p *entity.Pendapatan) error {
	return s.repo.Update(p)
}

func (s *pendapatanService) Delete(id uint) error {
	return s.repo.Delete(id)
}
