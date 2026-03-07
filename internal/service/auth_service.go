package service

import (
	"errors"
	"spbu_go/internal/entity"
	"spbu_go/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo}
}

func (s *authService) Login(username, password string) (*entity.User, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}
