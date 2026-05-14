package service

import (
	"spbu_go/internal/dto"
	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

type PartnerService interface {
	GetAll() ([]entity.Partner, error)
	GetActive() ([]entity.Partner, error)
	GetInactive() ([]entity.Partner, error)
	Datatable(req dto.DatatableRequest, isActive bool) (int64, int64, []entity.Partner, error)
	GetByID(id uint) (entity.Partner, error)
	Create(partner *entity.Partner) error
	Update(partner *entity.Partner) error
	Delete(id uint) error
	Restore(id uint) error
}

type partnerService struct {
	repo repository.PartnerRepository
}

func NewPartnerService(repo repository.PartnerRepository) PartnerService {
	return &partnerService{repo}
}

func (s *partnerService) GetAll() ([]entity.Partner, error) {
	return s.repo.FindAll()
}

func (s *partnerService) GetActive() ([]entity.Partner, error) {
	return s.repo.FindActive()
}

func (s *partnerService) GetInactive() ([]entity.Partner, error) {
	return s.repo.FindInactive()
}

func (s *partnerService) Datatable(req dto.DatatableRequest, isActive bool) (int64, int64, []entity.Partner, error) {
	return s.repo.Datatable(req, isActive)
}

func (s *partnerService) GetByID(id uint) (entity.Partner, error) {
	return s.repo.FindByID(id)
}

func (s *partnerService) Create(partner *entity.Partner) error {
	return s.repo.Create(partner)
}

func (s *partnerService) Update(partner *entity.Partner) error {
	return s.repo.Update(partner)
}

func (s *partnerService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *partnerService) Restore(id uint) error {
	return s.repo.Restore(id)
}
