package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"spbu_go/internal/dto"
	"spbu_go/internal/entity"
	"spbu_go/internal/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type PartnerHandler struct {
	partnerService service.PartnerService
}

func NewPartnerHandler(partnerService service.PartnerService) *PartnerHandler {
	return &PartnerHandler{partnerService}
}

// Index page for active partners
func (h *PartnerHandler) Index(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")

	c.HTML(http.StatusOK, "partner/index.html", gin.H{
		"Title":      "Data Partner",
		"ActiveMenu": "partner",
		"IsArchive":  false,
		"User":       user,
		"Favicon":    favicon,
	})
}

// Datatable endpoint for Server-Side processing
func (h *PartnerHandler) Datatable(c *gin.Context) {
	var req dto.DatatableRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isActive := c.Query("type") != "archive"

	total, filtered, partners, err := h.partnerService.Datatable(req, isActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data dari server"})
		return
	}

	// Format phone numbers for UI display
	for i := range partners {
		if strings.HasPrefix(partners[i].Phone, "+62") {
			partners[i].Phone = "0" + strings.TrimPrefix(partners[i].Phone, "+62")
		}
	}

	c.JSON(http.StatusOK, dto.DatatableResponse{
		Draw:            req.Draw,
		RecordsTotal:    total,
		RecordsFiltered: filtered,
		Data:            partners,
	})
}

// Archive page for inactive partners
func (h *PartnerHandler) Archive(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")

	c.HTML(http.StatusOK, "partner/index.html", gin.H{
		"Title":      "Data Partner (Tidak Aktif)",
		"ActiveMenu": "partner",
		"IsArchive":  true,
		"User":       user,
		"Favicon":    favicon,
	})
}

// Create new partner
func (h *PartnerHandler) Create(c *gin.Context) {
	var partner entity.Partner
	if err := c.ShouldBind(&partner); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Input tidak valid"})
		return
	}

	// Apply Formatting Rules
	caser := cases.Title(language.Indonesian)
	partner.Name = caser.String(strings.ToLower(strings.TrimSpace(partner.Name)))

	if partner.ContactPerson != "" {
		partner.ContactPerson = caser.String(strings.ToLower(strings.TrimSpace(partner.ContactPerson)))
	}

	if strings.HasPrefix(partner.Phone, "0") {
		partner.Phone = "+62" + strings.TrimPrefix(partner.Phone, "0")
	}

	// Get logged in user ID for UpdatedBy
	if userVal, exists := c.Get("user"); exists {
		if u, ok := userVal.(*entity.User); ok {
			partner.UpdatedBy = &u.ID
		}
	}

	if err := h.partnerService.Create(&partner); err != nil {
		log.Printf("Error creating partner: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal menyimpan partner"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Partner berhasil ditambahkan"})
}

// Update existing partner
func (h *PartnerHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	partner, err := h.partnerService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "Partner tidak ditemukan"})
		return
	}

	if err := c.ShouldBind(&partner); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Input tidak valid"})
		return
	}

	// Apply Formatting Rules
	caser := cases.Title(language.Indonesian)
	partner.Name = caser.String(strings.ToLower(strings.TrimSpace(partner.Name)))

	if partner.ContactPerson != "" {
		partner.ContactPerson = caser.String(strings.ToLower(strings.TrimSpace(partner.ContactPerson)))
	}

	if strings.HasPrefix(partner.Phone, "0") {
		partner.Phone = "+62" + strings.TrimPrefix(partner.Phone, "0")
	}

	// Update the updater ID
	if userVal, exists := c.Get("user"); exists {
		if u, ok := userVal.(*entity.User); ok {
			partner.UpdatedBy = &u.ID
		}
	}

	if err := h.partnerService.Update(&partner); err != nil {
		log.Printf("Error updating partner: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal mengupdate partner"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Partner berhasil diupdate"})
}

// Soft Delete
func (h *PartnerHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	if err := h.partnerService.Delete(uint(id)); err != nil {
		log.Printf("Error deleting partner: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal menghapus partner"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Partner berhasil dinonaktifkan"})
}

// Restore Soft Deleted Partner
func (h *PartnerHandler) Restore(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	if err := h.partnerService.Restore(uint(id)); err != nil {
		log.Printf("Error restoring partner: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal mengembalikan partner"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Partner berhasil diaktifkan kembali"})
}
