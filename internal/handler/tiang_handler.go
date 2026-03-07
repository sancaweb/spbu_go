package handler

import (
	"net/http"
	"spbu_go/internal/entity"
	"spbu_go/internal/service"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type TiangHandler struct {
	service    service.TiangService
	bbmService service.BBMService
}

func NewTiangHandler(service service.TiangService, bbmService service.BBMService) *TiangHandler {
	return &TiangHandler{service, bbmService}
}

func (h *TiangHandler) Index(c *gin.Context) {
	tiangs, err := h.service.GetAll()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"Error": err.Error()})
		return
	}

	bbms, err := h.bbmService.GetActive()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"Error": err.Error()})
		return
	}

	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	c.HTML(http.StatusOK, "master/tiang/index.html", gin.H{
		"Tiangs":     tiangs,
		"BBMs":       bbms,
		"User":       user,
		"Favicon":    favicon,
		"Title":      "Master Data Tiang",
		"ActiveMenu": "master_tiang",
	})
}

func (h *TiangHandler) Create(c *gin.Context) {
	var tiang entity.Tiang

	// Auto-capitalize first letter of name
	name := c.PostForm("name")
	if len(name) > 0 {
		name = strings.ToUpper(string(name[0])) + name[1:]
	}
	tiang.Name = name
	tiang.Slug = c.PostForm("slug")

	// Get current user ID for updated_by
	userVal, exists := c.Get("user")
	if exists {
		if u, ok := userVal.(entity.User); ok {
			tiang.UpdatedBy = &u.ID
		} else if uPtr, ok := userVal.(*entity.User); ok {
			tiang.UpdatedBy = &uPtr.ID
		}
	}

	if err := h.service.Create(&tiang); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Tiang created successfully"})
}

func (h *TiangHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	tiang, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "Tiang not found"})
		return
	}

	// Auto-capitalize first letter of name
	name := c.PostForm("name")
	if len(name) > 0 {
		name = strings.ToUpper(string(name[0])) + name[1:]
	}
	tiang.Name = name
	tiang.Slug = c.PostForm("slug")

	// Get current user ID for updated_by
	userVal, exists := c.Get("user")
	if exists {
		if u, ok := userVal.(entity.User); ok {
			tiang.UpdatedBy = &u.ID
		} else if uPtr, ok := userVal.(*entity.User); ok {
			tiang.UpdatedBy = &uPtr.ID
		}
	}

	if err := h.service.Update(&tiang); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Tiang updated successfully"})
}

func (h *TiangHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Tiang deleted successfully"})
}
