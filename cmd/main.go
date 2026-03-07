package main

import (
	"log"
	"spbu_go/config"
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

	// Auto Migrate (enabled for easier setup)
	// Auto Migrate (enabled for easier setup)
	// if err := database.DB.AutoMigrate(&entity.Permission{}); err != nil {
	// 	log.Fatal("Failed to migrate Permission:", err)
	// }

	// Manual Migration (safe — CREATE TABLE IF NOT EXISTS is idempotent)
	// Create BBM Table
	err := database.DB.Exec(`CREATE TABLE IF NOT EXISTS bbm (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		margin DECIMAL(15, 2) NOT NULL DEFAULT 0,
		price DECIMAL(15, 2) NOT NULL DEFAULT 0,
		stock DECIMAL(20, 2) NOT NULL DEFAULT 0,
		reward_percent DECIMAL(5, 2) NOT NULL DEFAULT 0,
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_by INT NULL REFERENCES users(id) ON DELETE SET NULL,
		deleted_at TIMESTAMP NULL
	)`).Error
	if err != nil {
		log.Printf("Failed to create BBM table: %v", err)
	}

	// Create Tiang Table
	err = database.DB.Exec(`CREATE TABLE IF NOT EXISTS tiang (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		slug VARCHAR(255) NOT NULL UNIQUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_by INT NULL REFERENCES users(id) ON DELETE SET NULL,
		deleted_at TIMESTAMP NULL
	)`).Error
	if err != nil {
		log.Printf("Failed to create Tiang table: %v", err)
	}

	// Create Nozzles Table
	err = database.DB.Exec(`CREATE TABLE IF NOT EXISTS nozzles (
		id SERIAL PRIMARY KEY,
		tiang_id INT NOT NULL REFERENCES tiang(id) ON DELETE CASCADE,
		description VARCHAR(255),
		bbm_id INT NOT NULL REFERENCES bbm(id) ON DELETE CASCADE,
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_by INT NULL REFERENCES users(id) ON DELETE SET NULL,
		deleted_at TIMESTAMP NULL
	)`).Error
	if err != nil {
		log.Printf("Failed to create Nozzles table: %v", err)
	}

	// Create Settings Table
	err = database.DB.Exec(`CREATE TABLE IF NOT EXISTS settings (
		id SERIAL PRIMARY KEY,
		setting_name VARCHAR(100) NOT NULL UNIQUE,
		setting_value VARCHAR(255) NOT NULL DEFAULT '',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP NULL
	)`).Error
	if err != nil {
		log.Printf("Failed to create Settings table: %v", err)
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

	// Migrate BBM stock column from decimal to bigint
	database.DB.Exec(`DO $$ BEGIN
		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='bbm' AND column_name='stock' AND data_type != 'bigint') THEN
			ALTER TABLE bbm ALTER COLUMN stock TYPE bigint USING stock::bigint;
		END IF;
	END $$`)

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

	log.Println("Manual migration completed")

	// 3. Seed Database
	seeders.Seed()
	// Make sure to run the SQL migrations manually if you disable this.

	// 3. Seed Database
	// seeders.Seed()

	// 4. Setup Dependency Injection
	// Repositories
	userRepo := repository.NewUserRepository(database.DB)
	roleRepo := repository.NewRoleRepository(database.DB)
	bbmRepo := repository.NewBBMRepository(database.DB)
	tiangRepo := repository.NewTiangRepository(database.DB)
	nozzleRepo := repository.NewNozzleRepository(database.DB)
	permissionRepo := repository.NewPermissionRepository(database.DB)
	settingRepo := repository.NewSettingRepository(database.DB)

	// Services
	userService := service.NewUserService(userRepo)
	roleService := service.NewRoleService(roleRepo)
	authService := service.NewAuthService(userRepo)
	bbmService := service.NewBBMService(bbmRepo)
	tiangService := service.NewTiangService(tiangRepo)
	nozzleService := service.NewNozzleService(nozzleRepo)
	permissionService := service.NewPermissionService(permissionRepo)
	settingService := service.NewSettingService(settingRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService, roleService, permissionService)
	roleHandler := handler.NewRoleHandler(roleService)
	bbmHandler := handler.NewBBMHandler(bbmService, settingService)
	tiangHandler := handler.NewTiangHandler(tiangService, bbmService)
	nozzleHandler := handler.NewNozzleHandler(nozzleService)
	settingHandler := handler.NewSettingHandler(settingService)

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
		}
	}

	// 6. Run Server
	log.Printf("Server running on port %s", cfg.AppPort)
	r.Run(":" + cfg.AppPort)
}
