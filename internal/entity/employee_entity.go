package entity

import (
	"time"

	"gorm.io/gorm"
)

type Jabatan struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	KodeJabatan  string         `gorm:"type:varchar(10);not null" json:"kode_jabatan" form:"kode_jabatan"`
	NamaJabatan  string         `gorm:"type:varchar(50);not null" json:"nama_jabatan" form:"nama_jabatan"`
	RewardPersen float64        `gorm:"type:decimal(5,2);default:0" json:"reward_persen" form:"reward_persen"`
	IsActive     bool           `gorm:"default:true" json:"is_active" form:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Jabatan) TableName() string {
	return "jabatan"
}

type Pendapatan struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	NamaPendapatan string        `gorm:"type:varchar(100);not null" json:"nama_pendapatan" form:"nama_pendapatan"`
	Tipe          string         `gorm:"type:varchar(10);not null;default:'nominal'" json:"tipe" form:"tipe"` // nominal | persen
	Nilai         int64          `gorm:"type:bigint;default:0" json:"nilai" form:"nilai"`
	Deskripsi     string         `gorm:"type:text" json:"deskripsi" form:"deskripsi"`
	IsActive      bool           `gorm:"default:true" json:"is_active" form:"is_active"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	UpdatedBy     *uint          `json:"updated_by"`
	Updater       *User          `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Pendapatan) TableName() string {
	return "pendapatan"
}

type Potongan struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	KodePotongan string         `gorm:"type:varchar(10);not null;uniqueIndex:uni_potongan_kode_potongan" json:"kode_potongan" form:"kode_potongan"`
	NamaPotongan string         `gorm:"type:varchar(100);not null" json:"nama_potongan" form:"nama_potongan"`
	Tipe         string         `gorm:"type:varchar(10);not null;default:'nominal'" json:"tipe" form:"tipe"` // nominal | persen
	Nilai        int64          `gorm:"type:bigint;default:0" json:"nilai" form:"nilai"`
	Deskripsi    string         `gorm:"type:text" json:"deskripsi" form:"deskripsi"`
	IsActive     bool           `gorm:"default:true" json:"is_active" form:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	UpdatedBy    *uint          `json:"updated_by"`
	Updater      *User          `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Potongan) TableName() string {
	return "potongan"
}

type Karyawan struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	NIK             string         `gorm:"type:varchar(20);uniqueIndex;not null" json:"nik" form:"nik"`
	NamaLengkap     string         `gorm:"type:varchar(100);not null" json:"nama_lengkap" form:"nama_lengkap"`
	GajiPokok       int64          `gorm:"type:bigint;default:0" json:"gaji_pokok" form:"gaji_pokok"`
	Alamat          string         `gorm:"type:text" json:"alamat" form:"alamat"`
	TempatLahir     string         `gorm:"type:varchar(50)" json:"tempat_lahir" form:"tempat_lahir"`
	TanggalLahir    time.Time      `gorm:"type:date" json:"tanggal_lahir" form:"tanggal_lahir"`
	StatusNikah     string         `gorm:"type:varchar(15)" json:"status_nikah" form:"status_nikah"` // lajang, menikah, cerai
	JumlahAnak      int            `gorm:"default:0" json:"jumlah_anak" form:"jumlah_anak"`
	JabatanID       *uint          `gorm:"index" json:"jabatan_id" form:"jabatan_id"`
	Jabatan         *Jabatan       `gorm:"foreignKey:JabatanID" json:"jabatan,omitempty"`
	JenisKelamin    string         `gorm:"type:varchar(1)" json:"jenis_kelamin" form:"jenis_kelamin"` // L / P
	Agama           string         `gorm:"type:varchar(20)" json:"agama" form:"agama"`
	NoHP            string         `gorm:"type:varchar(20)" json:"no_hp" form:"no_hp"`
	Pendidikan      string         `gorm:"type:varchar(20)" json:"pendidikan" form:"pendidikan"`
	TglPengangkatan time.Time      `gorm:"type:date" json:"tgl_pengangkatan" form:"tgl_pengangkatan"`
	TglKeluar       *time.Time     `gorm:"type:date" json:"tgl_keluar" form:"tgl_keluar"`
	Foto            string         `gorm:"type:varchar(255)" json:"foto" form:"foto"`
	IsActive        bool           `gorm:"default:true" json:"is_active" form:"is_active"`
	Pendapatans     []Pendapatan   `gorm:"many2many:karyawan_pendapatan;" json:"pendapatans,omitempty"`
	Potongans       []Potongan     `gorm:"many2many:karyawan_potongan;" json:"potongans,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	UpdatedBy       *uint          `json:"updated_by"`
	Updater         *User          `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Karyawan) TableName() string {
	return "karyawan"
}

// KaryawanPendapatan — junction table for Karyawan <-> Pendapatan (many2many)
type KaryawanPendapatan struct {
	KaryawanID   uint `gorm:"primaryKey"`
	PendapatanID uint `gorm:"primaryKey"`
}

func (KaryawanPendapatan) TableName() string {
	return "karyawan_pendapatan"
}

// KaryawanPotongan — junction table for Karyawan <-> Potongan (many2many)
type KaryawanPotongan struct {
	KaryawanID uint `gorm:"primaryKey"`
	PotonganID uint `gorm:"primaryKey"`
}

func (KaryawanPotongan) TableName() string {
	return "karyawan_potongan"
}
