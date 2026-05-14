package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"spbu_go/internal/dto"
	"spbu_go/internal/entity"
	"spbu_go/internal/service"

	"github.com/gin-gonic/gin"
)

// titleCase converts a string so each word starts with a capital letter
func titleCase(s string) string {
	words := strings.Fields(strings.TrimSpace(s))
	for i, w := range words {
		if len(w) > 0 {
			runes := []rune(strings.ToLower(w))
			runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

// sanitizePhone strips all non-digit characters from a phone number string
func sanitizePhone(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, s)
}

type KaryawanHandler struct {
	karyawanService  service.KaryawanService
	jabatanService   service.JabatanService
	pendapatanService service.PendapatanService
	potonganService  service.PotonganService
}

func NewKaryawanHandler(
	karyawanService service.KaryawanService,
	jabatanService service.JabatanService,
	pendapatanService service.PendapatanService,
	potonganService service.PotonganService,
) *KaryawanHandler {
	return &KaryawanHandler{karyawanService, jabatanService, pendapatanService, potonganService}
}

// kompGajiItem is a lightweight struct for JSON-encoding komponen gaji in the template
type kompGajiItem struct {
	ID    uint   `json:"id"`
	Nama  string `json:"nama"`
	Tipe  string `json:"tipe"`
	Nilai int64  `json:"nilai"`
}

// Index — halaman list karyawan aktif
func (h *KaryawanHandler) Index(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	jabatans, _ := h.jabatanService.GetAll()
	pendapatans, _ := h.pendapatanService.GetActive()
	potongans, _ := h.potonganService.GetActive()

	// Build lightweight JSON for template (avoids Go template syntax inside <script>)
	pItems := make([]kompGajiItem, len(pendapatans))
	for i, p := range pendapatans {
		pItems[i] = kompGajiItem{ID: p.ID, Nama: p.NamaPendapatan, Tipe: p.Tipe, Nilai: p.Nilai}
	}
	qItems := make([]kompGajiItem, len(potongans))
	for i, p := range potongans {
		qItems[i] = kompGajiItem{ID: p.ID, Nama: p.NamaPotongan, Tipe: p.Tipe, Nilai: p.Nilai}
	}
	pJSON, _ := json.Marshal(pItems)
	qJSON, _ := json.Marshal(qItems)

	c.HTML(http.StatusOK, "master/employee/index.html", gin.H{
		"Title":          "Data Karyawan",
		"ActiveMenu":     "master_employee",
		"IsArchive":      false,
		"User":           user,
		"Favicon":        favicon,
		"Jabatans":       jabatans,
		"Pendapatans":    pendapatans,
		"Potongans":      potongans,
		"PendapatansJSON": string(pJSON),
		"PotongansJSON":  string(qJSON),
	})
}

// Archive — halaman karyawan tidak aktif
func (h *KaryawanHandler) Archive(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")

	c.HTML(http.StatusOK, "master/employee/index.html", gin.H{
		"Title":      "Data Karyawan (Tidak Aktif)",
		"ActiveMenu": "master_employee",
		"IsArchive":  true,
		"User":       user,
		"Favicon":    favicon,
		"Jabatans":   []entity.Jabatan{},
	})
}

// Datatable — server-side datatable endpoint
func (h *KaryawanHandler) Datatable(c *gin.Context) {
	var req dto.DatatableRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isActive := c.Query("type") != "archive"
	total, filtered, data, err := h.karyawanService.Datatable(req, isActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data"})
		return
	}

	// Build response — sertakan nama jabatan untuk ditampilkan
	type Row struct {
		ID              uint    `json:"id"`
		NIK             string  `json:"nik"`
		NamaLengkap     string  `json:"nama_lengkap"`
		JenisKelamin    string  `json:"jenis_kelamin"`
		NoHP            string  `json:"no_hp"`
		NamaJabatan     string  `json:"nama_jabatan"`
		StatusNikah     string  `json:"status_nikah"`
		TglPengangkatan string  `json:"tgl_pengangkatan"`
		IsActive        bool    `json:"is_active"`
		Foto            string  `json:"foto"`
	}

	rows := make([]Row, len(data))
	for i, k := range data {
		namaJabatan := "-"
		if k.Jabatan != nil {
			namaJabatan = k.Jabatan.NamaJabatan
		}
		rows[i] = Row{
			ID:              k.ID,
			NIK:             k.NIK,
			NamaLengkap:     k.NamaLengkap,
			JenisKelamin:    k.JenisKelamin,
			NoHP:            k.NoHP,
			NamaJabatan:     namaJabatan,
			StatusNikah:     k.StatusNikah,
			TglPengangkatan: k.TglPengangkatan.Format("02/01/2006"),
			IsActive:        k.IsActive,
			Foto:            k.Foto,
		}
	}

	c.JSON(http.StatusOK, dto.DatatableResponse{
		Draw:            req.Draw,
		RecordsTotal:    total,
		RecordsFiltered: filtered,
		Data:            rows,
	})
}

// GetOne — get single karyawan data as JSON (for edit modal)
func (h *KaryawanHandler) GetOne(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	k, err := h.karyawanService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "Karyawan tidak ditemukan"})
		return
	}

	tglLahirStr := ""
	if !k.TanggalLahir.IsZero() {
		tglLahirStr = k.TanggalLahir.Format("2006-01-02")
	}
	tglAngkatStr := ""
	if !k.TglPengangkatan.IsZero() {
		tglAngkatStr = k.TglPengangkatan.Format("2006-01-02")
	}
	tglKeluarStr := ""
	if k.TglKeluar != nil && !k.TglKeluar.IsZero() {
		tglKeluarStr = k.TglKeluar.Format("2006-01-02")
	}

	// Build assigned IDs
	pendapatanIDs := make([]uint, len(k.Pendapatans))
	for i, p := range k.Pendapatans {
		pendapatanIDs[i] = p.ID
	}
	potonganIDs := make([]uint, len(k.Potongans))
	for i, p := range k.Potongans {
		potonganIDs[i] = p.ID
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data": gin.H{
			"id":               k.ID,
			"nik":              k.NIK,
			"nama_lengkap":     k.NamaLengkap,
			"gaji_pokok":       k.GajiPokok,
			"alamat":           k.Alamat,
			"tempat_lahir":     k.TempatLahir,
			"tanggal_lahir":    tglLahirStr,
			"status_nikah":     k.StatusNikah,
			"jumlah_anak":      k.JumlahAnak,
			"jabatan_id":       k.JabatanID,
			"jenis_kelamin":    k.JenisKelamin,
			"agama":            k.Agama,
			"no_hp":            k.NoHP,
			"pendidikan":       k.Pendidikan,
			"tgl_pengangkatan": tglAngkatStr,
			"tgl_keluar":       tglKeluarStr,
			"foto":             k.Foto,
			"is_active":        k.IsActive,
			"nama_jabatan":     func() string {
				if k.Jabatan != nil {
					return k.Jabatan.NamaJabatan
				}
				return ""
			}(),
			"pendapatan_ids":   pendapatanIDs,
			"potongan_ids":     potonganIDs,
			"pendapatans":      k.Pendapatans,
			"potongans":        k.Potongans,
		},
	})
}

// validateKaryawan performs backend validation and returns a map of field → error message.
// Returns nil when all fields are valid.
func validateKaryawan(
	nik, namaLengkap, jenisKelamin, statusNikah, noHP, tglPengangkatanStr, gajiStr string,
	jumlahAnak int,
) map[string]string {
	errors := map[string]string{}

	// NIK — wajib, 3–20 karakter, alfanumerik
	if nik == "" {
		errors["nik"] = "NIK wajib diisi"
	} else if len(nik) < 3 || len(nik) > 20 {
		errors["nik"] = "NIK harus antara 3–20 karakter"
	}

	// Nama Lengkap — wajib, 3–100 karakter
	if namaLengkap == "" {
		errors["nama_lengkap"] = "Nama lengkap wajib diisi"
	} else if len(namaLengkap) < 3 || len(namaLengkap) > 100 {
		errors["nama_lengkap"] = "Nama lengkap harus antara 3–100 karakter"
	}

	// Jenis Kelamin — wajib, L atau P
	if jenisKelamin == "" {
		errors["jenis_kelamin"] = "Jenis kelamin wajib dipilih"
	} else if jenisKelamin != "L" && jenisKelamin != "P" {
		errors["jenis_kelamin"] = "Jenis kelamin tidak valid"
	}

	// Status Nikah — wajib, nilai tertentu
	validStatus := map[string]bool{"lajang": true, "menikah": true, "cerai": true}
	if statusNikah == "" {
		errors["status_nikah"] = "Status nikah wajib dipilih"
	} else if !validStatus[statusNikah] {
		errors["status_nikah"] = "Status nikah tidak valid"
	}

	// Nomor HP — jika diisi, harus dimulai dengan 62 dan 10–15 digit
	if noHP != "" {
		if !strings.HasPrefix(noHP, "62") {
			errors["no_hp"] = "Nomor HP harus diawali 62"
		} else if len(noHP) < 10 || len(noHP) > 15 {
			errors["no_hp"] = "Nomor HP harus antara 10–15 digit"
		}
	}

	// Jumlah Anak — tidak boleh negatif
	if jumlahAnak < 0 {
		errors["jumlah_anak"] = "Jumlah anak tidak boleh negatif"
	}

	// Gaji Pokok — wajib, > 0
	gajiClean := strings.ReplaceAll(gajiStr, ".", "")
	gajiClean = strings.ReplaceAll(gajiClean, ",", ".")
	if gaji, err := strconv.ParseInt(gajiClean, 10, 64); err != nil || gaji <= 0 {
		errors["gaji_pokok"] = "Gaji pokok wajib diisi dan harus lebih dari 0"
	}

	// Tanggal Pengangkatan — wajib
	if tglPengangkatanStr == "" {
		errors["tgl_pengangkatan"] = "Tanggal pengangkatan wajib diisi"
	} else if _, err := time.Parse("2006-01-02", tglPengangkatanStr); err != nil {
		errors["tgl_pengangkatan"] = "Format tanggal pengangkatan tidak valid"
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

// Create — simpan karyawan baru
func (h *KaryawanHandler) Create(c *gin.Context) {
	var k entity.Karyawan

	k.NIK = strings.TrimSpace(c.PostForm("nik"))
	k.NamaLengkap = titleCase(c.PostForm("nama_lengkap"))
	k.Alamat = strings.TrimSpace(c.PostForm("alamat"))
	k.TempatLahir = titleCase(c.PostForm("tempat_lahir"))
	k.StatusNikah = c.PostForm("status_nikah")
	k.JenisKelamin = c.PostForm("jenis_kelamin")
	k.Agama = c.PostForm("agama")
	k.NoHP = sanitizePhone(c.PostForm("no_hp"))
	k.Pendidikan = c.PostForm("pendidikan")
	k.IsActive = true

	gajiStr := strings.ReplaceAll(c.PostForm("gaji_pokok"), ".", "")
	gajiStr = strings.ReplaceAll(gajiStr, ",", ".")
	if gaji, err := strconv.ParseInt(gajiStr, 10, 64); err == nil {
		k.GajiPokok = gaji
	}

	if jumlahAnak, err := strconv.Atoi(c.PostForm("jumlah_anak")); err == nil {
		k.JumlahAnak = jumlahAnak
	}

	// Backend validation
	if validationErrors := validateKaryawan(
		k.NIK, k.NamaLengkap, k.JenisKelamin, k.StatusNikah, k.NoHP,
		c.PostForm("tgl_pengangkatan"), c.PostForm("gaji_pokok"), k.JumlahAnak,
	); validationErrors != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status":  false,
			"message": "Periksa kembali data yang diisi",
			"errors":  validationErrors,
		})
		return
	}

	if jabatanIDStr := c.PostForm("jabatan_id"); jabatanIDStr != "" {
		if jabatanID, err := strconv.ParseUint(jabatanIDStr, 10, 32); err == nil {
			uid := uint(jabatanID)
			k.JabatanID = &uid
		}
	}

	if tglLahir, err := time.Parse("2006-01-02", c.PostForm("tanggal_lahir")); err == nil {
		k.TanggalLahir = tglLahir
	}
	if tglAngkat, err := time.Parse("2006-01-02", c.PostForm("tgl_pengangkatan")); err == nil {
		k.TglPengangkatan = tglAngkat
	}
	if tglKeluarStr := c.PostForm("tgl_keluar"); tglKeluarStr != "" {
		if tglKeluar, err := time.Parse("2006-01-02", tglKeluarStr); err == nil {
			k.TglKeluar = &tglKeluar
		}
	}
	// Set is_active from form, then override if tgl_keluar has already passed
	k.IsActive = c.PostForm("is_active") == "true"
	if k.TglKeluar != nil && !k.TglKeluar.After(time.Now()) {
		k.IsActive = false
	}

	// Upload foto
	if fotoPath, err := handleFotoUpload(c, "foto", 0); err == nil && fotoPath != "" {
		k.Foto = fotoPath
	}

	if userVal, exists := c.Get("user"); exists {
		if u, ok := userVal.(*entity.User); ok {
			k.UpdatedBy = &u.ID
		}
	}

	if err := h.karyawanService.Create(&k); err != nil {
		log.Printf("Error creating karyawan: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal menyimpan data karyawan"})
		return
	}

	// Set pendapatan and potongan associations
	h.syncKomponenGaji(c, k.ID)

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Karyawan berhasil ditambahkan"})
}

// Update — update data karyawan
func (h *KaryawanHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	k, err := h.karyawanService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": false, "message": "Karyawan tidak ditemukan"})
		return
	}

	k.NIK = strings.TrimSpace(c.PostForm("nik"))
	k.NamaLengkap = titleCase(c.PostForm("nama_lengkap"))
	k.Alamat = strings.TrimSpace(c.PostForm("alamat"))
	k.TempatLahir = titleCase(c.PostForm("tempat_lahir"))
	k.StatusNikah = c.PostForm("status_nikah")
	k.JenisKelamin = c.PostForm("jenis_kelamin")
	k.Agama = c.PostForm("agama")
	k.NoHP = sanitizePhone(c.PostForm("no_hp"))
	k.Pendidikan = c.PostForm("pendidikan")

	gajiStr := strings.ReplaceAll(c.PostForm("gaji_pokok"), ".", "")
	gajiStr = strings.ReplaceAll(gajiStr, ",", ".")
	if gaji, err := strconv.ParseInt(gajiStr, 10, 64); err == nil {
		k.GajiPokok = gaji
	}

	if jumlahAnak, err := strconv.Atoi(c.PostForm("jumlah_anak")); err == nil {
		k.JumlahAnak = jumlahAnak
	}

	// Backend validation
	if validationErrors := validateKaryawan(
		k.NIK, k.NamaLengkap, k.JenisKelamin, k.StatusNikah, k.NoHP,
		c.PostForm("tgl_pengangkatan"), c.PostForm("gaji_pokok"), k.JumlahAnak,
	); validationErrors != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status":  false,
			"message": "Periksa kembali data yang diisi",
			"errors":  validationErrors,
		})
		return
	}

	// Reset jabatan_id
	k.JabatanID = nil
	if jabatanIDStr := c.PostForm("jabatan_id"); jabatanIDStr != "" {
		if jabatanID, err := strconv.ParseUint(jabatanIDStr, 10, 32); err == nil {
			uid := uint(jabatanID)
			k.JabatanID = &uid
		}
	}

	if tglLahir, err := time.Parse("2006-01-02", c.PostForm("tanggal_lahir")); err == nil {
		k.TanggalLahir = tglLahir
	}
	if tglAngkat, err := time.Parse("2006-01-02", c.PostForm("tgl_pengangkatan")); err == nil {
		k.TglPengangkatan = tglAngkat
	}
	k.TglKeluar = nil
	if tglKeluarStr := c.PostForm("tgl_keluar"); tglKeluarStr != "" {
		if tglKeluar, err := time.Parse("2006-01-02", tglKeluarStr); err == nil {
			k.TglKeluar = &tglKeluar
		}
	}
	// Set is_active from form, then override if tgl_keluar has already passed
	k.IsActive = c.PostForm("is_active") == "true"
	if k.TglKeluar != nil && !k.TglKeluar.After(time.Now()) {
		k.IsActive = false
	}

	// Upload foto baru jika ada
	if fotoPath, err := handleFotoUpload(c, "foto", k.ID); err == nil && fotoPath != "" {
		k.Foto = fotoPath
	}

	if userVal, exists := c.Get("user"); exists {
		if u, ok := userVal.(*entity.User); ok {
			k.UpdatedBy = &u.ID
		}
	}

	if err := h.karyawanService.Update(&k); err != nil {
		log.Printf("Error updating karyawan: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal mengupdate data karyawan"})
		return
	}

	// Sync pendapatan and potongan associations
	h.syncKomponenGaji(c, k.ID)

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Data karyawan berhasil diupdate"})
}

// Delete — soft delete karyawan
func (h *KaryawanHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	if err := h.karyawanService.Delete(uint(id)); err != nil {
		log.Printf("Error deleting karyawan: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal menonaktifkan karyawan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Karyawan berhasil dinonaktifkan"})
}

// Restore — restore karyawan yang dinonaktifkan
func (h *KaryawanHandler) Restore(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	if err := h.karyawanService.Restore(uint(id)); err != nil {
		log.Printf("Error restoring karyawan: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal mengaktifkan karyawan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Karyawan berhasil diaktifkan kembali"})
}

// syncKomponenGaji — helper to sync pendapatan/potongan many2many associations from form data
func (h *KaryawanHandler) syncKomponenGaji(c *gin.Context, karyawanID uint) {
	// Parse pendapatan_ids[]
	pendapatanStrs := c.PostFormArray("pendapatan_ids[]")
	var pendapatanIDs []uint
	for _, s := range pendapatanStrs {
		if id, err := strconv.ParseUint(s, 10, 32); err == nil {
			pendapatanIDs = append(pendapatanIDs, uint(id))
		}
	}
	if err := h.karyawanService.SetPendapatans(karyawanID, pendapatanIDs); err != nil {
		log.Printf("Error setting pendapatans for karyawan %d: %v", karyawanID, err)
	}

	// Parse potongan_ids[]
	potonganStrs := c.PostFormArray("potongan_ids[]")
	var potonganIDs []uint
	for _, s := range potonganStrs {
		if id, err := strconv.ParseUint(s, 10, 32); err == nil {
			potonganIDs = append(potonganIDs, uint(id))
		}
	}
	if err := h.karyawanService.SetPotongans(karyawanID, potonganIDs); err != nil {
		log.Printf("Error setting potongans for karyawan %d: %v", karyawanID, err)
	}
}

// handleFotoUpload — helper untuk upload foto karyawan
func handleFotoUpload(c *gin.Context, fieldName string, karyawanID uint) (string, error) {
	file, err := c.FormFile(fieldName)
	if err != nil {
		return "", err // tidak ada file — bukan error kritis
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[ext] {
		return "", fmt.Errorf("tipe file tidak diizinkan")
	}

	uploadDir := "./static/uploads/karyawan"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%d_%d%s", karyawanID, time.Now().Unix(), ext)
	savePath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		return "", err
	}

	return "/static/uploads/karyawan/" + filename, nil
}
