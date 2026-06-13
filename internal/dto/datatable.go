package dto

// DatatableRequest standardizes the parameters sent by jQuery DataTables
type DatatableRequest struct {
	Draw   int `form:"draw"`
	Start  int `form:"start"`
	Length int `form:"length"`
	Search struct {
		Value string `form:"value"`
		Regex string `form:"regex"`
	} `form:"search"`
	Order []struct {
		Column int    `form:"column"`
		Dir    string `form:"dir"`
	} `form:"order"`
	Columns []struct {
		Data       string `form:"data"`
		Name       string `form:"name"`
		Searchable string `form:"searchable"`
		Orderable  string `form:"orderable"`
		Search     struct {
			Value string `form:"value"`
			Regex string `form:"regex"`
		} `form:"search"`
	} `form:"columns"`

	// Custom filter fields for modules that extend the DataTables POST payload.
	FilterNoForm           string `form:"filter_no_form"`
	FilterTanggal          string `form:"filter_tanggal"` // YYYY-MM-DD
	FilterShiftID          uint   `form:"filter_shift_id"`
	FilterPartnerID        uint   `form:"filter_partner_id"`
	FilterWaktuMulai       string `form:"filter_waktu_mulai"` // YYYY-MM-DD
	FilterWaktuAkhir       string `form:"filter_waktu_akhir"` // YYYY-MM-DD
	FilterStatusPembayaran string `form:"filter_status_pembayaran"`
	FilterStatusKesesuaian string `form:"filter_status_kesesuaian"` // sesuai|tidak_sesuai
}

// DatatableResponse standardizes the JSON response expected by jQuery DataTables
type DatatableResponse struct {
	Draw            int         `json:"draw"`
	RecordsTotal    int64       `json:"recordsTotal"`
	RecordsFiltered int64       `json:"recordsFiltered"`
	Data            interface{} `json:"data"`
	Error           string      `json:"error,omitempty"`
}

// StokDODTRow — data row untuk server-side datatable Stok DO.
type StokDODTRow struct {
	DetailID     uint64 `gorm:"column:detail_id" json:"detail_id"`
	PenebusanID  uint64 `gorm:"column:penebusan_id" json:"penebusan_id"`
	NoPenebusan  string `gorm:"column:no_penebusan" json:"no_penebusan"`
	NoSO         string `gorm:"column:no_so" json:"no_so"`
	TglPenebusan string `gorm:"column:tgl_penebusan" json:"tgl_penebusan"`
	JenisBBM     string `gorm:"column:jenis_bbm" json:"jenis_bbm"`
	JmlLiter     int64  `gorm:"column:jml_liter" json:"jml_liter"`
	QtyTerkirim  int64  `gorm:"column:qty_terkirim" json:"qty_terkirim"`
	SisaLiter    int64  `gorm:"column:sisa_liter" json:"sisa_liter"`
	StatusKirim  string `gorm:"column:status_kirim" json:"status_kirim"`
}

// StokDOSummary — ringkasan statistik untuk halaman Stok DO.
type StokDOSummary struct {
	TotalSO    int64 `gorm:"column:total_so"`
	TotalItems int64 `gorm:"column:total_items"`
	Selesai    int64 `gorm:"column:selesai"`
	Belum      int64 `gorm:"column:belum"`
}

// PenjualanDTRow — data row untuk server-side datatable Penjualan BBM.
type PenjualanDTRow struct {
	ID                 uint64 `gorm:"column:id" json:"id"`
	NoPenjualan        string `gorm:"column:no_penjualan" json:"no_penjualan"`
	Tgl                string `gorm:"column:tgl" json:"tgl"`
	WaktuMulai         string `gorm:"column:waktu_mulai" json:"waktu_mulai"`
	WaktuAkhir         string `gorm:"column:waktu_akhir" json:"waktu_akhir"`
	ShiftName          string `gorm:"column:shift_name" json:"shift_name"`
	TotalRpTotalisator int64  `gorm:"column:total_rp_totalisator" json:"total_rp_totalisator"`
	TotalPenerimaan    int64  `gorm:"column:total_penerimaan" json:"total_penerimaan"`
	AktualUang         int64  `gorm:"column:aktual_uang" json:"aktual_uang"`
	Selisih            int64  `gorm:"column:selisih" json:"selisih"`
}
