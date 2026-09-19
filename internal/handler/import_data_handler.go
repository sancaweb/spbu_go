package handler

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"spbu_go/internal/dto"
	"spbu_go/internal/entity"
	"spbu_go/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ImportDataHandler struct {
	importService service.ImportDataService
}

func NewImportDataHandler(importService service.ImportDataService) *ImportDataHandler {
	return &ImportDataHandler{importService: importService}
}

func (h *ImportDataHandler) Index(c *gin.Context) {
	connection, err := h.importService.GetConnection()
	if err != nil {
		log.Printf("Gagal membaca konfigurasi koneksi legacy: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"Error": "Gagal membaca konfigurasi import"})
		return
	}
	connectionJSON, err := json.Marshal(connection)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"Error": "Gagal menyiapkan halaman import"})
		return
	}

	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	c.HTML(http.StatusOK, "settings/import_data.html", gin.H{
		// This JSON is produced by encoding/json from server-side values; marking
		// it as JS prevents html/template from escaping the object expression.
		"ConnectionJSON": template.JS(connectionJSON),
		"User":           user,
		"Favicon":        favicon,
		"Title":          "Import Data",
		"ActiveMenu":     "import_data",
	})
}

func (h *ImportDataHandler) SaveConnection(c *gin.Context) {
	if err := h.importService.SaveConnection(connectionInput(c)); err != nil {
		log.Printf("Gagal menyimpan konfigurasi koneksi legacy: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": publicImportError(err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Koneksi sistem lama berhasil disimpan"})
}

func (h *ImportDataHandler) TestConnection(c *gin.Context) {
	if err := h.importService.TestConnection(c.Request.Context(), connectionInput(c)); err != nil {
		log.Printf("Uji koneksi legacy gagal: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Koneksi gagal. Periksa driver, host, port, database, username, dan password."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Koneksi ke sistem lama berhasil"})
}

func (h *ImportDataHandler) PreviewBBM(c *gin.Context) {
	rows, err := h.importService.PreviewBBM(c.Request.Context())
	if err != nil {
		log.Printf("Preview import BBM gagal: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": publicImportError(err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "data": rows, "count": len(rows)})
}

func (h *ImportDataHandler) DownloadBBMData(c *gin.Context) {
	rows, err := h.importService.PreviewBBM(c.Request.Context())
	if err != nil {
		log.Printf("Gagal mengambil data BBM dari sistem lama: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": publicImportError(err)})
		return
	}
	content, err := buildLegacyBBMXLSX(rows)
	if err != nil {
		log.Printf("Gagal membuat Excel data BBM: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal membuat file Excel data BBM"})
		return
	}
	c.Header("Content-Disposition", `attachment; filename="data_bbm_sistem_lama.xlsx"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", content)
}

func (h *ImportDataHandler) DownloadBBMTemplate(c *gin.Context) {
	content, err := buildLegacyBBMTemplateXLSX()
	if err != nil {
		log.Printf("Gagal membuat template Excel BBM: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal membuat template Excel BBM"})
		return
	}
	c.Header("Content-Disposition", `attachment; filename="template_import_bbm.xlsx"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", content)
}

func (h *ImportDataHandler) ImportBBM(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, importDataUploadMaxBytes)
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "File Excel BBM wajib dipilih"})
		return
	}
	if extension := strings.ToLower(filepath.Ext(fileHeader.Filename)); extension != ".xlsx" {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Format file harus Excel (.xlsx)"})
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "File Excel BBM tidak dapat dibaca"})
		return
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, importDataUploadMaxBytes+1))
	if err != nil || int64(len(content)) > importDataUploadMaxBytes {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Ukuran file import maksimal 5 MB"})
		return
	}
	rows, err := parseLegacyBBMXLSX(content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": err.Error()})
		return
	}

	var updatedBy *uint
	if userValue, exists := c.Get("user"); exists {
		switch user := userValue.(type) {
		case entity.User:
			updatedBy = &user.ID
		case *entity.User:
			if user != nil {
				updatedBy = &user.ID
			}
		}
	}

	result, err := h.importService.ImportBBMRows(c.Request.Context(), rows, updatedBy)
	if err != nil {
		log.Printf("Import BBM gagal: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": publicImportError(err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Import master BBM berhasil",
		"result":  result,
	})
}

func connectionInput(c *gin.Context) dto.LegacyConnectionInput {
	return dto.LegacyConnectionInput{
		Driver:       c.PostForm("driver"),
		Host:         c.PostForm("host"),
		Port:         c.PostForm("port"),
		DatabaseName: c.PostForm("database_name"),
		Username:     c.PostForm("username"),
		Password:     c.PostForm("password"),
		SSLMode:      c.PostForm("ssl_mode"),
	}
}

func publicImportError(err error) string {
	switch {
	case errors.Is(err, service.ErrLegacyConnectionNotConfigured):
		return "Koneksi sistem lama belum disimpan"
	case errors.Is(err, service.ErrUnsupportedLegacyDriver):
		return "Driver database belum didukung. Gunakan MySQL/MariaDB untuk import BBM."
	case errors.Is(err, gorm.ErrRecordNotFound):
		return "Konfigurasi koneksi belum ditemukan"
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return "Proses import dibatalkan atau melewati batas waktu"
	default:
		message := err.Error()
		for _, prefix := range []string{
			"host, database, dan username",
			"parameter koneksi",
			"port harus",
			"mode SSL",
			"kolom wajib",
			"baris ",
			"file Excel",
			"nama BBM",
			"nilai numerik BBM",
			"BBM dengan id",
		} {
			if strings.HasPrefix(message, prefix) {
				return message
			}
		}
		return "Operasi import gagal. Periksa konfigurasi koneksi dan format data."
	}
}
