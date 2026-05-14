package service

import (
	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

type WalletService interface {
	GetAll() ([]entity.Wallet, error)
	GetByID(id uint) (*entity.Wallet, error)
	Create(wallet *entity.Wallet) error
	Update(id uint, wallet *entity.Wallet) error
	Delete(id uint) error
}

type walletService struct {
	repo repository.WalletRepository
}

func NewWalletService(repo repository.WalletRepository) WalletService {
	return &walletService{repo}
}

func (s *walletService) GetAll() ([]entity.Wallet, error) {
	return s.repo.FindAll()
}

func (s *walletService) GetByID(id uint) (*entity.Wallet, error) {
	return s.repo.FindByID(id)
}

func (s *walletService) Create(wallet *entity.Wallet) error {
	return s.repo.Create(wallet)
}

func (s *walletService) Update(id uint, wallet *entity.Wallet) error {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	existing.WalletName = wallet.WalletName
	existing.IsDefault = wallet.IsDefault
	existing.Description = wallet.Description
	existing.Saldo = wallet.Saldo
	existing.UpdatedBy = wallet.UpdatedBy
	return s.repo.Update(existing)
}

func (s *walletService) Delete(id uint) error {
	return s.repo.Delete(id)
}
