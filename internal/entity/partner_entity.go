package entity

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrPartnerNotArchived = errors.New("Hanya partner dalam arsip yang dapat dihapus permanen")
	ErrPartnerHasPiutang  = errors.New("Partner tidak dapat dihapus permanen karena sudah memiliki transaksi piutang")
	ErrPartnerReferenced  = errors.New("Partner tidak dapat dihapus permanen karena masih digunakan oleh data lain")
)

type Partner struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"type:varchar(255);not null" json:"name" form:"name"`
	ContactPerson string         `gorm:"type:varchar(255)" json:"contact_person" form:"contact_person"`
	Phone         string         `gorm:"type:varchar(50)" json:"phone" form:"phone"`
	Address       string         `gorm:"type:text" json:"address" form:"address"`
	StartDate     *time.Time     `gorm:"column:start_date;type:date" json:"start_date" form:"start_date" time_format:"2006-01-02" time_utc:"1"`
	EndDate       *time.Time     `gorm:"column:end_date;type:date" json:"end_date" form:"end_date" time_format:"2006-01-02" time_utc:"1"`
	IsActive      bool           `gorm:"column:isactive;not null;default:true" json:"isactive" form:"isactive"`
	AutoOff       bool           `gorm:"column:auto_off;not null;default:false" json:"auto_off" form:"auto_off"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	UpdatedBy     *uint          `json:"updated_by"`
	Updater       *User          `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Partner) TableName() string {
	return "partners"
}
