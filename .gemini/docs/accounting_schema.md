# Skema Financial Accounting — SPBU Go

> Dokumen ini adalah referensi resmi untuk sistem pencatatan keuangan SPBU Go.  
> Diperbarui: Mei 2026

---

## 1. Prinsip Dasar — Double-Entry Bookkeeping

Setiap transaksi keuangan dicatat dengan **minimal dua baris jurnal**:

- Satu atau lebih baris **Debit** (Dr)
- Satu atau lebih baris **Kredit** (Cr)

Syarat mutlak: **Σ Debit = Σ Kredit** per transaksi.

Nilai disimpan sebagai **integer Rupiah** (bigint) — tidak ada desimal untuk menghindari floating-point error.

---

## 2. Tabel Database Inti

| Tabel             | Fungsi                                                    |
| ----------------- | --------------------------------------------------------- |
| `coa_types`       | Kelompok akun (1=Aset s/d 6=Beban)                        |
| `coas`            | Chart of Accounts — daftar semua akun buku besar          |
| `coa_mappings`    | Mapping: jenis transaksi + role semantik → akun COA       |
| `journal_entries` | Buku besar — setiap baris = satu kaki jurnal double-entry |

---

## 3. Chart of Accounts (COA)

### Kelompok 1 — Aset (Normal Balance: Debit)

| Kode     | Nama                                   | Header | Keterangan                              |
| -------- | -------------------------------------- | ------ | --------------------------------------- |
| **1100** | **Aset Lancar**                        | ✓      |                                         |
| 1101     | Kas                                    |        | Kas tunai operasional SPBU              |
| 1102     | Bank Mandiri                           |        | Rekening Bank Mandiri                   |
| 1103     | Bank BCA                               |        | Rekening Bank BCA                       |
| 1110     | Piutang Dagang B2B                     |        | Piutang penjualan BBM kredit ke partner |
| 1111     | Piutang Kasbon Karyawan                |        | Pinjaman kasbon yang belum dilunasi     |
| **1120** | **Persediaan BBM**                     | ✓      | Kelompok persediaan bahan bakar         |
| 1121     | Persediaan BBM — Pertalite             |        | Stok di tangki                          |
| 1122     | Persediaan BBM — Pertamax              |        | Stok di tangki                          |
| 1123     | Persediaan BBM — Pertamax Turbo        |        | Stok di tangki                          |
| 1124     | Persediaan BBM — Pertamina Dex         |        | Stok di tangki                          |
| 1125     | Persediaan BBM — Dexlite               |        | Stok di tangki                          |
| 112X     | Persediaan BBM — [nama baru]           |        | Di-generate via tombol "Generate COA"   |
| 1131     | Uang Muka Penebusan Pertamina          |        | Advance payment ke Pertamina            |
| **1200** | **Aset Tetap**                         | ✓      |                                         |
| 1201     | Tanah                                  |        |                                         |
| 1202     | Bangunan                               |        |                                         |
| 1203     | Mesin & Peralatan Dispenser            |        |                                         |
| 1204     | Inventaris Kantor                      |        |                                         |
| 1211     | Akumulasi Penyusutan Bangunan          |        | Contra asset                            |
| 1212     | Akumulasi Penyusutan Mesin & Dispenser |        | Contra asset                            |
| 1213     | Akumulasi Penyusutan Inventaris        |        | Contra asset                            |

### Kelompok 2 — Kewajiban (Normal Balance: Kredit)

| Kode | Nama                        | Keterangan                      |
| ---- | --------------------------- | ------------------------------- |
| 2102 | Hutang BPJS Kesehatan       | Iuran yang belum dibayarkan     |
| 2103 | Hutang BPJS Ketenagakerjaan | Iuran yang belum dibayarkan     |
| 2104 | Hutang Gaji Karyawan        | Gaji dihitung belum dibayar     |
| 2105 | Hutang Pajak PPh 21         | Pajak karyawan belum disetorkan |

> Catatan: Tidak ada "Hutang ke Pertamina" karena penebusan BBM dibayar online
> di muka (advance payment), bukan kredit. Akun 1131 (Uang Muka) yang dipakai.

### Kelompok 3 — Modal / Ekuitas (Normal Balance: Kredit)

| Kode | Nama                 | Keterangan                   |
| ---- | -------------------- | ---------------------------- |
| 3101 | Modal Disetor        | Modal pemilik                |
| 3102 | Laba Ditahan         | Akumulasi laba tahun lalu    |
| 3103 | Ikhtisar Laba / Rugi | Akun penutup (closing entry) |

### Kelompok 4 — Pendapatan (Normal Balance: Kredit)

| Kode     | Nama                                             | Header | Keterangan |
| -------- | ------------------------------------------------ | ------ | ---------- |
| **4100** | **Pendapatan Penjualan BBM — Tunai**             | ✓      |            |
| 4101     | Pendapatan Penjualan Pertalite — Tunai           |        |            |
| 4102     | Pendapatan Penjualan Pertamax — Tunai            |        |            |
| 4103     | Pendapatan Penjualan Pertamax Turbo — Tunai      |        |            |
| 4104     | Pendapatan Penjualan Pertamina Dex — Tunai       |        |            |
| 4105     | Pendapatan Penjualan Dexlite — Tunai             |        |            |
| **4110** | **Pendapatan Penjualan BBM — Kredit B2B**        | ✓      |            |
| 4111     | Pendapatan Penjualan Pertalite — Kredit B2B      |        |            |
| 4112     | Pendapatan Penjualan Pertamax — Kredit B2B       |        |            |
| 4113     | Pendapatan Penjualan Pertamax Turbo — Kredit B2B |        |            |
| 4114     | Pendapatan Penjualan Pertamina Dex — Kredit B2B  |        |            |
| 4115     | Pendapatan Penjualan Dexlite — Kredit B2B        |        |            |
| **4120** | **Pendapatan Lain-lain**                         | ✓      |            |
| 4121     | Pendapatan Bunga Bank                            |        | Jasa giro  |
| 4122     | Pendapatan Non-BBM Lainnya                       |        | Insidental |

### Kelompok 5 — Harga Pokok Penjualan / HPP (Normal Balance: Debit)

| Kode     | Nama                                        | Header | Keterangan                  |
| -------- | ------------------------------------------- | ------ | --------------------------- |
| **5100** | **Harga Pokok Penjualan BBM**               | ✓      |                             |
| 5101     | HPP Pertalite                               |        | Harga dasar × liter terjual |
| 5102     | HPP Pertamax                                |        |                             |
| 5103     | HPP Pertamax Turbo                          |        |                             |
| 5104     | HPP Pertamina Dex                           |        |                             |
| 5105     | HPP Dexlite                                 |        |                             |
| **5110** | **Biaya Pengadaan BBM**                     | ✓      | Biaya terkait penebusan     |
| 5111     | Biaya Admin Bank Penebusan — Pertalite      |        |                             |
| 5112     | Biaya Admin Bank Penebusan — Pertamax       |        |                             |
| 5113     | Biaya Admin Bank Penebusan — Pertamax Turbo |        |                             |
| 5114     | Biaya Admin Bank Penebusan — Pertamina Dex  |        |                             |
| 5115     | Biaya Admin Bank Penebusan — Dexlite        |        |                             |
| **5120** | **Selisih / Penyusutan BBM**                | ✓      |                             |
| 5121     | Selisih & Penyusutan Pertalite              |        | Susut / selisih takaran     |
| 5122     | Selisih & Penyusutan Pertamax               |        |                             |
| 5123     | Selisih & Penyusutan Pertamax Turbo         |        |                             |
| 5124     | Selisih & Penyusutan Pertamina Dex          |        |                             |
| 5125     | Selisih & Penyusutan Dexlite                |        |                             |

### Kelompok 6 — Beban Operasional (Normal Balance: Debit)

| Kode     | Nama                                       | Header | Keterangan                    |
| -------- | ------------------------------------------ | ------ | ----------------------------- |
| **6100** | **Beban Personalia**                       | ✓      |                               |
| 6101     | Beban Gaji Pokok                           |        | Gaji bulanan seluruh karyawan |
| 6102     | Beban Tunjangan Karyawan                   |        | Makan, transport, jabatan     |
| 6103     | Beban BPJS Kesehatan — Pemberi Kerja       |        | Tanggungan perusahaan         |
| 6104     | Beban BPJS Ketenagakerjaan — Pemberi Kerja |        | Tanggungan perusahaan         |
| 6105     | Beban Reward Karyawan                      |        | Bonus dari % penjualan BBM    |
| **6200** | **Beban Operasional Umum**                 | ✓      |                               |
| 6201     | Beban Listrik                              |        |                               |
| 6202     | Beban Air                                  |        |                               |
| 6203     | Beban Telepon & Internet                   |        |                               |
| 6204     | Beban Pemeliharaan & Perbaikan Dispenser   |        |                               |
| 6205     | Beban Administrasi Bank                    |        | Biaya rekening operasional    |
| 6206     | Beban ATK & Perlengkapan                   |        |                               |
| 6207     | Beban Penyusutan Aset Tetap                |        | Alokasi penyusutan            |

---

## 4. Jenis Transaksi & Jurnal yang Dihasilkan

### 4.1 Penebusan BBM ke Pertamina (`trans_type = "penebusan"`)

Penebusan adalah pembayaran **advance (uang muka)** ke Pertamina secara online.
Tidak ada hutang — uang langsung keluar, BBM belum tiba.

**Mapping Roles:**

| Role              | D/K    | Per BBM? | Akun Default                            |
| ----------------- | ------ | -------- | --------------------------------------- |
| `debit_uang_muka` | Debit  | Tidak    | 1131 Uang Muka Penebusan Pertamina      |
| `debit_adm_bank`  | Debit  | **Ya**   | 511X Biaya Admin Bank Penebusan — [BBM] |
| `kredit_bank`     | Kredit | Tidak    | 1103 Bank BCA (atau wallet dipilih)     |

**Jurnal yang diposting saat status = CO:**

```
Dr  1131  Uang Muka Penebusan Pertamina    = subtotal + total_ppn [+ sisa adm tak termapping]
Dr  5111  Biaya Admin Bank — Pertalite     = adm_bank × (subtotal Pertalite / total subtotal)
Dr  5112  Biaya Admin Bank — Pertamax      = adm_bank × (subtotal Pertamax / total subtotal)
    Cr  1103  Bank BCA                     = total_bayar
━━━ Debit = Kredit ✓
```

**Saat BBM tiba (modul kedatangan_bbm — belum diimplementasi):**

```
Dr  112X  Persediaan BBM — [Pertalite]     = liter × harga_dasar
Dr  112X  Persediaan BBM — [Pertamax]      = liter × harga_dasar
    Cr  1131  Uang Muka Penebusan          = Σ nilai BBM yang tiba
```

---

### 4.2 Penjualan Tunai (`trans_type = "penjualan_tunai"`)

**Mapping Roles:**

| Role                | D/K    | Per BBM? | Akun Default                            |
| ------------------- | ------ | -------- | --------------------------------------- |
| `debit_kas`         | Debit  | Tidak    | 1101 Kas                                |
| `kredit_persediaan` | Kredit | **Ya**   | 112X Persediaan BBM — [BBM]             |
| `kredit_pendapatan` | Kredit | **Ya**   | 410X Pendapatan Penjualan [BBM] — Tunai |

**Jurnal:**

```
Dr  1101  Kas                              = total penjualan
    Cr  1121  Persediaan BBM — Pertalite   = liter × harga_dasar  (HPP)
    Cr  4101  Pendapatan Pertalite Tunai   = total - HPP  (margin)
━━━ Debit = Kredit ✓
```

---

### 4.3 Penjualan Kredit B2B (`trans_type = "penjualan_kredit"`)

**Mapping Roles:**

| Role                | D/K    | Per BBM? | Akun Default                                 |
| ------------------- | ------ | -------- | -------------------------------------------- |
| `debit_piutang`     | Debit  | Tidak    | 1110 Piutang Dagang B2B                      |
| `kredit_persediaan` | Kredit | **Ya**   | 112X Persediaan BBM — [BBM]                  |
| `kredit_pendapatan` | Kredit | **Ya**   | 411X Pendapatan Penjualan [BBM] — Kredit B2B |

**Jurnal:**

```
Dr  1110  Piutang Dagang B2B              = total tagihan
    Cr  1121  Persediaan BBM — Pertalite  = HPP
    Cr  4111  Pendapatan Pertalite B2B    = margin
━━━ Debit = Kredit ✓
```

---

### 4.4 Kedatangan BBM (`trans_type = "kedatangan_bbm"`)

Stok masuk dari Pertamina — melunasi uang muka.

**Mapping Roles:**

| Role               | D/K    | Per BBM? | Akun Default                       |
| ------------------ | ------ | -------- | ---------------------------------- |
| `debit_persediaan` | Debit  | **Ya**   | 112X Persediaan BBM — [BBM]        |
| `kredit_uang_muka` | Kredit | Tidak    | 1131 Uang Muka Penebusan Pertamina |

---

### 4.5 Pelunasan Piutang B2B (`trans_type = "pelunasan_piutang"`)

**Mapping Roles:**

| Role             | D/K    | Akun Default            |
| ---------------- | ------ | ----------------------- |
| `debit_kas`      | Debit  | 1101 Kas                |
| `kredit_piutang` | Kredit | 1110 Piutang Dagang B2B |

---

### 4.6 Penggajian Karyawan (`trans_type = "payroll"`)

**Mapping Roles:**

| Role                 | D/K    | Akun Default                     |
| -------------------- | ------ | -------------------------------- |
| `debit_gaji`         | Debit  | 6101 Beban Gaji Pokok            |
| `debit_tunjangan`    | Debit  | 6102 Beban Tunjangan Karyawan    |
| `kredit_hutang_gaji` | Kredit | 2104 Hutang Gaji Karyawan        |
| `kredit_bpjs_kes`    | Kredit | 2102 Hutang BPJS Kesehatan       |
| `kredit_bpjs_tk`     | Kredit | 2103 Hutang BPJS Ketenagakerjaan |
| `kredit_pph21`       | Kredit | 2105 Hutang Pajak PPh 21         |

**Jurnal:**

```
Dr  6101  Beban Gaji Pokok               = total gaji pokok
Dr  6102  Beban Tunjangan                = total tunjangan
    Cr  2104  Hutang Gaji Karyawan       = gaji bersih dibayarkan
    Cr  2102  Hutang BPJS Kesehatan      = potongan BPJS Kes karyawan
    Cr  2103  Hutang BPJS TK             = potongan BPJS TK karyawan
    Cr  2105  Hutang Pajak PPh 21        = potongan PPh 21
━━━ Debit = Kredit ✓
```

---

### 4.7 Kasbon Karyawan (`trans_type = "kasbon"`)

**Mapping Roles:**

| Role                   | D/K    | Akun Default                 |
| ---------------------- | ------ | ---------------------------- |
| `debit_piutang_kasbon` | Debit  | 1111 Piutang Kasbon Karyawan |
| `kredit_kas`           | Kredit | 1101 Kas                     |

---

### 4.8 Cash In — Penerimaan Lainnya (`trans_type = "cash_in"`)

**Mapping Roles:**

| Role          | D/K    | Akun Default                |
| ------------- | ------ | --------------------------- |
| `debit_kas`   | Debit  | 1101 Kas                    |
| `kredit_akun` | Kredit | _(dikonfigurasi per kasus)_ |

---

### 4.9 Cash Out — Pengeluaran Lainnya (`trans_type = "cash_out"`)

**Mapping Roles:**

| Role         | D/K    | Akun Default                |
| ------------ | ------ | --------------------------- |
| `debit_akun` | Debit  | _(dikonfigurasi per kasus)_ |
| `kredit_kas` | Kredit | 1101 Kas                    |

---

## 5. Modul Accounting (`AccountingService`)

### Lokasi

`internal/service/accounting_service.go`

### Interface

```go
type AccountingService interface {
    PostTransaction(req PostTransactionRequest) error
    RePostTransaction(req PostTransactionRequest) error   // idempotent
    ReverseTransaction(refType string, refID uint) error
    GetJournalByRef(refType string, refID uint) ([]entity.JournalEntry, error)
    GetLedger(coaID uint, limit int) ([]entity.JournalEntry, error)
    ResolveCOA(transType, role string, bbmID *uint) (uint, error)
}
```

### Cara Integrasi ke Modul Transaksi Baru

```go
// 1. Inject AccountingService ke dalam struct service
type myTrxService struct {
    repo       repository.MyTrxRepository
    accounting service.AccountingService   // ← tambahkan
}

func NewMyTrxService(repo repository.MyTrxRepository, acc service.AccountingService) MyTrxService {
    return &myTrxService{repo: repo, accounting: acc}
}

// 2. Panggil saat dokumen di-complete (RePostTransaction = idempotent)
func (s *myTrxService) PostJournal(trx *entity.MyTrx, createdBy *uint) error {
    return s.accounting.RePostTransaction(service.PostTransactionRequest{
        TransType: "penjualan_tunai",
        RefType:   "penjualan",
        RefID:     uint(trx.ID),
        TransDate: trx.TglJual,
        CreatedBy: createdBy,
        Lines: []service.JournalLine{
            {Role: "debit_kas",         WalletID: &walletID, Debit:  trx.Total},
            {Role: "kredit_persediaan", BBMID: &bbmID,       Credit: trx.HPP},
            {Role: "kredit_pendapatan", BBMID: &bbmID,       Credit: trx.Total - trx.HPP},
        },
    })
}

// 3. Panggil saat dokumen dihapus / void
func (s *myTrxService) ReverseJournal(id uint64) error {
    return s.accounting.ReverseTransaction("penjualan", uint(id))
}

// 4. Wiring di cmd/main.go (accountingService sudah tersedia global)
myTrxService := service.NewMyTrxService(myTrxRepo, accountingService)
```

### Guard Balance

`AccountingService.buildEntries()` secara otomatis memvalidasi:

- `Σ Debit = Σ Kredit` — jika tidak balance, error dikembalikan sebelum INSERT
- Setiap `JournalLine` hanya boleh mengisi Debit **atau** Kredit, tidak keduanya
- COA Mapping yang belum dikonfigurasi menghasilkan error yang mengarahkan user ke menu COA Mapping

---

## 6. Generate COA Otomatis per BBM

Saat BBM baru ditambahkan, tekan tombol **"Generate COA"** di menu Master → BBM.
Sistem akan otomatis membuat:

| Prefix | Akun yang dibuat                         |
| ------ | ---------------------------------------- |
| 112X   | Persediaan BBM — [Nama BBM]              |
| 410X   | Pendapatan Penjualan [Nama] — Tunai      |
| 411X   | Pendapatan Penjualan [Nama] — Kredit B2B |
| 510X   | HPP [Nama BBM]                           |
| 511X   | Biaya Admin Bank Penebusan — [Nama]      |
| 512X   | Selisih & Penyusutan [Nama]              |

Dan sekaligus membuat **COA Mapping** untuk:

- `penebusan / debit_adm_bank` → 511X
- `penjualan_tunai / kredit_pendapatan` → 410X
- `penjualan_kredit / kredit_pendapatan` → 411X
- `penjualan_tunai / kredit_persediaan` → 112X
- `penjualan_kredit / kredit_persediaan` → 112X
- `kedatangan_bbm / debit_persediaan` → 112X

---

## 7. Laporan Keuangan yang Dapat Dihasilkan

Dari data `journal_entries`, laporan berikut dapat digenerate:

| Laporan                          | Query Dasar                                        |
| -------------------------------- | -------------------------------------------------- |
| **Buku Besar per Akun**          | `WHERE coa_id = ? ORDER BY trans_date`             |
| **Neraca Saldo (Trial Balance)** | `GROUP BY coa_id, SUM(debit), SUM(credit)`         |
| **Laba Rugi**                    | Filter COA tipe 4 (Pendapatan) dan 5+6 (Beban+HPP) |
| **Neraca**                       | Filter COA tipe 1 (Aset), 2 (Kewajiban), 3 (Modal) |
| **Mutasi per Dokumen**           | `WHERE ref_type = ? AND ref_id = ?`                |

> Laporan ini belum diimplementasi sebagai endpoint — referensikan dokumen ini
> saat membangun modul laporan keuangan.

---

## 8. Status Dokumen Transaksi

| Kode | Label    | Accounting Effect                                    |
| ---- | -------- | ---------------------------------------------------- |
| `DR` | Draft    | **Tidak ada** — hanya disimpan, belum posting jurnal |
| `CO` | Complete | **Ada** — jurnal diposting ke `journal_entries`      |

Saat dokumen di-**void/hapus**: jurnal otomatis di-**reverse** (semua baris jurnal dihapus).
Saat dokumen di-**edit** lalu CO: jurnal lama di-**reverse** lalu **re-post** yang baru (idempotent).
