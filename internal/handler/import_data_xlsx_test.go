package handler

import (
	"strings"
	"testing"

	"spbu_go/internal/dto"
)

func TestLegacyBBMXLSXRoundTrip(t *testing.T) {
	content, err := buildLegacyBBMXLSX([]dto.LegacyBBMRow{
		{SourceID: 12, Name: "Pertalite & Turbo", Margin: 100.5, Price: 10000, Stock: 1234, RewardPercent: 1.25, IsActive: true},
		{SourceID: 13, Name: "Solar <Dex>", Margin: 0, Price: 6800, Stock: 0, RewardPercent: 0, IsActive: false},
	})
	if err != nil {
		t.Fatalf("buildLegacyBBMXLSX returned error: %v", err)
	}
	rows, err := parseLegacyBBMXLSX(content)
	if err != nil {
		t.Fatalf("parseLegacyBBMXLSX returned error: %v", err)
	}
	if len(rows) != 2 || rows[0].SourceID != 12 || rows[0].Name != "Pertalite & Turbo" || rows[0].Margin != 100.5 || !rows[0].IsActive {
		t.Fatalf("unexpected first row: %+v", rows)
	}
	if rows[1].Name != "Solar <Dex>" || rows[1].IsActive {
		t.Fatalf("unexpected second row: %+v", rows)
	}
}

func TestLegacyBBMImportHeadersAndValidation(t *testing.T) {
	content, err := buildImportXLSX("BBM", []string{"nama_bbm", "harga_jual", "stok_liter", "reward_persen", "status"}, [][]string{
		{"Avtur", "12.500,50", "100,5", "2", "aktif"},
	})
	if err != nil {
		t.Fatal(err)
	}
	rows, err := parseLegacyBBMXLSX(content)
	if err != nil {
		t.Fatalf("alternative headers should be accepted: %v", err)
	}
	if len(rows) != 1 || rows[0].Name != "Avtur" || rows[0].Price != 12500.5 || rows[0].Stock != 100.5 || !rows[0].IsActive {
		t.Fatalf("unexpected normalized row: %+v", rows)
	}

	invalid, err := buildImportXLSX("BBM", legacyBBMHeaders, [][]string{{"1", "", "0", "100", "0", "0", "true"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parseLegacyBBMXLSX(invalid); err == nil || !strings.Contains(err.Error(), "nama BBM kosong") {
		t.Fatalf("expected empty-name validation, got %v", err)
	}
}

func TestLegacyBBMTemplateHasNoImportRows(t *testing.T) {
	content, err := buildLegacyBBMTemplateXLSX()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parseLegacyBBMXLSX(content); err == nil || !strings.Contains(err.Error(), "belum memiliki data") {
		t.Fatalf("expected template-only file to require data, got %v", err)
	}
}
