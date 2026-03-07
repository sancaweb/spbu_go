package entity

import (
	"time"

	"gorm.io/gorm"
)

type Tiang struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name" form:"name"`
	Slug      string         `gorm:"type:varchar(255);unique;not null" json:"slug" form:"slug"`
	Nozzles   []Nozzle       `gorm:"foreignKey:TiangID" json:"nozzles"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	UpdatedBy *uint          `json:"updated_by"`
	Updater   *User          `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Tiang) TableName() string {
	return "tiang"
}

type Nozzle struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TiangID     uint           `gorm:"not null" json:"tiang_id" form:"tiang_id"`
	Tiang       *Tiang         `gorm:"foreignKey:TiangID" json:"tiang,omitempty"`
	Description string         `gorm:"type:varchar(255)" json:"description" form:"description"`
	BBMID       uint           `gorm:"not null" json:"bbm_id" form:"bbm_id"`
	BBM         *BBM           `gorm:"foreignKey:BBMID" json:"bbm,omitempty"`
	IsActive    bool           `gorm:"default:true" json:"is_active" form:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	UpdatedBy   *uint          `json:"updated_by"`
	Updater     *User          `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Nozzle) TableName() string {
	return "nozzles"
}
