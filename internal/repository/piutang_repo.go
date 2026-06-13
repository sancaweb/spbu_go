package repository

import (
	"fmt"
	"strings"
	"time"

	"spbu_go/internal/dto"
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

// ─── PiutangRepository ────────────────────────────────────────────────────────

type PiutangRepository interface {
	FindAll() ([]entity.TrxPiutang, error)
	FindByID(id uint64) (*entity.TrxPiutang, error)
	FindByPenjualanID(penjualanID uint64) ([]entity.TrxPiutang, error)
	Datatable(req dto.DatatableRequest) (int64, int64, []entity.TrxPiutang, error)
	DatatableRows(req dto.DatatableRequest) (int64, int64, []PiutangDTRow, error)
	DatatableDetailRows(req dto.DatatableRequest) (int64, int64, []PiutangDetailDTRow, error)
	DatatableRekapRows(req dto.DatatableRequest) (int64, int64, []PiutangRekapDTRow, error)
	Summary() (PiutangSummary, error)
	SummaryByMonth(month string) (PiutangSummary, error)
	GroupedRekapByMonth(month string) ([]PiutangGroupedDate, PiutangGroupedGrandTotal, error)
	Create(p *entity.TrxPiutang) error
	Update(p *entity.TrxPiutang) error
	MarkPaid(id uint64, updatedBy *uint) error
	Delete(id uint64) error
}

type piutangRepository struct {
	db *gorm.DB
}

func NewPiutangRepository(db *gorm.DB) PiutangRepository {
	return &piutangRepository{db: db}
}

func (r *piutangRepository) FindAll() ([]entity.TrxPiutang, error) {
	var list []entity.TrxPiutang
	err := r.db.
		Preload("Partner").
		Preload("Penjualan").
		Preload("Details").
		Preload("Details.BBM").
		Order("created DESC").
		Find(&list).Error
	return list, err
}

func (r *piutangRepository) FindByID(id uint64) (*entity.TrxPiutang, error) {
	var p entity.TrxPiutang
	err := r.db.
		Preload("Partner").
		Preload("Penjualan").
		Preload("Penjualan.Shift").
		Preload("Details").
		Preload("Details.BBM").
		Preload("Creator").
		Preload("Updater").
		First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *piutangRepository) FindByPenjualanID(penjualanID uint64) ([]entity.TrxPiutang, error) {
	var list []entity.TrxPiutang
	err := r.db.
		Where("penjualan_id = ?", penjualanID).
		Preload("Partner").
		Preload("Details").
		Preload("Details.BBM").
		Find(&list).Error
	return list, err
}

func (r *piutangRepository) Datatable(req dto.DatatableRequest) (int64, int64, []entity.TrxPiutang, error) {
	var list []entity.TrxPiutang
	var total, filtered int64

	base := r.db.Model(&entity.TrxPiutang{})
	base.Count(&total)

	if req.Search.Value != "" {
		s := "%" + req.Search.Value + "%"
		base = base.Joins("LEFT JOIN partners ON partners.id = trx_piutang.pelanggan_id").
			Where("partners.name ILIKE ? OR trx_piutang.status ILIKE ?", s, s)
	}
	base.Count(&filtered)

	// Ordering
	orderCol := "trx_piutang.created"
	orderDir := "DESC"
	if len(req.Order) > 0 {
		if req.Order[0].Dir == "asc" {
			orderDir = "ASC"
		}
		switch req.Order[0].Column {
		case 1:
			orderCol = "trx_piutang.created"
		case 2:
			orderCol = "partners.name"
		case 3:
			orderCol = "trx_piutang.total_tagihan"
		case 4:
			orderCol = "trx_piutang.status"
		}
	}

	limit := req.Length
	if limit <= 0 {
		limit = 10
	}

	err := base.
		Preload("Partner").
		Preload("Penjualan").
		Preload("Penjualan.Shift").
		Order(orderCol + " " + orderDir).
		Offset(req.Start).
		Limit(limit).
		Find(&list).Error

	return total, filtered, list, err
}

// DatatableRows — raw query for piutang datatable with joined partner/penjualan info.
func (r *piutangRepository) DatatableRows(req dto.DatatableRequest) (int64, int64, []PiutangDTRow, error) {
	baseSQL := `
		FROM trx_piutang pt
		JOIN partners pa ON pa.id = pt.pelanggan_id
		JOIN trx_penjualan pj ON pj.id_penjualan = pt.penjualan_id
		JOIN shifts sh ON sh.id = pj.shift_id
		LEFT JOIN (
			SELECT piutang_id, COALESCE(SUM(total_line), 0) AS total_detail
			FROM trx_piutang_detail
			GROUP BY piutang_id
		) dsum ON dsum.piutang_id = pt.id_piutang
	`
	var conditions []string
	var args []interface{}

	if req.Search.Value != "" {
		s := "%" + req.Search.Value + "%"
		conditions = append(conditions, "(pa.name ILIKE ? OR pt.status ILIKE ?)")
		args = append(args, s, s)
	}

	if req.FilterPartnerID > 0 {
		conditions = append(conditions, "pt.pelanggan_id = ?")
		args = append(args, req.FilterPartnerID)
	}
	if v := strings.TrimSpace(req.FilterWaktuMulai); v != "" {
		conditions = append(conditions, "DATE(pt.created) >= ?")
		args = append(args, v)
	}
	if v := strings.TrimSpace(req.FilterWaktuAkhir); v != "" {
		conditions = append(conditions, "DATE(pt.created) <= ?")
		args = append(args, v)
	}
	if v := strings.TrimSpace(req.FilterStatusPembayaran); v != "" {
		conditions = append(conditions, "pt.status = ?")
		args = append(args, v)
	}
	if v := strings.TrimSpace(req.FilterStatusKesesuaian); v != "" {
		switch v {
		case "sesuai":
			conditions = append(conditions, "COALESCE(dsum.total_detail, 0) = pt.total_tagihan")
		case "tidak_sesuai":
			conditions = append(conditions, "COALESCE(dsum.total_detail, 0) <> pt.total_tagihan")
		}
	}

	whereSQL := ""
	if len(conditions) > 0 {
		whereSQL = " WHERE " + strings.Join(conditions, " AND ")
	}

	var total, filtered int64
	r.db.Raw("SELECT COUNT(*) " + baseSQL).Scan(&total)
	if whereSQL != "" {
		r.db.Raw("SELECT COUNT(*) "+baseSQL+whereSQL, args...).Scan(&filtered)
	} else {
		filtered = total
	}

	orderCol := "pt.created"
	orderDir := "DESC"
	if len(req.Order) > 0 {
		if req.Order[0].Dir == "asc" {
			orderDir = "ASC"
		}
		switch req.Order[0].Column {
		case 1:
			orderCol = "pt.created"
		case 2:
			orderCol = "pa.name"
		case 3:
			orderCol = "pt.total_tagihan"
		case 4:
			orderCol = "pt.status"
		}
	}

	limit := req.Length
	if limit <= 0 {
		limit = 10
	}

	selectSQL := fmt.Sprintf(`
		SELECT
			pt.id_piutang,
			TO_CHAR(pt.created, 'YYYY-MM-DD')   AS tgl,
			pa.name                              AS partner_name,
			pj.no_penjualan,
			sh.shift_name,
			pt.total_tagihan,
			pt.status,
			pt.isinvoiced,
			(COALESCE(dsum.total_detail, 0) = pt.total_tagihan) AS is_matched,
			TO_CHAR(pt.updated, 'YYYY-MM-DD HH24:MI') AS updated_at
		%s%s
		ORDER BY %s %s
		LIMIT %d OFFSET %d
	`, baseSQL, whereSQL, orderCol, orderDir, limit, req.Start)

	var rows []PiutangDTRow
	err := r.db.Raw(selectSQL, args...).Scan(&rows).Error
	return total, filtered, rows, err
}

// DatatableDetailRows returns voucher-level piutang rows.
func (r *piutangRepository) DatatableDetailRows(req dto.DatatableRequest) (int64, int64, []PiutangDetailDTRow, error) {
	baseSQL := `
		FROM trx_piutang_detail d
		JOIN trx_piutang pt ON pt.id_piutang = d.piutang_id
		JOIN partners pa ON pa.id = pt.pelanggan_id
		JOIN trx_penjualan pj ON pj.id_penjualan = pt.penjualan_id
		JOIN bbm b ON b.id = d.bbm_id
	`
	var conditions []string
	var args []interface{}
	if req.Search.Value != "" {
		s := "%" + req.Search.Value + "%"
		conditions = append(conditions, "(pa.name ILIKE ? OR pj.no_penjualan ILIKE ? OR d.no_voucher ILIKE ? OR d.no_pol ILIKE ? OR d.driver_name ILIKE ? OR b.name ILIKE ?)")
		args = append(args, s, s, s, s, s, s)
	}
	if req.FilterPartnerID > 0 {
		conditions = append(conditions, "pt.pelanggan_id = ?")
		args = append(args, req.FilterPartnerID)
	}
	if v := strings.TrimSpace(req.FilterWaktuMulai); v != "" {
		conditions = append(conditions, "DATE(pt.created) >= ?")
		args = append(args, v)
	}
	if v := strings.TrimSpace(req.FilterWaktuAkhir); v != "" {
		conditions = append(conditions, "DATE(pt.created) <= ?")
		args = append(args, v)
	}
	if v := strings.TrimSpace(req.FilterStatusPembayaran); v != "" {
		conditions = append(conditions, "pt.status = ?")
		args = append(args, v)
	}
	whereSQL := ""
	if len(conditions) > 0 {
		whereSQL = " WHERE " + strings.Join(conditions, " AND ")
	}

	var total, filtered int64
	r.db.Raw("SELECT COUNT(*) " + baseSQL).Scan(&total)
	if whereSQL != "" {
		r.db.Raw("SELECT COUNT(*) "+baseSQL+whereSQL, args...).Scan(&filtered)
	} else {
		filtered = total
	}

	orderCol := "pt.created"
	orderDir := "DESC"
	if len(req.Order) > 0 {
		if req.Order[0].Dir == "asc" {
			orderDir = "ASC"
		}
		switch req.Order[0].Column {
		case 1:
			orderCol = "pt.created"
		case 2:
			orderCol = "pa.name"
		case 3:
			orderCol = "pj.no_penjualan"
		case 4:
			orderCol = "d.no_voucher"
		case 7:
			orderCol = "d.qty_liter"
		case 9:
			orderCol = "d.total_line"
		}
	}

	limit := req.Length
	if limit <= 0 {
		limit = 10
	}
	selectSQL := fmt.Sprintf(`
		SELECT
			d.id_piutang_detail,
			pt.id_piutang,
			TO_CHAR(pt.created, 'YYYY-MM-DD') AS tgl,
			pa.name AS partner_name,
			pj.no_penjualan,
			d.no_voucher,
			d.no_pol,
			d.driver_name,
			b.name AS bbm_name,
			d.harga_bbm,
			d.qty_liter,
			d.total_line,
			pt.status
		%s%s
		ORDER BY %s %s
		LIMIT %d OFFSET %d
	`, baseSQL, whereSQL, orderCol, orderDir, limit, req.Start)

	var rows []PiutangDetailDTRow
	err := r.db.Raw(selectSQL, args...).Scan(&rows).Error
	return total, filtered, rows, err
}

// DatatableRekapRows returns partner-level piutang summary rows.
func (r *piutangRepository) DatatableRekapRows(req dto.DatatableRequest) (int64, int64, []PiutangRekapDTRow, error) {
	baseSQL := `
		FROM trx_piutang pt
		JOIN partners pa ON pa.id = pt.pelanggan_id
	`
	var conditions []string
	var args []interface{}
	if req.Search.Value != "" {
		s := "%" + req.Search.Value + "%"
		conditions = append(conditions, "pa.name ILIKE ?")
		args = append(args, s)
	}

	month := strings.TrimSpace(req.FilterWaktuMulai)
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	conditions = append(conditions, "TO_CHAR(pt.created, 'YYYY-MM') = ?")
	args = append(args, month)
	whereSQL := ""
	if len(conditions) > 0 {
		whereSQL = " WHERE " + strings.Join(conditions, " AND ")
	}

	countSQL := "SELECT COUNT(*) FROM (SELECT pa.id " + baseSQL + whereSQL + " GROUP BY pa.id) x"
	var total, filtered int64
	r.db.Raw("SELECT COUNT(*) FROM (SELECT pa.id " + baseSQL + " GROUP BY pa.id) x").Scan(&total)
	if whereSQL != "" {
		r.db.Raw(countSQL, args...).Scan(&filtered)
	} else {
		filtered = total
	}

	orderCol := "pa.name"
	orderDir := "ASC"
	if len(req.Order) > 0 {
		if req.Order[0].Dir == "desc" {
			orderDir = "DESC"
		}
		switch req.Order[0].Column {
		case 1:
			orderCol = "pa.name"
		case 2:
			orderCol = "total_piutang"
		case 3:
			orderCol = "total_unpaid"
		case 4:
			orderCol = "total_paid"
		case 5:
			orderCol = "nilai_unpaid"
		case 6:
			orderCol = "nilai_paid"
		case 7:
			orderCol = "nilai_total"
		}
	}

	limit := req.Length
	if limit <= 0 {
		limit = 10
	}
	selectSQL := fmt.Sprintf(`
		SELECT
			pa.id AS partner_id,
			pa.name AS partner_name,
			COUNT(pt.id_piutang) AS total_piutang,
			SUM(CASE WHEN pt.status = 'unpaid' THEN 1 ELSE 0 END) AS total_unpaid,
			SUM(CASE WHEN pt.status = 'paid' THEN 1 ELSE 0 END) AS total_paid,
			COALESCE(SUM(CASE WHEN pt.status = 'unpaid' THEN pt.total_tagihan ELSE 0 END), 0) AS nilai_unpaid,
			COALESCE(SUM(CASE WHEN pt.status = 'paid' THEN pt.total_tagihan ELSE 0 END), 0) AS nilai_paid,
			COALESCE(SUM(pt.total_tagihan), 0) AS nilai_total
		%s%s
		GROUP BY pa.id, pa.name
		ORDER BY %s %s
		LIMIT %d OFFSET %d
	`, baseSQL, whereSQL, orderCol, orderDir, limit, req.Start)

	var rows []PiutangRekapDTRow
	err := r.db.Raw(selectSQL, args...).Scan(&rows).Error
	return total, filtered, rows, err
}

func (r *piutangRepository) Summary() (PiutangSummary, error) {
	return r.SummaryByMonth(time.Now().Format("2006-01"))
}

func (r *piutangRepository) SummaryByMonth(month string) (PiutangSummary, error) {
	if strings.TrimSpace(month) == "" {
		month = time.Now().Format("2006-01")
	}
	var s PiutangSummary
	err := r.db.Raw(`
		SELECT
			COUNT(*) AS total_piutang,
			COALESCE(SUM(CASE WHEN status = 'unpaid' THEN 1 ELSE 0 END), 0) AS total_unpaid,
			COALESCE(SUM(CASE WHEN status = 'paid' THEN 1 ELSE 0 END), 0) AS total_paid,
			COALESCE(SUM(CASE WHEN status = 'unpaid' THEN total_tagihan ELSE 0 END), 0) AS nilai_unpaid,
			COALESCE(SUM(CASE WHEN status = 'paid' THEN total_tagihan ELSE 0 END), 0) AS nilai_paid,
			COALESCE(SUM(total_tagihan), 0) AS nilai_total
		FROM trx_piutang
		WHERE TO_CHAR(created, 'YYYY-MM') = ?
	`, month).Scan(&s).Error
	return s, err
}

func (r *piutangRepository) GroupedRekapByMonth(month string) ([]PiutangGroupedDate, PiutangGroupedGrandTotal, error) {
	if strings.TrimSpace(month) == "" {
		month = time.Now().Format("2006-01")
	}

	type row struct {
		Tanggal             string `gorm:"column:tanggal"`
		TglSort             string `gorm:"column:tgl_sort"`
		NoUrut              int64  `gorm:"column:no_urut"`
		PiutangPelanggan    string `gorm:"column:piutang_pelanggan"`
		PiutangTagihan      int64  `gorm:"column:piutang_tagihan"`
		BayarPelanggan      string `gorm:"column:bayar_pelanggan"`
		BayarNominal        int64  `gorm:"column:bayar_nominal"`
		BayarPeriodePiutang string `gorm:"column:bayar_periode_piutang"`
	}

	var rows []row
	err := r.db.Raw(`
		SELECT
			TO_CHAR(pt.created, 'DD-Mon-YYYY') AS tanggal,
			TO_CHAR(pt.created, 'YYYY-MM-DD') AS tgl_sort,
			ROW_NUMBER() OVER (PARTITION BY DATE(pt.created) ORDER BY pt.id_piutang) AS no_urut,
			pa.name AS piutang_pelanggan,
			pt.total_tagihan AS piutang_tagihan,
			CASE WHEN pt.status = 'paid' THEN pa.name ELSE '-' END AS bayar_pelanggan,
			CASE WHEN pt.status = 'paid' THEN pt.total_tagihan ELSE 0 END AS bayar_nominal,
			CASE WHEN pt.status = 'paid' THEN TO_CHAR(pt.created, 'Mon YYYY') ELSE '-' END AS bayar_periode_piutang
		FROM trx_piutang pt
		JOIN partners pa ON pa.id = pt.pelanggan_id
		WHERE TO_CHAR(pt.created, 'YYYY-MM') = ?
		ORDER BY DATE(pt.created) ASC, pt.id_piutang ASC
	`, month).Scan(&rows).Error
	if err != nil {
		return nil, PiutangGroupedGrandTotal{}, err
	}

	groupedMap := map[string]*PiutangGroupedDate{}
	var order []string
	grand := PiutangGroupedGrandTotal{}

	for _, r := range rows {
		if _, ok := groupedMap[r.TglSort]; !ok {
			groupedMap[r.TglSort] = &PiutangGroupedDate{
				TanggalLabel: r.Tanggal,
				Rows:         []PiutangGroupedRow{},
			}
			order = append(order, r.TglSort)
		}
		g := groupedMap[r.TglSort]
		g.Rows = append(g.Rows, PiutangGroupedRow{
			No:                  r.NoUrut,
			PiutangPelanggan:    r.PiutangPelanggan,
			PiutangTagihan:      r.PiutangTagihan,
			BayarPelanggan:      r.BayarPelanggan,
			BayarNominal:        r.BayarNominal,
			BayarPeriodePiutang: r.BayarPeriodePiutang,
		})
		g.TotalPiutang += r.PiutangTagihan
		g.TotalPembayaran += r.BayarNominal
		grand.GrandTotalPiutang += r.PiutangTagihan
		grand.GrandTotalPembayaran += r.BayarNominal
	}

	var grouped []PiutangGroupedDate
	for _, k := range order {
		grouped = append(grouped, *groupedMap[k])
	}
	return grouped, grand, nil
}

func (r *piutangRepository) Create(p *entity.TrxPiutang) error {
	return r.db.Omit("Partner", "Penjualan", "Creator", "Updater", "Details.BBM", "Details.Penjualan", "Details.Creator", "Details.Updater").
		Create(p).Error
}

func (r *piutangRepository) Update(p *entity.TrxPiutang) error {
	return r.db.Omit("Partner", "Penjualan", "Creator", "Updater").Save(p).Error
}

func (r *piutangRepository) MarkPaid(id uint64, updatedBy *uint) error {
	return r.db.Model(&entity.TrxPiutang{}).
		Where("id_piutang = ?", id).
		Updates(map[string]interface{}{
			"status":     entity.PiutangPaid,
			"updated_by": updatedBy,
		}).Error
}

func (r *piutangRepository) Delete(id uint64) error {
	// Hard delete (piutang tidak menggunakan soft delete)
	return r.db.Where("id_piutang = ?", id).Delete(&entity.TrxPiutang{}).Error
}

// ─── PiutangDTRow — raw row for datatable ─────────────────────────────────────

type PiutangDTRow struct {
	IDPiutang    uint64 `gorm:"column:id_piutang" json:"id_piutang"`
	Tgl          string `gorm:"column:tgl" json:"tgl"`
	PartnerName  string `gorm:"column:partner_name" json:"partner_name"`
	NoPenjualan  string `gorm:"column:no_penjualan" json:"no_penjualan"`
	ShiftName    string `gorm:"column:shift_name" json:"shift_name"`
	TotalTagihan int64  `gorm:"column:total_tagihan" json:"total_tagihan"`
	Status       string `gorm:"column:status" json:"status"`
	IsInvoiced   bool   `gorm:"column:isinvoiced" json:"isinvoiced"`
	IsMatched    bool   `gorm:"column:is_matched" json:"is_matched"`
	UpdatedAt    string `gorm:"column:updated_at" json:"updated_at"`
}

type PiutangDetailDTRow struct {
	IDPiutangDetail uint64 `gorm:"column:id_piutang_detail" json:"id_piutang_detail"`
	IDPiutang       uint64 `gorm:"column:id_piutang" json:"id_piutang"`
	Tgl             string `gorm:"column:tgl" json:"tgl"`
	PartnerName     string `gorm:"column:partner_name" json:"partner_name"`
	NoPenjualan     string `gorm:"column:no_penjualan" json:"no_penjualan"`
	NoVoucher       string `gorm:"column:no_voucher" json:"no_voucher"`
	NoPol           string `gorm:"column:no_pol" json:"no_pol"`
	DriverName      string `gorm:"column:driver_name" json:"driver_name"`
	BBMName         string `gorm:"column:bbm_name" json:"bbm_name"`
	HargaBBM        int64  `gorm:"column:harga_bbm" json:"harga_bbm"`
	QtyLiter        int64  `gorm:"column:qty_liter" json:"qty_liter"`
	TotalLine       int64  `gorm:"column:total_line" json:"total_line"`
	Status          string `gorm:"column:status" json:"status"`
}

type PiutangRekapDTRow struct {
	PartnerID    uint   `gorm:"column:partner_id" json:"partner_id"`
	PartnerName  string `gorm:"column:partner_name" json:"partner_name"`
	TotalPiutang int64  `gorm:"column:total_piutang" json:"total_piutang"`
	TotalUnpaid  int64  `gorm:"column:total_unpaid" json:"total_unpaid"`
	TotalPaid    int64  `gorm:"column:total_paid" json:"total_paid"`
	NilaiUnpaid  int64  `gorm:"column:nilai_unpaid" json:"nilai_unpaid"`
	NilaiPaid    int64  `gorm:"column:nilai_paid" json:"nilai_paid"`
	NilaiTotal   int64  `gorm:"column:nilai_total" json:"nilai_total"`
}

type PiutangSummary struct {
	TotalPiutang int64 `gorm:"column:total_piutang" json:"total_piutang"`
	TotalUnpaid  int64 `gorm:"column:total_unpaid" json:"total_unpaid"`
	TotalPaid    int64 `gorm:"column:total_paid" json:"total_paid"`
	NilaiUnpaid  int64 `gorm:"column:nilai_unpaid" json:"nilai_unpaid"`
	NilaiPaid    int64 `gorm:"column:nilai_paid" json:"nilai_paid"`
	NilaiTotal   int64 `gorm:"column:nilai_total" json:"nilai_total"`
}

type PiutangGroupedRow struct {
	No                  int64  `json:"no"`
	PiutangPelanggan    string `json:"piutang_pelanggan"`
	PiutangTagihan      int64  `json:"piutang_tagihan"`
	BayarPelanggan      string `json:"bayar_pelanggan"`
	BayarNominal        int64  `json:"bayar_nominal"`
	BayarPeriodePiutang string `json:"bayar_periode_piutang"`
}

type PiutangGroupedDate struct {
	TanggalLabel    string              `json:"tanggal_label"`
	Rows            []PiutangGroupedRow `json:"rows"`
	TotalPiutang    int64               `json:"total_piutang"`
	TotalPembayaran int64               `json:"total_pembayaran"`
}

type PiutangGroupedGrandTotal struct {
	GrandTotalPiutang    int64 `json:"grand_total_piutang"`
	GrandTotalPembayaran int64 `json:"grand_total_pembayaran"`
}
