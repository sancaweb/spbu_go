package repository

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

type legacy6846Fixture struct {
	Source struct {
		DetailPenjualan []struct {
			BBMID    int    `json:"jenisBbm"`
			JmlLiter string `json:"jmlLiter"`
		} `json:"tb_detail_penjualan"`
		Kedatangan []struct {
			BBMID    int    `json:"jenisBbm"`
			JmlLiter string `json:"jmlLiter"`
		} `json:"tb_kedatangan_bbm"`
	} `json:"source"`
	Expected struct {
		Penyusutan []struct {
			BBMID                  int     `json:"bbm_id"`
			FirstStock             float64 `json:"first_stock"`
			ReceiptLiters          float64 `json:"receipt_liters"`
			DensityTestLiters      float64 `json:"density_test_liters"`
			TotalPenjualanLiters   float64 `json:"total_penjualan_liters"`
			ExpectedEndstockBooked float64 `json:"expected_endstock_booked"`
			ExpectedEndstockActual float64 `json:"expected_endstock_actual"`
			ExpectedPenyusutan     float64 `json:"expected_penyusutan"`
		} `json:"penyusutan"`
	} `json:"expected"`
}

func TestPenyusutanSample6846MatchesLegacySnapshots(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file path")
	}
	fixturePath := filepath.Join(filepath.Dir(filename), "..", "handler", "testdata", "penjualan_6846.json")
	content, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read legacy fixture: %v", err)
	}

	var fixture legacy6846Fixture
	if err := json.Unmarshal(content, &fixture); err != nil {
		t.Fatalf("decode legacy fixture: %v", err)
	}

	sales := map[int]float64{}
	for _, detail := range fixture.Source.DetailPenjualan {
		liters, err := strconv.ParseFloat(detail.JmlLiter, 64)
		if err != nil {
			t.Fatalf("parse sales liters for BBM %d: %v", detail.BBMID, err)
		}
		sales[detail.BBMID] += liters
	}
	receipts := map[int]float64{}
	for _, receipt := range fixture.Source.Kedatangan {
		liters, err := strconv.ParseFloat(receipt.JmlLiter, 64)
		if err != nil {
			t.Fatalf("parse receipt liters for BBM %d: %v", receipt.BBMID, err)
		}
		receipts[receipt.BBMID] += liters
	}

	for _, expected := range fixture.Expected.Penyusutan {
		booked := calculateEndstockBooked(
			expected.FirstStock,
			expected.ReceiptLiters,
			expected.TotalPenjualanLiters,
			expected.DensityTestLiters,
		)
		actual := expected.ExpectedEndstockActual
		loss := actual - booked

		if math.Abs(sales[expected.BBMID]-expected.TotalPenjualanLiters) > 0.000001 {
			t.Fatalf("BBM %d sales mismatch: got %.8f, want %.8f", expected.BBMID, sales[expected.BBMID], expected.TotalPenjualanLiters)
		}
		if math.Abs(receipts[expected.BBMID]-expected.ReceiptLiters) > 0.000001 {
			t.Fatalf("BBM %d receipt mismatch: got %.8f, want %.8f", expected.BBMID, receipts[expected.BBMID], expected.ReceiptLiters)
		}
		if math.Abs(booked-expected.ExpectedEndstockBooked) > 0.000001 {
			t.Fatalf("BBM %d booked stock mismatch: got %.8f, want %.8f", expected.BBMID, booked, expected.ExpectedEndstockBooked)
		}
		if math.Abs(loss-expected.ExpectedPenyusutan) > 0.000001 {
			t.Fatalf("BBM %d penyusutan mismatch: got %.8f, want %.8f", expected.BBMID, loss, expected.ExpectedPenyusutan)
		}
	}
}
