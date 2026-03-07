package handler

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"spbu_go/internal/service"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/image/draw"
)

type SettingHandler struct {
	settingService service.SettingService
}

func NewSettingHandler(settingService service.SettingService) *SettingHandler {
	return &SettingHandler{settingService}
}

func (h *SettingHandler) Index(c *gin.Context) {
	settings, _ := h.settingService.GetAll()

	// Build a map for easy access in template
	settingsMap := make(map[string]string)
	for _, s := range settings {
		settingsMap[s.SettingName] = s.SettingValue
	}

	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	c.HTML(http.StatusOK, "settings/index.html", gin.H{
		"Settings":   settingsMap,
		"User":       user,
		"Favicon":    favicon,
		"Title":      "Site Settings",
		"ActiveMenu": "site_settings",
	})
}

func (h *SettingHandler) Update(c *gin.Context) {
	name := c.PostForm("setting_name")
	value := c.PostForm("setting_value")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Setting name is required"})
		return
	}

	if err := h.settingService.Set(name, value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Setting updated successfully"})
}

func (h *SettingHandler) UploadFavicon(c *gin.Context) {
	file, err := c.FormFile("favicon")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "No file uploaded"})
		return
	}

	// Validate file type
	ext := filepath.Ext(file.Filename)
	allowed := map[string]bool{".png": true, ".ico": true, ".jpg": true, ".jpeg": true, ".svg": true, ".webp": true}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Invalid file type. Allowed: png, ico, jpg, svg, webp"})
		return
	}

	// Ensure upload directory exists
	uploadDir := "./static/uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Failed to create upload directory"})
		return
	}

	// For raster images (png, jpg, jpeg), resize to 32x32
	if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
		// Open uploaded file
		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Failed to read uploaded file"})
			return
		}
		defer src.Close()

		// Decode image
		srcImg, _, err := image.Decode(src)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Failed to decode image"})
			return
		}

		// Resize to 32x32 using high-quality CatmullRom interpolation
		const faviconSize = 32
		dstImg := image.NewRGBA(image.Rect(0, 0, faviconSize, faviconSize))
		draw.CatmullRom.Scale(dstImg, dstImg.Bounds(), srcImg, srcImg.Bounds(), draw.Over, nil)

		// Always save as PNG for best favicon quality
		filename := fmt.Sprintf("favicon_%d.png", time.Now().Unix())
		savePath := filepath.Join(uploadDir, filename)

		outFile, err := os.Create(savePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Failed to save resized image"})
			return
		}
		defer outFile.Close()

		if err := png.Encode(outFile, dstImg); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Failed to encode resized image"})
			return
		}

		faviconURL := "/static/uploads/" + filename
		if err := h.settingService.Set("favicon", faviconURL); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Failed to save setting"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": true, "message": "Favicon updated (resized to 32x32)", "favicon_url": faviconURL})
		return
	}

	// For non-raster (ico, svg, webp) — save as-is
	filename := fmt.Sprintf("favicon_%d%s", time.Now().Unix(), ext)
	savePath := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Failed to save file"})
		return
	}

	faviconURL := "/static/uploads/" + filename
	if err := h.settingService.Set("favicon", faviconURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Failed to save setting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Favicon updated successfully", "favicon_url": faviconURL})
}

// Blank import to register JPEG decoder
var _ = jpeg.DefaultQuality
