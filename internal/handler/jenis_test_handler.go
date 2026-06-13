package handler

import (
	"net/http"
	"strconv"
	"strings"

	"spbu_go/internal/entity"
	"spbu_go/internal/service"

	"github.com/gin-gonic/gin"
)

// JenisTestHandler — CRUD master data jenis test.
type JenisTestHandler struct {
	svc service.JenisTestService
}

func NewJenisTestHandler(svc service.JenisTestService) *JenisTestHandler {
	return &JenisTestHandler{svc}
}

func parseJenisTestActive(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "false", "0", "off", "no":
		return false
	default:
		return true
	}
}

// Index — halaman list jenis test.
func (h *JenisTestHandler) Index(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	list, _ := h.svc.GetAll()

	c.HTML(http.StatusOK, "master/jenis-test/index.html", gin.H{
		"Title":      "Data Jenis Test",
		"ActiveMenu": "master_jenis_test",
		"User":       user,
		"Favicon":    favicon,
		"JenisTests": list,
	})
}

// Create — simpan jenis test baru.
func (h *JenisTestHandler) Create(c *gin.Context) {
	namaTest := strings.TrimSpace(c.PostForm("nama_test"))
	deskripsi := strings.TrimSpace(c.PostForm("deskripsi"))
	isActive := parseJenisTestActive(c.PostForm("is_active"))

	if namaTest == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Nama Test tidak boleh kosong"})
		return
	}

	j := &entity.JenisTest{
		NamaTest:  namaTest,
		Deskripsi: deskripsi,
		IsActive:  isActive,
	}
	if err := h.svc.Create(j); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal menyimpan: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Jenis Test berhasil ditambahkan", "data": j})
}

// Update — edit jenis test.
func (h *JenisTestHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ID tidak valid"})
		return
	}

	j, err := h.svc.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Data tidak ditemukan"})
		return
	}

	j.NamaTest = strings.TrimSpace(c.PostForm("nama_test"))
	j.Deskripsi = strings.TrimSpace(c.PostForm("deskripsi"))
	j.IsActive = parseJenisTestActive(c.PostForm("is_active"))

	if j.NamaTest == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Nama Test tidak boleh kosong"})
		return
	}

	if err := h.svc.Update(&j); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengupdate: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Jenis Test berhasil diupdate", "data": j})
}

// ToggleActive — toggle is_active tanpa reload halaman.
func (h *JenisTestHandler) ToggleActive(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ID tidak valid"})
		return
	}

	j, err := h.svc.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Data tidak ditemukan"})
		return
	}

	j.IsActive = !j.IsActive
	if err := h.svc.Update(&j); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengupdate: " + err.Error()})
		return
	}

	statusText := "dinonaktifkan"
	if j.IsActive {
		statusText = "diaktifkan"
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Jenis Test berhasil " + statusText, "data": gin.H{"is_active": j.IsActive}})
}

// Delete — hapus jenis test (soft delete).
func (h *JenisTestHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ID tidak valid"})
		return
	}

	if err := h.svc.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal menghapus: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Jenis Test berhasil dihapus"})
}
