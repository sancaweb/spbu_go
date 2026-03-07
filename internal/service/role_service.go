package service

import (
	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

type roleService struct {
	repo repository.RoleRepository
}

func NewRoleService(repo repository.RoleRepository) RoleService {
	return &roleService{repo}
}

func (s *roleService) GetAll() ([]entity.Role, error) {
	return s.repo.FindAll()
}

func (s *roleService) Create(role *entity.Role) error {
	return s.repo.Create(role)
}

func (s *roleService) Update(id uint, role *entity.Role) error {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	existing.Name = role.Name
	existing.Code = role.Code
	return s.repo.Update(existing)
}

func (s *roleService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *roleService) UpdatePermissions(roleID uint, permissionIDs []uint) error {
	return s.repo.UpdatePermissions(roleID, permissionIDs)
}
