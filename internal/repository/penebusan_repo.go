package repository

import (
	"fmt"
	"strings"

	"spbu_go/internal/dto"
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

type PenebusanRepository interface {
	FindAll() ([]entity.TrxPenebusan, error)
	FindByID(id uint64) (*entity.TrxPenebusan, error)
	Datatable(req dto.DatatableRequest) (int64, int64, []entity.TrxPenebusan, error)
	Create(p *entity.TrxPenebusan) error
	Update(p *entity.TrxPenebusan) error
	Delete(id uint64) error
	// FindStokDO returns all CO penebusan that have a no_so, with details preloaded.
	// Used by the /transaction/stok-do page.
	FindStokDO() ([]entity.TrxPenebusan, error)
	// DatatableStokDO returns paginated stok-do rows for server-side DataTables.
	DatatableStokDO(req dto.DatatableRequest) (int64, int64, []dto.StokDODTRow, error)
	// GetStokDOSummary returns aggregate summary stats for the stok-do page header cards.
	GetStokDOSummary() dto.StokDOSummary
	// UpdateDetailQtyTerkirim sets qty_terkirim on a single detail row.
	UpdateDetailQtyTerkirim(detailID uint64, qty int64) error
}

type penebusanRepository struct {
	db *gorm.DB
}

func NewPenebusanRepository(db *gorm.DB) PenebusanRepository {
	return &penebusanRepository{db: db}
}

func (r *penebusanRepository) FindAll() ([]entity.TrxPenebusan, error) {
	var list []entity.TrxPenebusan
	result := r.db.
		Preload("Wallet").
		Preload("Updater").
		Preload("Details").
		Preload("Details.BBM").
		Order("tgl_penebusan DESC, id DESC").
		Find(&list)
	return list, result.Error
}

func (r *penebusanRepository) Datatable(req dto.DatatableRequest) (int64, int64, []entity.TrxPenebusan, error) {
	var list []entity.TrxPenebusan
	var total, filtered int64

	base := r.db.Model(&entity.TrxPenebusan{})
	base.Count(&total)

	if req.Search.Value != "" {
		s := "%" + req.Search.Value + "%"
		base = base.Where("no_penebusan ILIKE ? OR no_so ILIKE ?", s, s)
	}
	base.Count(&filtered)

	// Ordering
	orderCol := "tgl_penebusan"
	orderDir := "DESC"
	if len(req.Order) > 0 {
		if req.Order[0].Dir == "asc" {
			orderDir = "ASC"
		}
		switch req.Order[0].Column {
		case 1:
			orderCol = "tgl_penebusan"
		case 2:
			orderCol = "no_penebusan"
		case 3:
			orderCol = "no_so"
		case 4:
			orderCol = "adm_bank"
		case 5:
			orderCol = "total_bayar"
		case 6:
			orderCol = "status"
		}
	}

	limit := req.Length
	if limit <= 0 {
		limit = 10
	}

	err := base.
		Preload("Wallet").
		Order(orderCol + " " + orderDir + ", id DESC").
		Offset(req.Start).
		Limit(limit).
		Find(&list).Error

	return total, filtered, list, err
}

func (r *penebusanRepository) FindByID(id uint64) (*entity.TrxPenebusan, error) {
	var p entity.TrxPenebusan
	err := r.db.
		Preload("Wallet").
		Preload("Updater").
		Preload("Details").
		Preload("Details.BBM").
		First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *penebusanRepository) Create(p *entity.TrxPenebusan) error {
	return r.db.Omit("Wallet", "Updater", "Details.BBM").Create(p).Error
}

func (r *penebusanRepository) Update(p *entity.TrxPenebusan) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Omit("Wallet", "Updater", "Details", "Details.BBM").Save(p).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("penebusan_id = ?", p.ID).Delete(&entity.TrxPenebusanDetail{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(p.Details) > 0 {
		for i := range p.Details {
			p.Details[i].PenebusanID = p.ID
		}
		if err := tx.Omit("BBM").Create(&p.Details).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *penebusanRepository) Delete(id uint64) error {
	return r.db.Delete(&entity.TrxPenebusan{}, id).Error
}

// FindStokDO returns all CO penebusan with no_so, ordered by tgl_penebusan DESC.
// Preloads Details and Details.BBM for stok-do display.
func (r *penebusanRepository) FindStokDO() ([]entity.TrxPenebusan, error) {
	var list []entity.TrxPenebusan
	err := r.db.
		Where("status = ? AND no_so IS NOT NULL AND no_so != ''", entity.PenebusanComplete).
		Preload("Details").
		Preload("Details.BBM").
		Order("tgl_penebusan DESC, id DESC").
		Find(&list).Error
	return list, err
}

// UpdateDetailQtyTerkirim sets qty_terkirim for a single detail row.
func (r *penebusanRepository) UpdateDetailQtyTerkirim(detailID uint64, qty int64) error {
	return r.db.Model(&entity.TrxPenebusanDetail{}).
		Where("id = ?", detailID).
		Update("qty_terkirim", qty).Error
}

// GetStokDOSummary returns aggregate stats for the stok-do page header cards.
func (r *penebusanRepository) GetStokDOSummary() dto.StokDOSummary {
	var s dto.StokDOSummary
	r.db.Raw(`
		SELECT
			COUNT(DISTINCT p.no_so)   AS total_so,
			COUNT(d.id)               AS total_items,
			COUNT(CASE WHEN d.qty_terkirim > 0 AND d.jml_liter <= d.qty_terkirim THEN 1 END) AS selesai,
			COUNT(CASE WHEN d.qty_terkirim = 0 OR  d.jml_liter > d.qty_terkirim THEN 1 END) AS belum
		FROM trx_penebusan_detail d
		JOIN trx_penebusan p ON p.id = d.penebusan_id
			AND p.status = 'CO'
			AND p.no_so IS NOT NULL AND p.no_so != ''
			AND p.deleted_at IS NULL
	`).Scan(&s)
	return s
}

// DatatableStokDO returns paginated stok-do rows with search, sort, and column filter support.
func (r *penebusanRepository) DatatableStokDO(req dto.DatatableRequest) (int64, int64, []dto.StokDODTRow, error) {
	baseSQL := `
		FROM trx_penebusan_detail d
		JOIN trx_penebusan p ON p.id = d.penebusan_id
			AND p.status = 'CO'
			AND p.no_so IS NOT NULL AND p.no_so != ''
			AND p.deleted_at IS NULL
		JOIN bbm b ON b.id = d.bbm_id
	`

	var conditions []string
	var args []interface{}

	// Global search
	if req.Search.Value != "" {
		s := "%" + req.Search.Value + "%"
		conditions = append(conditions, "(p.no_so ILIKE ? OR p.no_penebusan ILIKE ? OR b.name ILIKE ?)")
		args = append(args, s, s, s)
	}

	// Column 5 search — status filter (Selesai / Belum)
	if len(req.Columns) > 5 && req.Columns[5].Search.Value != "" {
		switch req.Columns[5].Search.Value {
		case "Selesai":
			conditions = append(conditions, "(d.qty_terkirim > 0 AND d.jml_liter <= d.qty_terkirim)")
		case "Belum":
			conditions = append(conditions, "(d.qty_terkirim = 0 OR d.jml_liter > d.qty_terkirim)")
		}
	}

	whereSQL := ""
	if len(conditions) > 0 {
		whereSQL = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Total count (no filters)
	var total int64
	r.db.Raw("SELECT COUNT(*) " + baseSQL).Scan(&total)

	// Filtered count
	filtered := total
	if whereSQL != "" {
		r.db.Raw("SELECT COUNT(*) "+baseSQL+whereSQL, args...).Scan(&filtered)
	}

	// Order
	orderCol := "p.no_so"
	orderDir := "ASC"
	if len(req.Order) > 0 {
		if req.Order[0].Dir == "desc" {
			orderDir = "DESC"
		}
		switch req.Order[0].Column {
		case 0:
			orderCol = "p.no_so"
		case 1:
			orderCol = "b.name"
		case 2:
			orderCol = "d.jml_liter"
		case 3:
			orderCol = "d.qty_terkirim"
		case 4:
			orderCol = "(d.jml_liter - d.qty_terkirim)"
		}
	}

	limit := req.Length
	if limit <= 0 {
		limit = 25
	}

	selectSQL := `
		SELECT
			d.id                                     AS detail_id,
			p.id                                     AS penebusan_id,
			p.no_penebusan,
			COALESCE(p.no_so, '')                   AS no_so,
			TO_CHAR(p.tgl_penebusan, 'YYYY-MM-DD')  AS tgl_penebusan,
			b.name                                   AS jenis_bbm,
			d.jml_liter,
			d.qty_terkirim,
			GREATEST(d.jml_liter - d.qty_terkirim, 0) AS sisa_liter,
			CASE WHEN d.qty_terkirim > 0 AND d.jml_liter <= d.qty_terkirim
				THEN 'Selesai' ELSE 'Belum' END     AS status_kirim
	`

	orderSQL := fmt.Sprintf(" ORDER BY %s %s", orderCol, orderDir)
	limitSQL := fmt.Sprintf(" LIMIT %d OFFSET %d", limit, req.Start)

	var rows []dto.StokDODTRow
	err := r.db.Raw(selectSQL+baseSQL+whereSQL+orderSQL+limitSQL, args...).Scan(&rows).Error
	return total, filtered, rows, err
}
