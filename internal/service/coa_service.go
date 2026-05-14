package service

import (
	"errors"
	"time"

	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

// ─── COAService ──────────────────────────────────────────────────────────────

type COAService interface {
	GetAllGrouped() ([]entity.COAType, error)
	GetDetailAccounts() ([]entity.COA, error)
	GetByID(id uint) (*entity.COA, error)
	GetByCode(code string) (*entity.COA, error)
	Create(coa *entity.COA) error
	Update(coa *entity.COA) error
	Delete(id uint) error
	GetTransactions(coaID uint) ([]entity.JournalEntry, *entity.COA, error)
}

type coaService struct {
	coaRepo     repository.COARepository
	journalRepo repository.JournalEntryRepository
}

func NewCOAService(coaRepo repository.COARepository, journalRepo repository.JournalEntryRepository) COAService {
	return &coaService{coaRepo, journalRepo}
}

func (s *coaService) GetAllGrouped() ([]entity.COAType, error) {
	return s.coaRepo.FindAllGrouped()
}

func (s *coaService) GetDetailAccounts() ([]entity.COA, error) {
	return s.coaRepo.FindDetailOnly()
}

func (s *coaService) GetByID(id uint) (*entity.COA, error) {
	return s.coaRepo.FindByID(id)
}

func (s *coaService) GetByCode(code string) (*entity.COA, error) {
	return s.coaRepo.FindByCode(code)
}

func (s *coaService) Create(coa *entity.COA) error {
	return s.coaRepo.Create(coa)
}

func (s *coaService) Update(coa *entity.COA) error {
	existing, err := s.coaRepo.FindByID(coa.ID)
	if err != nil {
		return err
	}
	// System accounts: only allow name, description, is_active update
	if existing.IsSystem {
		existing.Name = coa.Name
		existing.Description = coa.Description
		existing.IsActive = coa.IsActive
		existing.UpdatedBy = coa.UpdatedBy
		return s.coaRepo.Update(existing)
	}
	return s.coaRepo.Update(coa)
}

func (s *coaService) Delete(id uint) error {
	existing, err := s.coaRepo.FindByID(id)
	if err != nil {
		return err
	}
	if existing.IsSystem {
		return errors.New("akun sistem tidak dapat dihapus")
	}
	return s.coaRepo.Delete(id)
}

func (s *coaService) GetTransactions(coaID uint) ([]entity.JournalEntry, *entity.COA, error) {
	coa, err := s.coaRepo.FindByID(coaID)
	if err != nil {
		return nil, nil, err
	}
	entries, err := s.journalRepo.FindByCOA(coaID, 1000)
	return entries, coa, err
}

// ─── JournalService — shared across all transaction modules ──────────────────
// Inject into any handler/service that records financial transactions.

type JournalService interface {
	// Record posts a balanced set of journal entries in a single batch.
	Record(entries []entity.JournalEntry) error
	// RecordOne posts a single journal entry line.
	RecordOne(entry *entity.JournalEntry) error
	// BuildEntry is a helper to construct a JournalEntry value.
	BuildEntry(coaID uint, walletID *uint, debit, credit int64, desc, refType string, refID *uint, transDate time.Time, createdBy *uint) entity.JournalEntry
}

type journalService struct {
	repo repository.JournalEntryRepository
}

func NewJournalService(repo repository.JournalEntryRepository) JournalService {
	return &journalService{repo}
}

func (s *journalService) Record(entries []entity.JournalEntry) error {
	return s.repo.CreateBatch(entries)
}

func (s *journalService) RecordOne(entry *entity.JournalEntry) error {
	return s.repo.Create(entry)
}

func (s *journalService) BuildEntry(coaID uint, walletID *uint, debit, credit int64, desc, refType string, refID *uint, transDate time.Time, createdBy *uint) entity.JournalEntry {
	return entity.JournalEntry{
		COAID:       coaID,
		WalletID:    walletID,
		Debit:       debit,
		Credit:      credit,
		Description: desc,
		RefType:     refType,
		RefID:       refID,
		TransDate:   transDate,
		CreatedBy:   createdBy,
	}
}
