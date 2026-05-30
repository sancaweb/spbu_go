package entity

import "time"

// TrxPenjualan — Header transaksi penjualan BBM per shift.
//
// Alur jurnal (penjualan_tunai):
//
//	Dr  Kas / Penerimaan (debit_kas)                = total_rp_totalisator
//	    Cr  Persediaan BBM — [BBM] (kredit_persediaan) = jml_liter × (bbm_price - margin) per BBM
//	    Cr  Pendapatan Penjualan (kredit_pendapatan)   = jml_liter × margin per BBM
//
// Syarat balance: Σ Cr Persediaan + Σ Cr Pendapatan = Σ (harga_dasar + margin) × liter = total_rp_totalisator ✓
type TrxPenjualan struct {
	ID uint64 `gorm:"primaryKey;column:id_penjualan" json:"id"`

	// Nomor dokumen (auto-generated: PJL/YYYY/MM/NNNN)
	NoPenjualan string `gorm:"column:no_penjualan;type:varchar(25);uniqueIndex:uni_trx_penjualan_no_penjualan;not null" json:"no_penjualan"`

	// Shift & waktu
	ShiftID    uint      `gorm:"column:shift_id;not null;index" json:"shift_id"`
	Shift      *Shift    `gorm:"foreignKey:ShiftID" json:"shift,omitempty"`
	WaktuMulai time.Time `gorm:"column:waktu_mulai;type:timestamp;not null" json:"waktu_mulai"`
	WaktuAkhir time.Time `gorm:"column:waktu_akhir;type:timestamp;not null" json:"waktu_akhir"`

	// Kalkulasi (semua dalam Rupiah integer)
	TotalRpTotalisator int64 `gorm:"column:total_rp_totalisator;type:bigint;not null;default:0" json:"total_rp_totalisator"` // Σ jml_rupiah dari detail
	TotalPenerimaan    int64 `gorm:"column:total_penerimaan;type:bigint;not null;default:0" json:"total_penerimaan"`         // total seluruh penerimaan (tunai + non-tunai)
	AktualUang         int64 `gorm:"column:aktual_uang;type:bigint;not null;default:0" json:"aktual_uang"`                   // uang tunai fisik di kasir
	Selisih            int64 `gorm:"column:selisih;type:bigint;not null;default:0" json:"selisih"`                           // aktual_uang - total_rp_totalisator

	// Relasi detail
	Details []TrxPenjualanDetail `gorm:"foreignKey:PenjualanID;constraint:OnDelete:CASCADE" json:"details,omitempty"`

	// Audit
	Created   time.Time `gorm:"column:created;autoCreateTime" json:"created"`
	CreatedBy *uint     `gorm:"column:created_by" json:"created_by"`
	Creator   *User     `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Updated   time.Time `gorm:"column:updated;autoUpdateTime" json:"updated"`
	UpdatedBy *uint     `gorm:"column:updated_by" json:"updated_by"`
	Updater   *User     `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
}

func (TrxPenjualan) TableName() string { return "trx_penjualan" }

// TrxPenjualanDetail — Satu baris per nozzle pada transaksi penjualan.
type TrxPenjualanDetail struct {
	ID          uint64 `gorm:"primaryKey;column:detail_penjualan_id" json:"id"`
	PenjualanID uint64 `gorm:"column:penjualan_id;not null;index" json:"penjualan_id"`

	// Referensi tiang & nozzle
	TiangID  uint    `gorm:"column:tiang_id;not null" json:"tiang_id"`
	Tiang    *Tiang  `gorm:"foreignKey:TiangID" json:"tiang,omitempty"`
	NozzleID uint    `gorm:"column:nozzle_id;not null" json:"nozzle_id"`
	Nozzle   *Nozzle `gorm:"foreignKey:NozzleID" json:"nozzle,omitempty"`
	BBMID    uint    `gorm:"column:bbm_id;not null" json:"bbm_id"`
	BBM      *BBM    `gorm:"foreignKey:BBMID" json:"bbm,omitempty"`

	// Harga snapshot saat transaksi (integer Rp)
	BBMPrice int64 `gorm:"column:bbm_price;type:bigint;not null;default:0" json:"bbm_price"` // harga jual per liter
	Margin   int64 `gorm:"column:margin;type:bigint;not null;default:0" json:"margin"`       // margin per liter

	// Totalisator & volume
	TotalisatorAwal  int64 `gorm:"column:totalisator_awal;type:bigint;not null;default:0" json:"totalisator_awal"`
	TotalisatorAkhir int64 `gorm:"column:totalisator_akhir;type:bigint;not null;default:0" json:"totalisator_akhir"`
	JmlLiter         int64 `gorm:"column:jml_liter;type:bigint;not null;default:0" json:"jml_liter"`   // totalisator_akhir - totalisator_awal
	JmlRupiah        int64 `gorm:"column:jml_rupiah;type:bigint;not null;default:0" json:"jml_rupiah"` // jml_liter × bbm_price
}

func (TrxPenjualanDetail) TableName() string { return "trx_penjualan_detail" }
