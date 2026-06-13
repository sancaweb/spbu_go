package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"spbu_go/internal/dto"
	"spbu_go/internal/entity"
	"spbu_go/internal/service"
)

// PiutangHandler menangani HTTP request untuk modul piutang B2B.
type PiutangHandler struct {
	svc        service.PiutangService
	partnerSvc service.PartnerService
	penjSvc    service.PenjualanService
	bbmSvc     service.BBMService
}

func NewPiutangHandler(svc service.PiutangService, partnerSvc service.PartnerService, penjSvc service.PenjualanService, bbmSvc service.BBMService) *PiutangHandler {
	return &PiutangHandler{svc: svc, partnerSvc: partnerSvc, penjSvc: penjSvc, bbmSvc: bbmSvc}
}

// ─── Index — list halaman piutang ─────────────────────────────────────────────

func (h *PiutangHandler) Index(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")

	partners, _ := h.partnerSvc.GetActive()
	bbms, _ := h.bbmSvc.GetActive()

	c.HTML(http.StatusOK, "transaction/piutang/index.html", gin.H{
		"User":       user,
		"Favicon":    favicon,
		"Title":      "Piutang B2B",
		"ActiveMenu": "piutang_data",
		"Partners":   partners,
		"BBMs":       bbms,
	})
}

func (h *PiutangHandler) Rincian(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	partners, _ := h.partnerSvc.GetActive()

	c.HTML(http.StatusOK, "transaction/piutang/rincian.html", gin.H{
		"User":       user,
		"Favicon":    favicon,
		"Title":      "Rincian Piutang",
		"ActiveMenu": "piutang_rincian",
		"Partners":   partners,
	})
}

func (h *PiutangHandler) Rekap(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	c.HTML(http.StatusOK, "transaction/piutang/rekap.html", gin.H{
		"User":       user,
		"Favicon":    favicon,
		"Title":      "Rekap Piutang",
		"ActiveMenu": "piutang_rekap",
	})
}

// ─── Datatable — server-side JSON ─────────────────────────────────────────────

func (h *PiutangHandler) Datatable(c *gin.Context) {
	var req dto.DatatableRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.DatatableResponse{Error: err.Error()})
		return
	}

	total, filtered, rows, err := h.svc.DatatableRows(req)
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

func (h *PiutangHandler) DatatableRincian(c *gin.Context) {
	var req dto.DatatableRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.DatatableResponse{Error: err.Error()})
		return
	}

	total, filtered, rows, err := h.svc.DatatableDetailRows(req)
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

func (h *PiutangHandler) DatatableRekap(c *gin.Context) {
	var req dto.DatatableRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.DatatableResponse{Error: err.Error()})
		return
	}

	total, filtered, rows, err := h.svc.DatatableRekapRows(req)
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

func (h *PiutangHandler) Summary(c *gin.Context) {
	month := strings.TrimSpace(c.Query("bulan"))
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	summary, err := h.svc.SummaryByMonth(month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": summary})
}

func (h *PiutangHandler) RekapGrouped(c *gin.Context) {
	month := strings.TrimSpace(c.Query("bulan"))
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	groups, grand, err := h.svc.GroupedRekapByMonth(month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"bulan":       month,
			"groups":      groups,
			"grand_total": grand,
		},
	})
}

// ─── GetDetail — detail piutang (JSON) ────────────────────────────────────────

func (h *PiutangHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	p, err := h.svc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Piutang tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": p})
}

// ─── Create — simpan piutang baru ─────────────────────────────────────────────

type createPiutangDetailInput struct {
	PenjualanID uint64 `json:"penjualan_id"`
	NoVoucher   string `json:"no_voucher"`
	NoPol       string `json:"no_pol"`
	DriverName  string `json:"driver_name"`
	BBMID       uint   `json:"bbm_id"`
	HargaBBM    int64  `json:"harga_bbm"`
	Margin      int64  `json:"margin"`
	QtyLiter    int64  `json:"qty_liter"`
}

type createPiutangInput struct {
	PenjualanID  uint64                     `json:"penjualan_id" binding:"required"`
	PelangganID  uint                       `json:"pelanggan_id"`
	PartnerID    uint                       `json:"partner_id"`
	TotalTagihan int64                      `json:"total_tagihan"`
	Details      []createPiutangDetailInput `json:"details" binding:"required,min=1"`
}

func (h *PiutangHandler) Create(c *gin.Context) {
	var input createPiutangInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	if input.PelangganID == 0 {
		input.PelangganID = input.PartnerID
	}
	if input.PelangganID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Pelanggan wajib dipilih"})
		return
	}

	userRaw, _ := c.Get("user")
	var createdBy *uint
	if u, ok := userRaw.(*entity.User); ok {
		createdBy = &u.ID
	}

	piutang := entity.TrxPiutang{
		PenjualanID:  input.PenjualanID,
		PelangganID:  input.PelangganID,
		TotalTagihan: input.TotalTagihan,
	}
	for _, d := range input.Details {
		penjID := d.PenjualanID
		if penjID == 0 {
			penjID = input.PenjualanID
		}
		piutang.Details = append(piutang.Details, entity.TrxPiutangDetail{
			PenjualanID: penjID,
			NoVoucher:   d.NoVoucher,
			NoPol:       d.NoPol,
			DriverName:  d.DriverName,
			BBMID:       d.BBMID,
			HargaBBM:    d.HargaBBM,
			Margin:      d.Margin,
			QtyLiter:    d.QtyLiter,
			TotalLine:   d.HargaBBM * d.QtyLiter,
		})
	}

	if err := h.svc.Create(&piutang, createdBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Piutang berhasil disimpan",
		"data":    gin.H{"id_piutang": piutang.IDPiutang},
	})
}

// ─── Lunas — tandai piutang sebagai lunas ─────────────────────────────────────

func (h *PiutangHandler) Lunas(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ID tidak valid"})
		return
	}

	userRaw, _ := c.Get("user")
	var updatedBy *uint
	if u, ok := userRaw.(*entity.User); ok {
		updatedBy = &u.ID
	}

	if err := h.svc.Lunas(id, updatedBy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Piutang berhasil ditandai lunas",
	})
}

// ─── Delete — hapus piutang ───────────────────────────────────────────────────

func (h *PiutangHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ID tidak valid"})
		return
	}

	if err := h.svc.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Piutang berhasil dihapus",
	})
}
