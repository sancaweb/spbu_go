package repository

import (
	"fmt"
	"strings"

	"spbu_go/internal/dto"
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

// SOSisaRow — row untuk dropdown No SO (hanya yang masih ada sisa liter).
type SOSisaRow struct {
	PenebusanID  uint64 `gorm:"column:penebusan_id" json:"penebusan_id"`
	NoSO         string `gorm:"column:no_so" json:"no_so"`
	NoPenebusan  string `gorm:"column:no_penebusan" json:"no_penebusan"`
	TglPenebusan string `gorm:"column:tgl_penebusan" json:"tgl_penebusan"`
}

// BBMSisaRow — row untuk dropdown Jenis BBM berdasarkan SO yang dipilih.
type BBMSisaRow struct {
	DetailID    uint64 `gorm:"column:detail_id" json:"detail_id"`
	BBMID       uint   `gorm:"column:bbm_id" json:"bbm_id"`
	NamaBBM     string `gorm:"column:nama_bbm" json:"nama_bbm"`
	JmlLiter    int64  `gorm:"column:jml_liter" json:"jml_liter"`
	QtyTerkirim int64  `gorm:"column:qty_terkirim" json:"qty_terkirim"`
	SisaLiter   int64  `gorm:"column:sisa_liter" json:"sisa_liter"`
}

// KedatanganDTRow — row untuk server-side datatable kedatangan BBM.
type KedatanganDTRow struct {
	ID            uint64 `gorm:"column:id" json:"id"`
	NoSO          string `gorm:"column:no_so" json:"no_so"`
	NoLO          string `gorm:"column:no_lo" json:"no_lo"`
	TglKedatangan string `gorm:"column:tgl_kedatangan" json:"tgl_kedatangan"`
	ShiftName     string `gorm:"column:shift_name" json:"shift_name"`
	JenisBBM      string `gorm:"column:jenis_bbm" json:"jenis_bbm"`
	JmlLiter      int64  `gorm:"column:jml_liter" json:"jml_liter"`
	NamaDriver    string `gorm:"column:nama_driver" json:"nama_driver"`
	NoPol         string `gorm:"column:no_pol" json:"no_pol"`
	PenebusanID   uint64 `gorm:"column:penebusan_id" json:"penebusan_id"`
}

type KedatanganBBMRepository interface {
	Datatable(req dto.DatatableRequest) (int64, int64, []KedatanganDTRow, error)
	FindByID(id uint64) (*entity.TrxKedatanganBBM, error)
	Create(k *entity.TrxKedatanganBBM) error
	Update(k *entity.TrxKedatanganBBM) error
	Delete(id uint64) error
	// FindSOSisa returns penebusan (CO, has no_so) that still have unreceived liter.
	FindSOSisa() ([]SOSisaRow, error)
	// FindBBMSisaByPenebusan returns detail rows for a given penebusan that still have sisa liter.
	FindBBMSisaByPenebusan(penebusanID uint64) ([]BBMSisaRow, error)
}

type kedatanganBBMRepository struct {
	db *gorm.DB
}

func NewKedatanganBBMRepository(db *gorm.DB) KedatanganBBMRepository {
	return &kedatanganBBMRepository{db: db}
}

// FindSOSisa — penebusan CO dengan no_so, yang detailnya masih ada sisa liter belum terkirim.
func (r *kedatanganBBMRepository) FindSOSisa() ([]SOSisaRow, error) {
	var rows []SOSisaRow
	err := r.db.Raw(`
		SELECT
			p.id                                     AS penebusan_id,
			COALESCE(p.no_so, '')                   AS no_so,
			p.no_penebusan,
			TO_CHAR(p.tgl_penebusan, 'YYYY-MM-DD')  AS tgl_penebusan
		FROM trx_penebusan p
		WHERE p.status = 'CO'
		  AND p.no_so IS NOT NULL AND p.no_so != ''
		  AND p.deleted_at IS NULL
		  AND EXISTS (
			  SELECT 1 FROM trx_penebusan_detail d
			  WHERE d.penebusan_id = p.id
			    AND d.jml_liter > d.qty_terkirim
		  )
		ORDER BY p.tgl_penebusan DESC, p.no_so
	`).Scan(&rows).Error
	return rows, err
}

// FindBBMSisaByPenebusan — detail BBM dari penebusan tertentu yang masih ada sisa.
func (r *kedatanganBBMRepository) FindBBMSisaByPenebusan(penebusanID uint64) ([]BBMSisaRow, error) {
	var rows []BBMSisaRow
	err := r.db.Raw(`
		SELECT
			d.id                                              AS detail_id,
			d.bbm_id,
			b.name                                            AS nama_bbm,
			d.jml_liter,
			d.qty_terkirim,
			GREATEST(d.jml_liter - d.qty_terkirim, 0)        AS sisa_liter
		FROM trx_penebusan_detail d
		JOIN bbm b ON b.id = d.bbm_id
		WHERE d.penebusan_id = ?
		  AND d.jml_liter > d.qty_terkirim
		ORDER BY b.name
	`, penebusanID).Scan(&rows).Error
	return rows, err
}

func (r *kedatanganBBMRepository) Datatable(req dto.DatatableRequest) (int64, int64, []KedatanganDTRow, error) {
	baseSQL := `
		FROM trx_kedatangan_bbm k
		JOIN trx_penebusan p ON p.id = k.penebusan_id
		JOIN shifts s ON s.id = k.shift_id
		JOIN bbm b ON b.id = k.bbm_id
	`

	var conditions []string
	var args []interface{}

	if req.Search.Value != "" {
		q := "%" + req.Search.Value + "%"
		conditions = append(conditions, "(p.no_so ILIKE ? OR k.no_lo ILIKE ? OR b.name ILIKE ? OR k.nama_driver ILIKE ?)")
		args = append(args, q, q, q, q)
	}

	whereSQL := ""
	if len(conditions) > 0 {
		whereSQL = " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	r.db.Raw("SELECT COUNT(*) " + baseSQL).Scan(&total)

	filtered := total
	if whereSQL != "" {
		r.db.Raw("SELECT COUNT(*) "+baseSQL+whereSQL, args...).Scan(&filtered)
	}

	orderCol := "k.tgl_kedatangan"
	orderDir := "DESC"
	if len(req.Order) > 0 {
		if req.Order[0].Dir == "asc" {
			orderDir = "ASC"
		}
		switch req.Order[0].Column {
		case 1:
			orderCol = "p.no_so"
		case 2:
			orderCol = "k.no_lo"
		case 3:
			orderCol = "k.tgl_kedatangan"
		case 4:
			orderCol = "s.shift_name"
		case 5:
			orderCol = "b.name"
		case 6:
			orderCol = "k.jml_liter"
		}
	}

	limit := req.Length
	if limit <= 0 {
		limit = 25
	}

	selectSQL := `
		SELECT
			k.id_kedatangan_bbm                                          AS id,
			COALESCE(p.no_so, '')                                        AS no_so,
			k.no_lo,
			TO_CHAR(k.tgl_kedatangan, 'YYYY-MM-DD HH24:MI')             AS tgl_kedatangan,
			s.shift_name,
			b.name                                                       AS jenis_bbm,
			k.jml_liter,
			k.nama_driver,
			k.no_pol,
			k.penebusan_id
	`

	orderSQL := fmt.Sprintf(" ORDER BY %s %s", orderCol, orderDir)
	limitSQL := fmt.Sprintf(" LIMIT %d OFFSET %d", limit, req.Start)

	var rows []KedatanganDTRow
	err := r.db.Raw(selectSQL+baseSQL+whereSQL+orderSQL+limitSQL, args...).Scan(&rows).Error
	return total, filtered, rows, err
}

func (r *kedatanganBBMRepository) FindByID(id uint64) (*entity.TrxKedatanganBBM, error) {
	var k entity.TrxKedatanganBBM
	err := r.db.
		Preload("Penebusan").
		Preload("PenebusanDetail").
		Preload("Shift").
		Preload("BBM").
		First(&k, id).Error
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *kedatanganBBMRepository) Create(k *entity.TrxKedatanganBBM) error {
	return r.db.Omit("Penebusan", "PenebusanDetail", "Shift", "BBM", "Creator", "Updater").Create(k).Error
}

func (r *kedatanganBBMRepository) Update(k *entity.TrxKedatanganBBM) error {
	return r.db.Omit("Penebusan", "PenebusanDetail", "Shift", "BBM", "Creator", "Updater").Save(k).Error
}

func (r *kedatanganBBMRepository) Delete(id uint64) error {
	return r.db.Delete(&entity.TrxKedatanganBBM{}, id).Error
}
