package entity

import "time"

// TrxKedatanganBBM — transaksi pencatatan kedatangan/pengiriman BBM dari Pertamina.
// Setiap record mewakili satu kedatangan BBM untuk satu jenis BBM dari satu penebusan.
type TrxKedatanganBBM struct {
	ID                uint64    `gorm:"primaryKey;column:id_kedatangan_bbm" json:"id"`
	PenebusanID       uint64    `gorm:"not null;index;column:penebusan_id" json:"penebusan_id"`
	PenebusanDetailID uint64    `gorm:"not null;index;column:penebusan_detail_id" json:"penebusan_detail_id"`
	NoLO              string    `gorm:"type:varchar(50);not null;column:no_lo" json:"no_lo" form:"no_lo"`
	TglKedatangan     time.Time `gorm:"type:timestamp;not null;column:tgl_kedatangan" json:"tgl_kedatangan"`
	ShiftID           uint      `gorm:"not null;index;column:shift_id" json:"shift_id"`
	BBMID             uint      `gorm:"not null;index;column:bbm_id" json:"bbm_id"`
	JmlLiter          int64     `gorm:"type:bigint;not null;default:0;column:jml_liter" json:"jml_liter" form:"jml_liter"`
	NamaDriver        string    `gorm:"type:varchar(100);column:nama_driver" json:"nama_driver" form:"nama_driver"`
	NoPol             string    `gorm:"type:varchar(20);column:no_pol" json:"no_pol" form:"no_pol"`
	CreatedAt         time.Time `gorm:"column:created" json:"created"`
	UpdatedAt         time.Time `gorm:"column:updated" json:"updated"`
	CreatedBy         *uint     `gorm:"column:created_by" json:"created_by"`
	UpdatedBy         *uint     `gorm:"column:updated_by" json:"updated_by"`

	// Relations (for preload)
	Penebusan       *TrxPenebusan       `gorm:"foreignKey:PenebusanID" json:"penebusan,omitempty"`
	PenebusanDetail *TrxPenebusanDetail `gorm:"foreignKey:PenebusanDetailID" json:"penebusan_detail,omitempty"`
	Shift           *Shift              `gorm:"foreignKey:ShiftID" json:"shift,omitempty"`
	BBM             *BBM                `gorm:"foreignKey:BBMID" json:"bbm,omitempty"`
	Creator         *User               `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Updater         *User               `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
}

func (TrxKedatanganBBM) TableName() string {
	return "trx_kedatangan_bbm"
}
