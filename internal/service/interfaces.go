package service

import "spbu_go/internal/entity"

type UserService interface {
	GetAll() ([]entity.User, error)
	GetByID(id uint) (*entity.User, error)
	Create(user *entity.User) error
	Update(id uint, user *entity.User) error
	Delete(id uint) error
}

type RoleService interface {
	GetAll() ([]entity.Role, error)
	Create(role *entity.Role) error
	Update(id uint, role *entity.Role) error
	Delete(id uint) error
	UpdatePermissions(roleID uint, permissionIDs []uint) error
}

type AuthService interface {
	Login(username, password string) (*entity.User, error)
}

type PermissionService interface {
	GetAll() ([]entity.Permission, error)
}

type SettingService interface {
	Get(key string) (string, error)
	Set(key, value string) error
	GetInt(key string, defaultVal int) int
	GetAll() ([]entity.Setting, error)
}
