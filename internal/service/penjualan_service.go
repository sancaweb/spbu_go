package service

import (
	"fmt"
	"log"

	"spbu_go/internal/dto"
	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

// PenjualanService — business logic transaksi penjualan BBM.
type PenjualanService interface {
	GetAll() ([]entity.TrxPenjualan, error)
	GetByID(id uint64) (*entity.TrxPenjualan, error)
	Datatable(req dto.DatatableRequest) (int64, int64, []dto.PenjualanDTRow, error)
	Create(p *entity.TrxPenjualan) error
	Update(p *entity.TrxPenjualan) error
	Delete(id uint64) error
	// PostJournal memposting (atau re-posting) jurnal double-entry untuk satu penjualan.
	// Idempotent: reverse entries lama terlebih dahulu, lalu posting baru.
	PostJournal(p *entity.TrxPenjualan, createdBy *uint) error
	// ReverseJournal menghapus semua jurnal terkait satu penjualan (dipakai saat delete).
	ReverseJournal(id uint64) error
}

type penjualanService struct {
	repo       repository.PenjualanRepository
	accounting AccountingService
}

func NewPenjualanService(repo repository.PenjualanRepository, accounting AccountingService) PenjualanService {
	return &penjualanService{repo: repo, accounting: accounting}
}

func (s *penjualanService) GetAll() ([]entity.TrxPenjualan, error) {
	return s.repo.FindAll()
}

func (s *penjualanService) GetByID(id uint64) (*entity.TrxPenjualan, error) {
	return s.repo.FindByID(id)
}

func (s *penjualanService) Datatable(req dto.DatatableRequest) (int64, int64, []dto.PenjualanDTRow, error) {
	return s.repo.Datatable(req)
}

func (s *penjualanService) Create(p *entity.TrxPenjualan) error {
	return s.repo.Create(p)
}

func (s *penjualanService) Update(p *entity.TrxPenjualan) error {
	return s.repo.Update(p)
}

func (s *penjualanService) Delete(id uint64) error {
	return s.repo.Delete(id)
}

func (s *penjualanService) ReverseJournal(id uint64) error {
	return s.accounting.ReverseTransaction("penjualan", uint(id))
}

// PostJournal membangun dan memposting jurnal double-entry untuk satu penjualan BBM.
//
// Struktur jurnal (penjualan_tunai):
//
//	Dr  Kas/Penerimaan (debit_kas)                      = total_rp_totalisator
//	    Cr  Persediaan BBM — [BBM] (kredit_persediaan)  = jml_liter × harga_dasar  (per BBM)
//	    Cr  Pendapatan Penjualan (kredit_pendapatan)     = jml_liter × margin       (per BBM)
//
// Balance check: Σ Cr = Σ (harga_dasar + margin) × liter = Σ bbm_price × liter = total_rp_totalisator ✓
func (s *penjualanService) PostJournal(p *entity.TrxPenjualan, createdBy *uint) error {
	if p.ID == 0 {
		return fmt.Errorf("PostJournal: penjualan ID tidak valid")
	}
	if len(p.Details) == 0 {
		log.Printf("[penjualan] PostJournal: penjualan %d tidak memiliki detail, skip", p.ID)
		return nil
	}

	// Agregasi per BBM untuk menghindari baris jurnal duplikat
	type bbmAgg struct {
		bbmID      uint
		bbmName    string
		hpp        int64 // jml_liter × (bbm_price - margin)
		pendapatan int64 // jml_liter × margin
	}
	aggMap := map[uint]*bbmAgg{}
	for _, d := range p.Details {
		if _, ok := aggMap[d.BBMID]; !ok {
			name := fmt.Sprintf("BBM ID %d", d.BBMID)
			if d.BBM != nil {
				name = d.BBM.Name
			}
			aggMap[d.BBMID] = &bbmAgg{bbmID: d.BBMID, bbmName: name}
		}
		g := aggMap[d.BBMID]
		hargaDasar := d.BBMPrice - d.Margin
		g.hpp += d.JmlLiter * hargaDasar
		g.pendapatan += d.JmlLiter * d.Margin
	}

	var lines []JournalLine

	// Debit: Kas Tunai = total penjualan
	lines = append(lines, JournalLine{
		Role:  "debit_kas",
		Debit: p.TotalRpTotalisator,
		Desc:  fmt.Sprintf("Penjualan BBM %s", p.NoPenjualan),
	})

	// Kredit per BBM: Persediaan (HPP) + Pendapatan (Margin)
	for bbmID, g := range aggMap {
		id := bbmID
		if g.hpp > 0 {
			lines = append(lines, JournalLine{
				Role:   "kredit_persediaan",
				BBMID:  &id,
				Credit: g.hpp,
				Desc:   fmt.Sprintf("HPP %s — %s", g.bbmName, p.NoPenjualan),
			})
		}
		if g.pendapatan > 0 {
			lines = append(lines, JournalLine{
				Role:   "kredit_pendapatan",
				BBMID:  &id,
				Credit: g.pendapatan,
				Desc:   fmt.Sprintf("Pendapatan penjualan %s — %s", g.bbmName, p.NoPenjualan),
			})
		}
	}

	return s.accounting.RePostTransaction(PostTransactionRequest{
		TransType: "penjualan_tunai",
		RefType:   "penjualan",
		RefID:     uint(p.ID),
		Lines:     lines,
		TransDate: p.WaktuMulai,
		CreatedBy: createdBy,
	})
}
