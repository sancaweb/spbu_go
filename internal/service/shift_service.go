package service

import (
	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

type ShiftService interface {
	GetAll() ([]entity.Shift, error)
	GetByID(id uint) (entity.Shift, error)
	Create(shift *entity.Shift) error
	Update(shift *entity.Shift) error
	Delete(id uint) error
	IsUsed(id uint) (bool, error)
}

type shiftService struct {
	repo repository.ShiftRepository
}

func NewShiftService(repo repository.ShiftRepository) ShiftService {
	return &shiftService{repo}
}

func (s *shiftService) GetAll() ([]entity.Shift, error) {
	return s.repo.FindAll()
}

func (s *shiftService) GetByID(id uint) (entity.Shift, error) {
	return s.repo.FindByID(id)
}

func (s *shiftService) Create(shift *entity.Shift) error {
	return s.repo.Create(shift)
}

func (s *shiftService) Update(shift *entity.Shift) error {
	return s.repo.Update(shift)
}

func (s *shiftService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *shiftService) IsUsed(id uint) (bool, error) {
	return s.repo.IsUsed(id)
}
