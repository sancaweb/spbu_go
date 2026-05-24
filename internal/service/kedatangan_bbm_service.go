package service

import (
	"spbu_go/internal/dto"
	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

type KedatanganBBMService interface {
	Datatable(req dto.DatatableRequest) (int64, int64, []repository.KedatanganDTRow, error)
	GetByID(id uint64) (*entity.TrxKedatanganBBM, error)
	Create(k *entity.TrxKedatanganBBM) error
	Update(k *entity.TrxKedatanganBBM) error
	Delete(id uint64) error
	GetSOSisa() ([]repository.SOSisaRow, error)
	GetBBMSisaByPenebusan(penebusanID uint64) ([]repository.BBMSisaRow, error)
}

type kedatanganBBMService struct {
	repo repository.KedatanganBBMRepository
}

func NewKedatanganBBMService(repo repository.KedatanganBBMRepository) KedatanganBBMService {
	return &kedatanganBBMService{repo}
}

func (s *kedatanganBBMService) Datatable(req dto.DatatableRequest) (int64, int64, []repository.KedatanganDTRow, error) {
	return s.repo.Datatable(req)
}

func (s *kedatanganBBMService) GetByID(id uint64) (*entity.TrxKedatanganBBM, error) {
	return s.repo.FindByID(id)
}

func (s *kedatanganBBMService) Create(k *entity.TrxKedatanganBBM) error {
	return s.repo.Create(k)
}

func (s *kedatanganBBMService) Update(k *entity.TrxKedatanganBBM) error {
	return s.repo.Update(k)
}

func (s *kedatanganBBMService) Delete(id uint64) error {
	return s.repo.Delete(id)
}

func (s *kedatanganBBMService) GetSOSisa() ([]repository.SOSisaRow, error) {
	return s.repo.FindSOSisa()
}

func (s *kedatanganBBMService) GetBBMSisaByPenebusan(penebusanID uint64) ([]repository.BBMSisaRow, error) {
	return s.repo.FindBBMSisaByPenebusan(penebusanID)
}
