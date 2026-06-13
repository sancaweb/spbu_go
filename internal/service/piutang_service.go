package service

import (
	"fmt"
	"strings"
	"time"

	"spbu_go/internal/dto"
	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

// ─── PiutangService ───────────────────────────────────────────────────────────

// PiutangService mengelola piutang B2B dan efek akuntansinya.
//
// Saat Create (status=unpaid), jurnal double-entry (penjualan_kredit):
//
//	Dr  Piutang Dagang B2B (debit_piutang)                   = total_tagihan
//	    Cr  Persediaan BBM per jenis (kredit_persediaan+BBMID) = harga_dasar × qty
//	    Cr  Pendapatan Kredit per jenis (kredit_pendapatan+BBMID) = margin × qty
//
// Saat Lunas, jurnal (pelunasan_piutang):
//
//	Dr  Kas/Bank default (debit_kas)                          = total_tagihan
//	    Cr  Piutang Dagang B2B (kredit_piutang)               = total_tagihan
type PiutangService interface {
	GetAll() ([]entity.TrxPiutang, error)
	GetByID(id uint64) (*entity.TrxPiutang, error)
	GetByPenjualanID(penjualanID uint64) ([]entity.TrxPiutang, error)
	DatatableRows(req dto.DatatableRequest) (int64, int64, []repository.PiutangDTRow, error)
	DatatableDetailRows(req dto.DatatableRequest) (int64, int64, []repository.PiutangDetailDTRow, error)
	DatatableRekapRows(req dto.DatatableRequest) (int64, int64, []repository.PiutangRekapDTRow, error)
	Summary() (repository.PiutangSummary, error)
	SummaryByMonth(month string) (repository.PiutangSummary, error)
	GroupedRekapByMonth(month string) ([]repository.PiutangGroupedDate, repository.PiutangGroupedGrandTotal, error)
	Create(p *entity.TrxPiutang, createdBy *uint) error
	Delete(id uint64) error
	Lunas(id uint64, updatedBy *uint) error
}

type piutangService struct {
	repo       repository.PiutangRepository
	accounting AccountingService
}

func NewPiutangService(repo repository.PiutangRepository, accounting AccountingService) PiutangService {
	return &piutangService{repo: repo, accounting: accounting}
}

func (s *piutangService) GetAll() ([]entity.TrxPiutang, error) {
	return s.repo.FindAll()
}

func (s *piutangService) GetByID(id uint64) (*entity.TrxPiutang, error) {
	return s.repo.FindByID(id)
}

func (s *piutangService) GetByPenjualanID(penjualanID uint64) ([]entity.TrxPiutang, error) {
	return s.repo.FindByPenjualanID(penjualanID)
}

func (s *piutangService) DatatableRows(req dto.DatatableRequest) (int64, int64, []repository.PiutangDTRow, error) {
	return s.repo.DatatableRows(req)
}

func (s *piutangService) DatatableDetailRows(req dto.DatatableRequest) (int64, int64, []repository.PiutangDetailDTRow, error) {
	return s.repo.DatatableDetailRows(req)
}

func (s *piutangService) DatatableRekapRows(req dto.DatatableRequest) (int64, int64, []repository.PiutangRekapDTRow, error) {
	if strings.TrimSpace(req.FilterWaktuMulai) == "" {
		req.FilterWaktuMulai = time.Now().Format("2006-01")
	}
	return s.repo.DatatableRekapRows(req)
}

func (s *piutangService) Summary() (repository.PiutangSummary, error) {
	return s.repo.SummaryByMonth(time.Now().Format("2006-01"))
}

func (s *piutangService) SummaryByMonth(month string) (repository.PiutangSummary, error) {
	if strings.TrimSpace(month) == "" {
		month = time.Now().Format("2006-01")
	}
	return s.repo.SummaryByMonth(month)
}

func (s *piutangService) GroupedRekapByMonth(month string) ([]repository.PiutangGroupedDate, repository.PiutangGroupedGrandTotal, error) {
	if strings.TrimSpace(month) == "" {
		month = time.Now().Format("2006-01")
	}
	return s.repo.GroupedRekapByMonth(month)
}

// Create menyimpan piutang baru beserta detail, lalu memposting jurnal akuntansi.
func (s *piutangService) Create(p *entity.TrxPiutang, createdBy *uint) error {
	p.Status = entity.PiutangUnpaid
	p.CreatedBy = createdBy
	p.UpdatedBy = createdBy

	for i := range p.Details {
		p.Details[i].PenjualanID = p.PenjualanID
		p.Details[i].CreatedBy = createdBy
		p.Details[i].UpdatedBy = createdBy
		// TotalLine = harga_bbm × qty_liter
		p.Details[i].TotalLine = p.Details[i].HargaBBM * p.Details[i].QtyLiter
	}

	// Hitung total_tagihan dari detail (jika belum di-set)
	if p.TotalTagihan == 0 {
		var total int64
		for _, d := range p.Details {
			total += d.TotalLine
		}
		p.TotalTagihan = total
	}

	if err := s.repo.Create(p); err != nil {
		return fmt.Errorf("gagal menyimpan piutang: %w", err)
	}

	// Post jurnal akuntansi (penjualan_kredit)
	if err := s.postJournalCreate(p, createdBy); err != nil {
		// Jurnal gagal tidak batalkan transaksi (soft failure — log saja)
		// Operator dapat repost manual via COA Mapping fix
		return fmt.Errorf("piutang tersimpan, namun jurnal gagal: %w", err)
	}

	return nil
}

// postJournalCreate mem-posting jurnal double-entry saat piutang dibuat.
func (s *piutangService) postJournalCreate(p *entity.TrxPiutang, createdBy *uint) error {
	// Agregasi per BBM dari details
	type bbmAgg struct {
		hpp        int64 // harga_dasar × qty = (harga_bbm - margin) × qty
		pendapatan int64 // margin × qty
	}
	agg := map[uint]*bbmAgg{}
	for _, d := range p.Details {
		if _, ok := agg[d.BBMID]; !ok {
			agg[d.BBMID] = &bbmAgg{}
		}
		hargaDasar := d.HargaBBM - d.Margin
		agg[d.BBMID].hpp += hargaDasar * d.QtyLiter
		agg[d.BBMID].pendapatan += d.Margin * d.QtyLiter
	}

	lines := []JournalLine{
		// Dr Piutang Dagang B2B
		{Role: "debit_piutang", Debit: p.TotalTagihan},
	}

	// Cr Persediaan + Cr Pendapatan per BBM
	for bbmID, a := range agg {
		bid := bbmID
		lines = append(lines,
			JournalLine{Role: "kredit_persediaan", BBMID: &bid, Credit: a.hpp},
			JournalLine{Role: "kredit_pendapatan", BBMID: &bid, Credit: a.pendapatan},
		)
	}

	return s.accounting.RePostTransaction(PostTransactionRequest{
		TransType: "penjualan_kredit",
		RefType:   "piutang",
		RefID:     uint(p.IDPiutang),
		Lines:     lines,
		TransDate: p.Created,
		CreatedBy: createdBy,
	})
}

// Delete menghapus piutang (dan membalik jurnal).
func (s *piutangService) Delete(id uint64) error {
	// Balik jurnal penjualan_kredit
	_ = s.accounting.ReverseTransaction("piutang", uint(id))

	// Balik jurnal pelunasan jika ada
	_ = s.accounting.ReverseTransaction("piutang_pelunasan", uint(id))

	return s.repo.Delete(id)
}

// Lunas menandai piutang sebagai lunas dan memposting jurnal penerimaan.
func (s *piutangService) Lunas(id uint64, updatedBy *uint) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("piutang tidak ditemukan: %w", err)
	}
	if p.Status == entity.PiutangPaid {
		return fmt.Errorf("piutang sudah berstatus lunas")
	}

	if err := s.repo.MarkPaid(id, updatedBy); err != nil {
		return fmt.Errorf("gagal memperbarui status: %w", err)
	}

	// Post jurnal pelunasan (pelunasan_piutang)
	lines := []JournalLine{
		{Role: "debit_kas", Debit: p.TotalTagihan},
		{Role: "kredit_piutang", Credit: p.TotalTagihan},
	}
	if err := s.accounting.RePostTransaction(PostTransactionRequest{
		TransType: "pelunasan_piutang",
		RefType:   "piutang_pelunasan",
		RefID:     uint(id),
		Lines:     lines,
		TransDate: time.Now(),
		CreatedBy: updatedBy,
	}); err != nil {
		return fmt.Errorf("status lunas tersimpan, namun jurnal gagal: %w", err)
	}

	return nil
}
