package entity

import "time"

// COAMapping maps a transaction type + semantic role to a specific COA account.
// BBMID is nil for non-BBM-specific roles (e.g., debit_kas, debit_piutang).
// BBMID is set for per-BBM roles (e.g., kredit_pendapatan per Pertalite).
type COAMapping struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TransType string    `gorm:"type:varchar(50);not null;index" json:"trans_type"`
	Role      string    `gorm:"type:varchar(50);not null" json:"role"`
	Label     string    `gorm:"type:varchar(100);not null" json:"label"`
	COAID     uint      `gorm:"not null;index" json:"coa_id"`
	COA       *COA      `gorm:"foreignKey:COAID" json:"coa,omitempty"`
	BBMID     *uint     `gorm:"index" json:"bbm_id"`
	BBM       *BBM      `gorm:"foreignKey:BBMID" json:"bbm,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (COAMapping) TableName() string { return "coa_mappings" }
