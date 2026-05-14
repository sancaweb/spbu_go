package handler

import (
	"encoding/json"
	"html/template"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"spbu_go/internal/dto"
	"spbu_go/internal/entity"
	"spbu_go/internal/service"

	"github.com/gin-gonic/gin"
)

type PenebusanHandler struct {
	service        service.PenebusanService
	bbmService     service.BBMService
	walletService  service.WalletService
	settingService service.SettingService
}

func NewPenebusanHandler(
	service service.PenebusanService,
	bbmService service.BBMService,
	walletService service.WalletService,
	settingService service.SettingService,
) *PenebusanHandler {
	return &PenebusanHandler{
		service:        service,
		bbmService:     bbmService,
		walletService:  walletService,
		settingService: settingService,
	}
}

// Index — renders the page shell; dropdown data for form loaded server-side
func (h *PenebusanHandler) Index(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	bbms, _ := h.bbmService.GetActive()
	wallets, _ := h.walletService.GetAll()
	decimalPlaces := h.settingService.GetInt("stock_decimal_places", 0)
	taxMultiplier := "0.0025"
	if taxSetting, err := h.settingService.Get("penebusan_tax_multiplier"); err == nil {
		taxSetting = strings.TrimSpace(taxSetting)
		if taxSetting != "" {
			taxMultiplier = taxSetting
		}
	}

	// Default wallet for auto-select in create form
	var defaultWalletID uint
	for _, w := range wallets {
		if w.IsDefault {
			defaultWalletID = w.ID
			break
		}
	}

	// Build a slim JSON array for the JS BBM map (auto-fill harga)
	type bbmJSON struct {
		ID     uint    `json:"id"`
		Name   string  `json:"name"`
		Price  float64 `json:"price"`
		Margin float64 `json:"margin"`
	}
	bbmList := make([]bbmJSON, 0, len(bbms))
	for _, b := range bbms {
		bbmList = append(bbmList, bbmJSON{ID: b.ID, Name: b.Name, Price: b.Price, Margin: b.Margin})
	}
	bbmsRaw, _ := json.Marshal(bbmList)

	c.HTML(http.StatusOK, "transaction/penebusan/index.html", gin.H{
		"User":                   user,
		"Favicon":                favicon,
		"Title":                  "Penebusan BBM",
		"ActiveMenu":             "trans_penebusan",
		"BBMs":                   bbms,
		"Wallets":                wallets,
		"DefaultWalletID":        defaultWalletID,
		"StockDecimalPlaces":     decimalPlaces,
		"PenebusanTaxMultiplier": taxMultiplier,
		"BBMsJSON":               template.JS(bbmsRaw),
	})
}

// Datatable — server-side jQuery DataTables endpoint
func (h *PenebusanHandler) Datatable(c *gin.Context) {
	var req dto.DatatableRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	total, filtered, list, err := h.service.Datatable(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.DatatableResponse{
			Draw:  req.Draw,
			Error: "Gagal mengambil data dari server",
		})
		return
	}

	c.JSON(http.StatusOK, dto.DatatableResponse{
		Draw:            req.Draw,
		RecordsTotal:    total,
		RecordsFiltered: filtered,
		Data:            buildHeadersJSON(list),
	})
}

// Create — simpan transaksi penebusan (DR/CO)
func (h *PenebusanHandler) Create(c *gin.Context) {
	var req createPenebusanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Payload tidak valid"})
		return
	}

	var existing *entity.TrxPenebusan
	if req.ID > 0 {
		var err error
		existing, err = h.service.GetByID(req.ID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "Data penebusan tidak ditemukan"})
			return
		}
	}

	tglPenebusan, err := time.Parse("2006-01-02", strings.TrimSpace(req.TglPenebusan))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Tanggal penebusan tidak valid"})
		return
	}

	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Rincian BBM wajib diisi"})
		return
	}

	status := string(entity.PenebusanDraft)
	if strings.EqualFold(strings.TrimSpace(req.Action), "complete") {
		status = string(entity.PenebusanComplete)
	} else if existing != nil {
		// Save pada mode edit harus mempertahankan status saat ini
		status = existing.Status
	}

	multiplier := 0.0025
	if taxSetting, err := h.settingService.Get("penebusan_tax_multiplier"); err == nil {
		taxSetting = strings.TrimSpace(taxSetting)
		if strings.Contains(taxSetting, ",") {
			taxSetting = strings.ReplaceAll(taxSetting, ".", "")
			taxSetting = strings.Replace(taxSetting, ",", ".", 1)
		}
		if v, err := strconv.ParseFloat(taxSetting, 64); err == nil {
			multiplier = v
		}
	}

	trx := entity.TrxPenebusan{
		NoSO:         nil,
		TglPenebusan: tglPenebusan,
		Status:       status,
		Catatan:      strings.TrimSpace(req.Catatan),
		AdmBank:      req.AdmBank,
	}

	if strings.TrimSpace(req.NoSO) != "" {
		noSO := strings.TrimSpace(req.NoSO)
		trx.NoSO = &noSO
	}
	if req.WalletID != nil && *req.WalletID > 0 {
		trx.WalletID = req.WalletID
	}
	if strings.TrimSpace(req.TglBayar) != "" {
		tglBayar, err := time.Parse("2006-01-02", strings.TrimSpace(req.TglBayar))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Tanggal bayar tidak valid"})
			return
		}
		trx.TglBayar = &tglBayar
	} else {
		tglBayar := time.Now()
		trx.TglBayar = &tglBayar
	}

	for _, it := range req.Items {
		if it.BBMID == 0 || it.JmlLiter <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Rincian BBM tidak valid"})
			return
		}

		subtotal := it.HargaDasar * it.JmlLiter
		ppn := int64(math.Round((100.0 / 116.0) * float64(it.JmlLiter) * float64(it.HargaJual) * multiplier))
		total := subtotal + ppn

		detail := entity.TrxPenebusanDetail{
			BBMID:      it.BBMID,
			JmlLiter:   it.JmlLiter,
			HargaDasar: it.HargaDasar,
			HargaJual:  it.HargaJual,
			Margin:     it.HargaJual - it.HargaDasar,
			PPNPersen:  11,
			PPNRp:      ppn,
			Subtotal:   subtotal,
			Total:      total,
		}

		trx.Subtotal += subtotal
		trx.TotalPPN += ppn
		trx.Details = append(trx.Details, detail)
	}
	trx.TotalBayar = trx.Subtotal + trx.TotalPPN + trx.AdmBank

	if req.ID > 0 {
		trx.ID = existing.ID
		trx.NoPenebusan = existing.NoPenebusan
		if err := h.service.Update(&trx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal memperbarui penebusan"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":       true,
			"message":      "Penebusan berhasil diperbarui",
			"id":           trx.ID,
			"no_penebusan": trx.NoPenebusan,
			"doc_status":   trx.Status,
		})
		return
	}

	if err := h.service.Create(&trx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal menyimpan penebusan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       true,
		"message":      "Penebusan berhasil disimpan",
		"id":           trx.ID,
		"no_penebusan": trx.NoPenebusan,
		"doc_status":   trx.Status,
	})
}

// Delete — soft delete transaksi penebusan
func (h *PenebusanHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal menghapus penebusan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Penebusan berhasil dihapus"})
}

// GetDetail — returns header + detail lines for a single penebusan (AJAX)
func (h *PenebusanHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	p, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data tidak ditemukan"})
		return
	}

	noSO := ""
	if p.NoSO != nil {
		noSO = *p.NoSO
	}
	walletName := ""
	if p.Wallet != nil {
		walletName = p.Wallet.WalletName
	}

	header := headerJSON{
		ID:           p.ID,
		NoPenebusan:  p.NoPenebusan,
		NoSO:         noSO,
		TglPenebusan: p.TglPenebusan.Format("2006-01-02"),
		TglBayar:     formatDatePtr(p.TglBayar),
		WalletID:     p.WalletID,
		AdmBank:      p.AdmBank,
		Catatan:      p.Catatan,
		Subtotal:     p.Subtotal,
		TotalPPN:     p.TotalPPN,
		TotalBayar:   p.TotalBayar,
		Status:       p.Status,
		StatusLabel:  p.GetStatusLabel(),
		WalletName:   walletName,
	}

	details := make([]detailJSON, 0, len(p.Details))
	for _, d := range p.Details {
		bbmName := ""
		if d.BBM != nil {
			bbmName = d.BBM.Name
		}
		details = append(details, detailJSON{
			ID:         d.ID,
			BBMID:      d.BBMID,
			BBMName:    bbmName,
			JmlLiter:   d.JmlLiter,
			HargaDasar: d.HargaDasar,
			HargaJual:  d.HargaJual,
			Margin:     d.Margin,
			PPNPersen:  d.PPNPersen,
			PPNRp:      d.PPNRp,
			Subtotal:   d.Subtotal,
			Total:      d.Total,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"header":  header,
		"details": details,
	})
}

// ─── JSON DTOs ────────────────────────────────────────────────────────────────

type headerJSON struct {
	ID           uint64 `json:"id"`
	NoPenebusan  string `json:"no_penebusan"`
	NoSO         string `json:"no_so"`
	TglPenebusan string `json:"tgl_penebusan"`
	TglBayar     string `json:"tgl_bayar"`
	WalletID     *uint  `json:"wallet_id"`
	AdmBank      int64  `json:"adm_bank"`
	Catatan      string `json:"catatan"`
	Subtotal     int64  `json:"subtotal"`
	TotalPPN     int64  `json:"total_ppn"`
	TotalBayar   int64  `json:"total_bayar"`
	Status       string `json:"status"`
	StatusLabel  string `json:"status_label"`
	WalletName   string `json:"wallet_name"`
}

type detailJSON struct {
	ID         uint64  `json:"id"`
	BBMID      uint    `json:"bbm_id"`
	BBMName    string  `json:"bbm_name"`
	JmlLiter   int64   `json:"jml_liter"`
	HargaDasar int64   `json:"harga_dasar"`
	HargaJual  int64   `json:"harga_jual"`
	Margin     int64   `json:"margin"`
	PPNPersen  float64 `json:"ppn_persen"`
	PPNRp      int64   `json:"ppn_rp"`
	Subtotal   int64   `json:"subtotal"`
	Total      int64   `json:"total"`
}

type createPenebusanItemRequest struct {
	BBMID      uint  `json:"bbm_id"`
	JmlLiter   int64 `json:"jml_liter"`
	HargaDasar int64 `json:"harga_dasar"`
	HargaJual  int64 `json:"harga_jual"`
}

type createPenebusanRequest struct {
	ID           uint64                       `json:"id"`
	Action       string                       `json:"action"`
	NoSO         string                       `json:"no_so"`
	TglPenebusan string                       `json:"tgl_penebusan"`
	TglBayar     string                       `json:"tgl_bayar"`
	WalletID     *uint                        `json:"wallet_id"`
	AdmBank      int64                        `json:"adm_bank"`
	Catatan      string                       `json:"catatan"`
	Items        []createPenebusanItemRequest `json:"items"`
}

func buildHeadersJSON(list []entity.TrxPenebusan) []headerJSON {
	out := make([]headerJSON, 0, len(list))
	for _, p := range list {
		noSO := ""
		if p.NoSO != nil {
			noSO = *p.NoSO
		}
		walletName := ""
		if p.Wallet != nil {
			walletName = p.Wallet.WalletName
		}
		out = append(out, headerJSON{
			ID:           p.ID,
			NoPenebusan:  p.NoPenebusan,
			NoSO:         noSO,
			TglPenebusan: p.TglPenebusan.Format("2006-01-02"),
			TglBayar:     formatDatePtr(p.TglBayar),
			WalletID:     p.WalletID,
			AdmBank:      p.AdmBank,
			Catatan:      p.Catatan,
			Subtotal:     p.Subtotal,
			TotalPPN:     p.TotalPPN,
			TotalBayar:   p.TotalBayar,
			Status:       p.Status,
			StatusLabel:  p.GetStatusLabel(),
			WalletName:   walletName,
		})
	}
	return out
}

func formatDatePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}
