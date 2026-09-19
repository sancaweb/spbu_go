package handler

import (
	_ "embed"
	"encoding/json"
	"math"
	"testing"
)

//go:embed testdata/penjualan_6846.json
var penjualan6846Fixture []byte

type penjualan6846Expected struct {
	TotalRpTotalisator int64 `json:"total_rp_totalisator"`
	TotalPiutang       int64 `json:"total_piutang"`
	TotalPenerimaanNet int64 `json:"total_penerimaan_net"`
}

type penjualan6846TestFixture struct {
	Request  penjualanRequest      `json:"request"`
	Expected penjualan6846Expected `json:"expected"`
}

func TestPenjualanSample6846PreservesDecimalTotalisatorAndNetReceipt(t *testing.T) {
	var fixture penjualan6846TestFixture
	if err := json.Unmarshal(penjualan6846Fixture, &fixture); err != nil {
		t.Fatalf("decode legacy fixture: %v", err)
	}

	p, err := parsePenjualanRequest(fixture.Request, nil)
	if err != nil {
		t.Fatalf("parse sample request: %v", err)
	}

	if p.TotalRpTotalisator != fixture.Expected.TotalRpTotalisator {
		t.Fatalf("totalisator mismatch: got %d, want %d", p.TotalRpTotalisator, fixture.Expected.TotalRpTotalisator)
	}

	var totalPiutang int64
	for _, row := range fixture.Request.PiutangB2B {
		totalPiutang += row.TotalTagihan
	}
	net, err := calculateNetPenerimaan(fixture.Request.TotalPenerimaan, totalPiutang, 0)
	if err != nil {
		t.Fatalf("calculate net receipt: %v", err)
	}
	if totalPiutang != fixture.Expected.TotalPiutang {
		t.Fatalf("piutang total mismatch: got %d, want %d", totalPiutang, fixture.Expected.TotalPiutang)
	}
	if net != fixture.Expected.TotalPenerimaanNet {
		t.Fatalf("net receipt mismatch: got %d, want %d", net, fixture.Expected.TotalPenerimaanNet)
	}

	if len(p.Details) != len(fixture.Request.Details) {
		t.Fatalf("detail count mismatch: got %d, want %d", len(p.Details), len(fixture.Request.Details))
	}
	if math.Abs(p.Details[0].TotalisatorAwal-6917375.99) > 0.000001 || math.Abs(p.Details[0].JmlLiter-418.20) > 0.000001 {
		t.Fatalf("decimal totalisator was not preserved: %+v", p.Details[0])
	}
}
