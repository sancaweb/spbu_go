package repository

import (
	"fmt"
	"strings"

	"spbu_go/internal/dto"
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

// PenjualanRepository — kontrak akses data trx_penjualan.
type PenjualanRepository interface {
	FindAll() ([]entity.TrxPenjualan, error)
	FindByID(id uint64) (*entity.TrxPenjualan, error)
	Datatable(req dto.DatatableRequest) (int64, int64, []dto.PenjualanDTRow, error)
	Create(p *entity.TrxPenjualan) error
	Update(p *entity.TrxPenjualan) error
	Delete(id uint64) error
}

type penjualanRepository struct {
	db *gorm.DB
}

func NewPenjualanRepository(db *gorm.DB) PenjualanRepository {
	return &penjualanRepository{db: db}
}

// nextNoPenjualan generates nomor urut PJL/YYYY/MM/NNNN berdasarkan waktu_mulai.
func (r *penjualanRepository) nextNoPenjualan(p *entity.TrxPenjualan) string {
	prefix := fmt.Sprintf("PJL/%04d/%02d/", p.WaktuMulai.Year(), int(p.WaktuMulai.Month()))
	var maxSeq int
	r.db.Raw(
		`SELECT COALESCE(MAX(
			CASE WHEN no_penjualan ~ '^PJL/[0-9]{4}/[0-9]{2}/[0-9]+$'
			THEN SPLIT_PART(no_penjualan, '/', 4)::INT ELSE 0 END
		), 0) FROM trx_penjualan WHERE no_penjualan LIKE ?`,
		prefix+"%",
	).Scan(&maxSeq)
	return fmt.Sprintf("%s%04d", prefix, maxSeq+1)
}

func (r *penjualanRepository) FindAll() ([]entity.TrxPenjualan, error) {
	var list []entity.TrxPenjualan
	err := r.db.
		Preload("Shift").
		Preload("Creator").
		Preload("Updater").
		Preload("Details").
		Preload("Details.Tiang").
		Preload("Details.Nozzle").
		Preload("Details.BBM").
		Order("waktu_mulai DESC, id_penjualan DESC").
		Find(&list).Error
	return list, err
}

func (r *penjualanRepository) FindByID(id uint64) (*entity.TrxPenjualan, error) {
	var p entity.TrxPenjualan
	err := r.db.
		Preload("Shift").
		Preload("Creator").
		Preload("Updater").
		Preload("Details").
		Preload("Details.Tiang").
		Preload("Details.Nozzle").
		Preload("Details.BBM").
		First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *penjualanRepository) Datatable(req dto.DatatableRequest) (int64, int64, []dto.PenjualanDTRow, error) {
	baseSQL := `
		FROM trx_penjualan p
		LEFT JOIN shifts s ON s.id = p.shift_id
		LEFT JOIN users u ON u.id = p.created_by
	`
	var conditions []string
	var args []interface{}

	// Global search
	if req.Search.Value != "" {
		q := "%" + req.Search.Value + "%"
		conditions = append(conditions, "(p.no_penjualan ILIKE ? OR s.shift_name ILIKE ?)")
		args = append(args, q, q)
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

	// Ordering
	orderCol := "p.waktu_mulai"
	orderDir := "DESC"
	if len(req.Order) > 0 {
		if req.Order[0].Dir == "asc" {
			orderDir = "ASC"
		}
		switch req.Order[0].Column {
		case 1:
			orderCol = "p.no_penjualan"
		case 2:
			orderCol = "p.waktu_mulai"
		case 3:
			orderCol = "s.shift_name"
		case 4:
			orderCol = "p.total_rp_totalisator"
		case 5:
			orderCol = "p.total_penerimaan"
		case 6:
			orderCol = "p.aktual_uang"
		case 7:
			orderCol = "p.selisih"
		}
	}

	limit := req.Length
	if limit <= 0 {
		limit = 25
	}

	selectSQL := `
		SELECT
			p.id_penjualan                               AS id,
			p.no_penjualan,
			TO_CHAR(p.waktu_mulai, 'YYYY-MM-DD')         AS tgl,
			TO_CHAR(p.waktu_mulai, 'HH24:MI')            AS waktu_mulai,
			TO_CHAR(p.waktu_akhir,  'HH24:MI')           AS waktu_akhir,
			COALESCE(s.shift_name, '-')                  AS shift_name,
			p.total_rp_totalisator,
			p.total_penerimaan,
			p.aktual_uang,
			p.selisih
	`
	orderSQL := fmt.Sprintf(" ORDER BY %s %s, p.id_penjualan DESC", orderCol, orderDir)
	limitSQL := fmt.Sprintf(" LIMIT %d OFFSET %d", limit, req.Start)

	var rows []dto.PenjualanDTRow
	err := r.db.Raw(selectSQL+baseSQL+whereSQL+orderSQL+limitSQL, args...).Scan(&rows).Error
	return total, filtered, rows, err
}

func (r *penjualanRepository) Create(p *entity.TrxPenjualan) error {
	// Generate nomor dokumen
	p.NoPenjualan = r.nextNoPenjualan(p)

	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Omit("Shift", "Creator", "Updater", "Details.Tiang", "Details.Nozzle", "Details.BBM").
		Create(p).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Set PenjualanID pada detail setelah header tersimpan
	for i := range p.Details {
		p.Details[i].PenjualanID = p.ID
	}
	if len(p.Details) > 0 {
		if err := tx.Omit("Tiang", "Nozzle", "BBM").Create(&p.Details).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *penjualanRepository) Update(p *entity.TrxPenjualan) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Update header
	if err := tx.Omit("Shift", "Creator", "Updater", "Details").
		Save(p).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Hapus detail lama, simpan yang baru
	if err := tx.Where("penjualan_id = ?", p.ID).Delete(&entity.TrxPenjualanDetail{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if len(p.Details) > 0 {
		for i := range p.Details {
			p.Details[i].PenjualanID = p.ID
			p.Details[i].ID = 0 // reset PK agar tidak duplikat
		}
		if err := tx.Omit("Tiang", "Nozzle", "BBM").Create(&p.Details).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *penjualanRepository) Delete(id uint64) error {
	return r.db.Delete(&entity.TrxPenjualan{}, id).Error
}
