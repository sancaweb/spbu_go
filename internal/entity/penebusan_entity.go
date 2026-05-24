package entity

import (
	"time"

	"gorm.io/gorm"
)

// TrxPenebusanStatus mendefinisikan status dokumen transaksi.
//
//	DR → Draft (belum berdampak ke flow bisnis lanjutan)
//	CO → Complete (siap diproses ke flow bisnis berikutnya)
type TrxPenebusanStatus = string

const (
	PenebusanDraft    TrxPenebusanStatus = "DR"
	PenebusanComplete TrxPenebusanStatus = "CO"
)

// TrxPenebusan — Header penebusan BBM ke Pertamina.
//
// Alur jurnal:
//
//	Saat status → paid:
//	  Dr  Uang Muka Pertamina (1131)  = subtotal + total_ppn
//	  Cr  Bank/Wallet (wallet_id)     = subtotal + total_ppn
//	  Dr  Biaya Admin Bank (5110)     = adm_bank   [jika adm_bank > 0]
//	  Cr  Bank/Wallet (wallet_id)     = adm_bank   [jika adm_bank > 0]
//
//	Saat kedatangan BBM (modul terpisah):
//	  Dr  Persediaan BBM 112X         = nilai per BBM
//	  Cr  Uang Muka Pertamina (1131)  = nilai yang diterima
type TrxPenebusan struct {
	ID uint64 `gorm:"primaryKey" json:"id"`

	// Nomor dokumen (kolom DB: no_penebusan)
	NoPenebusan string  `gorm:"column:no_penebusan;type:varchar(25);uniqueIndex:uni_trx_penebusan_no_penebusan;not null" json:"no_penebusan"` // PNB/2026/04/0001
	NoSO        *string `gorm:"type:varchar(25)" json:"no_so"`                                                                                // dari Pertamina, nullable

	// Tanggal
	TglPenebusan time.Time  `gorm:"type:date;not null" json:"tgl_penebusan"`
	TglBayar     *time.Time `gorm:"type:date" json:"tgl_bayar"` // diisi saat status → paid

	// Pembayaran
	WalletID *uint   `gorm:"index" json:"wallet_id"`
	Wallet   *Wallet `gorm:"foreignKey:WalletID" json:"wallet,omitempty"`
	AdmBank  int64   `gorm:"type:bigint;not null;default:0" json:"adm_bank"` // biaya admin bank (Rp)

	// Status dokumen: DR | CO
	Status string `gorm:"type:varchar(2);not null;default:'DR';index" json:"status"`

	// Catatan
	Catatan string `gorm:"type:text" json:"catatan"`

	// Kalkulasi total (integer Rp)
	Subtotal   int64 `gorm:"type:bigint;not null;default:0" json:"subtotal"`    // Σ (harga_dasar × liter) tanpa PPN
	TotalPPN   int64 `gorm:"type:bigint;not null;default:0" json:"total_ppn"`   // Σ PPN semua item
	TotalBayar int64 `gorm:"type:bigint;not null;default:0" json:"total_bayar"` // subtotal + total_ppn + adm_bank

	// Relasi detail
	Details []TrxPenebusanDetail `gorm:"foreignKey:PenebusanID;constraint:OnDelete:CASCADE" json:"details,omitempty"`

	// Audit
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	UpdatedBy *uint          `json:"updated_by"`
	Updater   *User          `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (TrxPenebusan) TableName() string { return "trx_penebusan" }

// GetStatusLabel mengembalikan label status dokumen untuk kebutuhan display UI.
// DR -> Draft, CO -> Complete.
func (t TrxPenebusan) GetStatusLabel() string {
	if t.Status == string(PenebusanComplete) {
		return "Complete"
	}
	return "Draft"
}

// TrxPenebusanDetail — Item penebusan per jenis BBM.
//
// Semua harga adalah snapshot saat transaksi (tidak berubah meski harga master diupdate).
// Volume (JmlLiter) disimpan sebagai integer × 10^stock_decimal_places,
// konsisten dengan kolom stock di tabel bbm.
type TrxPenebusanDetail struct {
	ID          uint64 `gorm:"primaryKey" json:"id"`
	PenebusanID uint64 `gorm:"not null;index" json:"penebusan_id"`
	BBMID       uint   `gorm:"not null;index" json:"bbm_id"`
	BBM         *BBM   `gorm:"foreignKey:BBMID" json:"bbm,omitempty"`

	// Volume — disimpan sebagai integer × 10^stock_decimal_places
	JmlLiter    int64 `gorm:"type:bigint;not null" json:"jml_liter"`
	QtyTerkirim int64 `gorm:"type:bigint;not null;default:0" json:"qty_terkirim"` // total liter sudah terkirim dari Pertamina

	// Harga snapshot (Rp per liter, integer)
	HargaDasar int64 `gorm:"type:bigint;not null" json:"harga_dasar"`      // harga beli dari Pertamina
	HargaJual  int64 `gorm:"type:bigint;not null" json:"harga_jual"`       // harga jual SPBU saat ini
	Margin     int64 `gorm:"type:bigint;not null;default:0" json:"margin"` // harga_jual - harga_dasar

	// PPN
	PPNPersen float64 `gorm:"type:decimal(5,2);not null;default:0" json:"ppn_persen"` // misal: 11.00
	PPNRp     int64   `gorm:"type:bigint;not null;default:0" json:"ppn_rp"`           // subtotal × ppn_persen / 100

	// Kalkulasi baris (Rp)
	Subtotal int64 `gorm:"type:bigint;not null;default:0" json:"subtotal"` // harga_dasar × jml_liter
	Total    int64 `gorm:"type:bigint;not null;default:0" json:"total"`    // subtotal + ppn_rp

	// Audit (tidak ada soft delete — ikut CASCADE dari header)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (TrxPenebusanDetail) TableName() string { return "trx_penebusan_detail" }
