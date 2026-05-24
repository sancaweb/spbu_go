package service

import (
	"fmt"
	"strings"

	"spbu_go/internal/entity"
	"spbu_go/internal/repository"
)

// ─── TransTypeInfo describes a transaction type and its expected mapping roles ─

type MappingRole struct {
	Role  string
	Label string
	IsBBM bool   // true = one row per BBM
	DC    string // "D" = debit, "C" = kredit
}

type TransTypeInfo struct {
	Code  string
	Label string
	Roles []MappingRole
}

// AllTransTypes returns the canonical list of transaction types and their roles.
func AllTransTypes() []TransTypeInfo {
	return []TransTypeInfo{
		{
			Code:  "penjualan_tunai",
			Label: "Penjualan BBM — Tunai",
			Roles: []MappingRole{
				{Role: "debit_kas", Label: "Kas / Penerimaan Tunai", IsBBM: false, DC: "D"},
				{Role: "kredit_persediaan", Label: "Persediaan BBM (per jenis)", IsBBM: true, DC: "C"},
				{Role: "kredit_pendapatan", Label: "Pendapatan Penjualan Tunai (per jenis)", IsBBM: true, DC: "C"},
			},
		},
		{
			Code:  "penjualan_kredit",
			Label: "Penjualan BBM — Kredit B2B",
			Roles: []MappingRole{
				{Role: "debit_piutang", Label: "Piutang Dagang B2B", IsBBM: false, DC: "D"},
				{Role: "kredit_persediaan", Label: "Persediaan BBM (per jenis)", IsBBM: true, DC: "C"},
				{Role: "kredit_pendapatan", Label: "Pendapatan Penjualan Kredit (per jenis)", IsBBM: true, DC: "C"},
			},
		},
		{
			Code:  "penebusan",
			Label: "Penebusan / Pembelian BBM ke Pertamina",
			Roles: []MappingRole{
				{Role: "debit_uang_muka", Label: "Uang Muka Penebusan Pertamina", IsBBM: false, DC: "D"},
				{Role: "debit_adm_bank", Label: "Biaya Admin Bank Penebusan (per jenis BBM)", IsBBM: true, DC: "D"},
				{Role: "kredit_bank", Label: "Bank Pembayaran", IsBBM: false, DC: "C"},
			},
		},
		{
			Code:  "kedatangan_bbm",
			Label: "Kedatangan BBM (stok masuk)",
			Roles: []MappingRole{
				{Role: "debit_persediaan", Label: "Persediaan BBM (per jenis)", IsBBM: true, DC: "D"},
				{Role: "kredit_uang_muka", Label: "Uang Muka Penebusan Pertamina", IsBBM: false, DC: "C"},
			},
		},
		{
			Code:  "payroll",
			Label: "Penggajian Karyawan",
			Roles: []MappingRole{
				{Role: "debit_gaji", Label: "Beban Gaji Pokok", IsBBM: false, DC: "D"},
				{Role: "debit_tunjangan", Label: "Beban Tunjangan Karyawan", IsBBM: false, DC: "D"},
				{Role: "kredit_hutang_gaji", Label: "Hutang Gaji Karyawan", IsBBM: false, DC: "C"},
				{Role: "kredit_bpjs_kes", Label: "Hutang BPJS Kesehatan", IsBBM: false, DC: "C"},
				{Role: "kredit_bpjs_tk", Label: "Hutang BPJS Ketenagakerjaan", IsBBM: false, DC: "C"},
				{Role: "kredit_pph21", Label: "Hutang Pajak PPh 21", IsBBM: false, DC: "C"},
			},
		},
		{
			Code:  "kasbon",
			Label: "Kasbon Karyawan",
			Roles: []MappingRole{
				{Role: "debit_piutang_kasbon", Label: "Piutang Kasbon Karyawan", IsBBM: false, DC: "D"},
				{Role: "kredit_kas", Label: "Kas / Pembayaran Kasbon", IsBBM: false, DC: "C"},
			},
		},
		{
			Code:  "pelunasan_piutang",
			Label: "Pelunasan Piutang B2B",
			Roles: []MappingRole{
				{Role: "debit_kas", Label: "Kas / Bank Penerimaan", IsBBM: false, DC: "D"},
				{Role: "kredit_piutang", Label: "Piutang Dagang B2B", IsBBM: false, DC: "C"},
			},
		},
		{
			Code:  "cash_in",
			Label: "Cash In (Penerimaan Kas Lainnya)",
			Roles: []MappingRole{
				{Role: "debit_kas", Label: "Kas / Bank Penerima", IsBBM: false, DC: "D"},
				{Role: "kredit_akun", Label: "Akun Kredit Lawan", IsBBM: false, DC: "C"},
			},
		},
		{
			Code:  "cash_out",
			Label: "Cash Out (Pengeluaran Kas Lainnya)",
			Roles: []MappingRole{
				{Role: "debit_akun", Label: "Akun Beban / Debit Lawan", IsBBM: false, DC: "D"},
				{Role: "kredit_kas", Label: "Kas / Bank Sumber", IsBBM: false, DC: "C"},
			},
		},
	}
}

// ─── COAMappingService ───────────────────────────────────────────────────────

type COAMappingService interface {
	GetAll() ([]entity.COAMapping, error)
	GetByTransType(transType string) ([]entity.COAMapping, error)
	GetTransTypes() []TransTypeInfo
	// Upsert creates or updates a single mapping row
	Upsert(transType, role, label string, coaID uint, bbmID *uint) error
	// GenerateCOAForBBM creates the standard COA accounts + default mappings for a new BBM
	GenerateCOAForBBM(bbm *entity.BBM, createdBy *uint) error
	// GetBBMMappings returns all mappings for a specific BBM (for the delete guard)
	GetBBMMappings(bbmID uint) ([]entity.COAMapping, error)
}

type coaMappingService struct {
	repo    repository.COAMappingRepository
	coaRepo repository.COARepository
}

func NewCOAMappingService(repo repository.COAMappingRepository, coaRepo repository.COARepository) COAMappingService {
	return &coaMappingService{repo, coaRepo}
}

func (s *coaMappingService) GetAll() ([]entity.COAMapping, error) {
	return s.repo.FindAll()
}

func (s *coaMappingService) GetByTransType(transType string) ([]entity.COAMapping, error) {
	return s.repo.FindByTransType(transType)
}

func (s *coaMappingService) GetTransTypes() []TransTypeInfo {
	return AllTransTypes()
}

func (s *coaMappingService) Upsert(transType, role, label string, coaID uint, bbmID *uint) error {
	m := &entity.COAMapping{
		TransType: transType,
		Role:      role,
		Label:     label,
		COAID:     coaID,
		BBMID:     bbmID,
	}
	return s.repo.Upsert(m)
}

func (s *coaMappingService) GetBBMMappings(bbmID uint) ([]entity.COAMapping, error) {
	return s.repo.FindByBBM(bbmID)
}

// GenerateCOAForBBM auto-creates COA accounts and mapping rows for a newly added BBM.
// It finds the next available code in each relevant group and creates:
//   - 112X  Persediaan BBM — [Name]
//   - 410X  Pendapatan Penjualan [Name] — Tunai
//   - 411X  Pendapatan Penjualan [Name] — Kredit B2B
//   - 510X  HPP [Name]
//   - 512X  Selisih & Penyusutan [Name]
//   - 511X  Biaya Admin Bank Penebusan — [Name]
func (s *coaMappingService) GenerateCOAForBBM(bbm *entity.BBM, createdBy *uint) error {
	name := bbm.Name

	type coaGen struct {
		prefix   string // e.g. "112"
		typeID   func() uint
		name     string
		desc     string
		isHdr    bool
		isSystem bool
	}

	// Load COAType IDs by code from db first
	// We'll do this inline via FindByCode approach — but COAType has its own repo.
	// For simplicity, use the COA repo's FindByCode to look up header accounts and derive typeID.
	lookupTypeIDFromCode := func(headerCode string) uint {
		hdr, err := s.coaRepo.FindByCode(headerCode)
		if err != nil || hdr == nil {
			return 0
		}
		return hdr.COATypeID
	}

	// Each entry: (prefix for max-code scan, header code for typeID, account name, desc, isSystem)
	type entry struct {
		prefix     string
		headerCode string
		acctName   string
		acctDesc   string
		isSystem   bool
		// which roles to create mappings for after creation
		transType  string
		role       string
		mappingLbl string
	}

	entries := []entry{
		{
			prefix:     "112",
			headerCode: "1120",
			acctName:   fmt.Sprintf("Persediaan BBM \u2014 %s", name),
			acctDesc:   fmt.Sprintf("Stok %s di tangki SPBU", name),
			isSystem:   true,
			transType:  "", role: "", // will map multiple
		},
		{
			prefix:     "410",
			headerCode: "4100",
			acctName:   fmt.Sprintf("Pendapatan Penjualan %s \u2014 Tunai", name),
			acctDesc:   fmt.Sprintf("Omzet %s dari totalisator nozzle (tunai)", name),
			isSystem:   true,
			transType:  "penjualan_tunai", role: "kredit_pendapatan",
			mappingLbl: fmt.Sprintf("Pendapatan Penjualan %s — Tunai", name),
		},
		{
			prefix:     "411",
			headerCode: "4110",
			acctName:   fmt.Sprintf("Pendapatan Penjualan %s \u2014 Kredit B2B", name),
			acctDesc:   fmt.Sprintf("Omzet %s ke partner (kredit/piutang)", name),
			isSystem:   true,
			transType:  "penjualan_kredit", role: "kredit_pendapatan",
			mappingLbl: fmt.Sprintf("Pendapatan Penjualan %s — Kredit B2B", name),
		},
		{
			prefix:     "510",
			headerCode: "5100",
			acctName:   fmt.Sprintf("HPP %s", name),
			acctDesc:   fmt.Sprintf("Harga dasar Pertamina × liter terjual (%s)", name),
			isSystem:   true,
			transType:  "", role: "",
		},
		{
			prefix:     "511",
			headerCode: "5110",
			acctName:   fmt.Sprintf("Biaya Admin Bank Penebusan \u2014 %s", name),
			acctDesc:   fmt.Sprintf("Biaya admin bank saat penebusan %s online", name),
			isSystem:   true,
			transType:  "penebusan", role: "debit_adm_bank",
			mappingLbl: fmt.Sprintf("Biaya Admin Bank Penebusan \u2014 %s", name),
		},
		{
			prefix:     "512",
			headerCode: "5120",
			acctName:   fmt.Sprintf("Selisih & Penyusutan %s", name),
			acctDesc:   fmt.Sprintf("Susut / selisih takaran %s", name),
			isSystem:   true,
			transType:  "", role: "",
		},
	}

	// Helper: pad code to 4 digits
	padCode := func(n int) string {
		return fmt.Sprintf("%04d", n)
	}

	// For persediaan (prefix 112), we need to map kredit_persediaan for penjualan_tunai,
	// penjualan_kredit, and debit_persediaan for kedatangan_bbm.
	// We store the generated persediaan COA id after creation.
	var persediaanCOAID uint

	for i, e := range entries {
		typeID := lookupTypeIDFromCode(e.headerCode)
		if typeID == 0 {
			continue
		}

		// Determine next code
		maxCode, _ := s.coaRepo.FindMaxCodeByPrefix(e.prefix)
		nextCode := maxCode + 1
		// Ensure nextCode stays within 4-digit range for the prefix group
		codeStr := padCode(nextCode)
		// If generated code is shorter than prefix logic expects, override sensibly
		if !strings.HasPrefix(codeStr, e.prefix) {
			// start from prefix + "1"
			baseInt := 0
			for _, ch := range e.prefix {
				baseInt = baseInt*10 + int(ch-'0')
			}
			codeStr = padCode(baseInt*10 + 1)
		}

		coa := entity.COA{
			COATypeID:   typeID,
			Code:        codeStr,
			Name:        e.acctName,
			Description: e.acctDesc,
			IsHeader:    false,
			IsSystem:    e.isSystem,
			IsActive:    true,
			UpdatedBy:   createdBy,
		}
		if err := s.coaRepo.Create(&coa); err != nil {
			return fmt.Errorf("gagal membuat COA %s: %w", codeStr, err)
		}

		if i == 0 {
			persediaanCOAID = coa.ID
		}

		// Create mapping if role specified
		if e.role != "" {
			bbmID := bbm.ID
			if err := s.repo.Upsert(&entity.COAMapping{
				TransType: e.transType,
				Role:      e.role,
				Label:     e.mappingLbl,
				COAID:     coa.ID,
				BBMID:     &bbmID,
			}); err != nil {
				return fmt.Errorf("gagal membuat mapping %s/%s: %w", e.transType, e.role, err)
			}
		}
	}

	// Create persediaan mappings (multi-transtype)
	if persediaanCOAID > 0 {
		bbmID := bbm.ID
		persediaanMappings := []entity.COAMapping{
			{TransType: "penjualan_tunai", Role: "kredit_persediaan",
				Label: fmt.Sprintf("Persediaan BBM — %s", name), COAID: persediaanCOAID, BBMID: &bbmID},
			{TransType: "penjualan_kredit", Role: "kredit_persediaan",
				Label: fmt.Sprintf("Persediaan BBM — %s", name), COAID: persediaanCOAID, BBMID: &bbmID},
			{TransType: "kedatangan_bbm", Role: "debit_persediaan",
				Label: fmt.Sprintf("Persediaan BBM — %s", name), COAID: persediaanCOAID, BBMID: &bbmID},
		}
		for i := range persediaanMappings {
			if err := s.repo.Upsert(&persediaanMappings[i]); err != nil {
				return fmt.Errorf("gagal membuat mapping persediaan: %w", err)
			}
		}
	}

	return nil
}
