package entity

import (
	"time"

	"gorm.io/gorm"
)

// COAType — Kelompok akun (1=Aset, 2=Kewajiban, 3=Modal, 4=Pendapatan, 5=HPP, 6=Beban)
type COAType struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Code          string         `gorm:"type:varchar(10);uniqueIndex;not null" json:"code"`
	Name          string         `gorm:"type:varchar(100);not null" json:"name"`
	NormalBalance string         `gorm:"type:varchar(6);not null;default:'debit'" json:"normal_balance"` // debit | credit
	Description   string         `gorm:"type:text" json:"description"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	COAs          []COA          `gorm:"foreignKey:COATypeID" json:"coas,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (COAType) TableName() string { return "coa_types" }

// COA — Chart of Account (akun buku besar)
type COA struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	COATypeID   uint           `gorm:"not null;index" json:"coa_type_id"`
	COAType     *COAType       `gorm:"foreignKey:COATypeID" json:"coa_type,omitempty"`
	Code        string         `gorm:"type:varchar(10);uniqueIndex;not null" json:"code"`
	Name        string         `gorm:"type:varchar(200);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	IsHeader    bool           `gorm:"default:false" json:"is_header"` // header = pengelompokan saja, tidak bisa dipilih di transaksi
	IsSystem    bool           `gorm:"default:false" json:"is_system"` // system = tidak bisa dihapus user
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	UpdatedBy   *uint          `json:"updated_by"`
	Updater     *User          `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (COA) TableName() string { return "coas" }

// JournalEntry — Baris jurnal double-entry. Setiap transaksi menghasilkan ≥2 baris.
// Prinsip: total Debit semua baris = total Kredit semua baris per RefType+RefID.
type JournalEntry struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	COAID       uint      `gorm:"not null;index" json:"coa_id"`
	COA         *COA      `gorm:"foreignKey:COAID" json:"coa,omitempty"`
	WalletID    *uint     `gorm:"index" json:"wallet_id"`
	Wallet      *Wallet   `gorm:"foreignKey:WalletID" json:"wallet,omitempty"`
	Debit       int64     `gorm:"default:0" json:"debit"`
	Credit      int64     `gorm:"default:0" json:"credit"`
	Description string    `gorm:"type:varchar(500)" json:"description"`
	// RefType: penjualan | penebusan | kedatangan | kasbon | cash_in | cash_out | payroll | manual
	RefType   string    `gorm:"type:varchar(50)" json:"ref_type"`
	RefID     *uint     `json:"ref_id"` // ID sumber transaksi
	TransDate time.Time `gorm:"not null" json:"trans_date"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy *uint     `json:"created_by"`
	Creator   *User     `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (JournalEntry) TableName() string { return "journal_entries" }
