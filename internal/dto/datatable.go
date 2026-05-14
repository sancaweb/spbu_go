package dto

// DatatableRequest standardizes the parameters sent by jQuery DataTables
type DatatableRequest struct {
	Draw   int `form:"draw"`
	Start  int `form:"start"`
	Length int `form:"length"`
	Search struct {
		Value string `form:"value"`
		Regex string `form:"regex"`
	} `form:"search"`
	Order []struct {
		Column int    `form:"column"`
		Dir    string `form:"dir"`
	} `form:"order"`
	Columns []struct {
		Data       string `form:"data"`
		Name       string `form:"name"`
		Searchable string `form:"searchable"`
		Orderable  string `form:"orderable"`
		Search     struct {
			Value string `form:"value"`
			Regex string `form:"regex"`
		} `form:"search"`
	} `form:"columns"`
}

// DatatableResponse standardizes the JSON response expected by jQuery DataTables
type DatatableResponse struct {
	Draw            int         `json:"draw"`
	RecordsTotal    int64       `json:"recordsTotal"`
	RecordsFiltered int64       `json:"recordsFiltered"`
	Data            interface{} `json:"data"`
	Error           string      `json:"error,omitempty"`
}
