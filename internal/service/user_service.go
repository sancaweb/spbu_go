package service

import (
	"spbu_go/internal/entity"
	"spbu_go/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo}
}

func (s *userService) GetAll() ([]entity.User, error) {
	return s.repo.FindAll()
}

func (s *userService) GetByID(id uint) (*entity.User, error) {
	return s.repo.FindByID(id)
}

func (s *userService) Create(user *entity.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	return s.repo.Create(user)
}

func (s *userService) Update(id uint, user *entity.User) error {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	existing.FirstName = user.FirstName
	existing.LastName = user.LastName
	existing.Username = user.Username
	existing.Email = user.Email
	existing.Phone = user.Phone
	existing.IsActive = user.IsActive
	existing.RoleID = user.RoleID

	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		existing.Password = string(hashedPassword)
	}

	return s.repo.Update(existing)
}

func (s *userService) Delete(id uint) error {
	return s.repo.Delete(id)
}
