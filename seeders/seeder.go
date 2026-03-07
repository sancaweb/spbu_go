package seeders

import (
	"log"
	"spbu_go/internal/entity"
	"spbu_go/pkg/database"

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

	log.Println("Database seeded successfully")
}
