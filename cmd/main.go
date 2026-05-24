package main

import (
	"log"
	"spbu_go/config"
	"spbu_go/internal/entity"
	"spbu_go/internal/handler"
	"spbu_go/internal/middleware"
	"spbu_go/internal/repository"
	"spbu_go/internal/server"
	"spbu_go/internal/service"
	"spbu_go/pkg/database"
	"spbu_go/seeders"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load Config
	cfg := config.LoadConfig()

	// 2. Database Connection
	database.Connect()

	// Manual Migration (safe — CREATE TABLE IF NOT EXISTS is idempotent)
	// Create BBM Table
	err := database.DB.Exec(`CREATE TABLE IF NOT EXISTS bbm (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		margin DECIMAL(15,2) DEFAULT 0,
		price DECIMAL(15,2) DEFAULT 0,
		stock BIGINT DEFAULT 0,
		reward_percent DECIMAL(5,2) DEFAULT 0,
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_by INT NULL REFERENCES users(id) ON DELETE SET NULL,
		deleted_at TIMESTAMP NULL
	)`)
	if err.Error != nil {
		log.Printf("Failed to create BBM table: %v", err.Error)
	}

	// Create Tiang Table
	err = database.DB.Exec(`CREATE TABLE IF NOT EXISTS tiang (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		slug VARCHAR(255) UNIQUE NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_by INT NULL REFERENCES users(id) ON DELETE SET NULL,
		deleted_at TIMESTAMP NULL
	)`)
	if err.Error != nil {
		log.Printf("Failed to create Tiang table: %v", err.Error)
	}

	// Create Nozzle Table
	err = database.DB.Exec(`CREATE TABLE IF NOT EXISTS nozzles (
		id SERIAL PRIMARY KEY,
		tiang_id INT NOT NULL REFERENCES tiang(id) ON DELETE CASCADE,
		description VARCHAR(255),
		bbm_id INT NOT NULL REFERENCES bbm(id) ON DELETE RESTRICT,
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_by INT NULL REFERENCES users(id) ON DELETE SET NULL,
		deleted_at TIMESTAMP NULL
	)`)
	if err.Error != nil {
		log.Printf("Failed to create Nozzles table: %v", err.Error)
	}

	// Create Settings Table
	err = database.DB.Exec(`CREATE TABLE IF NOT EXISTS settings (
		id SERIAL PRIMARY KEY,
		setting_name VARCHAR(100) UNIQUE NOT NULL,
		setting_value VARCHAR(255) NOT NULL DEFAULT '',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP NULL
	)`)
	if err.Error != nil {
		log.Printf("Failed to create Settings table: %v", err.Error)
	}

	// Migrate old settings columns if they exist
	database.DB.Exec(`DO $$ BEGIN
		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='settings' AND column_name='key') THEN
			ALTER TABLE settings RENAME COLUMN "key" TO setting_name;
		END IF;
		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='settings' AND column_name='value') THEN
			ALTER TABLE settings RENAME COLUMN "value" TO setting_value;
		END IF;
	END $$`)

	// Ensure named unique constraints exist so GORM AutoMigrate can manage them cleanly
	database.DB.Exec(`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'uni_jabatan_kode_jabatan') THEN
			ALTER TABLE jabatan ADD CONSTRAINT uni_jabatan_kode_jabatan UNIQUE (kode_jabatan);
		END IF;
	END $$`)
	database.DB.Exec(`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'uni_karyawan_nik') THEN
			ALTER TABLE karyawan ADD CONSTRAINT uni_karyawan_nik UNIQUE (nik);
		END IF;
	END $$`)
	// coa_types.code — rename auto-generated constraint to GORM-expected name
	database.DB.Exec(`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'uni_coa_types_code') THEN
			IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'coa_types_code_key') THEN
				ALTER TABLE coa_types RENAME CONSTRAINT coa_types_code_key TO uni_coa_types_code;
			ELSE
				ALTER TABLE coa_types ADD CONSTRAINT uni_coa_types_code UNIQUE (code);
			END IF;
		END IF;
	END $$`)
	// coas.code — rename auto-generated constraint to GORM-expected name
	database.DB.Exec(`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'uni_coas_code') THEN
			IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'coas_code_key') THEN
				ALTER TABLE coas RENAME CONSTRAINT coas_code_key TO uni_coas_code;
			ELSE
				ALTER TABLE coas ADD CONSTRAINT uni_coas_code UNIQUE (code);
			END IF;
		END IF;
	END $$`)

	// Migrate BBM stock column from decimal to bigint
	database.DB.Exec(`DO $$ BEGIN
		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='bbm' AND column_name='stock' AND data_type != 'bigint') THEN
			ALTER TABLE bbm ALTER COLUMN stock TYPE bigint USING stock::bigint;
		END IF;
	END $$`)

	// Create partners table
	database.DB.Exec(`CREATE TABLE IF NOT EXISTS partners (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		contact_person VARCHAR(255),
		phone VARCHAR(50),
		address TEXT,
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_by INT NULL REFERENCES users(id) ON DELETE SET NULL,
		deleted_at TIMESTAMP NULL
	)`)

	// Create jabatan table
	database.DB.Exec(`CREATE TABLE IF NOT EXISTS jabatan (
		id SERIAL PRIMARY KEY,
		kode_jabatan VARCHAR(10) UNIQUE NOT NULL,
		nama_jabatan VARCHAR(50) NOT NULL,
		reward_persen DECIMAL(5,2) DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP NULL
	)`)
	database.DB.Exec(`ALTER TABLE jabatan ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT TRUE`)

	// Create pendapatan table
	database.DB.Exec(`CREATE TABLE IF NOT EXISTS pendapatan (
		id SERIAL PRIMARY KEY,
		nama_pendapatan VARCHAR(100) NOT NULL,
		tipe VARCHAR(10) NOT NULL DEFAULT 'nominal',
		nilai BIGINT DEFAULT 0,
		deskripsi TEXT,
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_by INT NULL REFERENCES users(id) ON DELETE SET NULL,
		deleted_at TIMESTAMP NULL
	)`)

	// Create potongan table
	database.DB.Exec(`CREATE TABLE IF NOT EXISTS potongan (
		id SERIAL PRIMARY KEY,
		kode_potongan VARCHAR(10) NOT NULL,
		nama_potongan VARCHAR(100) NOT NULL,
		tipe VARCHAR(10) NOT NULL DEFAULT 'nominal',
		nilai BIGINT DEFAULT 0,
		deskripsi TEXT,
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_by INT NULL REFERENCES users(id) ON DELETE SET NULL,
		deleted_at TIMESTAMP NULL
	)`)
	// Create karyawan table
	database.DB.Exec(`CREATE TABLE IF NOT EXISTS karyawan (
		id SERIAL PRIMARY KEY,
		nik VARCHAR(20) UNIQUE NOT NULL,
		nama_lengkap VARCHAR(100) NOT NULL,
		gaji_pokok BIGINT DEFAULT 0,
		alamat TEXT,
		tempat_lahir VARCHAR(50),
		tanggal_lahir DATE,
		status_nikah VARCHAR(15),
		jumlah_anak INT DEFAULT 0,
		jabatan_id INT NULL REFERENCES jabatan(id) ON DELETE SET NULL,
		jenis_kelamin VARCHAR(1),
		agama VARCHAR(20),
		no_hp VARCHAR(20),
		pendidikan VARCHAR(20),
		tgl_pengangkatan DATE,
		tgl_keluar DATE NULL,
		foto VARCHAR(255),
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_by INT NULL REFERENCES users(id) ON DELETE SET NULL,
		deleted_at TIMESTAMP NULL
	)`)

	// Create karyawan_pendapatan junction table
	database.DB.Exec(`CREATE TABLE IF NOT EXISTS karyawan_pendapatan (
		karyawan_id INT NOT NULL REFERENCES karyawan(id) ON DELETE CASCADE,
		pendapatan_id INT NOT NULL REFERENCES pendapatan(id) ON DELETE CASCADE,
		PRIMARY KEY (karyawan_id, pendapatan_id)
	)`)

	// Create karyawan_potongan junction table
	database.DB.Exec(`CREATE TABLE IF NOT EXISTS karyawan_potongan (
		karyawan_id INT NOT NULL REFERENCES karyawan(id) ON DELETE CASCADE,
		potongan_id INT NOT NULL REFERENCES potongan(id) ON DELETE CASCADE,
		PRIMARY KEY (karyawan_id, potongan_id)
	)`)

	// Migrate Users table: add new fields
	database.DB.Exec(`DO $$ BEGIN
		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='name') THEN
			ALTER TABLE users RENAME COLUMN name TO first_name;
		END IF;
	END $$`)
	database.DB.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS last_name VARCHAR(50) DEFAULT ''`)
	database.DB.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS email VARCHAR(254) DEFAULT ''`)
	database.DB.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(20) DEFAULT ''`)
	database.DB.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS active BOOLEAN DEFAULT true`)

	// Create coa_types table
	database.DB.Exec(`CREATE TABLE IF NOT EXISTS coa_types (
		id SERIAL PRIMARY KEY,
		code VARCHAR(10) UNIQUE NOT NULL,
		name VARCHAR(100) NOT NULL,
		normal_balance VARCHAR(6) NOT NULL DEFAULT 'debit',
		description TEXT,
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP NULL
	)`)

	// Create coas table
	database.DB.Exec(`CREATE TABLE IF NOT EXISTS coas (
		id SERIAL PRIMARY KEY,
		coa_type_id INT NOT NULL REFERENCES coa_types(id),
		code VARCHAR(10) UNIQUE NOT NULL,
		name VARCHAR(200) NOT NULL,
		description TEXT,
		is_header BOOLEAN DEFAULT FALSE,
		is_system BOOLEAN DEFAULT FALSE,
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_by INT NULL REFERENCES users(id) ON DELETE SET NULL,
		deleted_at TIMESTAMP NULL
	)`)

	// Create journal_entries table (double-entry bookkeeping)
	database.DB.Exec(`CREATE TABLE IF NOT EXISTS journal_entries (
		id BIGSERIAL PRIMARY KEY,
		coa_id INT NOT NULL REFERENCES coas(id),
		wallet_id INT NULL REFERENCES wallets(id) ON DELETE SET NULL,
		debit BIGINT NOT NULL DEFAULT 0,
		credit BIGINT NOT NULL DEFAULT 0,
		description VARCHAR(500),
		ref_type VARCHAR(50),
		ref_id BIGINT NULL,
		trans_date TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		created_by INT NULL REFERENCES users(id) ON DELETE SET NULL
	)`)

	// Create coa_mappings table
	database.DB.Exec(`CREATE TABLE IF NOT EXISTS coa_mappings (
		id SERIAL PRIMARY KEY,
		trans_type VARCHAR(50) NOT NULL,
		role VARCHAR(50) NOT NULL,
		label VARCHAR(100) NOT NULL,
		coa_id INT NOT NULL REFERENCES coas(id),
		bbm_id INT NULL REFERENCES bbm(id) ON DELETE SET NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)

	// Create trx_penebusan table (header penebusan BBM ke Pertamina)
	database.DB.Exec(`CREATE TABLE IF NOT EXISTS trx_penebusan (
		id              BIGSERIAL       PRIMARY KEY,
		no_penebusan    VARCHAR(25)     NOT NULL,
		no_so           VARCHAR(25)     NULL,
		tgl_penebusan   DATE            NOT NULL,
		tgl_bayar       DATE            NULL,
		wallet_id       INT             NULL REFERENCES wallets(id) ON DELETE SET NULL,
		adm_bank        BIGINT          NOT NULL DEFAULT 0,
		status          VARCHAR(2)      NOT NULL DEFAULT 'DR',
		catatan         TEXT            NULL,
		subtotal        BIGINT          NOT NULL DEFAULT 0,
		total_ppn       BIGINT          NOT NULL DEFAULT 0,
		total_bayar     BIGINT          NOT NULL DEFAULT 0,
		created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_by      INT             NULL REFERENCES users(id) ON DELETE SET NULL,
		deleted_at      TIMESTAMP       NULL,
		CONSTRAINT uni_trx_penebusan_no_penebusan UNIQUE (no_penebusan)
	)`)

	// Ensure named unique constraint for trx_penebusan.no_penebusan
	// Guard against both pg_constraint (CONSTRAINT) and pg_class (INDEX) to avoid
	// "relation already exists" when GORM AutoMigrate created a unique index first.
	database.DB.Exec(`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'uni_trx_penebusan_no_penebusan')
		   AND NOT EXISTS (SELECT 1 FROM pg_class WHERE relname = 'uni_trx_penebusan_no_penebusan') THEN
			IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'trx_penebusan_no_penebusan_key') THEN
				ALTER TABLE trx_penebusan RENAME CONSTRAINT trx_penebusan_no_penebusan_key TO uni_trx_penebusan_no_penebusan;
			ELSE
				ALTER TABLE trx_penebusan ADD CONSTRAINT uni_trx_penebusan_no_penebusan UNIQUE (no_penebusan);
			END IF;
		END IF;
	END $$`)

	// Auto-generate nomor penebusan: PNB/YYYY/MM/increment
	database.DB.Exec(`CREATE OR REPLACE FUNCTION fn_trx_penebusan_set_no_penebusan()
	RETURNS TRIGGER AS $$
	DECLARE
		next_seq INT;
		prefix TEXT;
	BEGIN
		IF NEW.no_penebusan IS NULL OR btrim(NEW.no_penebusan) = '' THEN
			prefix := 'PNB/' || to_char(COALESCE(NEW.tgl_penebusan, CURRENT_DATE), 'YYYY/MM/') ;

			SELECT COALESCE(
				MAX(
					CASE
						WHEN no_penebusan ~ '^PNB/[0-9]{4}/[0-9]{2}/[0-9]+$' THEN split_part(no_penebusan, '/', 4)::INT
						ELSE 0
					END
				),
				0
			) + 1
			INTO next_seq
			FROM trx_penebusan
			WHERE no_penebusan LIKE prefix || '%';

			NEW.no_penebusan := prefix || lpad(next_seq::TEXT, 4, '0');
		END IF;

		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql`)

	database.DB.Exec(`DROP TRIGGER IF EXISTS trg_trx_penebusan_set_no_penebusan ON trx_penebusan`)
	database.DB.Exec(`CREATE TRIGGER trg_trx_penebusan_set_no_penebusan
	BEFORE INSERT ON trx_penebusan
	FOR EACH ROW
	EXECUTE FUNCTION fn_trx_penebusan_set_no_penebusan()`)

	// Normalize status values to DR/CO and enforce default
	database.DB.Exec(`DO $$ BEGIN
		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='trx_penebusan' AND column_name='status') THEN
			ALTER TABLE trx_penebusan ALTER COLUMN status TYPE VARCHAR(2);
			ALTER TABLE trx_penebusan ALTER COLUMN status SET DEFAULT 'DR';
			UPDATE trx_penebusan
			SET status = CASE
				WHEN UPPER(status) IN ('DR', 'DRAFT', 'CANCELLED') THEN 'DR'
				ELSE 'CO'
			END
			WHERE status IS NOT NULL;
		END IF;
	END $$`)

	// Create trx_penebusan_detail table (item per jenis BBM)
	database.DB.Exec(`CREATE TABLE IF NOT EXISTS trx_penebusan_detail (
		id              BIGSERIAL       PRIMARY KEY,
		penebusan_id    BIGINT          NOT NULL REFERENCES trx_penebusan(id) ON DELETE CASCADE,
		bbm_id          INT             NOT NULL REFERENCES bbm(id) ON DELETE RESTRICT,
		jml_liter       BIGINT          NOT NULL,
		harga_dasar     BIGINT          NOT NULL,
		harga_jual      BIGINT          NOT NULL,
		margin          BIGINT          NOT NULL DEFAULT 0,
		ppn_persen      DECIMAL(5,2)    NOT NULL DEFAULT 0,
		ppn_rp          BIGINT          NOT NULL DEFAULT 0,
		subtotal        BIGINT          NOT NULL DEFAULT 0,
		total           BIGINT          NOT NULL DEFAULT 0,
		created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)

	// Add qty_terkirim column to trx_penebusan_detail (tracking pengiriman BBM dari Pertamina)
	database.DB.Exec(`ALTER TABLE trx_penebusan_detail ADD COLUMN IF NOT EXISTS qty_terkirim BIGINT NOT NULL DEFAULT 0`)

	// Create shifts table (master shift kerja)
	database.DB.Exec(`CREATE TABLE IF NOT EXISTS shifts (
		id         SERIAL PRIMARY KEY,
		shift_name VARCHAR(100) NOT NULL,
		shift_time VARCHAR(50),
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)

	// Create trx_kedatangan_bbm table
	database.DB.Exec(`CREATE TABLE IF NOT EXISTS trx_kedatangan_bbm (
		id_kedatangan_bbm   BIGSERIAL       PRIMARY KEY,
		penebusan_id        BIGINT          NOT NULL REFERENCES trx_penebusan(id) ON DELETE RESTRICT,
		penebusan_detail_id BIGINT          NOT NULL REFERENCES trx_penebusan_detail(id) ON DELETE RESTRICT,
		no_lo               VARCHAR(50)     NOT NULL,
		tgl_kedatangan      TIMESTAMP       NOT NULL,
		shift_id            INT             NOT NULL REFERENCES shifts(id) ON DELETE RESTRICT,
		bbm_id              INT             NOT NULL REFERENCES bbm(id) ON DELETE RESTRICT,
		jml_liter           BIGINT          NOT NULL DEFAULT 0,
		nama_driver         VARCHAR(100),
		no_pol              VARCHAR(20),
		created             TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated             TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
		created_by          INT             NULL REFERENCES users(id) ON DELETE SET NULL,
		updated_by          INT             NULL REFERENCES users(id) ON DELETE SET NULL
	)`)

	log.Println("Manual migration completed")

	// Auto Migrate (enabled for easier setup)
	if err := database.DB.AutoMigrate(
		&entity.Setting{},
		&entity.User{},
		&entity.Role{},
		&entity.Permission{},
		&entity.BBM{},
		&entity.Tiang{},
		&entity.Nozzle{},
		&entity.Partner{},
		&entity.Jabatan{},
		&entity.Karyawan{},
		&entity.KaryawanPendapatan{},
		&entity.KaryawanPotongan{},
		&entity.Pendapatan{},
		&entity.Potongan{},
		&entity.Wallet{},
		&entity.COAType{},
		&entity.COA{},
		&entity.JournalEntry{},
		&entity.COAMapping{},
		&entity.TrxPenebusan{},
		&entity.TrxPenebusanDetail{},
		&entity.Shift{},
		&entity.TrxKedatanganBBM{},
	); err != nil {
		log.Fatalf("Gagal melakukan migrasi database: %v", err)
	}

	// 3. Seed Database
	seeders.Seed()
	// Make sure to run the SQL migrations manually if you disable this.

	// 4. Setup Dependency Injection
	// Repositories
	userRepo := repository.NewUserRepository(database.DB)
	roleRepo := repository.NewRoleRepository(database.DB)
	bbmRepo := repository.NewBBMRepository(database.DB)
	tiangRepo := repository.NewTiangRepository(database.DB)
	nozzleRepo := repository.NewNozzleRepository(database.DB)
	permissionRepo := repository.NewPermissionRepository(database.DB)
	settingRepo := repository.NewSettingRepository(database.DB)
	partnerRepo := repository.NewPartnerRepository(database.DB)
	karyawanRepo := repository.NewKaryawanRepository(database.DB)
	jabatanRepo := repository.NewJabatanRepository(database.DB)
	pendapatanRepo := repository.NewPendapatanRepository(database.DB)
	potonganRepo := repository.NewPotonganRepository(database.DB)
	walletRepo := repository.NewWalletRepository(database.DB)
	coaTypeRepo := repository.NewCOATypeRepository(database.DB)
	coaRepo := repository.NewCOARepository(database.DB)
	journalRepo := repository.NewJournalEntryRepository(database.DB)
	coaMappingRepo := repository.NewCOAMappingRepository(database.DB)
	penebusanRepo := repository.NewPenebusanRepository(database.DB)

	// Services
	userService := service.NewUserService(userRepo)
	roleService := service.NewRoleService(roleRepo)
	authService := service.NewAuthService(userRepo)
	bbmService := service.NewBBMService(bbmRepo)
	tiangService := service.NewTiangService(tiangRepo)
	nozzleService := service.NewNozzleService(nozzleRepo)
	permissionService := service.NewPermissionService(permissionRepo)
	settingService := service.NewSettingService(settingRepo)
	partnerService := service.NewPartnerService(partnerRepo)
	karyawanService := service.NewKaryawanService(karyawanRepo)
	jabatanService := service.NewJabatanService(jabatanRepo)
	pendapatanService := service.NewPendapatanService(pendapatanRepo)
	potonganService := service.NewPotonganService(potonganRepo)
	walletService := service.NewWalletService(walletRepo)
	_ = coaTypeRepo // used internally by coaService via coaRepo
	coaService := service.NewCOAService(coaRepo, journalRepo)
	// accountingService — modul akuntansi generik, inject ke setiap service transaksi baru.
	accountingService := service.NewAccountingService(journalRepo, coaMappingRepo)
	coaMappingService := service.NewCOAMappingService(coaMappingRepo, coaRepo)
	penebusanService := service.NewPenebusanService(penebusanRepo, accountingService)
	shiftRepo := repository.NewShiftRepository(database.DB)
	shiftService := service.NewShiftService(shiftRepo)
	kedatanganRepo := repository.NewKedatanganBBMRepository(database.DB)
	kedatanganService := service.NewKedatanganBBMService(kedatanganRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService, roleService, permissionService)
	roleHandler := handler.NewRoleHandler(roleService)
	bbmHandler := handler.NewBBMHandler(bbmService, settingService, coaMappingService)
	tiangHandler := handler.NewTiangHandler(tiangService, bbmService)
	nozzleHandler := handler.NewNozzleHandler(nozzleService)
	settingHandler := handler.NewSettingHandler(settingService)
	partnerHandler := handler.NewPartnerHandler(partnerService)
	karyawanHandler := handler.NewKaryawanHandler(karyawanService, jabatanService, pendapatanService, potonganService)
	jabatanHandler := handler.NewJabatanHandler(jabatanService)
	pendapatanHandler := handler.NewPendapatanHandler(pendapatanService)
	potonganHandler := handler.NewPotonganHandler(potonganService)
	walletHandler := handler.NewWalletHandler(walletService)
	coaHandler := handler.NewCOAHandler(coaService)
	coaMappingHandler := handler.NewCOAMappingHandler(coaMappingService, coaService)
	penebusanHandler := handler.NewPenebusanHandler(penebusanService, bbmService, walletService, settingService)
	shiftHandler := handler.NewShiftHandler(shiftService)
	kedatanganBBMHandler := handler.NewKedatanganBBMHandler(kedatanganService, shiftService)

	// 5. Setup Router
	r := server.NewRouter()

	// Session Store
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	// Routes
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/home")
	})

	// Public Routes
	guest := r.Group("/")
	guest.Use(middleware.GuestOnly())
	{
		guest.GET("/login", authHandler.LoginView)
		guest.POST("/login", authHandler.Login)
	}

	// Protected Routes
	protected := r.Group("/")
	protected.Use(middleware.AuthRequired(userRepo))
	protected.Use(middleware.SettingsMiddleware(settingService))
	{
		protected.GET("/logout", authHandler.Logout)

		// Home (Shortcuts)
		protected.GET("/home", func(c *gin.Context) {
			user, _ := c.Get("user")
			favicon, _ := c.Get("favicon")
			c.HTML(200, "home.html", gin.H{
				"User":       user,
				"Favicon":    favicon,
				"Title":      "Home",
				"ActiveMenu": "home",
			})
		})

		// Dashboard (Charts)
		protected.GET("/dashboard", func(c *gin.Context) {
			user, _ := c.Get("user")
			favicon, _ := c.Get("favicon")
			bbms, _ := bbmService.GetAll()
			decimalPlaces := settingService.GetInt("stock_decimal_places", 0)
			c.HTML(200, "dashboard.html", gin.H{
				"User":               user,
				"Favicon":            favicon,
				"Title":              "Dashboard",
				"ActiveMenu":         "dashboard",
				"BBMs":               bbms,
				"StockDecimalPlaces": decimalPlaces,
			})
		})

		// User Routes
		users := protected.Group("/users")
		{
			users.GET("", userHandler.Index)
			users.GET("/create", userHandler.CreateView)
			users.POST("", userHandler.Create)
			users.GET("/:id/edit", userHandler.EditView)
			users.POST("/:id", userHandler.Update)
			users.POST("/:id/delete", userHandler.Delete)
		}

		// Roles
		protected.GET("/roles", roleHandler.Index)
		protected.GET("/roles/create", roleHandler.CreateView)
		protected.POST("/roles", roleHandler.Create)
		protected.POST("/roles/:id", roleHandler.Update)
		protected.POST("/roles/:id/delete", roleHandler.Delete)
		protected.POST("/roles/:id/permissions", roleHandler.UpdatePermissions)

		// Site Settings
		protected.GET("/settings", settingHandler.Index)
		protected.GET("/settings/value/:name", settingHandler.GetValue)
		protected.POST("/settings", settingHandler.Update)
		protected.POST("/settings/favicon", settingHandler.UploadFavicon)

		// Master Routes
		master := protected.Group("/master")
		{
			bbm := master.Group("/bbm")
			{
				bbm.GET("", bbmHandler.Index)
				bbm.POST("", bbmHandler.Create)
				bbm.POST("/:id", bbmHandler.Update)
				bbm.POST("/:id/delete", bbmHandler.Delete)
				bbm.POST("/:id/generate-coa", bbmHandler.GenerateCOA)
			}
			tiang := master.Group("/tiang")
			{
				tiang.GET("", tiangHandler.Index)
				tiang.POST("", tiangHandler.Create)
				tiang.POST("/:id", tiangHandler.Update)
				tiang.POST("/:id/delete", tiangHandler.Delete)
			}
			nozzle := master.Group("/nozzle")
			{
				nozzle.POST("", nozzleHandler.Create)
				nozzle.POST("/:id", nozzleHandler.Update)
				nozzle.POST("/:id/delete", nozzleHandler.Delete)
			}
			partner := master.Group("/partner")
			{
				partner.GET("", partnerHandler.Index)
				partner.GET("/archive", partnerHandler.Archive)
				partner.POST("/datatable", partnerHandler.Datatable)
				partner.POST("", partnerHandler.Create)
				partner.POST("/:id", partnerHandler.Update)
				partner.POST("/:id/delete", partnerHandler.Delete)
				partner.POST("/:id/restore", partnerHandler.Restore)
			}
			employee := master.Group("/employee")
			{
				employee.GET("", karyawanHandler.Index)
				employee.GET("/archive", karyawanHandler.Archive)
				employee.POST("/datatable", karyawanHandler.Datatable)
				employee.GET("/:id", karyawanHandler.GetOne)
				employee.POST("", karyawanHandler.Create)
				employee.POST("/:id", karyawanHandler.Update)
				employee.POST("/:id/delete", karyawanHandler.Delete)
				employee.POST("/:id/restore", karyawanHandler.Restore)
			}
			jabatan := master.Group("/jabatan")
			{
				jabatan.GET("", jabatanHandler.Index)
				jabatan.GET("/archive", jabatanHandler.Archive)
				jabatan.POST("", jabatanHandler.Create)
				jabatan.POST("/:id", jabatanHandler.Update)
				jabatan.POST("/:id/delete", jabatanHandler.Delete)
			}
			pendapatan := master.Group("/pendapatan")
			{
				pendapatan.GET("", pendapatanHandler.Index)
				pendapatan.GET("/archive", pendapatanHandler.Archive)
				pendapatan.POST("", pendapatanHandler.Create)
				pendapatan.POST("/:id", pendapatanHandler.Update)
				pendapatan.POST("/:id/delete", pendapatanHandler.Delete)
			}
			potongan := master.Group("/potongan")
			{
				potongan.GET("", potonganHandler.Index)
				potongan.GET("/archive", potonganHandler.Archive)
				potongan.POST("", potonganHandler.Create)
				potongan.POST("/:id", potonganHandler.Update)
				potongan.POST("/:id/delete", potonganHandler.Delete)
			}
			shift := master.Group("/shift")
			{
				shift.GET("", shiftHandler.Index)
				shift.POST("", shiftHandler.Create)
				shift.POST("/:id", shiftHandler.Update)
				shift.POST("/:id/delete", shiftHandler.Delete)
			}
			keuangan := master.Group("/keuangan")
			{
				walletRoutes := keuangan.Group("/wallet")
				{
					walletRoutes.GET("", walletHandler.Index)
					walletRoutes.POST("", walletHandler.Create)
					walletRoutes.POST("/:id", walletHandler.Update)
					walletRoutes.POST("/:id/delete", walletHandler.Delete)
				}
				coaRoutes := keuangan.Group("/coa")
				{
					coaRoutes.GET("", coaHandler.Index)
					coaRoutes.POST("", coaHandler.Create)
					coaRoutes.POST("/:id", coaHandler.Update)
					coaRoutes.POST("/:id/delete", coaHandler.Delete)
					coaRoutes.GET("/:id/transactions", coaHandler.Transactions)
				}
				coaMappingRoutes := keuangan.Group("/coa-mapping")
				{
					coaMappingRoutes.GET("", coaMappingHandler.Index)
					coaMappingRoutes.POST("/upsert", coaMappingHandler.Upsert)
				}
			}
		}

		// Transaction Routes
		transaction := protected.Group("/transaction")
		{
			transaction.GET("/penebusan", penebusanHandler.Index)
			transaction.GET("/penebusan/settings/value/:name", settingHandler.GetValue)
			transaction.POST("/penebusan", penebusanHandler.Create)
			transaction.POST("/penebusan/datatable", penebusanHandler.Datatable)
			transaction.GET("/penebusan/:id/detail", penebusanHandler.GetDetail)
			transaction.POST("/penebusan/:id/delete", penebusanHandler.Delete)

			// Kedatangan BBM — pencatatan kedatangan pengiriman dari Pertamina
			transaction.GET("/kedatangan-bbm", kedatanganBBMHandler.Index)
			transaction.POST("/kedatangan-bbm/datatable", kedatanganBBMHandler.Datatable)
			transaction.GET("/kedatangan-bbm/so-options", kedatanganBBMHandler.GetSOOptions)
			transaction.GET("/kedatangan-bbm/so/:penebusan_id/bbm", kedatanganBBMHandler.GetBBMByPenebusan)
			transaction.GET("/kedatangan-bbm/:id", kedatanganBBMHandler.GetOne)
			transaction.POST("/kedatangan-bbm", kedatanganBBMHandler.Create)
			transaction.POST("/kedatangan-bbm/:id", kedatanganBBMHandler.Update)
			transaction.POST("/kedatangan-bbm/:id/delete", kedatanganBBMHandler.Delete)

			// Stok DO — tracking pengiriman BBM per nomor SO
			transaction.GET("/stok-do", penebusanHandler.StokDOIndex)
			transaction.POST("/stok-do/datatable", penebusanHandler.StokDODatatable)
			transaction.POST("/stok-do/:detail_id/qty", penebusanHandler.UpdateDetailQty)
		}
	}

	// 6. Run Server
	log.Printf("Server running on port %s", cfg.AppPort)
	r.Run(":" + cfg.AppPort)
}
