package handler

import (
	"net/http"
	"strconv"
	"strings"

	"spbu_go/internal/entity"
	"spbu_go/internal/service"

	"github.com/gin-gonic/gin"
)

type PendapatanHandler struct {
	pendapatanService service.PendapatanService
}

func NewPendapatanHandler(pendapatanService service.PendapatanService) *PendapatanHandler {
	return &PendapatanHandler{pendapatanService}
}

// Index — halaman list pendapatan aktif
func (h *PendapatanHandler) Index(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	list, _ := h.pendapatanService.GetActive()

	c.HTML(http.StatusOK, "master/pendapatan/index.html", gin.H{
		"Title":       "Master Pendapatan",
		"ActiveMenu":  "master_pendapatan",
		"User":        user,
		"Favicon":     favicon,
		"Pendapatans": list,
		"IsArchive":   false,
	})
}

// Archive — halaman list pendapatan tidak aktif
func (h *PendapatanHandler) Archive(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	list, _ := h.pendapatanService.GetInactive()

	c.HTML(http.StatusOK, "master/pendapatan/index.html", gin.H{
		"Title":       "Pendapatan Tidak Aktif",
		"ActiveMenu":  "master_pendapatan_archive",
		"User":        user,
		"Favicon":     favicon,
		"Pendapatans": list,
		"IsArchive":   true,
	})
}

// Create — simpan pendapatan baru
func (h *PendapatanHandler) Create(c *gin.Context) {
	nama := strings.TrimSpace(c.PostForm("nama_pendapatan"))
	tipe := c.PostForm("tipe")
	deskripsi := strings.TrimSpace(c.PostForm("deskripsi"))

	if nama == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": false, "message": "Nama Pendapatan wajib diisi"})
		return
	}
	if tipe != "nominal" && tipe != "persen" {
		tipe = "nominal"
	}

	nilaiStr := strings.ReplaceAll(c.PostForm("nilai"), ".", "")
	nilaiStr = strings.ReplaceAll(nilaiStr, ",", ".")
	nilaiFloat, _ := strconv.ParseFloat(nilaiStr, 64)
	nilai := int64(nilaiFloat)

	p := entity.Pendapatan{
		NamaPendapatan: nama,
		Tipe:           tipe,
		Nilai:          nilai,
		Deskripsi:      deskripsi,
		IsActive:       c.PostForm("is_active") != "false",
	}

	if userVal, exists := c.Get("user"); exists {
		if u, ok := userVal.(*entity.User); ok {
			p.UpdatedBy = &u.ID
		}
	}

	if err := h.pendapatanService.Create(&p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal menyimpan data pendapatan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Pendapatan berhasil ditambahkan"})
}

// Update — update pendapatan
func (h *PendapatanHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	p, err := h.pendapatanService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "Data pendapatan tidak ditemukan"})
		return
	}

	nama := strings.TrimSpace(c.PostForm("nama_pendapatan"))
	tipe := c.PostForm("tipe")
	deskripsi := strings.TrimSpace(c.PostForm("deskripsi"))

	if nama == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": false, "message": "Nama Pendapatan wajib diisi"})
		return
	}
	if tipe != "nominal" && tipe != "persen" {
		tipe = "nominal"
	}

	nilaiStr := strings.ReplaceAll(c.PostForm("nilai"), ".", "")
	nilaiStr = strings.ReplaceAll(nilaiStr, ",", ".")
	nilaiFloat, _ := strconv.ParseFloat(nilaiStr, 64)

	p.NamaPendapatan = nama
	p.Tipe = tipe
	p.Nilai = int64(nilaiFloat)
	p.Deskripsi = deskripsi
	p.IsActive = c.PostForm("is_active") != "false"

	if userVal, exists := c.Get("user"); exists {
		if u, ok := userVal.(*entity.User); ok {
			p.UpdatedBy = &u.ID
		}
	}

	if err := h.pendapatanService.Update(&p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal mengupdate data pendapatan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Pendapatan berhasil diupdate"})
}

// Delete — hapus pendapatan (soft delete)
func (h *PendapatanHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	if err := h.pendapatanService.Delete(uint(id)); err != nil {
		if strings.Contains(err.Error(), "foreign key") || strings.Contains(err.Error(), "violates") {
			c.JSON(http.StatusConflict, gin.H{"status": false, "message": "Pendapatan tidak bisa dihapus karena sedang digunakan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal menghapus pendapatan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Pendapatan berhasil dihapus"})
}
