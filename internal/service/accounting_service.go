package service

// ─── AccountingService — Modul Akuntansi Generik ─────────────────────────────
//
// AccountingService adalah single entry-point untuk semua efek akuntansi
// (accounting effects) di seluruh modul transaksi.
//
// CARA PAKAI DI MODUL TRANSAKSI BARU
// ───────────────────────────────────
//
//  1. Inject AccountingService ke dalam struct service transaksi:
//
//     type myTrxService struct {
//         repo       repository.MyTrxRepository
//         accounting service.AccountingService   // ← tambahkan ini
//     }
//
//  2. Saat dokumen di-complete, panggil RePostTransaction (idempotent):
//
//     err := s.accounting.RePostTransaction(service.PostTransactionRequest{
//         TransType: "penjualan_tunai",           // sesuai AllTransTypes()
//         RefType:   "penjualan",                 // label untuk ref_type di jurnal
//         RefID:     uint(trx.ID),
//         TransDate: trx.TglJual,
//         CreatedBy: &userID,
//         Lines: []service.JournalLine{
//             {Role: "debit_kas",         WalletID: &walletID, Debit:  trx.Total},
//             {Role: "kredit_persediaan", BBMID: &bbmID,       Credit: trx.HPP},
//             {Role: "kredit_pendapatan", BBMID: &bbmID,       Credit: trx.Total - trx.HPP},
//         },
//     })
//
//  3. Saat dokumen dihapus / void, panggil ReverseTransaction:
//
//     s.accounting.ReverseTransaction("penjualan", uint(trx.ID))
//
// PRINSIP DOUBLE-ENTRY
// ────────────────────
//   Setiap transaksi menghasilkan ≥2 baris jurnal.
//   Syarat mutlak: Σ Debit = Σ Kredit  (dijaga oleh buildEntries di bawah)
//
// RESOLUSI AKUN
// ─────────────
//   AccountingService tidak meng-hardcode akun apapun.
//   Setiap JournalLine hanya menyebut "role" semantik (mis. "debit_kas").
//   Role → COA ID dicari dari tabel coa_mappings yang dikonfigurasi user.
//   Jika mapping belum dikonfigurasi, PostTransaction akan mengembalikan error
//   yang menjelaskan role mana yang perlu dikonfigurasi.

import (
	"fmt"
	"time"

	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

// ─── JournalLine ─────────────────────────────────────────────────────────────

// JournalLine merepresentasikan satu kaki (leg) dari jurnal double-entry.
// Isi salah satu dari Debit atau Credit (tidak keduanya).
// Role adalah kunci semantik yang akan di-resolve ke COA ID via coa_mappings.
type JournalLine struct {
	// Role — kunci semantik mapping, harus cocok dengan baris di coa_mappings.
	// Contoh: "debit_kas", "kredit_bank", "kredit_pendapatan", "debit_persediaan".
	Role string

	// BBMID — wajib diisi untuk role yang bersifat per-BBM (IsBBM: true).
	// Nil untuk role global seperti "debit_kas", "kredit_bank".
	BBMID *uint

	// WalletID — isi untuk baris kas/bank agar journal_entries.wallet_id terisi.
	// Digunakan untuk laporan mutasi per wallet.
	WalletID *uint

	// Debit / Credit — isi salah satu dengan nilai > 0, bukan keduanya.
	Debit  int64
	Credit int64

	// Desc — keterangan baris jurnal (muncul di buku besar).
	Desc string
}

// ─── PostTransactionRequest ───────────────────────────────────────────────────

// PostTransactionRequest membawa semua konteks yang dibutuhkan
// untuk memposting satu set jurnal yang balance.
type PostTransactionRequest struct {
	// TransType — kode jenis transaksi, sesuai AllTransTypes().
	// Contoh: "penebusan", "penjualan_tunai", "payroll".
	TransType string

	// RefType — label yang disimpan di journal_entries.ref_type.
	// Biasanya sama dengan TransType, atau nama tabel sumber.
	RefType string

	// RefID — ID dokumen sumber (journal_entries.ref_id).
	RefID uint

	// Lines — daftar kaki jurnal. Harus balance (Σ Debit = Σ Kredit).
	Lines []JournalLine

	TransDate time.Time
	CreatedBy *uint
}

// ─── AccountingService Interface ─────────────────────────────────────────────

// AccountingService adalah satu-satunya pintu masuk untuk semua efek akuntansi.
// Inject ke dalam service transaksi apapun yang menghasilkan jurnal.
type AccountingService interface {
	// PostTransaction memposting set jurnal balance untuk satu transaksi.
	// Mengembalikan error jika mapping tidak ditemukan atau jurnal tidak balance.
	PostTransaction(req PostTransactionRequest) error

	// RePostTransaction adalah versi idempotent dari PostTransaction.
	// Membalik jurnal lama (jika ada) lalu memposting yang baru.
	// Gunakan ini untuk alur Create-then-Complete dan Edit dokumen CO.
	RePostTransaction(req PostTransactionRequest) error

	// ReverseTransaction menghapus semua jurnal untuk referensi tertentu.
	// Panggil saat dokumen dihapus / di-void.
	ReverseTransaction(refType string, refID uint) error

	// GetJournalByRef mengembalikan semua baris jurnal untuk satu dokumen sumber.
	GetJournalByRef(refType string, refID uint) ([]entity.JournalEntry, error)

	// GetLedger mengembalikan histori jurnal untuk satu akun COA.
	GetLedger(coaID uint, limit int) ([]entity.JournalEntry, error)

	// ResolveCOA mencari COA ID untuk kombinasi trans_type + role + bbm_id.
	// Berguna jika service perlu tahu nomor akun sebelum memposting.
	ResolveCOA(transType, role string, bbmID *uint) (uint, error)
}

// ─── Implementation ───────────────────────────────────────────────────────────

type accountingService struct {
	journalRepo repository.JournalEntryRepository
	mappingRepo repository.COAMappingRepository
}

// NewAccountingService membuat AccountingService yang siap pakai.
// Inject ke dalam constructor service transaksi manapun.
func NewAccountingService(
	journalRepo repository.JournalEntryRepository,
	mappingRepo repository.COAMappingRepository,
) AccountingService {
	return &accountingService{
		journalRepo: journalRepo,
		mappingRepo: mappingRepo,
	}
}

func (s *accountingService) ResolveCOA(transType, role string, bbmID *uint) (uint, error) {
	m, err := s.mappingRepo.FindByTransTypeAndRole(transType, role, bbmID)
	if err != nil {
		if bbmID != nil {
			return 0, fmt.Errorf(
				"COA mapping '%s / %s' untuk BBM ID %d belum dikonfigurasi — "+
					"buka menu Master → Keuangan → COA Mapping",
				transType, role, *bbmID,
			)
		}
		return 0, fmt.Errorf(
			"COA mapping '%s / %s' belum dikonfigurasi — "+
				"buka menu Master → Keuangan → COA Mapping",
			transType, role,
		)
	}
	return m.COAID, nil
}

func (s *accountingService) PostTransaction(req PostTransactionRequest) error {
	entries, err := s.buildEntries(req)
	if err != nil {
		return err
	}
	return s.journalRepo.CreateBatch(entries)
}

func (s *accountingService) RePostTransaction(req PostTransactionRequest) error {
	// Balik jurnal lama terlebih dahulu (idempotent).
	if err := s.journalRepo.DeleteByRef(req.RefType, req.RefID); err != nil {
		return fmt.Errorf("gagal membalik jurnal lama: %w", err)
	}
	return s.PostTransaction(req)
}

func (s *accountingService) ReverseTransaction(refType string, refID uint) error {
	return s.journalRepo.DeleteByRef(refType, refID)
}

func (s *accountingService) GetJournalByRef(refType string, refID uint) ([]entity.JournalEntry, error) {
	return s.journalRepo.FindByRef(refType, refID)
}

func (s *accountingService) GetLedger(coaID uint, limit int) ([]entity.JournalEntry, error) {
	return s.journalRepo.FindByCOA(coaID, limit)
}

// buildEntries meng-resolve COA ID setiap line dan memvalidasi keseimbangan jurnal.
func (s *accountingService) buildEntries(req PostTransactionRequest) ([]entity.JournalEntry, error) {
	if len(req.Lines) == 0 {
		return nil, fmt.Errorf("PostTransaction: tidak ada journal line yang diberikan")
	}

	refID := req.RefID
	var totalDebit, totalCredit int64
	entries := make([]entity.JournalEntry, 0, len(req.Lines))

	for _, line := range req.Lines {
		if line.Debit == 0 && line.Credit == 0 {
			continue // lewati baris kosong
		}
		if line.Debit > 0 && line.Credit > 0 {
			return nil, fmt.Errorf(
				"JournalLine role='%s': tidak boleh mengisi Debit dan Credit sekaligus",
				line.Role,
			)
		}

		coaID, err := s.ResolveCOA(req.TransType, line.Role, line.BBMID)
		if err != nil {
			return nil, err
		}

		totalDebit += line.Debit
		totalCredit += line.Credit

		entries = append(entries, entity.JournalEntry{
			COAID:       coaID,
			WalletID:    line.WalletID,
			Debit:       line.Debit,
			Credit:      line.Credit,
			Description: line.Desc,
			RefType:     req.RefType,
			RefID:       &refID,
			TransDate:   req.TransDate,
			CreatedBy:   req.CreatedBy,
		})
	}

	// Guard: jurnal harus balance.
	if totalDebit != totalCredit {
		return nil, fmt.Errorf(
			"jurnal tidak balance untuk %s/%s (ref_id=%d): "+
				"total debit Rp %d ≠ total kredit Rp %d",
			req.TransType, req.RefType, req.RefID, totalDebit, totalCredit,
		)
	}

	return entries, nil
}
