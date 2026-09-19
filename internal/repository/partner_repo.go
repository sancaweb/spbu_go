package repository

import (
	"errors"

	"spbu_go/internal/dto"
	"spbu_go/internal/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PartnerRepository interface {
	FindAll() ([]entity.Partner, error)
	FindActive() ([]entity.Partner, error)
	FindInactive() ([]entity.Partner, error)
	Datatable(req dto.DatatableRequest, isActive bool) (int64, int64, []dto.PartnerDTRow, error)
	FindByID(id uint) (entity.Partner, error)
	Create(partner *entity.Partner) error
	Update(partner *entity.Partner) error
	Delete(id uint) error
	DeletePermanent(id uint) error
	Restore(id uint) error
}

type partnerRepository struct {
	db *gorm.DB
}

func NewPartnerRepository(db *gorm.DB) PartnerRepository {
	return &partnerRepository{db}
}

func (r *partnerRepository) FindAll() ([]entity.Partner, error) {
	var partners []entity.Partner
	if err := r.deactivateExpiredAutoOff(); err != nil {
		return nil, err
	}
	err := r.db.Preload("Updater").Order("name ASC").Find(&partners).Error
	return partners, err
}

func (r *partnerRepository) FindActive() ([]entity.Partner, error) {
	var partners []entity.Partner
	if err := r.deactivateExpiredAutoOff(); err != nil {
		return nil, err
	}
	err := r.db.Where("isactive = ?", true).Preload("Updater").Order("name ASC").Find(&partners).Error
	return partners, err
}

func (r *partnerRepository) FindInactive() ([]entity.Partner, error) {
	var partners []entity.Partner
	if err := r.deactivateExpiredAutoOff(); err != nil {
		return nil, err
	}
	// Unscoped to bypass soft delete, combined with isactive logic.
	err := r.db.Unscoped().Where("isactive = ? OR deleted_at IS NOT NULL", false).Preload("Updater").Order("name ASC").Find(&partners).Error
	return partners, err
}

func (r *partnerRepository) Datatable(req dto.DatatableRequest, isActive bool) (int64, int64, []dto.PartnerDTRow, error) {
	partners := make([]dto.PartnerDTRow, 0)
	var total int64
	var filtered int64
	if err := r.deactivateExpiredAutoOff(); err != nil {
		return 0, 0, nil, err
	}

	query := r.db.Model(&entity.Partner{})
	if isActive {
		query = query.Where("isactive = ?", true)
	} else {
		query = r.db.Unscoped().Model(&entity.Partner{}).Where("isactive = ? OR deleted_at IS NOT NULL", false)
	}

	if err := query.Count(&total).Error; err != nil {
		return 0, 0, nil, err
	}

	// Global Search
	searchValue := req.Search.Value
	if searchValue != "" {
		searchQuery := "%" + searchValue + "%"
		query = query.Where("name ILIKE ? OR contact_person ILIKE ?", searchQuery, searchQuery)
	}

	if err := query.Count(&filtered).Error; err != nil {
		return 0, 0, nil, err
	}

	// Ordering
	if len(req.Order) > 0 {
		orderColIndex := req.Order[0].Column
		orderDir := req.Order[0].Dir

		var colName string
		switch orderColIndex {
		case 0:
			colName = "name"
		case 1:
			colName = "contact_person"
		case 2:
			colName = "phone"
		case 3:
			colName = "start_date"
		case 4:
			colName = "end_date"
		case 5:
			colName = "isactive"
		case 6:
			colName = "auto_off"
		default:
			colName = "name"
		}

		if orderDir != "asc" && orderDir != "desc" {
			orderDir = "asc"
		}
		query = query.Order(colName + " " + orderDir)
	} else {
		query = query.Order("name ASC")
	}

	// Pagination
	if req.Length > 0 {
		query = query.Limit(req.Length).Offset(req.Start)
	}

	err := query.Select(`partners.*,
		((partners.isactive = FALSE OR partners.deleted_at IS NOT NULL)
		AND NOT EXISTS (
			SELECT 1 FROM trx_piutang WHERE trx_piutang.pelanggan_id = partners.id
		)) AS can_delete_permanent`).
		Preload("Updater").Find(&partners).Error
	return total, filtered, partners, err
}

func (r *partnerRepository) FindByID(id uint) (entity.Partner, error) {
	var partner entity.Partner
	if err := r.deactivateExpiredAutoOff(); err != nil {
		return partner, err
	}
	err := r.db.Preload("Updater").First(&partner, id).Error
	return partner, err
}

func (r *partnerRepository) Create(partner *entity.Partner) error {
	return r.db.Omit("Updater").Create(partner).Error
}

func (r *partnerRepository) Update(partner *entity.Partner) error {
	return r.db.Omit("Updater").Save(partner).Error
}

func (r *partnerRepository) Delete(id uint) error {
	// First mark as inactive to keep business logic consistent
	r.db.Model(&entity.Partner{}).Where("id = ?", id).Update("isactive", false)
	// Then soft delete
	return r.db.Delete(&entity.Partner{}, id).Error
}

func (r *partnerRepository) DeletePermanent(id uint) error {
	if id == 0 {
		return gorm.ErrRecordNotFound
	}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var partner entity.Partner
		// The row lock serializes restore/delete and blocks new FK references
		// until the piutang check and deletion have completed.
		if err := tx.Unscoped().Clauses(clause.Locking{Strength: "UPDATE"}).First(&partner, id).Error; err != nil {
			return err
		}
		if partner.IsActive && !partner.DeletedAt.Valid {
			return entity.ErrPartnerNotArchived
		}

		var hasPiutang bool
		if err := tx.Raw(`SELECT EXISTS (
			SELECT 1 FROM trx_piutang WHERE pelanggan_id = ?
		)`, id).Scan(&hasPiutang).Error; err != nil {
			return err
		}
		if hasPiutang {
			return entity.ErrPartnerHasPiutang
		}

		result := tx.Unscoped().Delete(&partner)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})

	var sqlError interface{ SQLState() string }
	if errors.Is(err, gorm.ErrForeignKeyViolated) ||
		(errors.As(err, &sqlError) && sqlError.SQLState() == "23503") {
		return entity.ErrPartnerReferenced
	}
	return err
}

func (r *partnerRepository) Restore(id uint) error {
	// Restore soft deleted row by setting deleted_at to null and isactive to true.
	return r.db.Unscoped().Model(&entity.Partner{}).Where("id = ?", id).Updates(map[string]interface{}{
		"deleted_at": nil,
		"isactive":   true,
	}).Error
}

// deactivateExpiredAutoOff persists the automatic end of a partner agreement.
// The end date remains active for the whole date and expires on the following day.
func (r *partnerRepository) deactivateExpiredAutoOff() error {
	return r.db.Model(&entity.Partner{}).
		Where("auto_off = ? AND end_date IS NOT NULL AND end_date < CURRENT_DATE AND isactive = ?", true, true).
		Update("isactive", false).Error
}
