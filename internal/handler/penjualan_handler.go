package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"spbu_go/internal/dto"
	"spbu_go/internal/entity"
	"spbu_go/internal/service"

	"github.com/gin-gonic/gin"
)

// ─── Request DTOs ─────────────────────────────────────────────────────────────

type penjualanDetailReq struct {
	NozzleID         uint  `json:"nozzle_id"`
	TiangID          uint  `json:"tiang_id"`
	BBMID            uint  `json:"bbm_id"`
	BBMPrice         int64 `json:"bbm_price"`
	Margin           int64 `json:"margin"`
	TotalisatorAwal  int64 `json:"totalisator_awal"`
	TotalisatorAkhir int64 `json:"totalisator_akhir"`
}

type penjualanRequest struct {
	ShiftID         uint                 `json:"shift_id"`
	WaktuMulai      string               `json:"waktu_mulai"` // "2006-01-02T15:04"
	WaktuAkhir      string               `json:"waktu_akhir"`
	TotalPenerimaan int64                `json:"total_penerimaan"`
	AktualUang      int64                `json:"aktual_uang"`
	Details         []penjualanDetailReq `json:"details"`
}

// ─── Form data helper (untuk template) ───────────────────────────────────────

type NozzleFormRow struct {
	NozzleID         uint   `json:"nozzle_id"`
	TiangID          uint   `json:"tiang_id"`
	TiangName        string `json:"tiang_name"`
	Description      string `json:"description"`
	BBMID            uint   `json:"bbm_id"`
	BBMName          string `json:"bbm_name"`
	BBMPrice         int64  `json:"bbm_price"`
	Margin           int64  `json:"margin"`
	TotalisatorAwal  int64  `json:"totalisator_awal"`
	TotalisatorAkhir int64  `json:"totalisator_akhir"`
	JmlLiter         int64  `json:"jml_liter"`
	JmlRupiah        int64  `json:"jml_rupiah"`
}

// ─── Handler struct ───────────────────────────────────────────────────────────

type PenjualanHandler struct {
	penjualanSvc service.PenjualanService
	tiangSvc     service.TiangService
	shiftSvc     service.ShiftService
	settingSvc   service.SettingService
}

func NewPenjualanHandler(
	penjualanSvc service.PenjualanService,
	tiangSvc service.TiangService,
	shiftSvc service.ShiftService,
	settingSvc service.SettingService,
) *PenjualanHandler {
	return &PenjualanHandler{
		penjualanSvc: penjualanSvc,
		tiangSvc:     tiangSvc,
		shiftSvc:     shiftSvc,
		settingSvc:   settingSvc,
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// buildNozzleRows membangun daftar nozzle untuk form, pre-filled dari detail (edit mode).
func (h *PenjualanHandler) buildNozzleRows(existingDetails []entity.TrxPenjualanDetail) (template.JS, error) {
	tiangs, err := h.tiangSvc.GetAll()
	if err != nil {
		return "", err
	}

	// Build lookup dari existing details (nozzle_id → detail)
	detailMap := map[uint]*entity.TrxPenjualanDetail{}
	for i := range existingDetails {
		detailMap[existingDetails[i].NozzleID] = &existingDetails[i]
	}

	var rows []NozzleFormRow
	for _, t := range tiangs {
		for _, n := range t.Nozzles {
			if !n.IsActive {
				continue
			}
			bbmName := ""
			var bbmPrice, margin int64
			if n.BBM != nil {
				bbmName = n.BBM.Name
				bbmPrice = int64(n.BBM.Price)
				margin = int64(n.BBM.Margin)
			}
			row := NozzleFormRow{
				NozzleID:    n.ID,
				TiangID:     t.ID,
				TiangName:   t.Name,
				Description: n.Description,
				BBMID:       n.BBMID,
				BBMName:     bbmName,
				BBMPrice:    bbmPrice,
				Margin:      margin,
			}
			// Pre-fill dari existing detail jika ada
			if d, ok := detailMap[n.ID]; ok {
				row.TotalisatorAwal = d.TotalisatorAwal
				row.TotalisatorAkhir = d.TotalisatorAkhir
				row.JmlLiter = d.JmlLiter
				row.JmlRupiah = d.JmlRupiah
				// Gunakan snapshot harga dari detail jika ada
				if d.BBMPrice > 0 {
					row.BBMPrice = d.BBMPrice
				}
				if d.Margin > 0 {
					row.Margin = d.Margin
				}
			}
			rows = append(rows, row)
		}
	}

	b, err := json.Marshal(rows)
	if err != nil {
		return "", err
	}
	return template.JS(b), nil
}

// parsePenjualanRequest validates + converts request to TrxPenjualan entity.
func parsePenjualanRequest(req penjualanRequest, userID *uint) (*entity.TrxPenjualan, error) {
	const layout = "2006-01-02T15:04"
	waktuMulai, err := time.ParseInLocation(layout, req.WaktuMulai, time.Local)
	if err != nil {
		return nil, fmt.Errorf("format waktu mulai tidak valid (%s)", req.WaktuMulai)
	}
	waktuAkhir, err := time.ParseInLocation(layout, req.WaktuAkhir, time.Local)
	if err != nil {
		return nil, fmt.Errorf("format waktu akhir tidak valid (%s)", req.WaktuAkhir)
	}
	if req.ShiftID == 0 {
		return nil, fmt.Errorf("shift harus dipilih")
	}
	if len(req.Details) == 0 {
		return nil, fmt.Errorf("minimal 1 baris nozzle harus memiliki data")
	}

	p := &entity.TrxPenjualan{
		ShiftID:         req.ShiftID,
		WaktuMulai:      waktuMulai,
		WaktuAkhir:      waktuAkhir,
		TotalPenerimaan: req.TotalPenerimaan,
		AktualUang:      req.AktualUang,
		UpdatedBy:       userID,
	}

	hasData := false
	for _, dr := range req.Details {
		jmlLiter := dr.TotalisatorAkhir - dr.TotalisatorAwal
		if jmlLiter <= 0 {
			continue // skip baris tanpa penjualan
		}
		hasData = true
		jmlRupiah := jmlLiter * dr.BBMPrice
		p.TotalRpTotalisator += jmlRupiah
		p.Details = append(p.Details, entity.TrxPenjualanDetail{
			TiangID:          dr.TiangID,
			NozzleID:         dr.NozzleID,
			BBMID:            dr.BBMID,
			BBMPrice:         dr.BBMPrice,
			Margin:           dr.Margin,
			TotalisatorAwal:  dr.TotalisatorAwal,
			TotalisatorAkhir: dr.TotalisatorAkhir,
			JmlLiter:         jmlLiter,
			JmlRupiah:        jmlRupiah,
		})
	}
	if !hasData {
		return nil, fmt.Errorf("tidak ada nozzle dengan data penjualan (totalisator akhir harus > awal)")
	}

	p.Selisih = p.AktualUang - p.TotalRpTotalisator
	return p, nil
}

// ─── Handlers ────────────────────────────────────────────────────────────────

// Index — daftar transaksi penjualan BBM (datatable).
func (h *PenjualanHandler) Index(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	c.HTML(http.StatusOK, "transaction/penjualan/index.html", gin.H{
		"User":       user,
		"Favicon":    favicon,
		"Title":      "Penjualan BBM",
		"ActiveMenu": "trans_penjualan",
	})
}

// Datatable — server-side datatable endpoint.
func (h *PenjualanHandler) Datatable(c *gin.Context) {
	var req dto.DatatableRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.DatatableResponse{Error: err.Error()})
		return
	}
	total, filtered, rows, err := h.penjualanSvc.Datatable(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.DatatableResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.DatatableResponse{
		Draw:            req.Draw,
		RecordsTotal:    total,
		RecordsFiltered: filtered,
		Data:            rows,
	})
}

// FormCreate — halaman form buat penjualan baru.
func (h *PenjualanHandler) FormCreate(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	decimalPlaces := h.settingSvc.GetInt("stock_decimal_places", 0)

	nozzleRowsJSON, err := h.buildNozzleRows(nil)
	if err != nil {
		log.Printf("[penjualan] FormCreate: buildNozzleRows error: %v", err)
	}

	shifts, _ := h.shiftSvc.GetAll()

	c.HTML(http.StatusOK, "transaction/penjualan/form.html", gin.H{
		"User":          user,
		"Favicon":       favicon,
		"Title":         "Buat Penjualan BBM",
		"ActiveMenu":    "trans_penjualan",
		"IsEdit":        false,
		"Shifts":        shifts,
		"NozzleRows":    nozzleRowsJSON,
		"DecimalPlaces": decimalPlaces,
		"Penjualan":     nil,
	})
}

// FormEdit — halaman form edit penjualan.
func (h *PenjualanHandler) FormEdit(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	decimalPlaces := h.settingSvc.GetInt("stock_decimal_places", 0)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"Error": "ID tidak valid"})
		return
	}

	p, err := h.penjualanSvc.GetByID(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"Error": "Data penjualan tidak ditemukan"})
		return
	}

	nozzleRowsJSON, err := h.buildNozzleRows(p.Details)
	if err != nil {
		log.Printf("[penjualan] FormEdit: buildNozzleRows error: %v", err)
	}

	shifts, _ := h.shiftSvc.GetAll()

	c.HTML(http.StatusOK, "transaction/penjualan/form.html", gin.H{
		"User":          user,
		"Favicon":       favicon,
		"Title":         "Edit Penjualan BBM",
		"ActiveMenu":    "trans_penjualan",
		"IsEdit":        true,
		"Shifts":        shifts,
		"NozzleRows":    nozzleRowsJSON,
		"DecimalPlaces": decimalPlaces,
		"Penjualan":     p,
	})
}

// GetDetail — JSON detail satu penjualan.
func (h *PenjualanHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	p, err := h.penjualanSvc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, p)
}

// Create — simpan penjualan baru (JSON POST).
func (h *PenjualanHandler) Create(c *gin.Context) {
	user, _ := c.Get("user")
	var userID *uint
	if u, ok := user.(*entity.User); ok && u != nil {
		uid := u.ID
		userID = &uid
	}

	var req penjualanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Format request tidak valid: " + err.Error()})
		return
	}

	p, err := parsePenjualanRequest(req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	p.CreatedBy = userID

	if err := h.penjualanSvc.Create(p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal menyimpan: " + err.Error()})
		return
	}

	// Post journal (non-blocking: error hanya di-log)
	if jErr := h.penjualanSvc.PostJournal(p, userID); jErr != nil {
		log.Printf("[penjualan] Create %d: PostJournal error: %v", p.ID, jErr)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Penjualan %s berhasil disimpan", p.NoPenjualan),
		"id":      p.ID,
	})
}

// Update — edit penjualan (JSON POST).
func (h *PenjualanHandler) Update(c *gin.Context) {
	user, _ := c.Get("user")
	var userID *uint
	if u, ok := user.(*entity.User); ok && u != nil {
		uid := u.ID
		userID = &uid
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ID tidak valid"})
		return
	}

	existing, err := h.penjualanSvc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Data tidak ditemukan"})
		return
	}

	var req penjualanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Format request tidak valid: " + err.Error()})
		return
	}

	p, err := parsePenjualanRequest(req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// Pertahankan field yang tidak berubah
	p.ID = existing.ID
	p.NoPenjualan = existing.NoPenjualan
	p.CreatedBy = existing.CreatedBy
	p.Created = existing.Created

	if err := h.penjualanSvc.Update(p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengupdate: " + err.Error()})
		return
	}

	// Re-post journal (idempotent: reverse lama + post baru)
	if jErr := h.penjualanSvc.PostJournal(p, userID); jErr != nil {
		log.Printf("[penjualan] Update %d: PostJournal error: %v", p.ID, jErr)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Penjualan %s berhasil diupdate", p.NoPenjualan),
		"id":      p.ID,
	})
}

// Delete — hapus penjualan (JSON POST).
func (h *PenjualanHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ID tidak valid"})
		return
	}

	// Reverse journal sebelum delete
	if jErr := h.penjualanSvc.ReverseJournal(id); jErr != nil {
		log.Printf("[penjualan] Delete %d: ReverseJournal error: %v", id, jErr)
	}

	if err := h.penjualanSvc.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal menghapus: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Data penjualan berhasil dihapus"})
}
