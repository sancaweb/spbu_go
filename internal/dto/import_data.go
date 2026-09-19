package dto

// LegacyConnectionInput is intentionally separate from the persisted entity so
// a plaintext password can only exist during the request that uses it.
type LegacyConnectionInput struct {
	Driver       string `form:"driver" json:"driver"`
	Host         string `form:"host" json:"host"`
	Port         string `form:"port" json:"port"`
	DatabaseName string `form:"database_name" json:"database_name"`
	Username     string `form:"username" json:"username"`
	Password     string `form:"password" json:"-"`
	SSLMode      string `form:"ssl_mode" json:"ssl_mode"`
}

type LegacyConnectionView struct {
	Driver       string `json:"driver"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DatabaseName string `json:"database_name"`
	Username     string `json:"username"`
	SSLMode      string `json:"ssl_mode"`
	HasPassword  bool   `json:"has_password"`
}

type LegacyBBMRow struct {
	SourceID      int64   `json:"source_id"`
	Name          string  `json:"name"`
	Margin        float64 `json:"margin"`
	Price         float64 `json:"price"`
	Stock         float64 `json:"stock"`
	RewardPercent float64 `json:"reward_percent"`
	IsActive      bool    `json:"is_active"`
}

type ImportResult struct {
	Read    int `json:"read"`
	Created int `json:"created"`
	Updated int `json:"updated"`
}
