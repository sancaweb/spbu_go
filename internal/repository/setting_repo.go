package repository

import (
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

type settingRepository struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) SettingRepository {
	return &settingRepository{db}
}

func (r *settingRepository) FindByKey(key string) (*entity.Setting, error) {
	var setting entity.Setting
	err := r.db.Where("setting_name = ?", key).First(&setting).Error
	return &setting, err
}

func (r *settingRepository) FindAll() ([]entity.Setting, error) {
	var settings []entity.Setting
	err := r.db.Order("setting_name").Find(&settings).Error
	return settings, err
}

func (r *settingRepository) Upsert(key, value string) error {
	var setting entity.Setting
	result := r.db.Where("setting_name = ?", key).First(&setting)
	if result.Error != nil {
		setting = entity.Setting{SettingName: key, SettingValue: value}
		return r.db.Create(&setting).Error
	}
	setting.SettingValue = value
	return r.db.Save(&setting).Error
}
