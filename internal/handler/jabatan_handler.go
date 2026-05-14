package handler

import (
	"net/http"
	"strconv"
	"strings"

	"spbu_go/internal/entity"
	"spbu_go/internal/service"

	"github.com/gin-gonic/gin"
)

type JabatanHandler struct {
	jabatanService service.JabatanService
}

func NewJabatanHandler(jabatanService service.JabatanService) *JabatanHandler {
	return &JabatanHandler{jabatanService}
}

// Index — halaman list jabatan aktif
func (h *JabatanHandler) Index(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	jabatans, _ := h.jabatanService.GetActive()

	c.HTML(http.StatusOK, "master/jabatan/index.html", gin.H{
		"Title":      "Data Jabatan",
		"ActiveMenu": "master_jabatan",
		"User":       user,
		"Favicon":    favicon,
		"Jabatans":   jabatans,
		"IsArchive":  false,
	})
}

// Archive — halaman list jabatan tidak aktif
func (h *JabatanHandler) Archive(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	jabatans, _ := h.jabatanService.GetInactive()

	c.HTML(http.StatusOK, "master/jabatan/index.html", gin.H{
		"Title":      "Jabatan Tidak Aktif",
		"ActiveMenu": "master_jabatan_archive",
		"User":       user,
		"Favicon":    favicon,
		"Jabatans":   jabatans,
		"IsArchive":  true,
	})
}

// Create — simpan jabatan baru
func (h *JabatanHandler) Create(c *gin.Context) {
	kode := strings.TrimSpace(strings.ToUpper(c.PostForm("kode_jabatan")))
	nama := strings.TrimSpace(c.PostForm("nama_jabatan"))

	if kode == "" || nama == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": false, "message": "Kode dan Nama Jabatan wajib diisi"})
		return
	}

	rewardStr := strings.ReplaceAll(c.PostForm("reward_persen"), ",", ".")
	reward, _ := strconv.ParseFloat(rewardStr, 64)

	jabatan := entity.Jabatan{
		KodeJabatan:  kode,
		NamaJabatan:  nama,
		RewardPersen: reward,
		IsActive:     c.PostForm("is_active") != "false",
	}

	if err := h.jabatanService.Create(&jabatan); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "uni_jabatan") {
			c.JSON(http.StatusConflict, gin.H{"status": false, "message": "Kode Jabatan sudah digunakan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal menyimpan data jabatan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Jabatan berhasil ditambahkan"})
}

// Update — update jabatan
func (h *JabatanHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	jabatan, err := h.jabatanService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "Jabatan tidak ditemukan"})
		return
	}

	kode := strings.TrimSpace(strings.ToUpper(c.PostForm("kode_jabatan")))
	nama := strings.TrimSpace(c.PostForm("nama_jabatan"))

	if kode == "" || nama == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": false, "message": "Kode dan Nama Jabatan wajib diisi"})
		return
	}

	rewardStr := strings.ReplaceAll(c.PostForm("reward_persen"), ",", ".")
	reward, _ := strconv.ParseFloat(rewardStr, 64)

	jabatan.KodeJabatan = kode
	jabatan.NamaJabatan = nama
	jabatan.RewardPersen = reward
	jabatan.IsActive = c.PostForm("is_active") != "false"

	if err := h.jabatanService.Update(&jabatan); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "uni_jabatan") {
			c.JSON(http.StatusConflict, gin.H{"status": false, "message": "Kode Jabatan sudah digunakan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal mengupdate data jabatan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Jabatan berhasil diupdate"})
}

// Delete — hapus jabatan
func (h *JabatanHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	if err := h.jabatanService.Delete(uint(id)); err != nil {
		if strings.Contains(err.Error(), "foreign key") || strings.Contains(err.Error(), "violates") {
			c.JSON(http.StatusConflict, gin.H{"status": false, "message": "Jabatan tidak bisa dihapus karena masih digunakan oleh karyawan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal menghapus jabatan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Jabatan berhasil dihapus"})
}
