package service

import (
	"spbu_go/internal/dto"
	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

type PenebusanService interface {
	GetAll() ([]entity.TrxPenebusan, error)
	GetByID(id uint64) (*entity.TrxPenebusan, error)
	Datatable(req dto.DatatableRequest) (int64, int64, []entity.TrxPenebusan, error)
	Create(p *entity.TrxPenebusan) error
	Update(p *entity.TrxPenebusan) error
	Delete(id uint64) error
}

type penebusanService struct {
	repo repository.PenebusanRepository
}

func NewPenebusanService(repo repository.PenebusanRepository) PenebusanService {
	return &penebusanService{repo: repo}
}

func (s *penebusanService) GetAll() ([]entity.TrxPenebusan, error) {
	return s.repo.FindAll()
}

func (s *penebusanService) GetByID(id uint64) (*entity.TrxPenebusan, error) {
	return s.repo.FindByID(id)
}

func (s *penebusanService) Datatable(req dto.DatatableRequest) (int64, int64, []entity.TrxPenebusan, error) {
	return s.repo.Datatable(req)
}

func (s *penebusanService) Create(p *entity.TrxPenebusan) error {
	return s.repo.Create(p)
}

func (s *penebusanService) Update(p *entity.TrxPenebusan) error {
	return s.repo.Update(p)
}

func (s *penebusanService) Delete(id uint64) error {
	return s.repo.Delete(id)
}
