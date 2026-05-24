package entity

import "time"

// Shift — master data shift kerja karyawan SPBU.
type Shift struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ShiftName string    `gorm:"type:varchar(100);not null" json:"shift_name" form:"shift_name"`
	ShiftTime string    `gorm:"type:varchar(50);not null" json:"shift_time" form:"shift_time"` // e.g. "07:00 - 15:00"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Shift) TableName() string {
	return "shifts"
}
