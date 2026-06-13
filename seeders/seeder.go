package seeders

import (
	"fmt"
	"log"
	"spbu_go/internal/entity"
	"spbu_go/pkg/database"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func Seed() {
	db := database.DB

	// Roles
	adminRole := entity.Role{Name: "Administrator", Code: "admin"}
	staffRole := entity.Role{Name: "Staff", Code: "staff"}

	if err := db.FirstOrCreate(&adminRole, entity.Role{Code: "admin"}).Error; err != nil {
		log.Printf("Failed to seed admin role: %v", err)
	}
	if err := db.FirstOrCreate(&staffRole, entity.Role{Code: "staff"}).Error; err != nil {
		log.Printf("Failed to seed staff role: %v", err)
	}

	// Permissions
	perms := []entity.Permission{
		{Name: "Manage Users", Code: "user_manage"},
		{Name: "Manage Roles", Code: "role_manage"},
		{Name: "View Dashboard", Code: "dashboard_view"},
	}

	for i := range perms {
		if err := db.Where(entity.Permission{Code: perms[i].Code}).FirstOrCreate(&perms[i]).Error; err != nil {
			log.Printf("Failed to seed permission %s: %v", perms[i].Code, err)
		}
	}

	// Assign Permissions to Admin
	if err := db.Model(&adminRole).Association("Permissions").Replace(perms); err != nil {
		log.Printf("Failed to assign permissions to admin: %v", err)
	}

	// Users
	password, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	adminUser := entity.User{
		FirstName: "Admin",
		LastName:  "User",
		Username:  "admin",
		Password:  string(password),
		Email:     "admin@spbu.com",
		IsActive:  true,
		RoleID:    adminRole.ID,
	}

	if err := db.FirstOrCreate(&adminUser, entity.User{Username: "admin"}).Error; err != nil {
		log.Printf("Failed to seed admin user: %v", err)
	}

	// BBM Data
	bbms := []entity.BBM{
		{Name: "Bio Solar", Margin: 280, Price: 6800, Stock: 652161, RewardPercent: 2, IsActive: true},
		{Name: "DEX", Margin: 705, Price: 13500, Stock: 1554321521, RewardPercent: 1, IsActive: true},
		{Name: "Pertalite", Margin: 385, Price: 10000, Stock: 3340426, RewardPercent: 4, IsActive: true},
		{Name: "Pertamax", Margin: 690, Price: 11800, Stock: 1621304313, RewardPercent: 1, IsActive: true},
		{Name: "Premium", Margin: 262, Price: 6450, Stock: 0, RewardPercent: 3, IsActive: false},
		{Name: "Pertalite Khusus", Margin: 360, Price: 7250, Stock: 0, RewardPercent: 5, IsActive: true},
		{Name: "Pertamax Turbo", Margin: 655, Price: 14500, Stock: 0, RewardPercent: 20, IsActive: true},
		{Name: "Dexlite", Margin: 615, Price: 13320, Stock: 0, RewardPercent: 30, IsActive: true},
	}

	for i := range bbms {
		if err := db.Where(entity.BBM{Name: bbms[i].Name}).FirstOrCreate(&bbms[i]).Error; err != nil {
			log.Printf("Failed to seed bbm %s: %v", bbms[i].Name, err)
		}
	}

	// Tiang Data
	tiangs := []entity.Tiang{
		{Name: "Tiang 1", Slug: "tiang1"},
		{Name: "Tiang 2", Slug: "tiang2"},
		{Name: "Tiang 3", Slug: "tiang3"},
		{Name: "Tiang 4", Slug: "tiang4"},
	}

	for i := range tiangs {
		if err := db.Where(entity.Tiang{Slug: tiangs[i].Slug}).FirstOrCreate(&tiangs[i]).Error; err != nil {
			log.Printf("Failed to seed tiang %s: %v", tiangs[i].Name, err)
		}
	}

	// Nozzle Data (Relies on IDs, assuming sequential seeding or lookups)
	// We'll look up IDs to be safe
	var t1, t2, t3, t4 entity.Tiang
	db.Where("slug = ?", "tiang1").First(&t1)
	db.Where("slug = ?", "tiang2").First(&t2)
	db.Where("slug = ?", "tiang3").First(&t3)
	db.Where("slug = ?", "tiang4").First(&t4)

	var b1, b2, b3, b4 entity.BBM
	db.Where("name = ?", "Bio Solar").First(&b1) // ID 1
	db.Where("name = ?", "DEX").First(&b2)       // ID 2
	db.Where("name = ?", "Pertalite").First(&b3) // ID 3
	db.Where("name = ?", "Pertamax").First(&b4)  // ID 4

	nozzles := []entity.Nozzle{
		{TiangID: t1.ID, Description: "1A", BBMID: b4.ID, IsActive: true},
		{TiangID: t1.ID, Description: "1B", BBMID: b4.ID, IsActive: true},
		{TiangID: t1.ID, Description: "1C", BBMID: b3.ID, IsActive: true},
		{TiangID: t1.ID, Description: "1D", BBMID: b3.ID, IsActive: true},
		{TiangID: t2.ID, Description: "2A", BBMID: b4.ID, IsActive: true},
		{TiangID: t2.ID, Description: "2B", BBMID: b4.ID, IsActive: true},
		{TiangID: t2.ID, Description: "2C", BBMID: b1.ID, IsActive: true},
		{TiangID: t2.ID, Description: "2D", BBMID: b1.ID, IsActive: true},
		{TiangID: t3.ID, Description: "3A", BBMID: b3.ID, IsActive: true},
		{TiangID: t3.ID, Description: "3B", BBMID: b3.ID, IsActive: true},
		{TiangID: t3.ID, Description: "3C", BBMID: b3.ID, IsActive: true},
		{TiangID: t3.ID, Description: "3D", BBMID: b3.ID, IsActive: true},
		{TiangID: t4.ID, Description: "4A", BBMID: b3.ID, IsActive: true},
		{TiangID: t4.ID, Description: "4B", BBMID: b3.ID, IsActive: true},
		{TiangID: t4.ID, Description: "4C", BBMID: b2.ID, IsActive: true},
		{TiangID: t4.ID, Description: "4D", BBMID: b2.ID, IsActive: true},
	}

	for i := range nozzles {
		// Use FirstOrCreate to avoid duplicates if re-seeding
		if err := db.Where(entity.Nozzle{Description: nozzles[i].Description, TiangID: nozzles[i].TiangID}).FirstOrCreate(&nozzles[i]).Error; err != nil {
			log.Printf("Failed to seed nozzle %s: %v", nozzles[i].Description, err)
		}
	}

	// --- Partner Seeder ---
	partners := []entity.Partner{
		{
			Name:          "PT. Energi Nusantara",
			ContactPerson: "Budi Santoso",
			Phone:         "081234567890",
			Address:       "Jl. Sudirman No. 123, Jakarta",
			IsActive:      true,
		},
		{
			Name:          "CV. Maju Jaya",
			ContactPerson: "Siti Aminah",
			Phone:         "089876543210",
			Address:       "Kawasan Industri Jababeka Blok A, Bekasi",
			IsActive:      true,
		},
		{
			Name:          "KopKar Pertamina",
			ContactPerson: "Andi Wijaya",
			Phone:         "085512345678",
			Address:       "Jl. Yos Sudarso, Plumpang, Jakarta Utara",
			IsActive:      false, // Inactive example
		},
	}

	for i := range partners {
		if err := db.Where(entity.Partner{Name: partners[i].Name}).FirstOrCreate(&partners[i]).Error; err != nil {
			log.Printf("Failed to seed partner %s: %v", partners[i].Name, err)
		}
	}

	// --- Shift Seeder ---
	seedShifts()

	// --- Penebusan Seeder --- (must be called after BBM seeding)
	seedPenebusan()

	// --- Employee & Jabatan Seeder ---
	seedEmployees()

	// --- Kedatangan BBM Seeder --- (must run after penebusan + shift seeding)
	seedKedatanganBBM()

	// --- Penjualan Seeder --- (must run after nozzle + shift seeding)
	seedPenjualan()

	// --- Piutang Seeder --- (must run after penjualan + partner + BBM seeding)
	seedPiutang()

	// --- Penyusutan Seeder --- (must run after BBM + shift seeding)
	seedPenyusutan()

	// --- Jenis Test Seeder ---
	seedJenisTest()

	log.Println("Database seeded successfully")
}

func parseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func parseDatePtr(s string) *time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}

func seedEmployees() {
	db := database.DB

	// Seed Jabatan
	jabatans := []entity.Jabatan{
		{KodeJabatan: "OPR", NamaJabatan: "Operator", RewardPersen: 2.5},
		{KodeJabatan: "KSR", NamaJabatan: "Kasir", RewardPersen: 2.0},
		{KodeJabatan: "SPV", NamaJabatan: "Supervisor", RewardPersen: 3.5},
		{KodeJabatan: "MNJ", NamaJabatan: "Manajer", RewardPersen: 5.0},
		{KodeJabatan: "SEC", NamaJabatan: "Security", RewardPersen: 1.5},
	}
	for i := range jabatans {
		db.Where(entity.Jabatan{KodeJabatan: jabatans[i].KodeJabatan}).FirstOrCreate(&jabatans[i])
	}

	// Reload jabatans to get IDs
	var jOpr, jKsr, jSpv, jMnj, jSec entity.Jabatan
	db.Where("kode_jabatan = ?", "OPR").First(&jOpr)
	db.Where("kode_jabatan = ?", "KSR").First(&jKsr)
	db.Where("kode_jabatan = ?", "SPV").First(&jSpv)
	db.Where("kode_jabatan = ?", "MNJ").First(&jMnj)
	db.Where("kode_jabatan = ?", "SEC").First(&jSec)

	karyawans := []entity.Karyawan{
		{NIK: "KRY001", NamaLengkap: "Budi Setiawan", GajiPokok: 2500000, Alamat: "Jl. Mawar No. 1, Bandung", TempatLahir: "Bandung", TanggalLahir: parseDate("1990-03-15"), StatusNikah: "menikah", JumlahAnak: 2, JabatanID: &jOpr.ID, JenisKelamin: "L", Agama: "Islam", NoHP: "081211110001", Pendidikan: "SMA", TglPengangkatan: parseDate("2018-01-10"), IsActive: true},
		{NIK: "KRY002", NamaLengkap: "Siti Rahayu", GajiPokok: 2500000, Alamat: "Jl. Melati No. 5, Bandung", TempatLahir: "Sumedang", TanggalLahir: parseDate("1995-07-22"), StatusNikah: "lajang", JumlahAnak: 0, JabatanID: &jOpr.ID, JenisKelamin: "P", Agama: "Islam", NoHP: "081211110002", Pendidikan: "SMA", TglPengangkatan: parseDate("2020-03-01"), IsActive: true},
		{NIK: "KRY003", NamaLengkap: "Ahmad Fauzi", GajiPokok: 2800000, Alamat: "Jl. Kenanga No. 10, Cimahi", TempatLahir: "Cimahi", TanggalLahir: parseDate("1988-11-05"), StatusNikah: "menikah", JumlahAnak: 3, JabatanID: &jKsr.ID, JenisKelamin: "L", Agama: "Islam", NoHP: "081211110003", Pendidikan: "D3", TglPengangkatan: parseDate("2016-06-15"), IsActive: true},
		{NIK: "KRY004", NamaLengkap: "Dewi Kusuma", GajiPokok: 2800000, Alamat: "Perum Griya Asri Blok B12, Bandung", TempatLahir: "Jakarta", TanggalLahir: parseDate("1993-02-28"), StatusNikah: "menikah", JumlahAnak: 1, JabatanID: &jKsr.ID, JenisKelamin: "P", Agama: "Kristen", NoHP: "081211110004", Pendidikan: "D3", TglPengangkatan: parseDate("2019-09-01"), IsActive: true},
		{NIK: "KRY005", NamaLengkap: "Hendra Gunawan", GajiPokok: 3500000, Alamat: "Jl. Cihampelas No. 88, Bandung", TempatLahir: "Garut", TanggalLahir: parseDate("1985-05-17"), StatusNikah: "menikah", JumlahAnak: 2, JabatanID: &jSpv.ID, JenisKelamin: "L", Agama: "Islam", NoHP: "081211110005", Pendidikan: "S1", TglPengangkatan: parseDate("2013-04-10"), IsActive: true},
		{NIK: "KRY006", NamaLengkap: "Rina Marliana", GajiPokok: 2500000, Alamat: "Jl. Antapani No. 3, Bandung", TempatLahir: "Tasikmalaya", TanggalLahir: parseDate("1997-09-10"), StatusNikah: "lajang", JumlahAnak: 0, JabatanID: &jOpr.ID, JenisKelamin: "P", Agama: "Islam", NoHP: "081211110006", Pendidikan: "SMA", TglPengangkatan: parseDate("2021-01-05"), IsActive: true},
		{NIK: "KRY007", NamaLengkap: "Dedi Kurniawan", GajiPokok: 2500000, Alamat: "Jl. Dago No. 20, Bandung", TempatLahir: "Bandung", TanggalLahir: parseDate("1992-12-01"), StatusNikah: "menikah", JumlahAnak: 1, JabatanID: &jOpr.ID, JenisKelamin: "L", Agama: "Islam", NoHP: "081211110007", Pendidikan: "SMA", TglPengangkatan: parseDate("2018-08-20"), IsActive: true},
		{NIK: "KRY008", NamaLengkap: "Fitri Handayani", GajiPokok: 2800000, Alamat: "Jl. Pajajaran No. 55, Bogor", TempatLahir: "Bogor", TanggalLahir: parseDate("1994-04-14"), StatusNikah: "menikah", JumlahAnak: 2, JabatanID: &jKsr.ID, JenisKelamin: "P", Agama: "Islam", NoHP: "081211110008", Pendidikan: "D3", TglPengangkatan: parseDate("2017-11-01"), IsActive: true},
		{NIK: "KRY009", NamaLengkap: "Ridwan Saputra", GajiPokok: 2500000, Alamat: "Jl. Raya Cileunyi No. 7, Bandung", TempatLahir: "Purwakarta", TanggalLahir: parseDate("1996-06-30"), StatusNikah: "lajang", JumlahAnak: 0, JabatanID: &jOpr.ID, JenisKelamin: "L", Agama: "Islam", NoHP: "081211110009", Pendidikan: "SMA", TglPengangkatan: parseDate("2022-02-14"), IsActive: true},
		{NIK: "KRY010", NamaLengkap: "Nadia Permata", GajiPokok: 2500000, Alamat: "Jl. Kiaracondong No. 12, Bandung", TempatLahir: "Bandung", TanggalLahir: parseDate("1999-01-25"), StatusNikah: "lajang", JumlahAnak: 0, JabatanID: &jOpr.ID, JenisKelamin: "P", Agama: "Islam", NoHP: "081211110010", Pendidikan: "SMA", TglPengangkatan: parseDate("2023-03-01"), IsActive: true},
		{NIK: "KRY011", NamaLengkap: "Agus Prabowo", GajiPokok: 2300000, Alamat: "Jl. Gedebage No. 4, Bandung", TempatLahir: "Surabaya", TanggalLahir: parseDate("1991-08-18"), StatusNikah: "cerai", JumlahAnak: 1, JabatanID: &jSec.ID, JenisKelamin: "L", Agama: "Islam", NoHP: "081211110011", Pendidikan: "SMA", TglPengangkatan: parseDate("2019-05-12"), IsActive: true},
		{NIK: "KRY012", NamaLengkap: "Yuli Astuti", GajiPokok: 2500000, Alamat: "Jl. Buah Batu No. 9, Bandung", TempatLahir: "Yogyakarta", TanggalLahir: parseDate("1993-10-07"), StatusNikah: "menikah", JumlahAnak: 1, JabatanID: &jOpr.ID, JenisKelamin: "P", Agama: "Katolik", NoHP: "081211110012", Pendidikan: "SMA", TglPengangkatan: parseDate("2020-07-01"), IsActive: true},
		{NIK: "KRY013", NamaLengkap: "Fajar Nugroho", GajiPokok: 5000000, Alamat: "Jl. Diponegoro No. 100, Bandung", TempatLahir: "Semarang", TanggalLahir: parseDate("1982-03-22"), StatusNikah: "menikah", JumlahAnak: 3, JabatanID: &jMnj.ID, JenisKelamin: "L", Agama: "Islam", NoHP: "081211110013", Pendidikan: "S1", TglPengangkatan: parseDate("2010-01-15"), IsActive: true},
		{NIK: "KRY014", NamaLengkap: "Mega Wulandari", GajiPokok: 3500000, Alamat: "Jl. Pasteur No. 25, Bandung", TempatLahir: "Bandung", TanggalLahir: parseDate("1987-12-10"), StatusNikah: "menikah", JumlahAnak: 2, JabatanID: &jSpv.ID, JenisKelamin: "P", Agama: "Islam", NoHP: "081211110014", Pendidikan: "S1", TglPengangkatan: parseDate("2015-08-01"), IsActive: true},
		{NIK: "KRY015", NamaLengkap: "Irfan Hakim", GajiPokok: 2300000, Alamat: "Jl. Soekarno Hatta No. 77, Bandung", TempatLahir: "Cianjur", TanggalLahir: parseDate("1998-05-05"), StatusNikah: "lajang", JumlahAnak: 0, JabatanID: &jSec.ID, JenisKelamin: "L", Agama: "Islam", NoHP: "081211110015", Pendidikan: "SMA", TglPengangkatan: parseDate("2022-09-01"), IsActive: true},
		{NIK: "KRY016", NamaLengkap: "Sri Wahyuni", GajiPokok: 2500000, Alamat: "Jl. Gatot Subroto No. 33, Bandung", TempatLahir: "Solo", TanggalLahir: parseDate("1996-02-14"), StatusNikah: "lajang", JumlahAnak: 0, JabatanID: &jOpr.ID, JenisKelamin: "P", Agama: "Islam", NoHP: "081211110016", Pendidikan: "SMA", TglPengangkatan: parseDate("2021-06-15"), IsActive: true},
		{NIK: "KRY017", NamaLengkap: "Bambang Susilo", GajiPokok: 2500000, Alamat: "Jl. Ahmad Yani No. 45, Cimahi", TempatLahir: "Cilacap", TanggalLahir: parseDate("1990-09-09"), StatusNikah: "menikah", JumlahAnak: 2, JabatanID: &jOpr.ID, JenisKelamin: "L", Agama: "Islam", NoHP: "081211110017", Pendidikan: "SMA", TglPengangkatan: parseDate("2017-03-20"), IsActive: true},
		{NIK: "KRY018", NamaLengkap: "Lia Amelia", GajiPokok: 2800000, Alamat: "Perum Cipageran Indah Blok C5, Cimahi", TempatLahir: "Cimahi", TanggalLahir: parseDate("1994-07-17"), StatusNikah: "menikah", JumlahAnak: 1, JabatanID: &jKsr.ID, JenisKelamin: "P", Agama: "Islam", NoHP: "081211110018", Pendidikan: "D3", TglPengangkatan: parseDate("2018-10-01"), IsActive: true},
		{NIK: "KRY019", NamaLengkap: "Rudi Hartono", GajiPokok: 2300000, Alamat: "Jl. Baros No. 8, Cimahi", TempatLahir: "Sukabumi", TanggalLahir: parseDate("1989-11-23"), StatusNikah: "cerai", JumlahAnak: 0, JabatanID: &jSec.ID, JenisKelamin: "L", Agama: "Islam", NoHP: "081211110019", Pendidikan: "SMA", TglPengangkatan: parseDate("2016-02-01"), TglKeluar: parseDatePtr("2024-12-31"), IsActive: false},
		{NIK: "KRY020", NamaLengkap: "Anggraeni Putri", GajiPokok: 2500000, Alamat: "Jl. Raya Padalarang No. 15, Bandung Barat", TempatLahir: "Bandung", TanggalLahir: parseDate("2000-04-20"), StatusNikah: "lajang", JumlahAnak: 0, JabatanID: &jOpr.ID, JenisKelamin: "P", Agama: "Islam", NoHP: "081211110020", Pendidikan: "SMA", TglPengangkatan: parseDate("2023-08-01"), IsActive: true},
	}

	for i := range karyawans {
		var existing entity.Karyawan
		// Use Unscoped so soft-deleted records are also found, preventing unique constraint violations
		if err := db.Unscoped().Where("nik = ?", karyawans[i].NIK).First(&existing).Error; err != nil {
			// Record doesn't exist at all — create it
			if err2 := db.Omit("Updater", "Jabatan", "Pendapatans", "Potongans").Create(&karyawans[i]).Error; err2 != nil {
				log.Printf("Failed to seed karyawan %s: %v", karyawans[i].NIK, err2)
			}
		}
	}

	log.Println("Employee data seeded successfully")

	// Seed Pendapatan
	pendapatans := []entity.Pendapatan{
		{NamaPendapatan: "Tunjangan Makan", Tipe: "nominal", Nilai: 150000, Deskripsi: "Tunjangan makan harian untuk seluruh karyawan", IsActive: true},
		{NamaPendapatan: "Tunjangan Transport", Tipe: "nominal", Nilai: 200000, Deskripsi: "Tunjangan biaya transportasi bulanan", IsActive: true},
		{NamaPendapatan: "Tunjangan Jabatan", Tipe: "nominal", Nilai: 500000, Deskripsi: "Tunjangan berdasarkan jabatan karyawan", IsActive: true},
		{NamaPendapatan: "Bonus Kinerja", Tipe: "persen", Nilai: 10, Deskripsi: "Bonus berdasarkan persentase gaji pokok atas pencapaian target", IsActive: true},
		{NamaPendapatan: "Tunjangan Kehadiran", Tipe: "nominal", Nilai: 100000, Deskripsi: "Tunjangan untuk karyawan dengan kehadiran penuh", IsActive: true},
		{NamaPendapatan: "Tunjangan Lembur", Tipe: "nominal", Nilai: 75000, Deskripsi: "Kompensasi per hari kerja lembur", IsActive: true},
		{NamaPendapatan: "Tunjangan Hari Raya", Tipe: "persen", Nilai: 100, Deskripsi: "THR setara satu bulan gaji pokok", IsActive: true},
		{NamaPendapatan: "Tunjangan Komunikasi", Tipe: "nominal", Nilai: 50000, Deskripsi: "Pulsa/paket data untuk kebutuhan operasional", IsActive: false},
	}
	for i := range pendapatans {
		db.Where(entity.Pendapatan{NamaPendapatan: pendapatans[i].NamaPendapatan}).FirstOrCreate(&pendapatans[i])
	}

	// Seed Potongan
	potongans := []entity.Potongan{
		{KodePotongan: "BPJSK", NamaPotongan: "Potongan BPJS Kesehatan", Tipe: "persen", Nilai: 1, Deskripsi: "Iuran BPJS Kesehatan 1% dari gaji pokok (karyawan)", IsActive: true},
		{KodePotongan: "BPJST", NamaPotongan: "Potongan BPJS Tenaga Kerja", Tipe: "persen", Nilai: 2, Deskripsi: "Iuran BPJS Ketenagakerjaan 2% dari gaji pokok (karyawan)", IsActive: true},
		{KodePotongan: "PPH21", NamaPotongan: "Potongan Pph 21", Tipe: "persen", Nilai: 5, Deskripsi: "Pajak Penghasilan Pasal 21 sesuai tarif berlaku", IsActive: true},
		{KodePotongan: "KASBON", NamaPotongan: "Cicilan Kasbon", Tipe: "nominal", Nilai: 0, Deskripsi: "Potongan cicilan kasbon / pinjaman karyawan (nilai dinamis)", IsActive: true},
		{KodePotongan: "ABSEN", NamaPotongan: "Potongan Absen", Tipe: "nominal", Nilai: 50000, Deskripsi: "Potongan per hari tidak hadir tanpa keterangan", IsActive: true},
		{KodePotongan: "KETRL", NamaPotongan: "Potongan Keterlambatan", Tipe: "nominal", Nilai: 25000, Deskripsi: "Potongan per kejadian terlambat masuk kerja", IsActive: false},
	}
	for i := range potongans {
		db.Where(entity.Potongan{KodePotongan: potongans[i].KodePotongan}).FirstOrCreate(&potongans[i])
	}

	// Seed Wallet
	wallets := []entity.Wallet{
		{WalletName: "KAS", IsDefault: true, Description: "Kas tunai operasional", Saldo: 0},
		{WalletName: "Bank Mandiri", IsDefault: false, Description: "Rekening Bank Mandiri", Saldo: 0},
		{WalletName: "Bank BCA", IsDefault: false, Description: "Rekening Bank BCA", Saldo: 0},
	}
	for i := range wallets {
		db.Where(entity.Wallet{WalletName: wallets[i].WalletName}).FirstOrCreate(&wallets[i])
	}
	log.Println("Wallet data seeded successfully")

	// ─── Seed COA Types ──────────────────────────────────────────────────────
	type coaTypeItem struct {
		Code          string
		Name          string
		NormalBalance string
		Description   string
	}
	coaTypeItems := []coaTypeItem{
		{"1", "Aset", "debit", "Harta dan kekayaan yang dimiliki perusahaan"},
		{"2", "Kewajiban", "credit", "Hutang dan kewajiban perusahaan kepada pihak lain"},
		{"3", "Modal / Ekuitas", "credit", "Modal pemilik dan laba ditahan"},
		{"4", "Pendapatan", "credit", "Pendapatan dari kegiatan operasional SPBU"},
		{"5", "Harga Pokok Penjualan", "debit", "Biaya langsung terkait produk BBM yang dijual"},
		{"6", "Beban Operasional", "debit", "Biaya operasional, personalia, dan administrasi"},
	}
	typeMap := map[string]uint{} // code -> ID
	for _, t := range coaTypeItems {
		ct := entity.COAType{Code: t.Code, Name: t.Name, NormalBalance: t.NormalBalance, Description: t.Description, IsActive: true}
		db.Where("code = ?", t.Code).FirstOrCreate(&ct)
		typeMap[t.Code] = ct.ID
	}

	// ─── Seed COA Accounts ───────────────────────────────────────────────────
	type coaItem struct {
		TypeCode    string
		Code        string
		Name        string
		Description string
		IsHeader    bool
		IsSystem    bool
	}
	coaItems := []coaItem{
		// ── 1. ASET ──
		{"1", "1100", "Aset Lancar", "Kelompok aset yang dapat dicairkan dalam setahun", true, true},
		{"1", "1101", "Kas", "Kas tunai operasional SPBU", false, true},
		{"1", "1102", "Bank Mandiri", "Rekening Bank Mandiri SPBU", false, true},
		{"1", "1103", "Bank BCA", "Rekening Bank BCA SPBU", false, true},
		{"1", "1110", "Piutang Dagang B2B", "Piutang penjualan BBM kredit ke partner", false, true},
		{"1", "1111", "Piutang Kasbon Karyawan", "Pinjaman kasbon yang belum dilunasi karyawan", false, true},
		{"1", "1120", "Persediaan BBM", "Kelompok persediaan bahan bakar minyak", true, true},
		{"1", "1121", "Persediaan BBM — Pertalite", "Stok Pertalite di tangki SPBU", false, true},
		{"1", "1122", "Persediaan BBM — Pertamax", "Stok Pertamax di tangki SPBU", false, true},
		{"1", "1123", "Persediaan BBM — Pertamax Turbo", "Stok Pertamax Turbo di tangki SPBU", false, true},
		{"1", "1124", "Persediaan BBM — Pertamina Dex", "Stok Pertamina Dex di tangki SPBU", false, true},
		{"1", "1125", "Persediaan BBM — Dexlite", "Stok Dexlite di tangki SPBU", false, true},
		{"1", "1131", "Uang Muka Penebusan Pertamina", "Pembayaran ke Pertamina sebelum BBM tiba", false, true},
		{"1", "1200", "Aset Tetap", "Kelompok aset tetap berwujud", true, true},
		{"1", "1201", "Tanah", "Nilai tanah lokasi SPBU", false, false},
		{"1", "1202", "Bangunan", "Nilai bangunan dan konstruksi SPBU", false, false},
		{"1", "1203", "Mesin & Peralatan Dispenser", "Mesin pompa BBM dan dispenser", false, false},
		{"1", "1204", "Inventaris Kantor", "Peralatan dan inventaris kantor operasional", false, false},
		{"1", "1211", "Akumulasi Penyusutan Bangunan", "Contra asset — penyusutan bangunan", false, false},
		{"1", "1212", "Akumulasi Penyusutan Mesin & Dispenser", "Contra asset — penyusutan mesin dispenser", false, false},
		{"1", "1213", "Akumulasi Penyusutan Inventaris", "Contra asset — penyusutan inventaris", false, false},

		// ── 2. KEWAJIBAN ──
		// 2101 Hutang ke Pertamina dihapus: penebusan dibayar online langsung (tidak ada hutang)
		{"2", "2102", "Hutang BPJS Kesehatan", "Iuran BPJS Kesehatan yang belum dibayarkan", false, true},
		{"2", "2103", "Hutang BPJS Ketenagakerjaan", "Iuran BPJS TK yang belum dibayarkan", false, true},
		{"2", "2104", "Hutang Gaji Karyawan", "Gaji karyawan yang sudah dihitung namun belum dibayar", false, true},
		{"2", "2105", "Hutang Pajak PPh 21", "Pajak penghasilan karyawan yang belum disetorkan ke kas negara", false, true},

		// ── 3. MODAL / EKUITAS ──
		{"3", "3101", "Modal Disetor", "Modal yang disetorkan pemilik/pemegang saham", false, true},
		{"3", "3102", "Laba Ditahan", "Akumulasi laba dari tahun-tahun sebelumnya", false, true},
		{"3", "3103", "Ikhtisar Laba / Rugi", "Akun penutup untuk perhitungan laba/rugi berjalan", false, true},

		// ── 4. PENDAPATAN ──
		{"4", "4100", "Pendapatan Penjualan BBM — Tunai", "Kelompok pendapatan dari penjualan tunai harian", true, true},
		{"4", "4101", "Pendapatan Penjualan Pertalite — Tunai", "Omzet Pertalite dari totalisator nozzle (tunai)", false, true},
		{"4", "4102", "Pendapatan Penjualan Pertamax — Tunai", "Omzet Pertamax dari totalisator nozzle (tunai)", false, true},
		{"4", "4103", "Pendapatan Penjualan Pertamax Turbo — Tunai", "Omzet Pertamax Turbo dari totalisator nozzle (tunai)", false, true},
		{"4", "4104", "Pendapatan Penjualan Pertamina Dex — Tunai", "Omzet Pertamina Dex dari totalisator nozzle (tunai)", false, true},
		{"4", "4105", "Pendapatan Penjualan Dexlite — Tunai", "Omzet Dexlite dari totalisator nozzle (tunai)", false, true},
		{"4", "4110", "Pendapatan Penjualan BBM — Kredit B2B", "Kelompok pendapatan dari penjualan kredit ke partner", true, true},
		{"4", "4111", "Pendapatan Penjualan Pertalite — Kredit B2B", "Omzet Pertalite ke partner (kredit/piutang)", false, true},
		{"4", "4112", "Pendapatan Penjualan Pertamax — Kredit B2B", "Omzet Pertamax ke partner (kredit/piutang)", false, true},
		{"4", "4113", "Pendapatan Penjualan Pertamax Turbo — Kredit B2B", "Omzet Pertamax Turbo ke partner (kredit/piutang)", false, true},
		{"4", "4114", "Pendapatan Penjualan Pertamina Dex — Kredit B2B", "Omzet Pertamina Dex ke partner (kredit/piutang)", false, true},
		{"4", "4115", "Pendapatan Penjualan Dexlite — Kredit B2B", "Omzet Dexlite ke partner (kredit/piutang)", false, true},
		{"4", "4120", "Pendapatan Lain-lain", "Kelompok pendapatan di luar penjualan BBM", true, false},
		{"4", "4121", "Pendapatan Bunga Bank", "Jasa giro/bunga dari rekening bank SPBU", false, false},
		{"4", "4122", "Pendapatan Non-BBM Lainnya", "Pendapatan insidental di luar kategori utama", false, false},

		// ── 5. HARGA POKOK PENJUALAN ──
		{"5", "5100", "Harga Pokok Penjualan BBM", "Kelompok HPP per jenis BBM", true, true},
		{"5", "5101", "HPP Pertalite", "Harga dasar Pertamina × liter terjual (Pertalite)", false, true},
		{"5", "5102", "HPP Pertamax", "Harga dasar Pertamina × liter terjual (Pertamax)", false, true},
		{"5", "5103", "HPP Pertamax Turbo", "Harga dasar Pertamina × liter terjual (Pertamax Turbo)", false, true},
		{"5", "5104", "HPP Pertamina Dex", "Harga dasar Pertamina × liter terjual (Pertamina Dex)", false, true},
		{"5", "5105", "HPP Dexlite", "Harga dasar Pertamina × liter terjual (Dexlite)", false, true},
		{"5", "5110", "Biaya Pengadaan BBM", "Kelompok biaya terkait penebusan/pengiriman BBM", true, true},
		{"5", "5111", "Biaya Admin Bank Penebusan — Pertalite", "Biaya admin bank saat penebusan Pertalite online", false, true},
		{"5", "5112", "Biaya Admin Bank Penebusan — Pertamax", "Biaya admin bank saat penebusan Pertamax online", false, true},
		{"5", "5113", "Biaya Admin Bank Penebusan — Pertamax Turbo", "Biaya admin bank saat penebusan Pertamax Turbo online", false, true},
		{"5", "5114", "Biaya Admin Bank Penebusan — Pertamina Dex", "Biaya admin bank saat penebusan Pertamina Dex online", false, true},
		{"5", "5115", "Biaya Admin Bank Penebusan — Dexlite", "Biaya admin bank saat penebusan Dexlite online", false, true},
		{"5", "5120", "Selisih / Penyusutan BBM", "Kelompok selisih takaran dan penguapan BBM", true, true},
		{"5", "5121", "Selisih & Penyusutan Pertalite", "Susut / selisih takaran Pertalite", false, true},
		{"5", "5122", "Selisih & Penyusutan Pertamax", "Susut / selisih takaran Pertamax", false, true},
		{"5", "5123", "Selisih & Penyusutan Pertamax Turbo", "Susut / selisih takaran Pertamax Turbo", false, true},
		{"5", "5124", "Selisih & Penyusutan Pertamina Dex", "Susut / selisih takaran Pertamina Dex", false, true},
		{"5", "5125", "Selisih & Penyusutan Dexlite", "Susut / selisih takaran Dexlite", false, true},

		// ── 6. BEBAN OPERASIONAL ──
		{"6", "6100", "Beban Personalia", "Kelompok biaya sumber daya manusia", true, false},
		{"6", "6101", "Beban Gaji Pokok", "Gaji pokok bulanan seluruh karyawan", false, true},
		{"6", "6102", "Beban Tunjangan Karyawan", "Tunjangan makan, transport, jabatan, dll.", false, true},
		{"6", "6103", "Beban BPJS Kesehatan — Pemberi Kerja", "Iuran BPJS Kesehatan tanggungan perusahaan", false, true},
		{"6", "6104", "Beban BPJS Ketenagakerjaan — Pemberi Kerja", "Iuran BPJS TK tanggungan perusahaan", false, true},
		{"6", "6105", "Beban Reward Karyawan", "Bonus/reward karyawan dari persentase penjualan BBM", false, true},
		{"6", "6200", "Beban Operasional Umum", "Kelompok biaya operasional kantor dan fasilitas", true, false},
		{"6", "6201", "Beban Listrik", "Tagihan listrik SPBU per bulan", false, false},
		{"6", "6202", "Beban Air", "Tagihan air SPBU per bulan", false, false},
		{"6", "6203", "Beban Telepon & Internet", "Tagihan telepon dan internet operasional", false, false},
		{"6", "6204", "Beban Pemeliharaan & Perbaikan Dispenser", "Servis dan perbaikan mesin dispenser BBM", false, false},
		{"6", "6205", "Beban Administrasi Bank", "Biaya administrasi rekening bank operasional", false, false},
		{"6", "6206", "Beban ATK & Perlengkapan", "Alat tulis kantor dan perlengkapan operasional", false, false},
		{"6", "6207", "Beban Penyusutan Aset Tetap", "Alokasi penyusutan bangunan, mesin, inventaris", false, false},
	}

	for _, item := range coaItems {
		typeID, ok := typeMap[item.TypeCode]
		if !ok {
			continue
		}
		coa := entity.COA{
			COATypeID:   typeID,
			Code:        item.Code,
			Name:        item.Name,
			Description: item.Description,
			IsHeader:    item.IsHeader,
			IsSystem:    item.IsSystem,
			IsActive:    true,
		}
		db.Where("code = ?", item.Code).FirstOrCreate(&coa)
	}
	log.Println("COA data seeded successfully")

	// ─── Seed COA Mappings (non-BBM roles only) ──────────────────────────────
	// BBM-specific mappings are created via GenerateCOAForBBM button per-BBM.
	// Here we seed only the generic (bbm_id IS NULL) role mappings.
	type mappingItem struct {
		TransType string
		Role      string
		Label     string
		COACode   string
	}
	mappingItems := []mappingItem{
		// penebusan
		{"penebusan", "debit_uang_muka", "Uang Muka Penebusan Pertamina", "1131"},
		{"penebusan", "kredit_bank", "Bank Pembayaran (Bank BCA)", "1103"},
		// kedatangan_bbm generic (kredit uang muka)
		{"kedatangan_bbm", "kredit_uang_muka", "Uang Muka Penebusan Pertamina", "1131"},
		// penjualan_tunai generic debit
		{"penjualan_tunai", "debit_kas", "Kas Tunai", "1101"},
		// penjualan_kredit generic debit
		{"penjualan_kredit", "debit_piutang", "Piutang Dagang B2B", "1110"},
		// pelunasan_piutang
		{"pelunasan_piutang", "debit_kas", "Kas / Bank Penerimaan", "1101"},
		{"pelunasan_piutang", "kredit_piutang", "Piutang Dagang B2B", "1110"},
		// kasbon
		{"kasbon", "debit_piutang_kasbon", "Piutang Kasbon Karyawan", "1111"},
		{"kasbon", "kredit_kas", "Kas Pembayaran Kasbon", "1101"},
		// payroll
		{"payroll", "debit_gaji", "Beban Gaji Pokok", "6101"},
		{"payroll", "debit_tunjangan", "Beban Tunjangan Karyawan", "6102"},
		{"payroll", "kredit_hutang_gaji", "Hutang Gaji Karyawan", "2104"},
		{"payroll", "kredit_bpjs_kes", "Hutang BPJS Kesehatan", "2102"},
		{"payroll", "kredit_bpjs_tk", "Hutang BPJS Ketenagakerjaan", "2103"},
		{"payroll", "kredit_pph21", "Hutang Pajak PPh 21", "2105"},
		// cash_in / cash_out
		{"cash_in", "debit_kas", "Kas / Bank Penerima", "1101"},
		{"cash_out", "kredit_kas", "Kas / Bank Sumber", "1101"},
	}
	for _, item := range mappingItems {
		var coa entity.COA
		if err := db.Where("code = ?", item.COACode).First(&coa).Error; err != nil {
			continue
		}
		mapping := entity.COAMapping{
			TransType: item.TransType,
			Role:      item.Role,
			Label:     item.Label,
			COAID:     coa.ID,
		}
		// Upsert: skip if already exists for this trans_type + role + bbm_id IS NULL
		var existing entity.COAMapping
		db.Where("trans_type = ? AND role = ? AND bbm_id IS NULL", item.TransType, item.Role).Limit(1).Find(&existing)
		if existing.ID == 0 {
			db.Omit("COA", "BBM").Create(&mapping)
		}
	}
	log.Println("COA Mapping defaults seeded successfully")

	// ─── Seed BBM-specific penebusan/debit_adm_bank mappings ─────────────────
	// Maps each seeded BBM to its pre-seeded 511X "Biaya Admin Bank Penebusan" COA account.
	// Bio Solar, DEX, Pertalite Khusus do not have 511X accounts seeded — users should
	// run "Generate COA" per BBM to create them.
	type bbmAdmItem struct {
		BBMName string
		COACode string
		Label   string
	}
	bbmAdmItems := []bbmAdmItem{
		{"Pertalite", "5111", "Biaya Admin Bank Penebusan \u2014 Pertalite"},
		{"Pertamax", "5112", "Biaya Admin Bank Penebusan \u2014 Pertamax"},
		{"Pertamax Turbo", "5113", "Biaya Admin Bank Penebusan \u2014 Pertamax Turbo"},
		{"DEX", "5114", "Biaya Admin Bank Penebusan \u2014 Pertamina Dex"},
		{"Dexlite", "5115", "Biaya Admin Bank Penebusan \u2014 Dexlite"},
	}
	for _, item := range bbmAdmItems {
		var bbm entity.BBM
		if err := db.Where("name = ?", item.BBMName).First(&bbm).Error; err != nil {
			continue
		}
		var coa entity.COA
		if err := db.Where("code = ?", item.COACode).First(&coa).Error; err != nil {
			continue
		}
		bbmID := bbm.ID
		var existing entity.COAMapping
		db.Where("trans_type = ? AND role = ? AND bbm_id = ?", "penebusan", "debit_adm_bank", bbmID).
			Limit(1).Find(&existing)
		if existing.ID == 0 {
			db.Omit("COA", "BBM").Create(&entity.COAMapping{
				TransType: "penebusan",
				Role:      "debit_adm_bank",
				Label:     item.Label,
				COAID:     coa.ID,
				BBMID:     &bbmID,
			})
		}
	}
	log.Println("BBM-specific penebusan/debit_adm_bank mappings seeded successfully")
}

// seedShifts — seed 3 shift kerja default.
func seedShifts() {
	db := database.DB
	shifts := []entity.Shift{
		{ShiftName: "Shift 1", ShiftTime: "07:00 - 15:00"},
		{ShiftName: "Shift 2", ShiftTime: "15:00 - 23:00"},
		{ShiftName: "Shift 3", ShiftTime: "23:00 - 07:00"},
	}
	for i := range shifts {
		db.Where(entity.Shift{ShiftName: shifts[i].ShiftName}).FirstOrCreate(&shifts[i])
	}
	log.Println("Shift data seeded successfully")
}

// seedKedatanganBBM — seed sample kedatangan BBM dari penebusan CO yang ada.
// Idempotent: skip jika sudah ada data.
func seedKedatanganBBM() {
	db := database.DB

	var count int64
	db.Model(&entity.TrxKedatanganBBM{}).Count(&count)
	if count > 0 {
		log.Printf("Kedatangan BBM sudah ada %d rows, skip seeding", count)
		return
	}

	// Ambil shift
	var shifts []entity.Shift
	db.Order("id ASC").Find(&shifts)
	if len(shifts) == 0 {
		log.Println("Kedatangan seeder: tidak ada shift, skip")
		return
	}

	// Ambil penebusan CO yang punya no_so, beserta detailnya
	var penebusanList []entity.TrxPenebusan
	db.Where("status = 'CO' AND no_so IS NOT NULL AND no_so != ''").
		Preload("Details").
		Preload("Details.BBM").
		Order("tgl_penebusan ASC").
		Limit(20).
		Find(&penebusanList)

	if len(penebusanList) == 0 {
		log.Println("Kedatangan seeder: tidak ada penebusan CO, skip")
		return
	}

	drivers := []string{"Supriyanto", "Budi Santosa", "Eko Prasetyo", "Hendra Jaya", "Dedy Kurniawan"}
	noPols := []string{"B 9234 XY", "D 4521 AB", "F 8832 CD", "B 1123 ZZ", "D 6670 EF"}

	seeded := 0
	for i, p := range penebusanList {
		for j, detail := range p.Details {
			// Seed sebagian liter (70-90% dari total, simulasikan partial delivery)
			pct := int64(70 + (i+j)%21) // 70..90
			jmlLiter := detail.JmlLiter * pct / 100
			if jmlLiter <= 0 {
				continue
			}

			shift := shifts[(i+j)%len(shifts)]
			noLO := fmt.Sprintf("LO/%04d/%02d/%03d",
				p.TglPenebusan.Year(), int(p.TglPenebusan.Month()), i*10+j+1)

			tglKedatangan := p.TglPenebusan.AddDate(0, 0, 1+j)

			k := entity.TrxKedatanganBBM{
				PenebusanID:       uint64(p.ID),
				PenebusanDetailID: uint64(detail.ID),
				NoLO:              noLO,
				TglKedatangan:     tglKedatangan,
				ShiftID:           shift.ID,
				BBMID:             detail.BBMID,
				JmlLiter:          jmlLiter,
				NamaDriver:        drivers[(i+j)%len(drivers)],
				NoPol:             noPols[(i+j)%len(noPols)],
			}

			if err := db.Omit("Penebusan", "PenebusanDetail", "Shift", "BBM", "Creator", "Updater").
				Create(&k).Error; err != nil {
				log.Printf("Failed to seed kedatangan LO=%s: %v", noLO, err)
				continue
			}

			// Update qty_terkirim pada penebusan_detail
			db.Model(&entity.TrxPenebusanDetail{}).
				Where("id = ?", detail.ID).
				Update("qty_terkirim", jmlLiter)

			seeded++
		}
	}
	log.Printf("Seeded %d kedatangan BBM records", seeded)
}

// seedPenebusan — dedicated seeder for sample penebusan data.
// Must be called AFTER BBM seeding so bbm table is populated.
// Akan top-up hingga total 50 records jika data saat ini < 50.
func seedPenebusan() {
	db := database.DB

	const targetCount = 50

	var currentCount int64
	db.Model(&entity.TrxPenebusan{}).Count(&currentCount)
	if currentCount >= targetCount {
		log.Printf("Penebusan sudah ada %d rows (target %d), skip seeding", currentCount, targetCount)
		return
	}

	// Ambil semua BBM aktif
	var bbmList []entity.BBM
	db.Where("is_active = ?", true).Order("id").Find(&bbmList)
	if len(bbmList) < 2 {
		log.Println("Penebusan seeder: tidak cukup data BBM aktif (minimal 2), skip")
		return
	}

	// Variasi status untuk distribusi realistis
	statuses := []string{
		entity.PenebusanComplete,
		entity.PenebusanComplete,
		entity.PenebusanComplete,
		entity.PenebusanComplete,
		entity.PenebusanComplete,
		entity.PenebusanComplete,
		entity.PenebusanDraft,
		entity.PenebusanDraft,
	}

	// Variasi jumlah liter per BBM
	literOptions := []int64{4000, 6000, 8000, 10000, 12000, 16000, 20000, 24000}

	// Variasi adm bank
	admOptions := []int64{10000, 12500, 15000, 17500, 20000}

	toSeed := int(targetCount - currentCount)
	startSeq := int(currentCount) + 1

	// Rentang tanggal: Jan 2025 — Apr 2026 (16 bulan)
	baseDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local)

	seeded := 0
	for i := 0; i < toSeed; i++ {
		seq := startSeq + i
		noPenebusan := fmt.Sprintf("PNB/%04d/%02d/%04d",
			baseDate.AddDate(0, i*16/toSeed, 0).Year(),
			int(baseDate.AddDate(0, i*16/toSeed, 0).Month()),
			seq,
		)
		tglPenebusan := baseDate.AddDate(0, 0, i*(480/toSeed))
		status := statuses[i%len(statuses)]
		admBank := admOptions[i%len(admOptions)]

		// Buat SO number untuk status complete
		var noSO *string
		if status == entity.PenebusanComplete {
			s := fmt.Sprintf("SO-%02d%02d-%04d", tglPenebusan.Year()%100, int(tglPenebusan.Month()), seq)
			noSO = &s
		}

		// Buat 1–3 detail BBM per header
		numDetails := (i % 3) + 1
		if numDetails > len(bbmList) {
			numDetails = len(bbmList)
		}

		var subtotal int64
		var details []entity.TrxPenebusanDetail
		for d := 0; d < numDetails; d++ {
			bbm := bbmList[d%len(bbmList)]
			jmlLiter := literOptions[(i+d)%len(literOptions)]
			hargaDasar := int64(bbm.Price) - int64(bbm.Margin)
			rowSubtotal := hargaDasar * jmlLiter
			subtotal += rowSubtotal
			details = append(details, entity.TrxPenebusanDetail{
				BBMID:      bbm.ID,
				JmlLiter:   jmlLiter,
				HargaDasar: hargaDasar,
				HargaJual:  int64(bbm.Price),
				Margin:     int64(bbm.Margin),
				PPNPersen:  11,
				Subtotal:   rowSubtotal,
			})
		}
		totalPPN := subtotal * 11 / 100
		totalBayar := subtotal + totalPPN + admBank

		header := entity.TrxPenebusan{
			NoPenebusan:  noPenebusan,
			NoSO:         noSO,
			TglPenebusan: tglPenebusan,
			AdmBank:      admBank,
			Status:       status,
			Catatan:      fmt.Sprintf("Penebusan BBM ke-%d", seq),
			Subtotal:     subtotal,
			TotalPPN:     totalPPN,
			TotalBayar:   totalBayar,
		}
		if status == entity.PenebusanComplete {
			tglBayar := tglPenebusan.AddDate(0, 0, 1)
			header.TglBayar = &tglBayar
		}

		if err := db.Omit("Wallet", "Updater").Create(&header).Error; err != nil {
			log.Printf("Failed to seed penebusan #%d: %v", seq, err)
			continue
		}
		for j := range details {
			details[j].PenebusanID = header.ID
			details[j].PPNRp = details[j].Subtotal * 11 / 100
			details[j].Total = details[j].Subtotal + details[j].PPNRp
			if err := db.Omit("BBM").Create(&details[j]).Error; err != nil {
				log.Printf("Failed to seed penebusan detail seq=%d bbm=%d: %v", seq, details[j].BBMID, err)
			}
		}
		seeded++
	}
	log.Printf("Seeded %d penebusan records (total now: %d)", seeded, currentCount+int64(seeded))
}

// seedPenjualan — seed sample data transaksi penjualan BBM per shift.
// Idempotent: skip jika sudah ada data.
func seedPenjualan() {
	db := database.DB

	var count int64
	db.Model(&entity.TrxPenjualan{}).Count(&count)
	if count > 0 {
		log.Printf("Penjualan sudah ada %d rows, skip seeding", count)
		return
	}

	// Ambil data referensi
	var shifts []entity.Shift
	db.Order("id ASC").Find(&shifts)
	if len(shifts) == 0 {
		log.Println("Penjualan seeder: tidak ada shift, skip")
		return
	}

	var nozzles []entity.Nozzle
	db.Where("is_active = ?", true).Preload("BBM").Preload("Tiang").Order("id ASC").Find(&nozzles)
	if len(nozzles) == 0 {
		log.Println("Penjualan seeder: tidak ada nozzle aktif, skip")
		return
	}

	// Buat 30 record penjualan: 10 hari × 3 shift
	baseDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	shiftTimes := []struct{ mulai, akhir int }{
		{7, 15},
		{15, 23},
		{23, 31}, // shift 3 melewati tengah malam
	}

	seeded := 0
	for day := 0; day < 10; day++ {
		tgl := baseDate.AddDate(0, 0, day)
		for si, shift := range shifts {
			if si >= len(shiftTimes) {
				break
			}
			st := shiftTimes[si]
			waktuMulai := time.Date(tgl.Year(), tgl.Month(), tgl.Day(), st.mulai, 0, 0, 0, time.Local)
			waktuAkhir := time.Date(tgl.Year(), tgl.Month(), tgl.Day(), st.akhir%24, 0, 0, 0, time.Local)
			if st.akhir >= 24 {
				waktuAkhir = waktuAkhir.AddDate(0, 0, 1)
			}

			// Buat detail per nozzle dengan data totalisator realistis
			var details []entity.TrxPenjualanDetail
			var totalRp int64
			base := int64((day*len(shifts)+si)*1000 + 50000) // totalisator awal berbeda tiap shift
			for ni, nz := range nozzles {
				if nz.BBM == nil {
					continue
				}
				// Volume penjualan bervariasi: 200-800 liter per nozzle
				jmlLiter := int64(200 + (day*len(nozzles)+ni)*37%600)
				price := int64(nz.BBM.Price)
				margin := int64(nz.BBM.Margin)
				totAwal := base + int64(ni)*100000
				totAkhir := totAwal + jmlLiter
				jmlRupiah := jmlLiter * price
				totalRp += jmlRupiah
				tiangID := uint(0)
				if nz.Tiang != nil {
					tiangID = nz.Tiang.ID
				}
				details = append(details, entity.TrxPenjualanDetail{
					TiangID:          tiangID,
					NozzleID:         nz.ID,
					BBMID:            nz.BBMID,
					BBMPrice:         price,
					Margin:           margin,
					TotalisatorAwal:  totAwal,
					TotalisatorAkhir: totAkhir,
					JmlLiter:         jmlLiter,
					JmlRupiah:        jmlRupiah,
				})
			}

			if len(details) == 0 {
				continue
			}

			// totalPenerimaan = totalRp (semua tunai), aktual sedikit variatif ±50.000
			delta := int64((day*3+si)%3-1) * 50000
			aktualUang := totalRp + delta
			selisih := aktualUang - totalRp

			header := entity.TrxPenjualan{
				ShiftID:            shift.ID,
				WaktuMulai:         waktuMulai,
				WaktuAkhir:         waktuAkhir,
				TotalRpTotalisator: totalRp,
				TotalPenerimaan:    totalRp,
				AktualUang:         aktualUang,
				Selisih:            selisih,
				Details:            details,
			}

			// Generate no_penjualan: PJL/YYYY/MM/seq
			seq := day*len(shifts) + si + 1
			header.NoPenjualan = fmt.Sprintf("PJL/%04d/%02d/%04d",
				waktuMulai.Year(), int(waktuMulai.Month()), seq)

			if err := db.Omit("Shift", "Creator", "Updater", "Details.Tiang", "Details.Nozzle", "Details.BBM").Create(&header).Error; err != nil {
				log.Printf("Failed to seed penjualan %s: %v", header.NoPenjualan, err)
				continue
			}
			seeded++
		}
	}
	log.Printf("Seeded %d penjualan records", seeded)
}

// seedPiutang — seed sample piutang B2B using real FK references:
// trx_penjualan.id_penjualan, partners.id, and bbm.id.
func seedPiutang() {
	db := database.DB

	var count int64
	db.Model(&entity.TrxPiutang{}).Count(&count)
	if count > 0 {
		log.Printf("Piutang sudah ada %d rows, skip seeding", count)
		return
	}

	var partners []entity.Partner
	db.Where("is_active = ?", true).Order("id ASC").Find(&partners)
	if len(partners) == 0 {
		log.Println("Piutang seeder: tidak ada partner aktif, skip")
		return
	}

	var penjualanList []entity.TrxPenjualan
	db.Preload("Details").
		Order("waktu_mulai ASC").
		Limit(12).
		Find(&penjualanList)
	if len(penjualanList) == 0 {
		log.Println("Piutang seeder: tidak ada penjualan, skip")
		return
	}

	drivers := []string{
		"Rizal Firmansyah",
		"Dadang Setiawan",
		"Yusuf Maulana",
		"Rudi Hartanto",
		"Agus Prasetyo",
		"Hendra Saputra",
	}
	noPols := []string{
		"B 9123 KQ",
		"D 8031 AL",
		"F 7712 MK",
		"B 1045 TX",
		"T 5529 HR",
		"Z 4088 YA",
	}

	var admin entity.User
	var createdBy *uint
	if err := db.Where("username = ?", "admin").First(&admin).Error; err == nil {
		createdBy = &admin.ID
	}

	seeded := 0
	for i, pjl := range penjualanList {
		if len(pjl.Details) == 0 {
			continue
		}

		partner := partners[i%len(partners)]
		createdAt := pjl.WaktuMulai.Add(time.Duration(30+i*7) * time.Minute)
		status := entity.PiutangUnpaid
		if i%4 == 0 {
			status = entity.PiutangPaid
		}

		detailCount := 2 + (i % 3)
		if detailCount > len(pjl.Details) {
			detailCount = len(pjl.Details)
		}

		piutang := entity.TrxPiutang{
			PenjualanID:  pjl.ID,
			PelangganID:  partner.ID,
			Status:       status,
			IsInvoiced:   i%3 != 1,
			Created:      createdAt,
			Updated:      createdAt,
			CreatedBy:    createdBy,
			UpdatedBy:    createdBy,
			TotalTagihan: 0,
		}

		for j := 0; j < detailCount; j++ {
			src := pjl.Details[(i+j)%len(pjl.Details)]
			qty := int64(20 + ((i+1)*(j+2)*7)%75)
			harga := src.BBMPrice
			if harga <= 0 {
				continue
			}
			totalLine := harga * qty
			piutang.TotalTagihan += totalLine
			piutang.Details = append(piutang.Details, entity.TrxPiutangDetail{
				PenjualanID: pjl.ID,
				NoVoucher:   fmt.Sprintf("VCR/%04d/%02d/%03d", pjl.WaktuMulai.Year(), int(pjl.WaktuMulai.Month()), i*10+j+1),
				NoPol:       noPols[(i+j)%len(noPols)],
				DriverName:  drivers[(i+j)%len(drivers)],
				BBMID:       src.BBMID,
				HargaBBM:    harga,
				Margin:      src.Margin,
				QtyLiter:    qty,
				TotalLine:   totalLine,
				Created:     createdAt,
				Updated:     createdAt,
				CreatedBy:   createdBy,
				UpdatedBy:   createdBy,
			})
		}

		if len(piutang.Details) == 0 || piutang.TotalTagihan == 0 {
			continue
		}

		if err := db.Omit(
			"Partner", "Penjualan", "Creator", "Updater",
			"Details.BBM", "Details.Penjualan", "Details.Creator", "Details.Updater",
		).Create(&piutang).Error; err != nil {
			log.Printf("Failed to seed piutang penjualan_id=%d partner_id=%d: %v", pjl.ID, partner.ID, err)
			continue
		}
		seeded++
	}

	log.Printf("Seeded %d piutang records", seeded)
}

// seedPenyusutan — seed sample data penyusutan/susut BBM.
// Idempotent: skip jika sudah ada data.
func seedPenyusutan() {
	db := database.DB

	var count int64
	db.Model(&entity.TrxPenyusutan{}).Count(&count)
	if count > 0 {
		log.Printf("Penyusutan sudah ada %d rows, skip seeding", count)
		return
	}

	// Ambil referensi
	var shifts []entity.Shift
	db.Order("id ASC").Find(&shifts)
	var bbmList []entity.BBM
	db.Where("is_active = ?", true).Order("id ASC").Find(&bbmList)
	if len(shifts) == 0 || len(bbmList) == 0 {
		log.Println("Penyusutan seeder: data referensi tidak cukup, skip")
		return
	}

	keterangans := []string{
		"Susut penguapan normal",
		"Selisih takaran nozzle",
		"Penyusutan akibat suhu tinggi",
		"Selisih totalisator vs dip test",
		"Kebocoran minor tangki",
	}

	baseDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	seeded := 0
	seq := 1
	for day := 0; day < 10; day++ {
		tgl := baseDate.AddDate(0, 0, day)
		// 1-2 record penyusutan per hari, per jenis BBM berbeda
		numRecords := 1 + day%2
		for r := 0; r < numRecords; r++ {
			bbm := bbmList[(day+r)%len(bbmList)]
			shift := shifts[(day+r)%len(shifts)]
			hargaDasar := int64(bbm.Price) - int64(bbm.Margin)
			// Penyusutan kecil: 5-50 liter
			jmlLiter := int64(5 + (day*3+r*7)%46)
			nilaiRupiah := jmlLiter * hargaDasar

			pst := entity.TrxPenyusutan{
				NoPenyusutan:  fmt.Sprintf("PST/%04d/%02d/%04d", tgl.Year(), int(tgl.Month()), seq),
				TglPenyusutan: tgl,
				ShiftID:       shift.ID,
				BBMID:         bbm.ID,
				JmlLiter:      jmlLiter,
				HargaDasar:    hargaDasar,
				NilaiRupiah:   nilaiRupiah,
				Keterangan:    keterangans[(day+r)%len(keterangans)],
			}

			if err := db.Omit("Shift", "BBM", "Creator", "Updater").Create(&pst).Error; err != nil {
				log.Printf("Failed to seed penyusutan %s: %v", pst.NoPenyusutan, err)
				continue
			}
			seq++
			seeded++
		}
	}
	log.Printf("Seeded %d penyusutan records", seeded)
}

// seedJenisTest — seed 10 jenis pengujian/kalibrasi BBM default.
// Idempotent: skip jika sudah ada data.
func seedJenisTest() {
	db := database.DB

	var count int64
	db.Model(&entity.JenisTest{}).Count(&count)
	if count > 0 {
		log.Printf("Jenis Test sudah ada %d rows, skip seeding", count)
		return
	}

	jenisTests := []entity.JenisTest{
		{NamaTest: "Density Pengawas", Deskripsi: "Pengujian densitas BBM oleh pengawas internal", IsActive: true},
		{NamaTest: "Density Teknisi", Deskripsi: "Pengujian densitas BBM oleh teknisi", IsActive: true},
		{NamaTest: "Density Metrologi", Deskripsi: "Pengujian densitas BBM oleh petugas metrologi legal", IsActive: true},
		{NamaTest: "Density Audit Pasti Pas", Deskripsi: "Pengujian densitas dalam rangka program audit Pasti Pas Pertamina", IsActive: true},
		{NamaTest: "Tera Pengawas", Deskripsi: "Tera ulang dispenser oleh pengawas internal SPBU", IsActive: true},
		{NamaTest: "Tera Teknisi", Deskripsi: "Tera ulang dispenser oleh teknisi dispenser", IsActive: true},
		{NamaTest: "Tera Metrologi", Deskripsi: "Tera ulang resmi oleh Dinas Metrologi Legal", IsActive: true},
		{NamaTest: "Tera Audit Pasti Pas", Deskripsi: "Tera ulang dispenser dalam rangka program audit Pasti Pas Pertamina", IsActive: true},
		{NamaTest: "Tes Nozzle", Deskripsi: "Pengujian akurasi takaran nozzle dispenser", IsActive: true},
		{NamaTest: "Tes Selang", Deskripsi: "Pengujian kebocoran dan kondisi selang dispenser", IsActive: true},
	}

	for i := range jenisTests {
		if err := db.Where(entity.JenisTest{NamaTest: jenisTests[i].NamaTest}).FirstOrCreate(&jenisTests[i]).Error; err != nil {
			log.Printf("Failed to seed jenis test %s: %v", jenisTests[i].NamaTest, err)
		}
	}
	log.Printf("Seeded %d jenis test records", len(jenisTests))
}
