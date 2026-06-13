package entity

import "time"

// TrxPenyusutan — pencatatan penyusutan / susut BBM harian per shift.
// Penyusutan terjadi karena penguapan, selisih takaran, atau kebocoran kecil.
type TrxPenyusutan struct {
	ID uint64 `gorm:"primaryKey;column:id_penyusutan" json:"id"`

	// Nomor dokumen (auto: PST/YYYY/MM/NNNN)
	NoPenyusutan string `gorm:"column:no_penyusutan;type:varchar(25);uniqueIndex:uni_trx_penyusutan_no;not null" json:"no_penyusutan"`

	// Referensi shift & tanggal
	TglPenyusutan time.Time `gorm:"column:tgl_penyusutan;type:date;not null" json:"tgl_penyusutan"`
	ShiftID       uint      `gorm:"column:shift_id;not null" json:"shift_id"`
	Shift         *Shift    `gorm:"foreignKey:ShiftID" json:"shift,omitempty"`

	// BBM yang mengalami penyusutan
	BBMID uint `gorm:"column:bbm_id;not null" json:"bbm_id"`
	BBM   *BBM `gorm:"foreignKey:BBMID" json:"bbm,omitempty"`

	// Volume & nilai
	JmlLiter    int64  `gorm:"column:jml_liter;type:bigint;not null;default:0" json:"jml_liter"`       // satuan: liter (integer, bisa × desimal)
	HargaDasar  int64  `gorm:"column:harga_dasar;type:bigint;not null;default:0" json:"harga_dasar"`   // harga dasar per liter saat itu
	NilaiRupiah int64  `gorm:"column:nilai_rupiah;type:bigint;not null;default:0" json:"nilai_rupiah"` // jml_liter × harga_dasar
	Keterangan  string `gorm:"column:keterangan;type:text" json:"keterangan"`

	// Audit
	Created   time.Time `gorm:"column:created;autoCreateTime" json:"created"`
	CreatedBy *uint     `gorm:"column:created_by" json:"created_by"`
	Creator   *User     `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Updated   time.Time `gorm:"column:updated;autoUpdateTime" json:"updated"`
	UpdatedBy *uint     `gorm:"column:updated_by" json:"updated_by"`
	Updater   *User     `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
}

func (TrxPenyusutan) TableName() string { return "trx_penyusutan" }
