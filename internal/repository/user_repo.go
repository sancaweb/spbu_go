package repository

import (
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db}
}

func (r *userRepository) FindAll() ([]entity.User, error) {
	var users []entity.User
	err := r.db.Preload("Role").Find(&users).Error
	return users, err
}

func (r *userRepository) FindByID(id uint) (*entity.User, error) {
	var user entity.User
	err := r.db.Preload("Role").First(&user, id).Error
	return &user, err
}

func (r *userRepository) FindByUsername(username string) (*entity.User, error) {
	var user entity.User
	err := r.db.Preload("Role.Permissions").Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *userRepository) Create(user *entity.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) Update(user *entity.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&entity.User{}, id).Error
}
