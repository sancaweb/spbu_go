package handler

import (
	"net/http"
	"spbu_go/internal/entity"
	"spbu_go/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type COAMappingHandler struct {
	mappingService service.COAMappingService
	coaService     service.COAService
}

func NewCOAMappingHandler(mappingService service.COAMappingService, coaService service.COAService) *COAMappingHandler {
	return &COAMappingHandler{mappingService, coaService}
}

func (h *COAMappingHandler) Index(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")

	transTypes := h.mappingService.GetTransTypes()
	allMappings, _ := h.mappingService.GetAll()
	detailCOAs, _ := h.coaService.GetDetailAccounts()

	// Build map: transType -> []COAMapping for fast lookup in template
	mappingMap := map[string][]entity.COAMapping{}
	for _, m := range allMappings {
		mappingMap[m.TransType] = append(mappingMap[m.TransType], m)
	}

	c.HTML(http.StatusOK, "master/keuangan/coa_mapping/index.html", gin.H{
		"User":        user,
		"Favicon":     favicon,
		"Title":       "COA Mapping",
		"ActiveMenu":  "keuangan_coa_mapping",
		"TransTypes":  transTypes,
		"MappingMap":  mappingMap,
		"DetailCOAs":  detailCOAs,
	})
}

// Upsert handles POST /master/keuangan/coa-mapping/upsert
// Saves a single mapping: trans_type, role, label, coa_id, bbm_id (optional)
func (h *COAMappingHandler) Upsert(c *gin.Context) {
	transType := c.PostForm("trans_type")
	role := c.PostForm("role")
	label := c.PostForm("label")
	coaIDStr := c.PostForm("coa_id")
	bbmIDStr := c.PostForm("bbm_id")

	coaID, _ := strconv.ParseUint(coaIDStr, 10, 64)
	if coaID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "COA harus dipilih"})
		return
	}

	var bbmID *uint
	if bbmIDStr != "" && bbmIDStr != "0" {
		id, _ := strconv.ParseUint(bbmIDStr, 10, 64)
		if id > 0 {
			uid := uint(id)
			bbmID = &uid
		}
	}

	if err := h.mappingService.Upsert(transType, role, label, uint(coaID), bbmID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Mapping disimpan"})
}
