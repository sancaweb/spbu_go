package repository

import (
	"spbu_go/internal/entity"

	"gorm.io/gorm"
)

type legacyDBConnectionRepository struct {
	db *gorm.DB
}

func NewLegacyDBConnectionRepository(db *gorm.DB) LegacyDBConnectionRepository {
	return &legacyDBConnectionRepository{db: db}
}

func (r *legacyDBConnectionRepository) Find() (*entity.LegacyDBConnection, error) {
	var connection entity.LegacyDBConnection
	err := r.db.Order("id ASC").First(&connection).Error
	return &connection, err
}

func (r *legacyDBConnectionRepository) Save(connection *entity.LegacyDBConnection) error {
	return r.db.Save(connection).Error
}
