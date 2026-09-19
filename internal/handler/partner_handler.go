package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"spbu_go/internal/dto"
	"spbu_go/internal/entity"
	"spbu_go/internal/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
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
	partner := entity.Partner{IsActive: true}
	if err := c.ShouldBind(&partner); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Input tidak valid"})
		return
	}
	if err := bindPartnerAgreementFields(c, &partner); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": err.Error()})
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
	if err := bindPartnerAgreementFields(c, &partner); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": err.Error()})
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

func (h *PartnerHandler) DeletePermanent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	if err := h.partnerService.DeletePermanent(uint(id)); err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "Partner tidak ditemukan"})
		case errors.Is(err, entity.ErrPartnerNotArchived), errors.Is(err, entity.ErrPartnerHasPiutang), errors.Is(err, entity.ErrPartnerReferenced):
			c.JSON(http.StatusConflict, gin.H{"status": false, "message": err.Error()})
		default:
			log.Printf("Error permanently deleting partner %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal menghapus partner secara permanen"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Partner berhasil dihapus permanen"})
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

const partnerDateFormat = "2006-01-02"

func bindPartnerAgreementFields(c *gin.Context, partner *entity.Partner) error {
	if value, exists := c.GetPostForm("start_date"); exists {
		date, err := parsePartnerDate(value)
		if err != nil {
			return fmt.Errorf("Start Date harus berformat YYYY-MM-DD")
		}
		partner.StartDate = date
	}

	if value, exists := c.GetPostForm("end_date"); exists {
		date, err := parsePartnerDate(value)
		if err != nil {
			return fmt.Errorf("End Date harus berformat YYYY-MM-DD")
		}
		partner.EndDate = date
	}

	if value, exists := c.GetPostForm("isactive"); exists {
		isActive, err := parsePartnerBoolean(value)
		if err != nil {
			return fmt.Errorf("Status partner tidak valid")
		}
		partner.IsActive = isActive
	}

	if value, exists := c.GetPostForm("auto_off"); exists {
		autoOff, err := parsePartnerBoolean(value)
		if err != nil {
			return fmt.Errorf("Auto off tidak valid")
		}
		partner.AutoOff = autoOff
	}

	if partner.StartDate != nil && partner.EndDate != nil && partner.EndDate.Before(*partner.StartDate) {
		return fmt.Errorf("End Date tidak boleh lebih awal dari Start Date")
	}
	if partner.AutoOff && partner.EndDate == nil {
		return fmt.Errorf("End Date wajib diisi jika Auto off diaktifkan")
	}

	return nil
}

func parsePartnerDate(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := time.ParseInLocation(partnerDateFormat, value, time.UTC)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parsePartnerBoolean(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "on", "yes":
		return true, nil
	case "false", "0", "off", "no":
		return false, nil
	default:
		return false, fmt.Errorf("nilai boolean tidak dikenal")
	}
}
