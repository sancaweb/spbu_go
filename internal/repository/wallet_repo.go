package repository

import (
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

type WalletRepository interface {
	FindAll() ([]entity.Wallet, error)
	FindByID(id uint) (*entity.Wallet, error)
	Create(wallet *entity.Wallet) error
	Update(wallet *entity.Wallet) error
	Delete(id uint) error
}

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepository{db}
}

func (r *walletRepository) FindAll() ([]entity.Wallet, error) {
	var wallets []entity.Wallet
	err := r.db.Order("is_default DESC, wallet_name ASC").Find(&wallets).Error
	return wallets, err
}

func (r *walletRepository) FindByID(id uint) (*entity.Wallet, error) {
	var wallet entity.Wallet
	err := r.db.First(&wallet, id).Error
	return &wallet, err
}

func (r *walletRepository) Create(wallet *entity.Wallet) error {
	return r.db.Omit("Updater").Create(wallet).Error
}

func (r *walletRepository) Update(wallet *entity.Wallet) error {
	return r.db.Omit("Updater").Save(wallet).Error
}

func (r *walletRepository) Delete(id uint) error {
	return r.db.Delete(&entity.Wallet{}, id).Error
}
