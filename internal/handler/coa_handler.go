package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"spbu_go/internal/entity"
	"spbu_go/internal/service"

	"github.com/gin-gonic/gin"
)

type COAHandler struct {
	coaService service.COAService
}

func NewCOAHandler(coaService service.COAService) *COAHandler {
	return &COAHandler{coaService}
}

// Index — list all COA accounts grouped by type
func (h *COAHandler) Index(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	types, _ := h.coaService.GetAllGrouped()

	c.HTML(http.StatusOK, "master/keuangan/coa/index.html", gin.H{
		"Title":      "Chart of Account (COA)",
		"ActiveMenu": "keuangan_coa",
		"User":       user,
		"Favicon":    favicon,
		"COATypes":   types,
	})
}

// Create — POST /master/keuangan/coa
func (h *COAHandler) Create(c *gin.Context) {
	userVal, _ := c.Get("user")

	coaTypeIDStr := c.PostForm("coa_type_id")
	coaTypeID, _ := strconv.ParseUint(coaTypeIDStr, 10, 64)
	code := strings.TrimSpace(c.PostForm("code"))
	name := strings.TrimSpace(c.PostForm("name"))
	description := strings.TrimSpace(c.PostForm("description"))
	isHeader := c.PostForm("is_header") == "true"

	if code == "" || name == "" || coaTypeID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Kode akun, nama, dan tipe akun wajib diisi"})
		return
	}

	coa := &entity.COA{
		COATypeID:   uint(coaTypeID),
		Code:        code,
		Name:        name,
		Description: description,
		IsHeader:    isHeader,
		IsActive:    true,
	}
	if u, ok := userVal.(entity.User); ok {
		coa.UpdatedBy = &u.ID
	}

	if err := h.coaService.Create(coa); err != nil {
		msg := "Gagal menyimpan akun COA"
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			msg = "Kode akun sudah digunakan, gunakan kode lain"
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Akun COA berhasil ditambahkan"})
}

// Update — POST /master/keuangan/coa/:id
func (h *COAHandler) Update(c *gin.Context) {
	userVal, _ := c.Get("user")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	coaTypeIDStr := c.PostForm("coa_type_id")
	coaTypeID, _ := strconv.ParseUint(coaTypeIDStr, 10, 64)
	code := strings.TrimSpace(c.PostForm("code"))
	name := strings.TrimSpace(c.PostForm("name"))
	description := strings.TrimSpace(c.PostForm("description"))
	isHeader := c.PostForm("is_header") == "true"
	isActive := c.PostForm("is_active") == "true"

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Nama akun wajib diisi"})
		return
	}

	coa := &entity.COA{
		ID:          uint(id),
		COATypeID:   uint(coaTypeID),
		Code:        code,
		Name:        name,
		Description: description,
		IsHeader:    isHeader,
		IsActive:    isActive,
	}
	if u, ok := userVal.(entity.User); ok {
		coa.UpdatedBy = &u.ID
	}

	if err := h.coaService.Update(coa); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal update COA: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Akun COA berhasil diperbarui"})
}

// Delete — POST /master/keuangan/coa/:id/delete
func (h *COAHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}
	if err := h.coaService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Akun COA berhasil dihapus"})
}

// JournalEntryWithBalance — journal entry enriched with running balance for display
type JournalEntryWithBalance struct {
	entity.JournalEntry
	Balance int64
}

// Transactions — GET /master/keuangan/coa/:id/transactions
func (h *COAHandler) Transactions(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Redirect(http.StatusFound, "/master/keuangan/coa")
		return
	}

	entries, coa, err := h.coaService.GetTransactions(uint(id))
	if err != nil {
		c.Redirect(http.StatusFound, "/master/keuangan/coa")
		return
	}

	// Determine normal balance direction
	normalBalance := "debit"
	if coa.COAType != nil {
		normalBalance = coa.COAType.NormalBalance
	}

	// Calculate running balance: entries are DESC (newest first).
	// Process oldest→newest, then reverse for display (newest first).
	n := len(entries)
	enriched := make([]JournalEntryWithBalance, n)
	var balance int64
	for i := n - 1; i >= 0; i-- {
		e := entries[i]
		if normalBalance == "debit" {
			balance += e.Debit - e.Credit
		} else {
			balance += e.Credit - e.Debit
		}
		enriched[n-1-i] = JournalEntryWithBalance{e, balance}
	}
	// Re-reverse so newest is first (index 0 = highest balance after latest tx)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		enriched[i], enriched[j] = enriched[j], enriched[i]
	}

	var totalDebit, totalCredit int64
	for _, e := range entries {
		totalDebit += e.Debit
		totalCredit += e.Credit
	}

	c.HTML(http.StatusOK, "master/keuangan/coa/transactions.html", gin.H{
		"Title":       coa.Code + " — " + coa.Name,
		"ActiveMenu":  "keuangan_coa",
		"User":        user,
		"Favicon":     favicon,
		"COA":         coa,
		"Entries":     enriched,
		"TotalDebit":  totalDebit,
		"TotalCredit": totalCredit,
		"CurrentTime": time.Now(),
	})
}
