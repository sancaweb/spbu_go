package repository

import (
	"errors"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"spbu_go/internal/dto"
	"spbu_go/internal/entity"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Opt in with SPBU_PARTNER_DB_TEST=1. All tables and fixtures are temporary
// on a dedicated connection whose search_path excludes application tables.
func newPartnerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	if os.Getenv("SPBU_PARTNER_DB_TEST") != "1" {
		t.Skip("set SPBU_PARTNER_DB_TEST=1 to run isolated PostgreSQL partner tests")
	}
	config, err := godotenv.Read(filepath.Join("..", "..", "config", ".env"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal("cannot read local database test configuration")
	}
	setting := func(key string) string {
		if value, ok := os.LookupEnv(key); ok {
			return value
		}
		return config[key]
	}
	databaseURL := url.URL{
		Scheme: "postgres",
		Host:   net.JoinHostPort(setting("DB_HOST"), setting("DB_PORT")),
		Path:   "/" + setting("DB_DATABASE"),
		User:   url.UserPassword(setting("DB_USERNAME"), setting("DB_PASSWORD")),
	}
	params := url.Values{
		"sslmode":         {"disable"},
		"connect_timeout": {"5"},
		"TimeZone":        {"UTC"},
		"search_path":     {"pg_temp"},
	}
	databaseURL.RawQuery = params.Encode()
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: databaseURL.String(), PreferSimpleProtocol: true,
	}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		var sqlError interface{ SQLState() string }
		if errors.As(err, &sqlError) {
			t.Fatalf("cannot connect to PostgreSQL (SQLSTATE %s); check local DB configuration", sqlError.SQLState())
		}
		var networkError *net.OpError
		if errors.As(err, &networkError) {
			t.Fatalf("cannot connect to PostgreSQL: %v", networkError.Err)
		}
		t.Fatalf("cannot connect to PostgreSQL (%T); check local DB configuration and server availability", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	for _, statement := range []string{
		`SET search_path TO pg_temp`,
		`SET statement_timeout TO '5s'`,
		`CREATE TEMP TABLE partners (
			id BIGINT PRIMARY KEY, name TEXT NOT NULL, contact_person TEXT, phone TEXT, address TEXT,
			start_date DATE, end_date DATE, isactive BOOLEAN NOT NULL, auto_off BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ, updated_by BIGINT, deleted_at TIMESTAMPTZ
		)`,
		`CREATE TEMP TABLE trx_piutang (
			id_piutang BIGINT PRIMARY KEY,
			pelanggan_id BIGINT NOT NULL REFERENCES partners(id) ON DELETE RESTRICT,
			status TEXT NOT NULL, total_tagihan BIGINT NOT NULL DEFAULT 0
		)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("prepare temporary test tables: %v", err)
		}
	}
	return db
}

func insertPartnerFixture(t *testing.T, db *gorm.DB, id uint, active, archived bool) {
	t.Helper()
	var deletedAt *time.Time
	if archived {
		now := time.Now()
		deletedAt = &now
	}
	if err := db.Exec(`INSERT INTO partners (id, name, isactive, deleted_at) VALUES (?, ?, ?, ?)`,
		id, "Partner Uji", active, deletedAt).Error; err != nil {
		t.Fatal(err)
	}
}

func insertPartnerPiutang(t *testing.T, db *gorm.DB, partnerID uint, status string) {
	t.Helper()
	if err := db.Exec(`INSERT INTO trx_piutang (id_piutang, pelanggan_id, status) VALUES (?, ?, ?)`,
		partnerID, partnerID, status).Error; err != nil {
		t.Fatal(err)
	}
}

func TestPartnerDeletePermanent(t *testing.T) {
	tests := []struct {
		name     string
		active   bool
		archived bool
		piutang  string
		wantErr  error
	}{
		{name: "soft deleted without piutang", archived: true},
		{name: "inactive without piutang"},
		{name: "soft deleted with legacy active flag", active: true, archived: true},
		{name: "active partner rejected", active: true, wantErr: entity.ErrPartnerNotArchived},
		{name: "unpaid piutang rejected", archived: true, piutang: "unpaid", wantErr: entity.ErrPartnerHasPiutang},
		{name: "paid piutang rejected", archived: true, piutang: "paid", wantErr: entity.ErrPartnerHasPiutang},
		{name: "inactive with header only piutang rejected", piutang: "unpaid", wantErr: entity.ErrPartnerHasPiutang},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := newPartnerTestDB(t)
			insertPartnerFixture(t, db, 1, test.active, test.archived)
			if test.piutang != "" {
				insertPartnerPiutang(t, db, 1, test.piutang)
			}
			err := NewPartnerRepository(db).DeletePermanent(1)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("got error %v, want %v", err, test.wantErr)
			}
			var remaining int64
			if err := db.Unscoped().Model(&entity.Partner{}).Where("id = ?", 1).Count(&remaining).Error; err != nil {
				t.Fatal(err)
			}
			if (remaining == 1) != (test.wantErr != nil) {
				t.Fatalf("remaining partner count: %d (expected retained: %v)", remaining, test.wantErr != nil)
			}
			if test.piutang != "" {
				var count int64
				if err := db.Table("trx_piutang").Count(&count).Error; err != nil || count != 1 {
					t.Fatalf("piutang was not preserved: count=%d, error=%v", count, err)
				}
			}
		})
	}
}

func TestPartnerDeletePermanentMissingID(t *testing.T) {
	db := newPartnerTestDB(t)
	for _, id := range []uint{0, 999} {
		if err := NewPartnerRepository(db).DeletePermanent(id); !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("id %d: expected not found, got %v", id, err)
		}
	}
}

func TestPartnerDatatableDeleteEligibility(t *testing.T) {
	db := newPartnerTestDB(t)
	insertPartnerFixture(t, db, 1, false, true)
	insertPartnerFixture(t, db, 2, false, false)
	insertPartnerFixture(t, db, 3, false, true)
	insertPartnerFixture(t, db, 4, false, false)
	insertPartnerFixture(t, db, 5, true, false)
	insertPartnerFixture(t, db, 6, true, true)
	insertPartnerPiutang(t, db, 3, "unpaid")
	insertPartnerPiutang(t, db, 4, "paid")
	if err := db.Exec(`CREATE TEMP TABLE users (id BIGINT PRIMARY KEY, first_name TEXT, deleted_at TIMESTAMPTZ);
		INSERT INTO users (id, first_name) VALUES (9, 'Operator Uji');
		UPDATE partners SET updated_by = 9, phone = '+628123456789', start_date = '2026-01-01' WHERE id = 1`).Error; err != nil {
		t.Fatal(err)
	}
	repo := NewPartnerRepository(db)
	total, filtered, rows, err := repo.Datatable(dto.DatatableRequest{Length: 10}, false)
	if err != nil || total != 5 || filtered != 5 || len(rows) != 5 {
		t.Fatalf("archive response: total=%d filtered=%d rows=%d error=%v", total, filtered, len(rows), err)
	}
	want := map[uint]bool{1: true, 2: true, 3: false, 4: false, 6: true}
	for _, row := range rows {
		eligible, exists := want[row.ID]
		if !exists || row.CanDeletePermanent != eligible {
			t.Fatalf("unexpected eligibility for partner %d: %v", row.ID, row.CanDeletePermanent)
		}
		if row.ID == 1 && (row.Updater == nil || row.Updater.FirstName != "Operator Uji" ||
			row.Phone != "+628123456789" || row.StartDate == nil || row.StartDate.Format("2006-01-02") != "2026-01-01") {
			t.Fatalf("existing partner fields or updater were lost: %+v", row)
		}
	}
	_, _, active, err := repo.Datatable(dto.DatatableRequest{Length: 10}, true)
	if err != nil || len(active) != 1 || active[0].CanDeletePermanent {
		t.Fatalf("active response: rows=%+v, error=%v", active, err)
	}
	req := dto.DatatableRequest{Length: 10}
	req.Search.Value = "no matching partner"
	_, filtered, rows, err = repo.Datatable(req, false)
	if err != nil || filtered != 0 || len(rows) != 0 {
		t.Fatalf("archive search escaped filters: filtered=%d rows=%d error=%v", filtered, len(rows), err)
	}
}

func TestPartnerDeletePermanentRechecksEligibility(t *testing.T) {
	for _, change := range []string{"new piutang", "restored partner"} {
		t.Run(change, func(t *testing.T) {
			db := newPartnerTestDB(t)
			insertPartnerFixture(t, db, 1, false, true)
			repo := NewPartnerRepository(db)
			_, _, rows, err := repo.Datatable(dto.DatatableRequest{Length: 10}, false)
			if err != nil || len(rows) != 1 || !rows[0].CanDeletePermanent {
				t.Fatalf("expected initially deletable partner: rows=%+v, error=%v", rows, err)
			}
			wantErr := entity.ErrPartnerHasPiutang
			if change == "new piutang" {
				insertPartnerPiutang(t, db, 1, "unpaid")
			} else {
				if err := repo.Restore(1); err != nil {
					t.Fatal(err)
				}
				wantErr = entity.ErrPartnerNotArchived
			}
			if err := repo.DeletePermanent(1); !errors.Is(err, wantErr) {
				t.Fatalf("stale eligibility was trusted: got %v, want %v", err, wantErr)
			}
			var count int64
			if err := db.Unscoped().Model(&entity.Partner{}).Count(&count).Error; err != nil || count != 1 {
				t.Fatalf("partner was not preserved: count=%d, error=%v", count, err)
			}
		})
	}
}

func TestPartnerDeletePermanentFailsClosed(t *testing.T) {
	db := newPartnerTestDB(t)
	insertPartnerFixture(t, db, 1, false, true)
	// Renaming only this connection's temporary table simulates a failed check.
	if err := db.Exec(`ALTER TABLE trx_piutang RENAME TO unavailable_piutang`).Error; err != nil {
		t.Fatal(err)
	}
	if err := NewPartnerRepository(db).DeletePermanent(1); err == nil {
		t.Fatal("expected failure when piutang cannot be verified")
	}
	var count int64
	if err := db.Unscoped().Model(&entity.Partner{}).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("partner was not preserved: count=%d, error=%v", count, err)
	}
}

func TestPartnerDeletePermanentRollsBackReferencedPartner(t *testing.T) {
	db := newPartnerTestDB(t)
	insertPartnerFixture(t, db, 1, false, true)
	if err := db.Exec(`CREATE TEMP TABLE partner_reference (partner_id BIGINT REFERENCES partners(id));
		INSERT INTO partner_reference (partner_id) VALUES (1)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := NewPartnerRepository(db).DeletePermanent(1); !errors.Is(err, entity.ErrPartnerReferenced) {
		t.Fatalf("expected reference protection, got %v", err)
	}
	var partner entity.Partner
	if err := db.Unscoped().First(&partner, 1).Error; err != nil || !partner.DeletedAt.Valid {
		t.Fatalf("archive was not preserved after rollback: partner=%+v, error=%v", partner, err)
	}
}
