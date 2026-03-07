package entity

import (
	"time"

	"gorm.io/gorm"
)

type BBM struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"type:varchar(255);not null" json:"name" form:"name"`
	Margin        float64        `gorm:"type:decimal(15,2);default:0" json:"margin" form:"margin"`
	Price         float64        `gorm:"type:decimal(15,2);default:0" json:"price" form:"price"`
	Stock         int64          `gorm:"type:bigint;default:0" json:"stock" form:"stock"`
	RewardPercent float64        `gorm:"type:decimal(5,2);default:0" json:"reward_percent" form:"reward_percent"`
	IsActive      bool           `gorm:"default:true" json:"is_active" form:"is_active"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	UpdatedBy     *uint          `json:"updated_by"`
	Updater       *User          `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (BBM) TableName() string {
	return "bbm"
}
