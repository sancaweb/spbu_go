package repository

import (
	"fmt"
	"strings"
	"time"

	"spbu_go/internal/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ─── Filter ──────────────────────────────────────────────────────────────────

type PenyusutanReportFilter struct {
	Mode  string
	Date  string // YYYY-MM-DD
	Month string // YYYY-MM
}

// ─── Summary (Daily & Monthly) ────────────────────────────────────────────────

type PenyusutanReportRow struct {
	Waktu            string  `json:"waktu"`
	JenisBBM         string  `json:"jenis_bbm"`
	StockAwal        float64 `json:"stock_awal"`
	TotalPenerimaan  float64 `json:"total_penerimaan"`
	Disp             string  `json:"disp"`
	TotalisatorAwal  float64 `json:"totalisator_awal"`
	TotalisatorAkhir float64 `json:"totalisator_akhir"`
	TotalPenjualan   float64 `json:"total_penjualan"` // SUM(jml_liter) semua shift untuk nozzle ini
}

type PenyusutanReportGroup struct {
	BBMID              uint                  `json:"bbm_id"`
	JenisBBM           string                `json:"jenis_bbm"`
	Rows               []PenyusutanReportRow `json:"rows"`
	FirstStock         float64               `json:"first_stock"`        // stok awal (dari record pertama hari/bulan itu)
	TotalPenerimaan    float64               `json:"total_penerimaan"`
	TotalPenjualan     float64               `json:"total_penjualan"`
	DensityTera        float64               `json:"density_tera"`
	PenjualanMinTera   float64               `json:"penjualan_min_tera"`
	StockAkhirAktual   float64               `json:"stock_akhir_aktual"`
	StockAkhirCatatan  float64               `json:"stock_akhir_catatan"` // dihitung on-the-fly: firstStock + penerimaan - penjualanMinTera
	Penyusutan         float64               `json:"penyusutan"`          // aktual - catatan (seperti SPBU lama)
	Persentasi         float64               `json:"persentasi"`
	LatestPenyusutanID uint64                `json:"latest_penyusutan_id"`
}

// ─── Per-Shift (Daily only) ──────────────────────────────────────────────────

// PenyusutanShiftNozzle adalah data satu baris nozzle dalam satu shift.
type PenyusutanShiftNozzle struct {
	Disp             string  `json:"disp"`
	TotalisatorAwal  float64 `json:"totalisator_awal"`
	TotalisatorAkhir float64 `json:"totalisator_akhir"`
	JmlLiter         float64 `json:"jml_liter"`
}

// PenyusutanShiftBBM adalah data satu BBM dalam satu shift, lengkap dengan kalkulasi penyusutan.
type PenyusutanShiftBBM struct {
	BBMID            uint                    `json:"bbm_id"`
	JenisBBM         string                  `json:"jenis_bbm"`
	PenyusutanID     uint64                  `json:"penyusutan_id"` // untuk tombol Edit Stok Aktual
	StokAwal         float64                 `json:"stok_awal"`
	TotalPenerimaan  float64                 `json:"total_penerimaan"`
	Nozzles          []PenyusutanShiftNozzle `json:"nozzles"`
	TotalPenjualan   float64                 `json:"total_penjualan"`
	DensityTera      float64                 `json:"density_tera"`
	PenjualanMinTera float64                 `json:"penjualan_min_tera"`
	StokAkhirAktual  float64                 `json:"stok_akhir_aktual"`
	StokAkhirCatatan float64                 `json:"stok_akhir_catatan"`
	Penyusutan       float64                 `json:"penyusutan"`
	Persentasi       float64                 `json:"persentasi"`
}

// PenyusutanShiftGroup adalah satu shift / satu penjualan.
type PenyusutanShiftGroup struct {
	PenjualanID uint64               `json:"penjualan_id"`
	NoForm      string               `json:"no_form"`
	ShiftName   string               `json:"shift_name"`
	WaktuMulai  string               `json:"waktu_mulai"`
	WaktuAkhir  string               `json:"waktu_akhir"`
	BBMGroups   []PenyusutanShiftBBM `json:"bbm_groups"`
}

// ─── Interface ───────────────────────────────────────────────────────────────

type PenyusutanRepository interface {
	UpsertFromPenjualan(p *entity.TrxPenjualan, actorID *uint) error
	GetReport(filter PenyusutanReportFilter) ([]PenyusutanReportGroup, error)
	GetShiftReport(date string) ([]PenyusutanShiftGroup, error)
	UpdateEndstockActual(id uint64, actual float64, updatedBy *uint) error
}

type penyusutanRepository struct {
	db *gorm.DB
}

func NewPenyusutanRepository(db *gorm.DB) PenyusutanRepository {
	return &penyusutanRepository{db: db}
}

// ─── UpsertFromPenjualan ─────────────────────────────────────────────────────
// Dipanggil saat laporan penjualan dibuat/diupdate.
// PENTING: endstock_actual TIDAK ikut diupdate pada conflict, agar nilai yang
// sudah diinput user secara manual tidak tertimpa.
func (r *penyusutanRepository) UpsertFromPenjualan(p *entity.TrxPenjualan, actorID *uint) error {
	if p == nil || p.ID == 0 {
		return fmt.Errorf("data penjualan tidak valid")
	}

	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	bbmAgg := map[uint]float64{}
	for _, d := range p.Details {
		bbmAgg[d.BBMID] += float64(d.JmlLiter)
	}

	if len(bbmAgg) == 0 {
		if err := tx.Where("penjualan_id = ?", p.ID).Delete(&entity.TrxPenyusutan{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		return tx.Commit().Error
	}

	bbmIDs := make([]uint, 0, len(bbmAgg))
	for bbmID := range bbmAgg {
		bbmIDs = append(bbmIDs, bbmID)
	}

	for bbmID, totalPenjualan := range bbmAgg {
		var totalPenerimaan float64
		tx.Raw(`
			SELECT COALESCE(SUM(k.jml_liter), 0)
			FROM trx_kedatangan_bbm k
			WHERE k.bbm_id = ? AND k.shift_id = ? AND k.tgl_kedatangan BETWEEN ? AND ?
		`, bbmID, p.ShiftID, p.WaktuMulai, p.WaktuAkhir).Scan(&totalPenerimaan)

		var densityTera float64
		tx.Raw(`
			SELECT COALESCE(SUM(pt.qty_liter), 0)
			FROM trx_penjualan_pengeluaran_test pt
			WHERE pt.penjualan_id = ? AND pt.bbm_id = ?
		`, p.ID, bbmID).Scan(&densityTera)

		// firstStock = endstock_actual dari record penyusutan sebelumnya untuk BBM ini
		var firstStock float64
		tx.Raw(`
			SELECT endstock_actual
			FROM trx_penyusutan
			WHERE bbm_id = ?
			  AND (waktu < ? OR (waktu = ? AND penjualan_id <> ?))
			ORDER BY waktu DESC, id_penyusutan DESC
			LIMIT 1
		`, bbmID, p.WaktuMulai, p.WaktuMulai, p.ID).Scan(&firstStock)

		if firstStock == 0 {
			tx.Raw(`SELECT COALESCE(stock, 0) FROM bbm WHERE id = ? LIMIT 1`, bbmID).Scan(&firstStock)
		}

		netSales := totalPenjualan - densityTera
		endBooked := firstStock + totalPenerimaan - netSales
		// endActual = endBooked hanya untuk INSERT baru.
		// Pada UPDATE (conflict), endstock_actual TIDAK ditimpa (lihat DoUpdates).
		endActual := endBooked

		row := entity.TrxPenyusutan{
			PenjualanID:    p.ID,
			NoForm:         p.NoPenjualan,
			ShiftID:        p.ShiftID,
			Waktu:          p.WaktuMulai,
			BBMID:          bbmID,
			FirstStock:     firstStock,
			EndstockActual: endActual,
			EndstockBooked: endBooked,
			CreatedBy:      actorID,
			UpdatedBy:      actorID,
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "penjualan_id"}, {Name: "bbm_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				// endstock_actual SENGAJA tidak ada di sini agar tidak overwrite input user
				"no_form":         row.NoForm,
				"shift_id":        row.ShiftID,
				"waktu":           row.Waktu,
				"first_stock":     row.FirstStock,
				"endstock_booked": row.EndstockBooked,
				"updated_by":      row.UpdatedBy,
				"updated":         time.Now(),
			}),
		}).Create(&row).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Where("penjualan_id = ? AND bbm_id NOT IN ?", p.ID, bbmIDs).Delete(&entity.TrxPenyusutan{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// ─── GetReport (Summary: Daily & Monthly) ────────────────────────────────────
func (r *penyusutanRepository) GetReport(filter PenyusutanReportFilter) ([]PenyusutanReportGroup, error) {
	mode := strings.ToLower(strings.TrimSpace(filter.Mode))
	if mode != "daily" && mode != "monthly" {
		return []PenyusutanReportGroup{}, nil
	}

	switch mode {
	case "daily":
		return r.getDailyReport(filter.Date)
	default:
		return r.getMonthlyReport(filter.Month)
	}
}

// ─── getDailyReport ───────────────────────────────────────────────────────────
// Logic daily:
// - Data dikelompokkan per BBM, lalu per nozzle di dalamnya
// - Totalisator Awal nozzle = totalisator_awal dari shift PERTAMA nozzle tsb di hari itu
// - Totalisator Akhir nozzle = totalisator_akhir dari shift TERAKHIR nozzle tsb di hari itu
// - Total Penjualan nozzle = Totalisator Akhir - Totalisator Awal (seharian penuh)
func (r *penyusutanRepository) getDailyReport(date string) ([]PenyusutanReportGroup, error) {
	if date == "" {
		return []PenyusutanReportGroup{}, nil
	}

	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, err
	}
	waktuMulaiStr := date + " 06:00:00"
	waktuSampaiStr := t.AddDate(0, 0, 1).Format("2006-01-02") + " 05:59:59"

	// Format tanggal untuk tampilan di kolom Waktu (DD Mon YYYY)
	waktuLabel := date
	if parsed, err := time.Parse("2006-01-02", date); err == nil {
		waktuLabel = parsed.Format("02 Jan 2006")
	}


	type nozzleRow struct {
		BBMID            uint    `gorm:"column:bbm_id"`
		BBMName          string  `gorm:"column:bbm_name"`
		NozzleID         uint    `gorm:"column:nozzle_id"`
		Disp             string  `gorm:"column:disp"`
		TotalisatorAwal  int64   `gorm:"column:totalisator_awal"`  // dari shift PERTAMA (untuk tampilan)
		TotalisatorAkhir int64   `gorm:"column:totalisator_akhir"` // dari shift TERAKHIR (untuk tampilan)
		JmlLiter         int64   `gorm:"column:jml_liter"`         // SUM dari semua shift (nilai akurat)
	}

	var nozzleRows []nozzleRow
	err = r.db.Raw(`
		SELECT
			d.bbm_id,
			COALESCE(b.name, '-')           AS bbm_name,
			d.nozzle_id,
			COALESCE(n.description, '-')    AS disp,
			-- Totalisator Awal dari shift PERTAMA per nozzle hari ini (hanya untuk tampilan)
			(
				SELECT d2.totalisator_awal
				FROM trx_penjualan_detail d2
				JOIN trx_penjualan p2 ON p2.id_penjualan = d2.penjualan_id
				WHERE d2.nozzle_id = d.nozzle_id
				  AND d2.bbm_id   = d.bbm_id
				  AND p2.waktu_mulai BETWEEN ? AND ?
				ORDER BY p2.waktu_mulai ASC
				LIMIT 1
			) AS totalisator_awal,
			-- Totalisator Akhir dari shift TERAKHIR per nozzle hari ini (hanya untuk tampilan)
			(
				SELECT d3.totalisator_akhir
				FROM trx_penjualan_detail d3
				JOIN trx_penjualan p3 ON p3.id_penjualan = d3.penjualan_id
				WHERE d3.nozzle_id = d.nozzle_id
				  AND d3.bbm_id   = d.bbm_id
				  AND p3.waktu_mulai BETWEEN ? AND ?
				ORDER BY p3.waktu_mulai DESC
				LIMIT 1
			) AS totalisator_akhir,
			-- Total penjualan = SUM(jml_liter) per nozzle dari semua shift hari ini (AKURAT)
			SUM(d.jml_liter)               AS jml_liter
		FROM trx_penjualan_detail d
		JOIN trx_penjualan p ON p.id_penjualan = d.penjualan_id
		LEFT JOIN bbm b       ON b.id = d.bbm_id
		LEFT JOIN nozzles n   ON n.id = d.nozzle_id
		WHERE p.waktu_mulai BETWEEN ? AND ?
		GROUP BY d.bbm_id, b.name, d.nozzle_id, n.description
		ORDER BY b.name ASC, n.description ASC
	`, waktuMulaiStr, waktuSampaiStr, waktuMulaiStr, waktuSampaiStr, waktuMulaiStr, waktuSampaiStr).Scan(&nozzleRows).Error
	if err != nil {
		return nil, err
	}
	if len(nozzleRows) == 0 {
		return []PenyusutanReportGroup{}, nil
	}


	// ── First stock per BBM (dari trx_penyusutan pertama di hari itu) ──
	type firstStockRow struct {
		BBMID      uint    `gorm:"column:bbm_id"`
		FirstStock float64 `gorm:"column:first_stock"`
	}
	firstStockMap := map[uint]float64{}
	var fsRows []firstStockRow
	if fsErr := r.db.Raw(`
		SELECT DISTINCT ON (ps.bbm_id)
			ps.bbm_id,
			ps.first_stock
		FROM trx_penyusutan ps
		JOIN trx_penjualan p ON p.id_penjualan = ps.penjualan_id
		WHERE p.waktu_mulai BETWEEN ? AND ?
		ORDER BY ps.bbm_id, ps.waktu ASC, ps.id_penjualan ASC
	`, waktuMulaiStr, waktuSampaiStr).Scan(&fsRows).Error; fsErr == nil {
		for _, fs := range fsRows {
			firstStockMap[fs.BBMID] = fs.FirstStock
		}
	}

	// ── Total penerimaan per BBM di hari ini ──
	receiveMap := map[uint]float64{}
	type receiveRow struct {
		BBMID uint    `gorm:"column:bbm_id"`
		Total float64 `gorm:"column:total"`
	}
	var receives []receiveRow
	if err2 := r.db.Raw(`
		SELECT bbm_id, COALESCE(SUM(jml_liter), 0) AS total
		FROM trx_kedatangan_bbm
		WHERE tgl_kedatangan BETWEEN ? AND ?
		GROUP BY bbm_id
	`, waktuMulaiStr, waktuSampaiStr).Scan(&receives).Error; err2 == nil {
		for _, it := range receives {
			receiveMap[it.BBMID] = it.Total
		}
	}

	// ── Density/Tera per BBM di hari ini ──
	teraMap := map[uint]float64{}
	type teraRow struct {
		BBMID uint    `gorm:"column:bbm_id"`
		Total float64 `gorm:"column:total"`
	}
	var teras []teraRow
	if err3 := r.db.Raw(`
		SELECT pt.bbm_id, COALESCE(SUM(pt.qty_liter), 0) AS total
		FROM trx_penjualan_pengeluaran_test pt
		JOIN trx_penjualan p ON p.id_penjualan = pt.penjualan_id
		WHERE p.waktu_mulai BETWEEN ? AND ?
		GROUP BY pt.bbm_id
	`, waktuMulaiStr, waktuSampaiStr).Scan(&teras).Error; err3 == nil {
		for _, it := range teras {
			teraMap[it.BBMID] = it.Total
		}
	}

	// ── Latest snapshot per BBM (stok aktual & ID penyusutan) ──
	type snapRow struct {
		ID             uint64  `gorm:"column:id_penyusutan"`
		BBMID          uint    `gorm:"column:bbm_id"`
		EndstockActual float64 `gorm:"column:endstock_actual"`
	}
	snapMap := map[uint]snapRow{}
	var snaps []snapRow
	if err4 := r.db.Raw(`
		SELECT DISTINCT ON (ps.bbm_id)
			ps.id_penyusutan,
			ps.bbm_id,
			ps.endstock_actual
		FROM trx_penyusutan ps
		JOIN trx_penjualan p ON p.id_penjualan = ps.penjualan_id
		WHERE p.waktu_mulai BETWEEN ? AND ?
		ORDER BY ps.bbm_id, ps.waktu DESC, ps.id_penjualan DESC
	`, waktuMulaiStr, waktuSampaiStr).Scan(&snaps).Error; err4 == nil {
		for _, s := range snaps {
			snapMap[s.BBMID] = s
		}
	}

	// ── Susun groups per BBM ──
	groupsMap := map[uint]*PenyusutanReportGroup{}
	order := make([]uint, 0)

	for _, nr := range nozzleRows {
		if _, ok := groupsMap[nr.BBMID]; !ok {
			groupsMap[nr.BBMID] = &PenyusutanReportGroup{
				BBMID:    nr.BBMID,
				JenisBBM: nr.BBMName,
				Rows:     []PenyusutanReportRow{},
			}
			order = append(order, nr.BBMID)
		}
		g := groupsMap[nr.BBMID]

		// Total penjualan nozzle = SUM(jml_liter) dari semua shift hari ini (AKURAT)
		// Catatan: jml_liter per baris sudah dihitung sebagai (totalisator_akhir - totalisator_awal)
		// saat input penjualan, sehingga ini adalah nilai yang benar terlepas dari
		// apakah totalisator antar shift berurutan atau tidak.
		g.Rows = append(g.Rows, PenyusutanReportRow{
			Waktu:            waktuLabel,
			JenisBBM:         nr.BBMName,
			StockAwal:        firstStockMap[nr.BBMID],
			TotalPenerimaan:  receiveMap[nr.BBMID],
			Disp:             nr.Disp,
			TotalisatorAwal:  float64(nr.TotalisatorAwal),
			TotalisatorAkhir: float64(nr.TotalisatorAkhir),
			TotalPenjualan:   float64(nr.JmlLiter),
		})
		g.TotalPenjualan += float64(nr.JmlLiter)
	}


	// ── Finalisasi kalkulasi per group BBM ──
	groups := make([]PenyusutanReportGroup, 0, len(order))
	for _, bbmID := range order {
		g := groupsMap[bbmID]
		g.FirstStock = firstStockMap[bbmID]
		g.TotalPenerimaan = receiveMap[bbmID]
		g.DensityTera = teraMap[bbmID]
		g.PenjualanMinTera = g.TotalPenjualan - g.DensityTera
		snap := snapMap[bbmID]
		g.LatestPenyusutanID = snap.ID
		g.StockAkhirAktual = snap.EndstockActual

		// StockAkhirCatatan dihitung on-the-fly:
		// stokAkhirCatatan = stokAwal + totalPenerimaan - penjualanMinTera
		g.StockAkhirCatatan = g.FirstStock + g.TotalPenerimaan - g.PenjualanMinTera

		// Penyusutan = Aktual - Catatan (sesuai SPBU lama)
		g.Penyusutan = g.StockAkhirAktual - g.StockAkhirCatatan

		if g.PenjualanMinTera != 0 {
			g.Persentasi = (g.Penyusutan / g.PenjualanMinTera) * 100
		}
		groups = append(groups, *g)
	}

	return groups, nil
}

// ─── getMonthlyReport ─────────────────────────────────────────────────────────
// Logic monthly: dikelompokkan per BBM, baris detail per penjualan (per nozzle per shift).
func (r *penyusutanRepository) getMonthlyReport(month string) ([]PenyusutanReportGroup, error) {
	if month == "" {
		return []PenyusutanReportGroup{}, nil
	}

	where := "TO_CHAR(p.waktu_mulai, 'YYYY-MM') = ?"
	args := []interface{}{month}

	// ── Baris detail (per nozzle per penjualan) ──
	type baseRow struct {
		PenjualanID      uint64 `gorm:"column:penjualan_id"`
		BBMID            uint   `gorm:"column:bbm_id"`
		BBMName          string `gorm:"column:bbm_name"`
		Waktu            string `gorm:"column:waktu"`
		Disp             string `gorm:"column:disp"`
		TotalisatorAwal  int64  `gorm:"column:totalisator_awal"`
		TotalisatorAkhir int64  `gorm:"column:totalisator_akhir"`
		TotalPenjualan   int64  `gorm:"column:total_penjualan"`
	}

	var rows []baseRow
	err := r.db.Raw(`
		SELECT
			d.penjualan_id,
			d.bbm_id,
			COALESCE(b.name, '-') AS bbm_name,
			TO_CHAR(p.waktu_mulai, 'DD Mon YYYY') AS waktu,
			COALESCE(n.description, '-') AS disp,
			d.totalisator_awal,
			d.totalisator_akhir,
			d.jml_liter AS total_penjualan
		FROM trx_penjualan_detail d
		JOIN trx_penjualan p ON p.id_penjualan = d.penjualan_id
		LEFT JOIN bbm b ON b.id = d.bbm_id
		LEFT JOIN nozzles n ON n.id = d.nozzle_id
		WHERE `+where+`
		ORDER BY b.name ASC, p.waktu_mulai ASC, n.description ASC
	`, args...).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return []PenyusutanReportGroup{}, nil
	}

	// ── First stock per BBM (record penyusutan pertama dalam periode = stok awal hari/bulan) ──
	type firstStockRow struct {
		BBMID      uint    `gorm:"column:bbm_id"`
		FirstStock float64 `gorm:"column:first_stock"`
	}
	firstStockMap := map[uint]float64{}
	var fsRows []firstStockRow
	fsErr := r.db.Raw(`
		SELECT DISTINCT ON (ps.bbm_id)
			ps.bbm_id,
			ps.first_stock
		FROM trx_penyusutan ps
		JOIN trx_penjualan p ON p.id_penjualan = ps.penjualan_id
		WHERE `+where+`
		ORDER BY ps.bbm_id, ps.waktu ASC, ps.id_penjualan ASC
	`, args...).Scan(&fsRows).Error
	if fsErr == nil {
		for _, fs := range fsRows {
			firstStockMap[fs.BBMID] = fs.FirstStock
		}
	}

	// ── Total penerimaan per BBM ──
	receiveMap := map[uint]float64{}
	type receiveRow struct {
		BBMID uint    `gorm:"column:bbm_id"`
		Total float64 `gorm:"column:total"`
	}
	var receives []receiveRow
	if err2 := r.db.Raw(`
		SELECT k.bbm_id, COALESCE(SUM(k.jml_liter), 0) AS total
		FROM trx_kedatangan_bbm k
		WHERE TO_CHAR(k.tgl_kedatangan, 'YYYY-MM') = ?
		GROUP BY k.bbm_id
	`, month).Scan(&receives).Error; err2 == nil {
		for _, it := range receives {
			receiveMap[it.BBMID] = it.Total
		}
	}

	// ── Density/Tera per BBM ──
	teraMap := map[uint]float64{}
	type teraRow struct {
		BBMID uint    `gorm:"column:bbm_id"`
		Total float64 `gorm:"column:total"`
	}
	var teras []teraRow
	if err3 := r.db.Raw(`
		SELECT pt.bbm_id, COALESCE(SUM(pt.qty_liter), 0) AS total
		FROM trx_penjualan_pengeluaran_test pt
		JOIN trx_penjualan p ON p.id_penjualan = pt.penjualan_id
		WHERE `+where+`
		GROUP BY pt.bbm_id
	`, args...).Scan(&teras).Error; err3 == nil {
		for _, it := range teras {
			teraMap[it.BBMID] = it.Total
		}
	}

	// ── Latest snapshot per BBM (untuk stok aktual & penyusutan ID) ──
	type snapRow struct {
		ID             uint64  `gorm:"column:id_penyusutan"`
		BBMID          uint    `gorm:"column:bbm_id"`
		EndstockActual float64 `gorm:"column:endstock_actual"`
	}
	snapMap := map[uint]snapRow{}
	var snaps []snapRow
	if err4 := r.db.Raw(`
		SELECT DISTINCT ON (ps.bbm_id)
			ps.id_penyusutan,
			ps.bbm_id,
			ps.endstock_actual
		FROM trx_penyusutan ps
		JOIN trx_penjualan p ON p.id_penjualan = ps.penjualan_id
		WHERE `+where+`
		ORDER BY ps.bbm_id, ps.waktu DESC, ps.id_penjualan DESC
	`, args...).Scan(&snaps).Error; err4 == nil {
		for _, s := range snaps {
			snapMap[s.BBMID] = s
		}
	}

	// ── Susun groups ──
	groupsMap := map[uint]*PenyusutanReportGroup{}
	order := make([]uint, 0)

	for _, rrow := range rows {
		if _, ok := groupsMap[rrow.BBMID]; !ok {
			groupsMap[rrow.BBMID] = &PenyusutanReportGroup{
				BBMID:    rrow.BBMID,
				JenisBBM: rrow.BBMName,
				Rows:     []PenyusutanReportRow{},
			}
			order = append(order, rrow.BBMID)
		}
		g := groupsMap[rrow.BBMID]
		g.Rows = append(g.Rows, PenyusutanReportRow{
			Waktu:            rrow.Waktu,
			JenisBBM:         rrow.BBMName,
			StockAwal:        firstStockMap[rrow.BBMID],
			TotalPenerimaan:  receiveMap[rrow.BBMID],
			Disp:             rrow.Disp,
			TotalisatorAwal:  float64(rrow.TotalisatorAwal),
			TotalisatorAkhir: float64(rrow.TotalisatorAkhir),
			TotalPenjualan:   float64(rrow.TotalPenjualan),
		})
		g.TotalPenjualan += float64(rrow.TotalPenjualan)
	}


	groups := make([]PenyusutanReportGroup, 0, len(order))
	for _, bbmID := range order {
		g := groupsMap[bbmID]
		g.FirstStock = firstStockMap[bbmID]
		g.TotalPenerimaan = receiveMap[bbmID]
		g.DensityTera = teraMap[bbmID]
		g.PenjualanMinTera = g.TotalPenjualan - g.DensityTera
		snap := snapMap[bbmID]
		g.LatestPenyusutanID = snap.ID
		g.StockAkhirAktual = snap.EndstockActual

		// StockAkhirCatatan dihitung on-the-fly (seperti SPBU lama):
		// stokAkhirCatatan = stokAwal + totalPenerimaan - penjualanMinTera
		g.StockAkhirCatatan = g.FirstStock + g.TotalPenerimaan - g.PenjualanMinTera

		// Penyusutan = Aktual - Catatan (sesuai SPBU lama):
		// negatif → stok berkurang dari catatan → beban penyusutan
		// positif → stok lebih dari catatan → pendapatan penyusutan
		g.Penyusutan = g.StockAkhirAktual - g.StockAkhirCatatan

		if g.PenjualanMinTera != 0 {
			g.Persentasi = (g.Penyusutan / g.PenjualanMinTera) * 100
		}
		groups = append(groups, *g)
	}

	return groups, nil
}

// ─── GetShiftReport (Per-Shift, Daily only) ──────────────────────────────────
// Mengembalikan data per shift / per penjualan untuk tanggal tertentu.
// Setiap shift berisi BBM groups dengan nozzle rows dan kalkulasi penyusutan per shift.
func (r *penyusutanRepository) GetShiftReport(date string) ([]PenyusutanShiftGroup, error) {
	if date == "" {
		return []PenyusutanShiftGroup{}, nil
	}

	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, err
	}
	waktuMulaiStr := date + " 06:00:00"
	waktuSampaiStr := t.AddDate(0, 0, 1).Format("2006-01-02") + " 05:59:59"

	// ── Ambil semua baris nozzle untuk tanggal ini ──
	type shiftNozzleRow struct {
		PenjualanID      uint64  `gorm:"column:penjualan_id"`
		NoForm           string  `gorm:"column:no_form"`
		ShiftName        string  `gorm:"column:shift_name"`
		WaktuMulai       string  `gorm:"column:waktu_mulai"`
		WaktuAkhir       string  `gorm:"column:waktu_akhir"`
		BBMID            uint    `gorm:"column:bbm_id"`
		BBMName          string  `gorm:"column:bbm_name"`
		Disp             string  `gorm:"column:disp"`
		TotalisatorAwal  float64 `gorm:"column:totalisator_awal"`
		TotalisatorAkhir float64 `gorm:"column:totalisator_akhir"`
		JmlLiter         float64 `gorm:"column:jml_liter"`
		PenyusutanID     uint64  `gorm:"column:penyusutan_id"`
		FirstStock       float64 `gorm:"column:first_stock"`
		EndstockActual   float64 `gorm:"column:endstock_actual"`
	}

	var nozzleRows []shiftNozzleRow
	err = r.db.Raw(`
		SELECT
			p.id_penjualan                                               AS penjualan_id,
			p.no_penjualan                                               AS no_form,
			COALESCE(s.shift_name, 'Shift ' || p.shift_id::text)              AS shift_name,
			TO_CHAR(p.waktu_mulai, 'DD Mon YYYY HH24:MI')               AS waktu_mulai,
			TO_CHAR(p.waktu_akhir, 'DD Mon YYYY HH24:MI')               AS waktu_akhir,
			d.bbm_id,
			COALESCE(b.name, '-')                                        AS bbm_name,
			COALESCE(n.description, '-')                                 AS disp,
			d.totalisator_awal,
			d.totalisator_akhir,
			d.jml_liter,
			COALESCE(ps.id_penyusutan, 0)                               AS penyusutan_id,
			COALESCE(ps.first_stock, 0)                                 AS first_stock,
			COALESCE(ps.endstock_actual, 0)                             AS endstock_actual
		FROM trx_penjualan p
		JOIN trx_penjualan_detail d ON d.penjualan_id = p.id_penjualan
		LEFT JOIN bbm b ON b.id = d.bbm_id
		LEFT JOIN nozzles n ON n.id = d.nozzle_id
		LEFT JOIN shifts s ON s.id = p.shift_id
		LEFT JOIN trx_penyusutan ps ON ps.penjualan_id = p.id_penjualan AND ps.bbm_id = d.bbm_id
		WHERE p.waktu_mulai BETWEEN ? AND ?
		ORDER BY p.waktu_mulai ASC, b.name ASC, n.description ASC
	`, waktuMulaiStr, waktuSampaiStr).Scan(&nozzleRows).Error
	if err != nil {
		return nil, err
	}
	if len(nozzleRows) == 0 {
		return []PenyusutanShiftGroup{}, nil
	}

	// ── Total penerimaan per BBM per shift ──
	type receiveKey struct {
		PenjualanID uint64
		BBMID       uint
	}
	receiveShiftMap := map[receiveKey]float64{}
	type receiveDBRow struct {
		PenjualanID uint64  `gorm:"column:id_penjualan"`
		BBMID       uint    `gorm:"column:bbm_id"`
		Total       float64 `gorm:"column:total"`
	}
	var receiveRows []receiveDBRow
	if err2 := r.db.Raw(`
		SELECT
			p.id_penjualan,
			k.bbm_id,
			COALESCE(SUM(k.jml_liter), 0) AS total
		FROM trx_kedatangan_bbm k
		JOIN trx_penjualan p ON k.shift_id = p.shift_id 
			AND k.tgl_kedatangan BETWEEN p.waktu_mulai AND p.waktu_akhir
		WHERE p.waktu_mulai BETWEEN ? AND ?
		GROUP BY p.id_penjualan, k.bbm_id
	`, waktuMulaiStr, waktuSampaiStr).Scan(&receiveRows).Error; err2 == nil {
		for _, rc := range receiveRows {
			receiveShiftMap[receiveKey{rc.PenjualanID, rc.BBMID}] = rc.Total
		}
	}

	// ── Density/Tera per (penjualan_id, bbm_id) ──
	type teraKey struct {
		PenjualanID uint64
		BBMID       uint
	}
	teraMap := map[teraKey]float64{}
	type teraDBRow struct {
		PenjualanID uint64  `gorm:"column:penjualan_id"`
		BBMID       uint    `gorm:"column:bbm_id"`
		Total       float64 `gorm:"column:total"`
	}
	var teras []teraDBRow
	if err3 := r.db.Raw(`
		SELECT pt.penjualan_id, pt.bbm_id, COALESCE(SUM(pt.qty_liter), 0) AS total
		FROM trx_penjualan_pengeluaran_test pt
		JOIN trx_penjualan p ON p.id_penjualan = pt.penjualan_id
		WHERE p.waktu_mulai BETWEEN ? AND ?
		GROUP BY pt.penjualan_id, pt.bbm_id
	`, waktuMulaiStr, waktuSampaiStr).Scan(&teras).Error; err3 == nil {
		for _, t := range teras {
			teraMap[teraKey{t.PenjualanID, t.BBMID}] = t.Total
		}
	}

	// ── Susun struktur per-shift ──
	type shiftKey struct {
		PenjualanID uint64
		BBMID       uint
	}

	// Urutan penjualan
	penjualanOrder := []uint64{}
	penjualanSeen := map[uint64]bool{}
	// Urutan BBM per penjualan
	bbmOrder := map[uint64][]uint{}
	bbmSeen := map[shiftKey]bool{}
	// Data sementara
	shiftMeta := map[uint64]PenyusutanShiftGroup{}
	bbmData := map[shiftKey]*PenyusutanShiftBBM{}

	for _, row := range nozzleRows {
		pID := row.PenjualanID
		bID := row.BBMID
		key := shiftKey{pID, bID}

		// Inisialisasi shift
		if !penjualanSeen[pID] {
			penjualanSeen[pID] = true
			penjualanOrder = append(penjualanOrder, pID)
			shiftMeta[pID] = PenyusutanShiftGroup{
				PenjualanID: pID,
				NoForm:      row.NoForm,
				ShiftName:   row.ShiftName,
				WaktuMulai:  row.WaktuMulai,
				WaktuAkhir:  row.WaktuAkhir,
			}
		}

		// Inisialisasi BBM dalam shift
		if !bbmSeen[key] {
			bbmSeen[key] = true
			bbmOrder[pID] = append(bbmOrder[pID], bID)
			bbmData[key] = &PenyusutanShiftBBM{
				BBMID:           bID,
				JenisBBM:        row.BBMName,
				PenyusutanID:    row.PenyusutanID,
				StokAwal:        row.FirstStock,
				TotalPenerimaan: receiveShiftMap[receiveKey{pID, bID}],
				Nozzles:         []PenyusutanShiftNozzle{},
			}
		}

		// Tambah nozzle row
		bd := bbmData[key]
		bd.Nozzles = append(bd.Nozzles, PenyusutanShiftNozzle{
			Disp:             row.Disp,
			TotalisatorAwal:  row.TotalisatorAwal,
			TotalisatorAkhir: row.TotalisatorAkhir,
			JmlLiter:         row.JmlLiter,
		})
		bd.TotalPenjualan += row.JmlLiter

		// Update stok aktual (dari snapshot penyusutan)
		bd.StokAkhirAktual = row.EndstockActual
	}

	// ── Finalisasi kalkulasi per BBM per shift ──
	result := make([]PenyusutanShiftGroup, 0, len(penjualanOrder))
	for _, pID := range penjualanOrder {
		meta := shiftMeta[pID]
		bbmGroups := make([]PenyusutanShiftBBM, 0, len(bbmOrder[pID]))

		for _, bID := range bbmOrder[pID] {
			key := shiftKey{pID, bID}
			bd := bbmData[key]
			tera := teraMap[teraKey{pID, bID}]
			bd.DensityTera = tera
			bd.PenjualanMinTera = bd.TotalPenjualan - tera

			// StokAkhirCatatan = stokAwal + penerimaan - penjualanMinTera
			bd.StokAkhirCatatan = bd.StokAwal + bd.TotalPenerimaan - bd.PenjualanMinTera

			// Penyusutan = Aktual - Catatan (sesuai SPBU lama)
			bd.Penyusutan = bd.StokAkhirAktual - bd.StokAkhirCatatan

			if bd.PenjualanMinTera != 0 {
				bd.Persentasi = (bd.Penyusutan / bd.PenjualanMinTera) * 100
			}

			bbmGroups = append(bbmGroups, *bd)
		}

		meta.BBMGroups = bbmGroups
		result = append(result, meta)
	}

	return result, nil
}

// ─── UpdateEndstockActual ────────────────────────────────────────────────────
// Update stok akhir aktual (hasil penghitungan fisik tangki di lapangan).
func (r *penyusutanRepository) UpdateEndstockActual(id uint64, actual float64, updatedBy *uint) error {
	return r.db.Model(&entity.TrxPenyusutan{}).Where("id_penyusutan = ?", id).Updates(map[string]interface{}{
		"endstock_actual": actual,
		"updated_by":      updatedBy,
		"updated":         time.Now(),
	}).Error
}
