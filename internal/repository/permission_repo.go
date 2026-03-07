package repository

import (
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db}
}

func (r *permissionRepository) FindAll() ([]entity.Permission, error) {
	var permissions []entity.Permission
	err := r.db.Find(&permissions).Error
	return permissions, err
}
