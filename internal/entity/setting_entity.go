package entity

import (
	"time"

	"gorm.io/gorm"
)

type Setting struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	SettingName  string         `gorm:"column:setting_name;type:varchar(100);unique;not null" json:"setting_name"`
	SettingValue string         `gorm:"column:setting_value;type:varchar(255);not null;default:''" json:"setting_value"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Setting) TableName() string {
	return "settings"
}
