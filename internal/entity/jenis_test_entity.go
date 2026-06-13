package entity

import (
	"time"

	"gorm.io/gorm"
)

// JenisTest — master data jenis pengujian/kalibrasi BBM.
// Contoh: Density Pengawas, Tera Metrologi, Tes Nozzle, dll.
type JenisTest struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	NamaTest  string         `gorm:"type:varchar(100);not null" json:"nama_test" form:"nama_test"`
	Deskripsi string         `gorm:"type:text" json:"deskripsi" form:"deskripsi"`
	IsActive  bool           `gorm:"default:true" json:"is_active" form:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (JenisTest) TableName() string { return "jenis_test" }
