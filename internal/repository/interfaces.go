package repository

import "spbu_go/internal/entity"

type UserRepository interface {
	FindAll() ([]entity.User, error)
	FindByID(id uint) (*entity.User, error)
	FindByUsername(username string) (*entity.User, error)
	Create(user *entity.User) error
	Update(user *entity.User) error
	Delete(id uint) error
}

type RoleRepository interface {
	FindAll() ([]entity.Role, error)
	FindByID(id uint) (*entity.Role, error)
	Create(role *entity.Role) error
	Update(role *entity.Role) error
	Delete(id uint) error
	UpdatePermissions(roleID uint, permissionIDs []uint) error
}

type PermissionRepository interface {
	FindAll() ([]entity.Permission, error)
}

type SettingRepository interface {
	FindByKey(key string) (*entity.Setting, error)
	FindAll() ([]entity.Setting, error)
	Upsert(key, value string) error
}
