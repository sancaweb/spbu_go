package service

import (
	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

type NozzleService interface {
	Create(nozzle *entity.Nozzle) error
	Update(nozzle *entity.Nozzle) error
	Delete(id uint) error
}

type nozzleService struct {
	repo repository.NozzleRepository
}

func NewNozzleService(repo repository.NozzleRepository) NozzleService {
	return &nozzleService{repo}
}

func (s *nozzleService) Create(nozzle *entity.Nozzle) error {
	return s.repo.Create(nozzle)
}

func (s *nozzleService) Update(nozzle *entity.Nozzle) error {
	return s.repo.Update(nozzle)
}

func (s *nozzleService) Delete(id uint) error {
	return s.repo.Delete(id)
}
