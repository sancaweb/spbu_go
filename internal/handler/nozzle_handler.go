package handler

import (
	"log"
	"net/http"
	"spbu_go/internal/entity"
	"spbu_go/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NozzleHandler struct {
	service service.NozzleService
}

func NewNozzleHandler(service service.NozzleService) *NozzleHandler {
	return &NozzleHandler{service}
}

func (h *NozzleHandler) Create(c *gin.Context) {
	var nozzle entity.Nozzle
	if err := c.ShouldBind(&nozzle); err != nil {
		log.Println("Nozzle Bind Failed, trying manual:", err)
		tiangID, _ := strconv.Atoi(c.PostForm("tiang_id"))
		bbmID, _ := strconv.Atoi(c.PostForm("bbm_id"))
		nozzle.TiangID = uint(tiangID)
		nozzle.BBMID = uint(bbmID)
		nozzle.Description = c.PostForm("description")
		nozzle.IsActive = c.PostForm("is_active") == "on"
	} else {
		if c.PostForm("is_active") == "on" {
			nozzle.IsActive = true
		}
	}

	// Get current user ID for updated_by
	userVal, exists := c.Get("user")
	if exists {
		if u, ok := userVal.(entity.User); ok {
			nozzle.UpdatedBy = &u.ID
		} else if uPtr, ok := userVal.(*entity.User); ok {
			nozzle.UpdatedBy = &uPtr.ID
		}
	}

	if err := h.service.Create(&nozzle); err != nil {
		log.Println("Failed to create nozzle:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Nozzle created successfully"})
}

func (h *NozzleHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Nozzle deleted successfully"})
}

func (h *NozzleHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var nozzle entity.Nozzle
	nozzle.ID = uint(id)

	bbmID, _ := strconv.Atoi(c.PostForm("bbm_id"))
	nozzle.BBMID = uint(bbmID)
	nozzle.Description = c.PostForm("description")
	nozzle.IsActive = c.PostForm("is_active") == "on"

	tiangID, _ := strconv.Atoi(c.PostForm("tiang_id"))
	nozzle.TiangID = uint(tiangID)

	// Get current user ID for updated_by
	userVal, exists := c.Get("user")
	if exists {
		if u, ok := userVal.(entity.User); ok {
			nozzle.UpdatedBy = &u.ID
		} else if uPtr, ok := userVal.(*entity.User); ok {
			nozzle.UpdatedBy = &uPtr.ID
		}
	}

	if err := h.service.Update(&nozzle); err != nil {
		log.Println("Failed to update nozzle:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Nozzle updated successfully"})
}
