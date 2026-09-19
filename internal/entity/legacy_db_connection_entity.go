package entity

import "time"

// LegacyDBConnection stores the connection profile used by the import module.
// The password is always persisted encrypted and is never serialized to JSON.
type LegacyDBConnection struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Name              string    `gorm:"type:varchar(100);not null;unique" json:"name"`
	Driver            string    `gorm:"type:varchar(20);not null" json:"driver"`
	Host              string    `gorm:"type:varchar(255);not null" json:"host"`
	Port              int       `gorm:"not null" json:"port"`
	DatabaseName      string    `gorm:"column:database_name;type:varchar(255);not null" json:"database_name"`
	Username          string    `gorm:"type:varchar(255);not null" json:"username"`
	EncryptedPassword string    `gorm:"column:encrypted_password;type:text;not null;default:''" json:"-"`
	SSLMode           string    `gorm:"column:ssl_mode;type:varchar(30);not null;default:'disable'" json:"ssl_mode"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (LegacyDBConnection) TableName() string {
	return "legacy_db_connections"
}
