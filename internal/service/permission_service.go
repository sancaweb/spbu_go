package service

import (
	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

type permissionService struct {
	repo repository.PermissionRepository
}

func NewPermissionService(repo repository.PermissionRepository) PermissionService {
	return &permissionService{repo}
}

func (s *permissionService) GetAll() ([]entity.Permission, error) {
	return s.repo.FindAll()
}
