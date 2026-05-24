package service

import (
	"fmt"
	"math"

	"spbu_go/internal/dto"
	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

type PenebusanService interface {
	GetAll() ([]entity.TrxPenebusan, error)
	GetByID(id uint64) (*entity.TrxPenebusan, error)
	Datatable(req dto.DatatableRequest) (int64, int64, []entity.TrxPenebusan, error)
	Create(p *entity.TrxPenebusan) error
	Update(p *entity.TrxPenebusan) error
	Delete(id uint64) error
	// GetStokDO returns all CO penebusan with no_so for the stok-do tracking page.
	GetStokDO() ([]entity.TrxPenebusan, error)
	// DatatableStokDO returns paginated stok-do rows for server-side DataTables.
	DatatableStokDO(req dto.DatatableRequest) (int64, int64, []dto.StokDODTRow, error)
	// GetStokDOSummary returns aggregate summary stats for the stok-do header cards.
	GetStokDOSummary() dto.StokDOSummary
	// UpdateQtyTerkirim sets qty_terkirim on a single detail row.
	UpdateQtyTerkirim(detailID uint64, qty int64) error
	// PostJournal posts (or re-posts) double-entry journal entries for a CO penebusan.
	// Safe to call repeatedly — it reverses existing entries first (idempotent).
	PostJournal(p *entity.TrxPenebusan, createdBy *uint) error
	// ReverseJournal removes all journal entries linked to a penebusan (used on delete).
	ReverseJournal(id uint64) error
}

type penebusanService struct {
	repo       repository.PenebusanRepository
	accounting AccountingService
}

func NewPenebusanService(
	repo repository.PenebusanRepository,
	accounting AccountingService,
) PenebusanService {
	return &penebusanService{repo: repo, accounting: accounting}
}

func (s *penebusanService) GetAll() ([]entity.TrxPenebusan, error) {
	return s.repo.FindAll()
}

func (s *penebusanService) GetByID(id uint64) (*entity.TrxPenebusan, error) {
	return s.repo.FindByID(id)
}

func (s *penebusanService) Datatable(req dto.DatatableRequest) (int64, int64, []entity.TrxPenebusan, error) {
	return s.repo.Datatable(req)
}

func (s *penebusanService) Create(p *entity.TrxPenebusan) error {
	return s.repo.Create(p)
}

func (s *penebusanService) Update(p *entity.TrxPenebusan) error {
	return s.repo.Update(p)
}

func (s *penebusanService) Delete(id uint64) error {
	return s.repo.Delete(id)
}

func (s *penebusanService) GetStokDO() ([]entity.TrxPenebusan, error) {
	return s.repo.FindStokDO()
}

func (s *penebusanService) DatatableStokDO(req dto.DatatableRequest) (int64, int64, []dto.StokDODTRow, error) {
	return s.repo.DatatableStokDO(req)
}

func (s *penebusanService) GetStokDOSummary() dto.StokDOSummary {
	return s.repo.GetStokDOSummary()
}

func (s *penebusanService) UpdateQtyTerkirim(detailID uint64, qty int64) error {
	return s.repo.UpdateDetailQtyTerkirim(detailID, qty)
}

func (s *penebusanService) ReverseJournal(id uint64) error {
	return s.accounting.ReverseTransaction("penebusan", uint(id))
}

// PostJournal membangun dan memposting jurnal double-entry untuk penebusan CO.
//
// Jurnal yang dihasilkan:
//
//	Dr  1131  Uang Muka Penebusan Pertamina  = subtotal + total_ppn [+ sisa adm tak termapping]
//	Dr  511X  Biaya Admin Bank — [BBM]        = adm_bank, prorata per share subtotal
//	    Cr  Bank (wallet)                     = total_bayar
func (s *penebusanService) PostJournal(p *entity.TrxPenebusan, createdBy *uint) error {
	if p.Status != entity.PenebusanComplete {
		return nil
	}
	if p.ID == 0 {
		return fmt.Errorf("PostJournal: penebusan ID tidak valid")
	}

	lines := s.buildPenebusanLines(p)

	return s.accounting.RePostTransaction(PostTransactionRequest{
		TransType: "penebusan",
		RefType:   "penebusan",
		RefID:     uint(p.ID),
		Lines:     lines,
		TransDate: p.TglPenebusan,
		CreatedBy: createdBy,
	})
}

// buildPenebusanLines menyusun daftar JournalLine untuk satu penebusan.
func (s *penebusanService) buildPenebusanLines(p *entity.TrxPenebusan) []JournalLine {
	noPenebusan := p.NoPenebusan
	var lines []JournalLine

	// ── Debit: Biaya Admin Bank per BBM (prorata by subtotal share) ──────────
	var totalAdmPosted int64
	if p.AdmBank > 0 && p.Subtotal > 0 {
		for i, detail := range p.Details {
			bbmID := detail.BBMID
			var admAmount int64
			if i == len(p.Details)-1 {
				// Baris terakhir mendapat sisa agar tidak ada drift rounding.
				admAmount = p.AdmBank - totalAdmPosted
			} else {
				admAmount = int64(math.Round(
					float64(p.AdmBank) * float64(detail.Subtotal) / float64(p.Subtotal),
				))
			}
			if admAmount <= 0 {
				continue
			}
			totalAdmPosted += admAmount

			bbmName := fmt.Sprintf("BBM ID %d", bbmID)
			if detail.BBM != nil {
				bbmName = detail.BBM.Name
			}
			lines = append(lines, JournalLine{
				Role:  "debit_adm_bank",
				BBMID: &bbmID,
				Debit: admAmount,
				Desc:  fmt.Sprintf("Biaya admin bank penebusan %s — %s", noPenebusan, bbmName),
			})
		}
	}

	// ── Debit: Uang Muka Pertamina = (subtotal + ppn) + sisa adm yg tak termapping ──
	admRemainder := p.AdmBank - totalAdmPosted
	uangMukaAmount := p.Subtotal + p.TotalPPN + admRemainder
	// Prepend agar uang muka tampil pertama di buku besar.
	lines = append([]JournalLine{{
		Role:  "debit_uang_muka",
		Debit: uangMukaAmount,
		Desc:  fmt.Sprintf("Uang muka penebusan BBM %s ke Pertamina", noPenebusan),
	}}, lines...)

	// ── Kredit: Bank / Wallet = total_bayar ─────────────────────────────────
	lines = append(lines, JournalLine{
		Role:     "kredit_bank",
		WalletID: p.WalletID,
		Credit:   p.TotalBayar,
		Desc:     fmt.Sprintf("Pembayaran penebusan BBM %s via bank", noPenebusan),
	})

	return lines
}
