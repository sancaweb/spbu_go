package service

import (
	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

type BBMService interface {
	GetAll() ([]entity.BBM, error)
	GetActive() ([]entity.BBM, error)
	GetByID(id uint) (entity.BBM, error)
	Create(bbm *entity.BBM) error
	Update(bbm *entity.BBM) error
	Delete(id uint) error
}

type bbmService struct {
	repo repository.BBMRepository
}

func NewBBMService(repo repository.BBMRepository) BBMService {
	return &bbmService{repo}
}

func (s *bbmService) GetAll() ([]entity.BBM, error) {
	return s.repo.FindAll()
}

func (s *bbmService) GetActive() ([]entity.BBM, error) {
	return s.repo.FindActive()
}

func (s *bbmService) GetByID(id uint) (entity.BBM, error) {
	return s.repo.FindByID(id)
}

func (s *bbmService) Create(bbm *entity.BBM) error {
	return s.repo.Create(bbm)
}

func (s *bbmService) Update(bbm *entity.BBM) error {
	return s.repo.Update(bbm)
}

func (s *bbmService) Delete(id uint) error {
	return s.repo.Delete(id)
}
