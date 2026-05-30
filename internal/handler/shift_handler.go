package handler

import (
	"net/http"
	"strconv"
	"strings"

	"spbu_go/internal/entity"
	"spbu_go/internal/service"

	"github.com/gin-gonic/gin"
)

type ShiftHandler struct {
	shiftService service.ShiftService
}

func NewShiftHandler(shiftService service.ShiftService) *ShiftHandler {
	return &ShiftHandler{shiftService}
}

// Index — halaman list shift
func (h *ShiftHandler) Index(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	shifts, _ := h.shiftService.GetAll()

	c.HTML(http.StatusOK, "master/shift/index.html", gin.H{
		"Title":      "Data Shift",
		"ActiveMenu": "master_shift",
		"User":       user,
		"Favicon":    favicon,
		"Shifts":     shifts,
	})
}

// Create — simpan shift baru
func (h *ShiftHandler) Create(c *gin.Context) {
	shiftName := strings.TrimSpace(c.PostForm("shift_name"))
	shiftTime := strings.TrimSpace(c.PostForm("shift_time"))

	if shiftName == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": false, "message": "Nama Shift wajib diisi"})
		return
	}

	shift := entity.Shift{
		ShiftName: shiftName,
		ShiftTime: shiftTime,
	}

	if err := h.shiftService.Create(&shift); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal menyimpan data shift"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Shift berhasil ditambahkan", "data": shift})
}

// Update — update shift
func (h *ShiftHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	shift, err := h.shiftService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "Shift tidak ditemukan"})
		return
	}

	shiftName := strings.TrimSpace(c.PostForm("shift_name"))
	shiftTime := strings.TrimSpace(c.PostForm("shift_time"))

	if shiftName == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": false, "message": "Nama Shift wajib diisi"})
		return
	}

	shift.ShiftName = shiftName
	shift.ShiftTime = shiftTime

	if err := h.shiftService.Update(&shift); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal mengupdate shift"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Shift berhasil diupdate", "data": shift})
}

// Delete — hapus shift
func (h *ShiftHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	used, err := h.shiftService.IsUsed(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal memeriksa data shift"})
		return
	}
	if used {
		c.JSON(http.StatusConflict, gin.H{
			"status":  false,
			"message": "Shift ini tidak dapat dihapus karena sudah memiliki data transaksi Kedatangan BBM yang terhubung. Hapus terlebih dahulu seluruh transaksi Kedatangan BBM yang menggunakan shift ini, kemudian coba hapus kembali.",
		})
		return
	}

	if err := h.shiftService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal menghapus shift"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Shift berhasil dihapus"})
}
