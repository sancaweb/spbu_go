package entity

import (
	"time"

	"gorm.io/gorm"
)

type Partner struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"type:varchar(255);not null" json:"name" form:"name"`
	ContactPerson string         `gorm:"type:varchar(255)" json:"contact_person" form:"contact_person"`
	Phone         string         `gorm:"type:varchar(50)" json:"phone" form:"phone"`
	Address       string         `gorm:"type:text" json:"address" form:"address"`
	IsActive      bool           `gorm:"default:true" json:"is_active" form:"is_active"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	UpdatedBy     *uint          `json:"updated_by"`
	Updater       *User          `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Partner) TableName() string {
	return "partners"
}
