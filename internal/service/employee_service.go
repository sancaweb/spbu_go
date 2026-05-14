package service

import (
	"spbu_go/internal/dto"
	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

type KaryawanService interface {
	GetAll() ([]entity.Karyawan, error)
	GetActive() ([]entity.Karyawan, error)
	GetInactive() ([]entity.Karyawan, error)
	GetByID(id uint) (entity.Karyawan, error)
	Datatable(req dto.DatatableRequest, isActive bool) (int64, int64, []entity.Karyawan, error)
	Create(karyawan *entity.Karyawan) error
	Update(karyawan *entity.Karyawan) error
	Delete(id uint) error
	Restore(id uint) error
	SetPendapatans(karyawanID uint, ids []uint) error
	SetPotongans(karyawanID uint, ids []uint) error
}

type karyawanService struct {
	repo repository.KaryawanRepository
}

func NewKaryawanService(repo repository.KaryawanRepository) KaryawanService {
	return &karyawanService{repo}
}

func (s *karyawanService) GetAll() ([]entity.Karyawan, error) {
	return s.repo.FindAll()
}

func (s *karyawanService) GetActive() ([]entity.Karyawan, error) {
	return s.repo.FindActive()
}

func (s *karyawanService) GetInactive() ([]entity.Karyawan, error) {
	return s.repo.FindInactive()
}

func (s *karyawanService) GetByID(id uint) (entity.Karyawan, error) {
	return s.repo.FindByID(id)
}

func (s *karyawanService) Datatable(req dto.DatatableRequest, isActive bool) (int64, int64, []entity.Karyawan, error) {
	return s.repo.Datatable(req, isActive)
}

func (s *karyawanService) Create(karyawan *entity.Karyawan) error {
	return s.repo.Create(karyawan)
}

func (s *karyawanService) Update(karyawan *entity.Karyawan) error {
	return s.repo.Update(karyawan)
}

func (s *karyawanService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *karyawanService) Restore(id uint) error {
	return s.repo.Restore(id)
}

func (s *karyawanService) SetPendapatans(karyawanID uint, ids []uint) error {
	return s.repo.SetPendapatans(karyawanID, ids)
}

func (s *karyawanService) SetPotongans(karyawanID uint, ids []uint) error {
	return s.repo.SetPotongans(karyawanID, ids)
}
