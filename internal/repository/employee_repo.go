package repository

import (
	"spbu_go/internal/dto"
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

type KaryawanRepository interface {
	FindAll() ([]entity.Karyawan, error)
	FindActive() ([]entity.Karyawan, error)
	FindInactive() ([]entity.Karyawan, error)
	FindByID(id uint) (entity.Karyawan, error)
	Datatable(req dto.DatatableRequest, isActive bool) (int64, int64, []entity.Karyawan, error)
	Create(karyawan *entity.Karyawan) error
	Update(karyawan *entity.Karyawan) error
	Delete(id uint) error
	Restore(id uint) error
	SetPendapatans(karyawanID uint, ids []uint) error
	SetPotongans(karyawanID uint, ids []uint) error
}

type karyawanRepository struct {
	db *gorm.DB
}

func NewKaryawanRepository(db *gorm.DB) KaryawanRepository {
	return &karyawanRepository{db}
}

func (r *karyawanRepository) FindAll() ([]entity.Karyawan, error) {
	var data []entity.Karyawan
	err := r.db.Preload("Jabatan").Preload("Updater").Order("nama_lengkap ASC").Find(&data).Error
	return data, err
}

func (r *karyawanRepository) FindActive() ([]entity.Karyawan, error) {
	var data []entity.Karyawan
	err := r.db.Where("is_active = ?", true).Preload("Jabatan").Preload("Updater").Order("nama_lengkap ASC").Find(&data).Error
	return data, err
}

func (r *karyawanRepository) FindInactive() ([]entity.Karyawan, error) {
	var data []entity.Karyawan
	err := r.db.Unscoped().Where("is_active = ? OR deleted_at IS NOT NULL", false).Preload("Jabatan").Preload("Updater").Order("nama_lengkap ASC").Find(&data).Error
	return data, err
}

func (r *karyawanRepository) FindByID(id uint) (entity.Karyawan, error) {
	var data entity.Karyawan
	err := r.db.Preload("Jabatan").Preload("Updater").Preload("Pendapatans").Preload("Potongans").First(&data, id).Error
	return data, err
}

func (r *karyawanRepository) Datatable(req dto.DatatableRequest, isActive bool) (int64, int64, []entity.Karyawan, error) {
	var data []entity.Karyawan
	var total, filtered int64

	query := r.db.Model(&entity.Karyawan{})
	if isActive {
		query = query.Where("is_active = ?", true)
	} else {
		query = r.db.Unscoped().Model(&entity.Karyawan{}).Where("is_active = ? OR deleted_at IS NOT NULL", false)
	}

	query.Count(&total)

	if req.Search.Value != "" {
		s := "%" + req.Search.Value + "%"
		query = query.Where("nik ILIKE ? OR nama_lengkap ILIKE ? OR no_hp ILIKE ?", s, s, s)
	}

	query.Count(&filtered)

	// Ordering
	colMap := map[int]string{
		0: "nik",
		1: "nama_lengkap",
		2: "no_hp",
	}
	orderCol := "nama_lengkap"
	orderDir := "asc"
	if len(req.Order) > 0 {
		if col, ok := colMap[req.Order[0].Column]; ok {
			orderCol = col
		}
		if req.Order[0].Dir == "desc" {
			orderDir = "desc"
		}
	}
	query = query.Order(orderCol + " " + orderDir)

	if req.Length > 0 {
		query = query.Limit(req.Length).Offset(req.Start)
	}

	err := query.Preload("Jabatan").Find(&data).Error
	return total, filtered, data, err
}

func (r *karyawanRepository) Create(karyawan *entity.Karyawan) error {
	return r.db.Omit("Jabatan", "Updater").Create(karyawan).Error
}

func (r *karyawanRepository) Update(karyawan *entity.Karyawan) error {
	return r.db.Omit("Jabatan", "Updater").Save(karyawan).Error
}

func (r *karyawanRepository) SetPendapatans(karyawanID uint, ids []uint) error {
	pendapatans := make([]entity.Pendapatan, len(ids))
	for i, id := range ids {
		pendapatans[i] = entity.Pendapatan{ID: id}
	}
	k := entity.Karyawan{}
	k.ID = karyawanID
	return r.db.Model(&k).Association("Pendapatans").Replace(pendapatans)
}

func (r *karyawanRepository) SetPotongans(karyawanID uint, ids []uint) error {
	potongans := make([]entity.Potongan, len(ids))
	for i, id := range ids {
		potongans[i] = entity.Potongan{ID: id}
	}
	k := entity.Karyawan{}
	k.ID = karyawanID
	return r.db.Model(&k).Association("Potongans").Replace(potongans)
}

func (r *karyawanRepository) Delete(id uint) error {
	r.db.Model(&entity.Karyawan{}).Where("id = ?", id).Update("is_active", false)
	return r.db.Delete(&entity.Karyawan{}, id).Error
}

func (r *karyawanRepository) Restore(id uint) error {
	return r.db.Unscoped().Model(&entity.Karyawan{}).Where("id = ?", id).Updates(map[string]interface{}{
		"deleted_at": nil,
		"is_active":  true,
	}).Error
}
