package dto

import "spbu_go/internal/entity"

type PartnerDTRow struct {
	entity.Partner
	CanDeletePermanent bool `gorm:"column:can_delete_permanent;->" json:"can_delete_permanent"`
}
