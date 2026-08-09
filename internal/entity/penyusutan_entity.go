package entity

import "time"

// TrxPenyusutan — snapshot laporan penyusutan BBM berbasis transaksi penjualan.
type TrxPenyusutan struct {
	ID uint64 `gorm:"primaryKey;column:id_penyusutan" json:"id_penyusutan"`

	// Referensi dokumen penjualan (header)
	PenjualanID uint64        `gorm:"column:penjualan_id;not null;index" json:"penjualan_id"`
	Penjualan   *TrxPenjualan `gorm:"foreignKey:PenjualanID" json:"penjualan,omitempty"`

	// Nomor form laporan penyusutan
	NoForm string `gorm:"column:no_form;type:varchar(50);not null" json:"no_form"`

	// Referensi shift & waktu pencatatan
	ShiftID uint      `gorm:"column:shift_id;not null" json:"shift_id"`
	Shift   *Shift    `gorm:"foreignKey:ShiftID" json:"shift,omitempty"`
	Waktu   time.Time `gorm:"column:waktu;type:timestamp;not null;index" json:"waktu"`

	// BBM yang mengalami penyusutan
	BBMID uint `gorm:"column:bbm_id;not null" json:"bbm_id"`
	BBM   *BBM `gorm:"foreignKey:BBMID" json:"bbm,omitempty"`

	// Snapshot stok
	FirstStock     float64 `gorm:"column:first_stock;type:numeric(20,2);not null;default:0" json:"first_stock"`
	EndstockActual float64 `gorm:"column:endstock_actual;type:numeric(20,2);not null;default:0" json:"endstock_actual"`
	EndstockBooked float64 `gorm:"column:endstock_booked;type:numeric(20,2);not null;default:0" json:"endstock_booked"`

	// Audit
	Created   time.Time `gorm:"column:created;autoCreateTime" json:"created"`
	CreatedBy *uint     `gorm:"column:created_by" json:"created_by"`
	Creator   *User     `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Updated   time.Time `gorm:"column:updated;autoUpdateTime" json:"updated"`
	UpdatedBy *uint     `gorm:"column:updated_by" json:"updated_by"`
	Updater   *User     `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
}

func (TrxPenyusutan) TableName() string { return "trx_penyusutan" }
