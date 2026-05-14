package entity

import (
	"time"

	"gorm.io/gorm"
)

type Wallet struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	WalletName  string         `gorm:"type:varchar(100);not null" json:"wallet_name" form:"wallet_name"`
	IsDefault   bool           `gorm:"default:false" json:"is_default" form:"is_default"`
	Description string         `gorm:"type:varchar(250)" json:"description" form:"description"`
	Saldo       int64          `gorm:"type:bigint;default:0" json:"saldo" form:"saldo"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	UpdatedBy   *uint          `json:"updated_by"`
	Updater     *User          `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Wallet) TableName() string {
	return "wallets"
}
