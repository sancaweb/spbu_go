package repository

import (
	"spbu_go/internal/dto"
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

type PartnerRepository interface {
	FindAll() ([]entity.Partner, error)
	FindActive() ([]entity.Partner, error)
	FindInactive() ([]entity.Partner, error)
	Datatable(req dto.DatatableRequest, isActive bool) (int64, int64, []entity.Partner, error)
	FindByID(id uint) (entity.Partner, error)
	Create(partner *entity.Partner) error
	Update(partner *entity.Partner) error
	Delete(id uint) error
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
	err := r.db.Preload("Updater").Order("name ASC").Find(&partners).Error
	return partners, err
}

func (r *partnerRepository) FindActive() ([]entity.Partner, error) {
	var partners []entity.Partner
	err := r.db.Where("is_active = ?", true).Preload("Updater").Order("name ASC").Find(&partners).Error
	return partners, err
}

func (r *partnerRepository) FindInactive() ([]entity.Partner, error) {
	var partners []entity.Partner
	// Unscoped to bypass soft delete, combined with is_active logic
	err := r.db.Unscoped().Where("is_active = ? OR deleted_at IS NOT NULL", false).Preload("Updater").Order("name ASC").Find(&partners).Error
	return partners, err
}

func (r *partnerRepository) Datatable(req dto.DatatableRequest, isActive bool) (int64, int64, []entity.Partner, error) {
	var partners []entity.Partner
	var total int64
	var filtered int64

	query := r.db.Model(&entity.Partner{})
	if isActive {
		query = query.Where("is_active = ?", true)
	} else {
		query = r.db.Unscoped().Model(&entity.Partner{}).Where("is_active = ? OR deleted_at IS NOT NULL", false)
	}

	query.Count(&total)

	// Global Search
	searchValue := req.Search.Value
	if searchValue != "" {
		searchQuery := "%" + searchValue + "%"
		query = query.Where("name ILIKE ? OR contact_person ILIKE ?", searchQuery, searchQuery)
	}

	query.Count(&filtered)

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

	err := query.Preload("Updater").Find(&partners).Error
	return total, filtered, partners, err
}

func (r *partnerRepository) FindByID(id uint) (entity.Partner, error) {
	var partner entity.Partner
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
	r.db.Model(&entity.Partner{}).Where("id = ?", id).Update("is_active", false)
	// Then soft delete
	return r.db.Delete(&entity.Partner{}, id).Error
}

func (r *partnerRepository) Restore(id uint) error {
	// Restore soft deleted row by setting deleted_at to null and is_active to true
	return r.db.Unscoped().Model(&entity.Partner{}).Where("id = ?", id).Updates(map[string]interface{}{
		"deleted_at": nil,
		"is_active":  true,
	}).Error
}
