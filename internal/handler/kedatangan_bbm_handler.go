package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"spbu_go/internal/dto"
	"spbu_go/internal/entity"
	"spbu_go/internal/service"

	"github.com/gin-gonic/gin"
)

type KedatanganBBMHandler struct {
	kedatanganService service.KedatanganBBMService
	shiftService      service.ShiftService
}

func NewKedatanganBBMHandler(
	kedatanganService service.KedatanganBBMService,
	shiftService service.ShiftService,
) *KedatanganBBMHandler {
	return &KedatanganBBMHandler{kedatanganService, shiftService}
}

// Index — halaman list kedatangan BBM
func (h *KedatanganBBMHandler) Index(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	shifts, _ := h.shiftService.GetAll()

	c.HTML(http.StatusOK, "transaction/kedatangan-bbm/index.html", gin.H{
		"Title":      "Kedatangan BBM",
		"ActiveMenu": "trx_kedatangan_bbm",
		"User":       user,
		"Favicon":    favicon,
		"Shifts":     shifts,
	})
}

// Datatable — server-side DataTables endpoint
func (h *KedatanganBBMHandler) Datatable(c *gin.Context) {
	var req dto.DatatableRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request tidak valid"})
		return
	}

	// Fallback: manually parse order info if ShouldBind didn't populate it
	// (Gin form binding sometimes doesn't handle nested arrays of structs reliably)
	if len(req.Order) == 0 {
		colStr := c.PostForm("order[0][column]")
		dir := c.PostForm("order[0][dir]")
		if colStr != "" {
			col, _ := strconv.Atoi(colStr)
			req.Order = append(req.Order, struct {
				Column int    `form:"column"`
				Dir    string `form:"dir"`
			}{Column: col, Dir: dir})
		}
	}

	total, filtered, rows, err := h.kedatanganService.Datatable(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.DatatableResponse{
			Draw:  req.Draw,
			Error: "Gagal memuat data",
		})
		return
	}

	c.JSON(http.StatusOK, dto.DatatableResponse{
		Draw:            req.Draw,
		RecordsTotal:    total,
		RecordsFiltered: filtered,
		Data:            rows,
	})
}

// GetSOOptions — JSON list penebusan CO dengan sisa liter > 0 (untuk dropdown No SO)
func (h *KedatanganBBMHandler) GetSOOptions(c *gin.Context) {
	rows, err := h.kedatanganService.GetSOSisa()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal memuat data SO"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "data": rows})
}

// GetBBMByPenebusan — JSON list jenis BBM dalam suatu penebusan yang masih ada sisa
func (h *KedatanganBBMHandler) GetBBMByPenebusan(c *gin.Context) {
	penebusanID, err := strconv.ParseUint(c.Param("penebusan_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID penebusan tidak valid"})
		return
	}

	rows, err := h.kedatanganService.GetBBMSisaByPenebusan(penebusanID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal memuat data BBM"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "data": rows})
}

// GetOne — JSON satu record kedatangan (untuk form edit)
func (h *KedatanganBBMHandler) GetOne(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	k, err := h.kedatanganService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "Data tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "data": k})
}

// Create — simpan data kedatangan baru
func (h *KedatanganBBMHandler) Create(c *gin.Context) {
	userRaw, _ := c.Get("user")
	userObj, _ := userRaw.(*entity.User)

	penebusanID, err := strconv.ParseUint(c.PostForm("penebusan_id"), 10, 64)
	if err != nil || penebusanID == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": false, "message": "No SO wajib dipilih"})
		return
	}

	penebusanDetailID, err := strconv.ParseUint(c.PostForm("penebusan_detail_id"), 10, 64)
	if err != nil || penebusanDetailID == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": false, "message": "Jenis BBM wajib dipilih"})
		return
	}

	noLO := strings.TrimSpace(c.PostForm("no_lo"))
	if noLO == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": false, "message": "No LO wajib diisi"})
		return
	}

	tglStr := strings.TrimSpace(c.PostForm("tgl_kedatangan"))
	tglKedatangan, err := time.Parse("2006-01-02 15:04", tglStr)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": false, "message": "Format tanggal tidak valid (YYYY-MM-DD HH:MM)"})
		return
	}

	shiftID, err := strconv.ParseUint(c.PostForm("shift_id"), 10, 32)
	if err != nil || shiftID == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": false, "message": "Shift wajib dipilih"})
		return
	}

	bbmID, err := strconv.ParseUint(c.PostForm("bbm_id"), 10, 32)
	if err != nil || bbmID == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": false, "message": "Jenis BBM wajib dipilih"})
		return
	}

	jmlLiter, err := strconv.ParseInt(c.PostForm("jml_liter"), 10, 64)
	if err != nil || jmlLiter <= 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": false, "message": "Jumlah liter harus lebih dari 0"})
		return
	}

	k := entity.TrxKedatanganBBM{
		PenebusanID:       penebusanID,
		PenebusanDetailID: penebusanDetailID,
		NoLO:              noLO,
		TglKedatangan:     tglKedatangan,
		ShiftID:           uint(shiftID),
		BBMID:             uint(bbmID),
		JmlLiter:          jmlLiter,
		NamaDriver:        strings.TrimSpace(c.PostForm("nama_driver")),
		NoPol:             strings.TrimSpace(strings.ToUpper(c.PostForm("no_pol"))),
	}

	if userObj != nil {
		uid := userObj.ID
		k.CreatedBy = &uid
		k.UpdatedBy = &uid
	}

	if err := h.kedatanganService.Create(&k); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal menyimpan data kedatangan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Data kedatangan berhasil disimpan"})
}

// Update — update data kedatangan
func (h *KedatanganBBMHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	k, err := h.kedatanganService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "Data tidak ditemukan"})
		return
	}

	userRaw, _ := c.Get("user")
	userObj, _ := userRaw.(*entity.User)

	noLO := strings.TrimSpace(c.PostForm("no_lo"))
	if noLO == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": false, "message": "No LO wajib diisi"})
		return
	}

	tglStr := strings.TrimSpace(c.PostForm("tgl_kedatangan"))
	tglKedatangan, err := time.Parse("2006-01-02 15:04", tglStr)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": false, "message": "Format tanggal tidak valid"})
		return
	}

	shiftID, err := strconv.ParseUint(c.PostForm("shift_id"), 10, 32)
	if err != nil || shiftID == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": false, "message": "Shift wajib dipilih"})
		return
	}

	jmlLiter, err := strconv.ParseInt(c.PostForm("jml_liter"), 10, 64)
	if err != nil || jmlLiter <= 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": false, "message": "Jumlah liter harus lebih dari 0"})
		return
	}

	k.NoLO = noLO
	k.TglKedatangan = tglKedatangan
	k.ShiftID = uint(shiftID)
	k.JmlLiter = jmlLiter
	k.NamaDriver = strings.TrimSpace(c.PostForm("nama_driver"))
	k.NoPol = strings.TrimSpace(strings.ToUpper(c.PostForm("no_pol")))

	if userObj != nil {
		uid := userObj.ID
		k.UpdatedBy = &uid
	}

	if err := h.kedatanganService.Update(k); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal mengupdate data kedatangan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Data kedatangan berhasil diupdate"})
}

// Delete — hapus data kedatangan
func (h *KedatanganBBMHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	if err := h.kedatanganService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal menghapus data kedatangan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Data berhasil dihapus"})
}
