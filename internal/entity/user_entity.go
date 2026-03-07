package entity

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	FirstName string         `gorm:"column:first_name;type:varchar(50)" json:"first_name" form:"first_name"`
	LastName  string         `gorm:"column:last_name;type:varchar(50)" json:"last_name" form:"last_name"`
	Username  string         `gorm:"unique;not null" json:"username"`
	Password  string         `gorm:"not null" json:"-"`
	Email     string         `gorm:"column:email;type:varchar(254)" json:"email" form:"email"`
	Phone     string         `gorm:"column:phone;type:varchar(20)" json:"phone" form:"phone"`
	IsActive  bool           `gorm:"column:active;default:true" json:"is_active" form:"is_active"`
	RoleID    uint           `gorm:"not null" json:"role_id"`
	Role      Role           `gorm:"foreignKey:RoleID" json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// FullName returns the display name
func (u User) FullName() string {
	if u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	return u.FirstName
}
