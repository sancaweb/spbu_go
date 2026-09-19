package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"spbu_go/internal/entity"
	"spbu_go/internal/helper"
	"spbu_go/internal/repository"
	"spbu_go/internal/service"

	"github.com/gin-gonic/gin"
)

type PenyusutanHandler struct {
	svc service.PenyusutanService
}

func NewPenyusutanHandler(svc service.PenyusutanService) *PenyusutanHandler {
	return &PenyusutanHandler{svc: svc}
}

func (h *PenyusutanHandler) Index(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")

	mode := strings.ToLower(strings.TrimSpace(c.Query("mode")))
	date := strings.TrimSpace(c.Query("date"))
	month := strings.TrimSpace(c.Query("month"))

	hasReport := false
	reportGroups := []repository.PenyusutanReportGroup{}
	shiftGroups := []repository.PenyusutanShiftGroup{}
	periodLabel := "-"

	switch mode {
	case "daily":
		hasReport = true
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}
		if parsed, err := time.Parse("2006-01-02", date); err == nil {
			periodLabel = parsed.Format("02 Jan 2006")
		}
		if groups, err := h.svc.GetReport(mode, date, ""); err == nil {
			reportGroups = groups
		}
		if groups, err := h.svc.GetShiftReport(date); err == nil {
			shiftGroups = groups
		}
	case "monthly":
		hasReport = true
		if month == "" {
			month = time.Now().Format("2006-01")
		}
		if parsed, err := time.Parse("2006-01", month); err == nil {
			periodLabel = parsed.Format("Jan 2006")
		}
		if groups, err := h.svc.GetReport(mode, "", month); err == nil {
			reportGroups = groups
		}
	}

	c.HTML(http.StatusOK, "transaction/penyusutan/index.html", gin.H{
		"User":         user,
		"Favicon":      favicon,
		"Title":        "Laporan Data Penyusutan",
		"ActiveMenu":   "trans_penyusutan",
		"Mode":         mode,
		"Date":         date,
		"Month":        month,
		"HasReport":    hasReport,
		"ReportGroups": reportGroups,
		"ShiftGroups":  shiftGroups,
		"PeriodLabel":  periodLabel,
	})
}

// UpdateEndstockActual — endpoint untuk update stok akhir aktual (penghitungan fisik tangki).
// POST /transaction/penyusutan/:id/actual
// Body JSON: { "endstock_actual": 1234.56 }
func (h *PenyusutanHandler) UpdateEndstockActual(c *gin.Context) {
	user, _ := c.Get("user")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ID tidak valid"})
		return
	}

	var req struct {
		EndstockActual float64 `json:"endstock_actual"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Format request tidak valid: " + err.Error()})
		return
	}

	var updatedBy *uint
	if u, ok := user.(*entity.User); ok && u != nil {
		uid := u.ID
		updatedBy = &uid
	}

	if err := h.svc.UpdateEndstockActual(id, req.EndstockActual, updatedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal menyimpan: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Stok aktual berhasil diperbarui"})
}

// ViewEdit — Menampilkan halaman form edit penyusutan per shift & BBM.
// Route: GET /transaction/penyusutan/view_edit/:token
func (h *PenyusutanHandler) ViewEdit(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	token := c.Param("token")

	bbmID, shiftID, penjualanID, err := helper.DecodePenyusutanToken(token)
	if err != nil {
		c.HTML(http.StatusBadRequest, "transaction/penyusutan/index.html", gin.H{
			"User":       user,
			"Favicon":    favicon,
			"Title":      "Laporan Data Penyusutan",
			"ActiveMenu": "trans_penyusutan",
			"Error":      "Token URL tidak valid atau kadaluarsa",
		})
		return
	}

	editData, err := h.svc.GetEditData(bbmID, shiftID, penjualanID)
	if err != nil {
		c.HTML(http.StatusNotFound, "transaction/penyusutan/index.html", gin.H{
			"User":       user,
			"Favicon":    favicon,
			"Title":      "Laporan Data Penyusutan",
			"ActiveMenu": "trans_penyusutan",
			"Error":      "Data penyusutan tidak ditemukan: " + err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "transaction/penyusutan/edit.html", gin.H{
		"User":       user,
		"Favicon":    favicon,
		"Title":      "Edit Data Penyusutan",
		"ActiveMenu": "trans_penyusutan",
		"Data":       editData,
	})
}

// SaveEdit — Memproses form submit edit penyusutan per shift & BBM.
// Route: POST /transaction/penyusutan/edit
func (h *PenyusutanHandler) SaveEdit(c *gin.Context) {
	user, _ := c.Get("user")
	token := c.PostForm("token")

	bbmID, shiftID, penjualanID, err := helper.DecodePenyusutanToken(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Token tidak valid"})
		return
	}

	stokAwalStr := strings.ReplaceAll(c.PostForm("stok_awal"), ",", "")
	stokAktualStr := strings.ReplaceAll(c.PostForm("stok_akhir_aktual"), ",", "")

	stokAwal, _ := strconv.ParseFloat(stokAwalStr, 64)
	stokAktual, _ := strconv.ParseFloat(stokAktualStr, 64)

	var updatedBy *uint
	if u, ok := user.(*entity.User); ok && u != nil {
		uid := u.ID
		updatedBy = &uid
	}

	if err := h.svc.SaveEditData(penjualanID, bbmID, shiftID, stokAwal, stokAktual, updatedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal menyimpan penyusutan: " + err.Error()})
		return
	}

	editData, err := h.svc.GetEditData(bbmID, shiftID, penjualanID)
	redirectDate := time.Now().Format("2006-01-02")
	if err == nil && editData != nil && editData.WaktuDate != "" {
		redirectDate = editData.WaktuDate
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"message":  "Data penyusutan berhasil disimpan",
		"redirect": "/transaction/penyusutan?mode=daily&date=" + redirectDate,
	})
}
