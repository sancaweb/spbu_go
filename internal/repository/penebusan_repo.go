package repository

import (
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
