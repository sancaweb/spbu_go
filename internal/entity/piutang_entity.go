package entity

import "time"

// ─── TrxPiutang — Header piutang B2B per partner per shift/penjualan ─────────
//
// Satu penjualan (shift) dapat menghasilkan beberapa piutang (satu per partner).
// Detail mencatat setiap voucher pengisian per kendaraan partner.
//
// Alur jurnal saat Create (penjualan_kredit):
//
//	Dr  Piutang Dagang B2B (debit_piutang)           = total_tagihan
//	    Cr  Persediaan BBM — [BBM] (kredit_persediaan) = harga_dasar × qty per BBM
//	    Cr  Pendapatan Penjualan Kredit (kredit_pendapatan) = margin × qty per BBM
//
// Alur jurnal saat Lunas (pelunasan_piutang):
//
//	Dr  Kas/Bank (debit_kas)                          = total_tagihan
//	    Cr  Piutang Dagang B2B (kredit_piutang)       = total_tagihan

const (
	PiutangUnpaid = "unpaid"
	PiutangPaid   = "paid"
)

type TrxPiutang struct {
	IDPiutang uint64 `gorm:"primaryKey;column:id_piutang" json:"id_piutang"`

	// Penjualan asal (shift yang menghasilkan piutang ini)
	PenjualanID uint64        `gorm:"column:penjualan_id;not null;index" json:"penjualan_id"`
	Penjualan   *TrxPenjualan `gorm:"foreignKey:PenjualanID" json:"penjualan,omitempty"`

	// Pelanggan kredit. Di aplikasi ini master pelanggan B2B disimpan di tabel partners.
	PelangganID uint     `gorm:"column:pelanggan_id;not null;index" json:"pelanggan_id"`
	Partner     *Partner `gorm:"foreignKey:PelangganID" json:"partner,omitempty"`

	// Nilai total tagihan (Rp integer)
	TotalTagihan int64 `gorm:"column:total_tagihan;type:bigint;not null;default:0" json:"total_tagihan"`

	// Status: PiutangUnpaid ("unpaid") | PiutangPaid ("paid")
	Status string `gorm:"column:status;type:varchar(10);not null;default:'unpaid'" json:"status"`

	// Flag: sudah dicetak invoice atau belum (fitur invoice terpisah)
	IsInvoiced bool `gorm:"column:isinvoiced;default:false" json:"isinvoiced"`

	// Relasi detail (voucher per kendaraan)
	Details []TrxPiutangDetail `gorm:"foreignKey:PiutangID;constraint:OnDelete:CASCADE" json:"details,omitempty"`

	// Audit
	Created   time.Time `gorm:"column:created;autoCreateTime" json:"created"`
	CreatedBy *uint     `gorm:"column:created_by" json:"created_by"`
	Creator   *User     `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Updated   time.Time `gorm:"column:updated;autoUpdateTime" json:"updated"`
	UpdatedBy *uint     `gorm:"column:updated_by" json:"updated_by"`
	Updater   *User     `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
}

func (TrxPiutang) TableName() string { return "trx_piutang" }

// ─── TrxPiutangDetail — Satu voucher pengisian per kendaraan ─────────────────

type TrxPiutangDetail struct {
	IDPiutangDetail uint64 `gorm:"primaryKey;column:id_piutang_detail" json:"id_piutang_detail"`

	// Parent piutang
	PiutangID uint64      `gorm:"column:piutang_id;not null;index" json:"piutang_id"`
	Piutang   *TrxPiutang `gorm:"foreignKey:PiutangID" json:"piutang,omitempty"`

	// Penjualan (harus sama dengan header, untuk tracing)
	PenjualanID uint64        `gorm:"column:penjualan_id;not null" json:"penjualan_id"`
	Penjualan   *TrxPenjualan `gorm:"foreignKey:PenjualanID" json:"penjualan,omitempty"`

	// Data kendaraan / voucher
	NoVoucher  string `gorm:"column:no_voucher;type:varchar(50)" json:"no_voucher"`
	NoPol      string `gorm:"column:no_pol;type:varchar(20)" json:"no_pol"`
	DriverName string `gorm:"column:driver_name;type:varchar(100)" json:"driver_name"`

	// BBM & kalkulasi
	BBMID     uint  `gorm:"column:bbm_id;not null" json:"bbm_id"`
	BBM       *BBM  `gorm:"foreignKey:BBMID" json:"bbm,omitempty"`
	HargaBBM  int64 `gorm:"column:harga_bbm;type:bigint;not null;default:0" json:"harga_bbm"` // harga jual per liter saat transaksi
	Margin    int64 `gorm:"column:margin;type:bigint;not null;default:0" json:"margin"`       // margin per liter (untuk jurnal HPP)
	QtyLiter  int64 `gorm:"column:qty_liter;type:bigint;not null;default:0" json:"qty_liter"`
	TotalLine int64 `gorm:"column:total_line;type:bigint;not null;default:0" json:"total_line"` // harga_bbm × qty_liter

	// Audit
	Created   time.Time `gorm:"column:created;autoCreateTime" json:"created"`
	CreatedBy *uint     `gorm:"column:created_by" json:"created_by"`
	Creator   *User     `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Updated   time.Time `gorm:"column:updated;autoUpdateTime" json:"updated"`
	UpdatedBy *uint     `gorm:"column:updated_by" json:"updated_by"`
	Updater   *User     `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
}

func (TrxPiutangDetail) TableName() string { return "trx_piutang_detail" }
