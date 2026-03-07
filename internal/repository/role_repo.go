package repository

import (
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db}
}

func (r *roleRepository) FindAll() ([]entity.Role, error) {
	var roles []entity.Role
	// Preload Permissions to show them in the list
	err := r.db.Preload("Permissions").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) FindByID(id uint) (*entity.Role, error) {
	var role entity.Role
	err := r.db.Preload("Permissions").First(&role, id).Error
	return &role, err
}

func (r *roleRepository) Create(role *entity.Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) Update(role *entity.Role) error {
	return r.db.Save(role).Error
}

func (r *roleRepository) Delete(id uint) error {
	return r.db.Delete(&entity.Role{}, id).Error
}

func (r *roleRepository) UpdatePermissions(roleID uint, permissionIDs []uint) error {
	var role entity.Role
	if err := r.db.First(&role, roleID).Error; err != nil {
		return err
	}

	var permissions []entity.Permission
	if len(permissionIDs) > 0 {
		if err := r.db.Find(&permissions, permissionIDs).Error; err != nil {
			return err
		}
	}

	// Replace association
	return r.db.Model(&role).Association("Permissions").Replace(permissions)
}
