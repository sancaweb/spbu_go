package handler

import (
	"net/http"
	"strconv"
	"strings"

	"spbu_go/internal/entity"
	"spbu_go/internal/service"

	"github.com/gin-gonic/gin"
)

type PotonganHandler struct {
	potonganService service.PotonganService
}

func NewPotonganHandler(potonganService service.PotonganService) *PotonganHandler {
	return &PotonganHandler{potonganService}
}

// Index — halaman list potongan aktif
func (h *PotonganHandler) Index(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	list, _ := h.potonganService.GetActive()

	c.HTML(http.StatusOK, "master/potongan/index.html", gin.H{
		"Title":     "Master Potongan",
		"ActiveMenu": "master_potongan",
		"User":      user,
		"Favicon":   favicon,
		"Potongans": list,
		"IsArchive": false,
	})
}

// Archive — halaman list potongan tidak aktif
func (h *PotonganHandler) Archive(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	list, _ := h.potonganService.GetInactive()

	c.HTML(http.StatusOK, "master/potongan/index.html", gin.H{
		"Title":     "Potongan Tidak Aktif",
		"ActiveMenu": "master_potongan_archive",
		"User":      user,
		"Favicon":   favicon,
		"Potongans": list,
		"IsArchive": true,
	})
}

// Create — simpan potongan baru
func (h *PotonganHandler) Create(c *gin.Context) {
	kode := strings.TrimSpace(strings.ToUpper(c.PostForm("kode_potongan")))
	nama := strings.TrimSpace(c.PostForm("nama_potongan"))
	tipe := c.PostForm("tipe")
	deskripsi := strings.TrimSpace(c.PostForm("deskripsi"))

	if kode == "" || nama == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": false, "message": "Kode dan Nama Potongan wajib diisi"})
		return
	}
	if tipe != "nominal" && tipe != "persen" {
		tipe = "nominal"
	}

	nilaiStr := strings.ReplaceAll(c.PostForm("nilai"), ".", "")
	nilaiStr = strings.ReplaceAll(nilaiStr, ",", ".")
	nilaiFloat, _ := strconv.ParseFloat(nilaiStr, 64)

	p := entity.Potongan{
		KodePotongan: kode,
		NamaPotongan: nama,
		Tipe:         tipe,
		Nilai:        int64(nilaiFloat),
		Deskripsi:    deskripsi,
		IsActive:     c.PostForm("is_active") != "false",
	}

	if userVal, exists := c.Get("user"); exists {
		if u, ok := userVal.(*entity.User); ok {
			p.UpdatedBy = &u.ID
		}
	}

	if err := h.potonganService.Create(&p); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "uni_potongan") {
			c.JSON(http.StatusConflict, gin.H{"status": false, "message": "Kode Potongan sudah digunakan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal menyimpan data potongan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Potongan berhasil ditambahkan"})
}

// Update — update potongan
func (h *PotonganHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	p, err := h.potonganService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "Data potongan tidak ditemukan"})
		return
	}

	kode := strings.TrimSpace(strings.ToUpper(c.PostForm("kode_potongan")))
	nama := strings.TrimSpace(c.PostForm("nama_potongan"))
	tipe := c.PostForm("tipe")
	deskripsi := strings.TrimSpace(c.PostForm("deskripsi"))

	if kode == "" || nama == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": false, "message": "Kode dan Nama Potongan wajib diisi"})
		return
	}
	if tipe != "nominal" && tipe != "persen" {
		tipe = "nominal"
	}

	nilaiStr := strings.ReplaceAll(c.PostForm("nilai"), ".", "")
	nilaiStr = strings.ReplaceAll(nilaiStr, ",", ".")
	nilaiFloat, _ := strconv.ParseFloat(nilaiStr, 64)

	p.KodePotongan = kode
	p.NamaPotongan = nama
	p.Tipe = tipe
	p.Nilai = int64(nilaiFloat)
	p.Deskripsi = deskripsi
	p.IsActive = c.PostForm("is_active") != "false"

	if userVal, exists := c.Get("user"); exists {
		if u, ok := userVal.(*entity.User); ok {
			p.UpdatedBy = &u.ID
		}
	}

	if err := h.potonganService.Update(&p); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "uni_potongan") {
			c.JSON(http.StatusConflict, gin.H{"status": false, "message": "Kode Potongan sudah digunakan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal mengupdate data potongan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Potongan berhasil diupdate"})
}

// Delete — hapus potongan (soft delete)
func (h *PotonganHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	if err := h.potonganService.Delete(uint(id)); err != nil {
		if strings.Contains(err.Error(), "foreign key") || strings.Contains(err.Error(), "violates") {
			c.JSON(http.StatusConflict, gin.H{"status": false, "message": "Potongan tidak bisa dihapus karena sedang digunakan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal menghapus potongan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Potongan berhasil dihapus"})
}
