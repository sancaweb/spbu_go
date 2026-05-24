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
