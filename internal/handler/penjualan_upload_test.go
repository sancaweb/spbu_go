package handler

import (
	"os"
	"strings"
	"testing"
)

func TestParseAndMergeSalesReportCSV(t *testing.T) {
	content := strings.NewReader("nozzle_id;nozzle;totalisator_awal;totalisator_akhir\n" +
		"1;Nozzle 1;1.234,50;2.000,75\n" +
		"2;Nozzle 2;100;250\n")

	uploaded, err := parseSalesReportCSV(content)
	if err != nil {
		t.Fatalf("parseSalesReportCSV returned error: %v", err)
	}
	if len(uploaded) != 2 {
		t.Fatalf("got %d uploaded rows, want 2", len(uploaded))
	}
	if uploaded[0].TotalisatorAwal != 1234.5 || uploaded[0].TotalisatorAkhir != 2000.75 {
		t.Fatalf("Indonesian decimal values were not parsed correctly: %+v", uploaded[0])
	}

	merged, err := mergeSalesReportRows([]NozzleFormRow{
		{NozzleID: 1, Description: "Nozzle 1", BBMPrice: 10000},
		{NozzleID: 2, Description: "Nozzle 2", BBMPrice: 15000},
	}, uploaded)
	if err != nil {
		t.Fatalf("mergeSalesReportRows returned error: %v", err)
	}
	if merged[0].JmlLiter != 766.25 || merged[0].JmlRupiah != 7662500 {
		t.Fatalf("first nozzle was calculated incorrectly: %+v", merged[0])
	}
	if merged[1].JmlLiter != 150 || merged[1].JmlRupiah != 2250000 {
		t.Fatalf("second nozzle was calculated incorrectly: %+v", merged[1])
	}
}

func TestMergeSalesReportRowsRequiresAllActiveNozzles(t *testing.T) {
	uploaded := []salesReportRow{{
		NozzleID:         1,
		TotalisatorAwal:  100,
		TotalisatorAkhir: 200,
	}}

	_, err := mergeSalesReportRows([]NozzleFormRow{
		{NozzleID: 1, Description: "Nozzle 1"},
		{NozzleID: 2, Description: "Nozzle 2"},
	}, uploaded)
	if err == nil || !strings.Contains(err.Error(), "semua nozzle aktif") {
		t.Fatalf("expected missing-nozzle validation error, got %v", err)
	}
}

func TestLegacySalesReportSampleIsUploadable(t *testing.T) {
	content, err := os.ReadFile("testdata/penjualan_6846_upload.csv")
	if err != nil {
		t.Fatalf("sample CSV could not be read: %v", err)
	}

	rows, err := parseSalesReportCSV(strings.NewReader(string(content)))
	if err != nil {
		t.Fatalf("sample CSV could not be parsed: %v", err)
	}
	if len(rows) != 16 {
		t.Fatalf("got %d sample rows, want 16", len(rows))
	}
	if rows[0].NozzleID != 1 || rows[0].TotalisatorAwal != 6917375.99 || rows[0].TotalisatorAkhir != 6917794.19 {
		t.Fatalf("first sample row does not match source transaction: %+v", rows[0])
	}
	if rows[15].NozzleID != 16 || rows[15].TotalisatorAwal != 495917.53 || rows[15].TotalisatorAkhir != 495952.43 {
		t.Fatalf("last sample row does not match source transaction: %+v", rows[15])
	}
}

func TestLegacySalesReportSampleXLSXIsUploadable(t *testing.T) {
	content, err := os.ReadFile("testdata/penjualan_6846_upload.xlsx")
	if err != nil {
		t.Fatalf("sample XLSX could not be read: %v", err)
	}

	rows, err := parseSalesReportXLSX(content)
	if err != nil {
		t.Fatalf("sample XLSX could not be parsed: %v", err)
	}
	if len(rows) != 16 {
		t.Fatalf("got %d sample Excel rows, want 16", len(rows))
	}
	if rows[0].NozzleID != 1 || rows[0].TotalisatorAwal != 6917375.99 || rows[0].TotalisatorAkhir != 6917794.19 {
		t.Fatalf("first sample Excel row does not match source transaction: %+v", rows[0])
	}
}

func TestSalesReportXLSXRoundTrip(t *testing.T) {
	content, err := buildSalesReportXLSX([]NozzleFormRow{
		{NozzleID: 1, Description: "Nozzle 1"},
		{NozzleID: 2, Description: "Nozzle 2"},
	})
	if err != nil {
		t.Fatalf("buildSalesReportXLSX returned error: %v", err)
	}

	rows, err := parseSalesReportXLSX(content)
	if err != nil {
		t.Fatalf("parseSalesReportXLSX returned error: %v", err)
	}
	if len(rows) != 2 || rows[0].NozzleID != 1 || rows[1].NozzleID != 2 {
		t.Fatalf("unexpected parsed Excel rows: %+v", rows)
	}
}
