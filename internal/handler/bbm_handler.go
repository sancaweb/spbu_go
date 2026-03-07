package handler

import (
	"log"
	"math"
	"net/http"
	"spbu_go/internal/entity"
	"spbu_go/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BBMHandler struct {
	service        service.BBMService
	settingService service.SettingService
}

func NewBBMHandler(service service.BBMService, settingService service.SettingService) *BBMHandler {
	return &BBMHandler{service, settingService}
}

func (h *BBMHandler) getStockDecimalPlaces() int {
	return h.settingService.GetInt("stock_decimal_places", 0)
}

func (h *BBMHandler) Index(c *gin.Context) {
	bbms, err := h.service.GetAll()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"Error": err.Error()})
		return
	}

	decimalPlaces := h.getStockDecimalPlaces()

	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	c.HTML(http.StatusOK, "master/bbm/index.html", gin.H{
		"BBMs":               bbms,
		"User":               user,
		"Favicon":            favicon,
		"Title":              "Master Data BBM",
		"ActiveMenu":         "master_bbm",
		"StockDecimalPlaces": decimalPlaces,
	})
}

func (h *BBMHandler) Create(c *gin.Context) {
	log.Println("BBM Create Request:", c.Request.PostForm)

	var bbm entity.BBM
	bbm.Name = c.PostForm("name")
	margin, _ := strconv.ParseFloat(c.PostForm("margin"), 64)
	bbm.Margin = margin
	price, _ := strconv.ParseFloat(c.PostForm("price"), 64)
	bbm.Price = price
	reward, _ := strconv.ParseFloat(c.PostForm("reward_percent"), 64)
	bbm.RewardPercent = reward
	bbm.IsActive = c.PostForm("is_active") == "on"

	// Convert stock: multiply by 10^precision to store as integer
	decimalPlaces := h.getStockDecimalPlaces()
	stockFloat, _ := strconv.ParseFloat(c.PostForm("stock"), 64)
	multiplier := math.Pow(10, float64(decimalPlaces))
	bbm.Stock = int64(math.Round(stockFloat * multiplier))

	log.Printf("BBM Struct after bind: %+v\n", bbm)

	userVal, exists := c.Get("user")
	if exists {
		if u, ok := userVal.(entity.User); ok {
			bbm.UpdatedBy = &u.ID
		} else if uPtr, ok := userVal.(*entity.User); ok {
			bbm.UpdatedBy = &uPtr.ID
		} else {
			log.Println("Warning: User in context is not entity.User or *entity.User")
		}
	}

	if err := h.service.Create(&bbm); err != nil {
		log.Println("Service Create failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "BBM created successfully"})
}

func (h *BBMHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("BBM Update Request for ID: %d, Data: %v\n", id, c.Request.PostForm)

	bbm, err := h.service.GetByID(uint(id))
	if err != nil {
		log.Println("BBM not found:", err)
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "BBM not found"})
		return
	}

	bbm.Name = c.PostForm("name")
	margin, _ := strconv.ParseFloat(c.PostForm("margin"), 64)
	bbm.Margin = margin
	price, _ := strconv.ParseFloat(c.PostForm("price"), 64)
	bbm.Price = price
	reward, _ := strconv.ParseFloat(c.PostForm("reward_percent"), 64)
	bbm.RewardPercent = reward
	bbm.IsActive = c.PostForm("is_active") == "on"

	// Convert stock: multiply by 10^precision to store as integer
	decimalPlaces := h.getStockDecimalPlaces()
	stockFloat, _ := strconv.ParseFloat(c.PostForm("stock"), 64)
	multiplier := math.Pow(10, float64(decimalPlaces))
	bbm.Stock = int64(math.Round(stockFloat * multiplier))

	userVal, exists := c.Get("user")
	if exists {
		if u, ok := userVal.(entity.User); ok {
			bbm.UpdatedBy = &u.ID
		} else if uPtr, ok := userVal.(*entity.User); ok {
			bbm.UpdatedBy = &uPtr.ID
		}
	}

	if err := h.service.Update(&bbm); err != nil {
		log.Println("Service Update failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "BBM updated successfully"})
}

func (h *BBMHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "BBM deleted successfully"})
}
