package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"time"

	"spbu_go/internal/dto"
	"spbu_go/internal/entity"
	"spbu_go/internal/helper"
	"spbu_go/internal/repository"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrLegacyConnectionNotConfigured = errors.New("koneksi sistem lama belum dikonfigurasi")
	ErrUnsupportedLegacyDriver       = errors.New("driver database sistem lama belum didukung")
)

type ImportDataService interface {
	GetConnection() (dto.LegacyConnectionView, error)
	SaveConnection(input dto.LegacyConnectionInput) error
	TestConnection(ctx context.Context, input dto.LegacyConnectionInput) error
	PreviewBBM(ctx context.Context) ([]dto.LegacyBBMRow, error)
	ImportBBM(ctx context.Context, updatedBy *uint) (dto.ImportResult, error)
	ImportBBMRows(ctx context.Context, rows []dto.LegacyBBMRow, updatedBy *uint) (dto.ImportResult, error)
}

type importDataService struct {
	connectionRepo repository.LegacyDBConnectionRepository
	targetDB       *gorm.DB
}

func NewImportDataService(connectionRepo repository.LegacyDBConnectionRepository, targetDB *gorm.DB) ImportDataService {
	return &importDataService{connectionRepo: connectionRepo, targetDB: targetDB}
}

func (s *importDataService) GetConnection() (dto.LegacyConnectionView, error) {
	connection, err := s.connectionRepo.Find()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.LegacyConnectionView{}, nil
		}
		return dto.LegacyConnectionView{}, err
	}
	return connectionView(connection), nil
}

func (s *importDataService) SaveConnection(input dto.LegacyConnectionInput) error {
	connection, password, err := s.connectionForInput(input, true)
	if err != nil {
		return err
	}
	encryptedPassword, err := helper.EncryptSecret(password)
	if err != nil {
		return err
	}
	connection.EncryptedPassword = encryptedPassword
	return s.connectionRepo.Save(connection)
}

func (s *importDataService) TestConnection(ctx context.Context, input dto.LegacyConnectionInput) error {
	connection, password, err := s.connectionForInput(input, true)
	if err != nil {
		return err
	}
	db, err := openLegacyDatabase(connection, password)
	if err != nil {
		return err
	}
	defer db.Close()
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return db.PingContext(pingCtx)
}

func (s *importDataService) PreviewBBM(ctx context.Context) ([]dto.LegacyBBMRow, error) {
	db, err := s.openSavedDatabase()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	queryCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	rows, err := db.QueryContext(queryCtx, `
		SELECT idBbm, namaBbm, margin, hargaJual, stokLiter, rewardPersen, status
		FROM tb_bbm
		ORDER BY namaBbm ASC`)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca tabel tb_bbm: %w", err)
	}
	defer rows.Close()

	result := make([]dto.LegacyBBMRow, 0)
	for rows.Next() {
		var raw [7]any
		if err := rows.Scan(&raw[0], &raw[1], &raw[2], &raw[3], &raw[4], &raw[5], &raw[6]); err != nil {
			return nil, fmt.Errorf("gagal membaca baris tb_bbm: %w", err)
		}
		row, err := parseLegacyBBMRow(raw)
		if err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gagal membaca hasil tb_bbm: %w", err)
	}
	return result, nil
}

func (s *importDataService) ImportBBM(ctx context.Context, updatedBy *uint) (dto.ImportResult, error) {
	rows, err := s.PreviewBBM(ctx)
	if err != nil {
		return dto.ImportResult{}, err
	}
	return s.ImportBBMRows(ctx, rows, updatedBy)
}

func (s *importDataService) ImportBBMRows(ctx context.Context, rows []dto.LegacyBBMRow, updatedBy *uint) (dto.ImportResult, error) {
	if err := validateLegacyBBMRows(rows); err != nil {
		return dto.ImportResult{}, err
	}

	result := dto.ImportResult{Read: len(rows)}
	err := s.targetDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, row := range rows {
			var bbm entity.BBM
			findErr := tx.Where("LOWER(name) = LOWER(?)", row.Name).First(&bbm).Error
			switch {
			case errors.Is(findErr, gorm.ErrRecordNotFound):
				bbm = entity.BBM{
					Name:          row.Name,
					Margin:        row.Margin,
					Price:         row.Price,
					Stock:         roundedInt64(row.Stock),
					RewardPercent: row.RewardPercent,
					IsActive:      row.IsActive,
					UpdatedBy:     updatedBy,
				}
				if err := tx.Create(&bbm).Error; err != nil {
					return fmt.Errorf("gagal membuat BBM %q: %w", row.Name, err)
				}
				result.Created++
			case findErr != nil:
				return fmt.Errorf("gagal mencari BBM %q: %w", row.Name, findErr)
			default:
				bbm.Name = row.Name
				bbm.Margin = row.Margin
				bbm.Price = row.Price
				bbm.Stock = roundedInt64(row.Stock)
				bbm.RewardPercent = row.RewardPercent
				bbm.IsActive = row.IsActive
				bbm.UpdatedBy = updatedBy
				if err := tx.Save(&bbm).Error; err != nil {
					return fmt.Errorf("gagal memperbarui BBM %q: %w", row.Name, err)
				}
				result.Updated++
			}
		}
		return nil
	})
	return result, err
}

func (s *importDataService) openSavedDatabase() (*sql.DB, error) {
	connection, err := s.connectionRepo.Find()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrLegacyConnectionNotConfigured
	}
	if err != nil {
		return nil, err
	}
	password, err := helper.DecryptSecret(connection.EncryptedPassword)
	if err != nil {
		return nil, err
	}
	return openLegacyDatabase(connection, password)
}

func (s *importDataService) connectionForInput(input dto.LegacyConnectionInput, allowStoredPassword bool) (*entity.LegacyDBConnection, string, error) {
	connection, err := normalizeConnectionInput(input)
	if err != nil {
		return nil, "", err
	}
	password := input.Password
	if strings.TrimSpace(password) == "" && allowStoredPassword {
		stored, findErr := s.connectionRepo.Find()
		if findErr == nil && stored.EncryptedPassword != "" {
			password, err = helper.DecryptSecret(stored.EncryptedPassword)
			if err != nil {
				return nil, "", err
			}
		}
	}
	if existing, findErr := s.connectionRepo.Find(); findErr == nil {
		connection.ID = existing.ID
	} else if !errors.Is(findErr, gorm.ErrRecordNotFound) {
		return nil, "", findErr
	}
	return connection, password, nil
}

func normalizeConnectionInput(input dto.LegacyConnectionInput) (*entity.LegacyDBConnection, error) {
	driver := strings.ToLower(strings.TrimSpace(input.Driver))
	if driver == "mariadb" {
		driver = "mysql"
	}
	if driver != "mysql" {
		return nil, ErrUnsupportedLegacyDriver
	}
	host := strings.TrimSpace(input.Host)
	databaseName := strings.TrimSpace(input.DatabaseName)
	username := strings.TrimSpace(input.Username)
	if host == "" || databaseName == "" || username == "" {
		return nil, errors.New("host, database, dan username wajib diisi")
	}
	if strings.ContainsAny(host, "/\\\x00") || strings.ContainsAny(databaseName, "\x00") || strings.ContainsAny(username, "\x00") {
		return nil, errors.New("parameter koneksi tidak valid")
	}
	portText := strings.TrimSpace(input.Port)
	if portText == "" {
		portText = "3306"
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return nil, errors.New("port harus berupa angka antara 1 sampai 65535")
	}
	sslMode := strings.ToLower(strings.TrimSpace(input.SSLMode))
	if sslMode == "" {
		sslMode = "disable"
	}
	if sslMode == "disable" {
		sslMode = "false"
	}
	if sslMode != "false" && sslMode != "true" && sslMode != "preferred" && sslMode != "skip-verify" {
		return nil, errors.New("mode SSL tidak valid")
	}
	return &entity.LegacyDBConnection{
		Name:         "legacy",
		Driver:       driver,
		Host:         host,
		Port:         port,
		DatabaseName: databaseName,
		Username:     username,
		SSLMode:      sslMode,
	}, nil
}

func openLegacyDatabase(connection *entity.LegacyDBConnection, password string) (*sql.DB, error) {
	tlsMode := strings.ToLower(strings.TrimSpace(connection.SSLMode))
	if tlsMode == "" || tlsMode == "disable" {
		tlsMode = "false"
	}
	config := mysql.Config{
		User:         connection.Username,
		Passwd:       password,
		Net:          "tcp",
		Addr:         net.JoinHostPort(connection.Host, strconv.Itoa(connection.Port)),
		DBName:       connection.DatabaseName,
		ParseTime:    true,
		TLSConfig:    tlsMode,
		Timeout:      10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 10 * time.Second,
		Params: map[string]string{
			"charset": "utf8mb4",
		},
	}
	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(1)
	return db, nil
}

func connectionView(connection *entity.LegacyDBConnection) dto.LegacyConnectionView {
	sslMode := strings.ToLower(strings.TrimSpace(connection.SSLMode))
	if sslMode == "" || sslMode == "disable" {
		sslMode = "false"
	}
	return dto.LegacyConnectionView{
		Driver:       connection.Driver,
		Host:         connection.Host,
		Port:         connection.Port,
		DatabaseName: connection.DatabaseName,
		Username:     connection.Username,
		SSLMode:      sslMode,
		HasPassword:  connection.EncryptedPassword != "",
	}
}

func parseLegacyBBMRow(raw [7]any) (dto.LegacyBBMRow, error) {
	sourceID, err := parseLegacyInt(raw[0])
	if err != nil {
		return dto.LegacyBBMRow{}, fmt.Errorf("idBbm tidak valid: %w", err)
	}
	name := strings.TrimSpace(parseLegacyText(raw[1]))
	if name == "" {
		return dto.LegacyBBMRow{}, fmt.Errorf("BBM dengan id %d memiliki nama kosong", sourceID)
	}
	margin, err := parseLegacyFloat(raw[2])
	if err != nil {
		return dto.LegacyBBMRow{}, fmt.Errorf("margin BBM %q tidak valid: %w", name, err)
	}
	price, err := parseLegacyFloat(raw[3])
	if err != nil {
		return dto.LegacyBBMRow{}, fmt.Errorf("harga BBM %q tidak valid: %w", name, err)
	}
	stock, err := parseLegacyFloat(raw[4])
	if err != nil {
		return dto.LegacyBBMRow{}, fmt.Errorf("stok BBM %q tidak valid: %w", name, err)
	}
	reward, err := parseLegacyFloat(raw[5])
	if err != nil {
		return dto.LegacyBBMRow{}, fmt.Errorf("reward BBM %q tidak valid: %w", name, err)
	}
	return dto.LegacyBBMRow{
		SourceID:      sourceID,
		Name:          name,
		Margin:        margin,
		Price:         price,
		Stock:         stock,
		RewardPercent: reward,
		IsActive:      parseLegacyBool(raw[6]),
	}, nil
}

func validateLegacyBBMRows(rows []dto.LegacyBBMRow) error {
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		key := strings.ToLower(strings.TrimSpace(row.Name))
		if _, exists := seen[key]; exists {
			return fmt.Errorf("nama BBM duplikat pada sistem lama: %q", row.Name)
		}
		seen[key] = struct{}{}
		if row.Margin < 0 || row.Price < 0 || row.Stock < 0 || row.RewardPercent < 0 {
			return fmt.Errorf("nilai numerik BBM %q tidak boleh negatif", row.Name)
		}
		if math.IsNaN(row.Margin) || math.IsInf(row.Margin, 0) ||
			math.IsNaN(row.Price) || math.IsInf(row.Price, 0) ||
			math.IsNaN(row.Stock) || math.IsInf(row.Stock, 0) ||
			math.IsNaN(row.RewardPercent) || math.IsInf(row.RewardPercent, 0) ||
			row.Price > math.MaxFloat64 || row.Stock > math.MaxInt64 {
			return fmt.Errorf("nilai BBM %q terlalu besar untuk sistem baru", row.Name)
		}
	}
	return nil
}

func roundedInt64(value float64) int64 {
	if value >= math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(math.Round(value))
}

func parseLegacyText(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case []byte:
		return string(typed)
	default:
		return fmt.Sprint(typed)
	}
}

func parseLegacyInt(value any) (int64, error) {
	parsed, err := parseLegacyFloat(value)
	if err != nil {
		return 0, err
	}
	if parsed != math.Trunc(parsed) || parsed > math.MaxInt64 || parsed < math.MinInt64 {
		return 0, errors.New("bukan bilangan bulat")
	}
	return int64(parsed), nil
}

func parseLegacyFloat(value any) (float64, error) {
	text := strings.TrimSpace(parseLegacyText(value))
	text = strings.TrimSuffix(text, ".")
	text = strings.ReplaceAll(text, ",", ".")
	if text == "" {
		return 0, nil
	}
	return strconv.ParseFloat(text, 64)
}

func parseLegacyBool(value any) bool {
	switch strings.ToLower(strings.TrimSpace(parseLegacyText(value))) {
	case "1", "true", "yes", "y", "aktif", "active":
		return true
	default:
		return false
	}
}
