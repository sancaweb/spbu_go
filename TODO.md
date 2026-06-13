# TODO - Perbaikan Endpoint `/transaction/piutang/rekap`

## Checklist Task User

- [ ] Tambahkan filter pemilihan periode bulan.
- [ ] Default data yang tampil adalah bulan berjalan.
- [ ] Card rekap existing dijadikan collapsible, default tertutup.
- [ ] Tambahkan card baru di bawah card existing untuk pengelompokan tabel per tanggal pada bulan dipilih.
- [ ] Tiap kelompok tabel memuat:
  - [ ] Tanggal
  - [ ] No
  - [ ] Data Piutang (Pelanggan, Tagihan)
  - [ ] Data Pembayaran Piutang (Pelanggan, Pembayaran, Periode Piutang)
- [ ] Pada akhir tiap kelompok tampil total Piutang dan total Pembayaran.
- [ ] Format tabel pengelompokan mengikuti referensi gambar user.
- [ ] Di akhir halaman tampil Grand Total Piutang dan Grand Total Pembayaran Piutang bulan dipilih.
- [ ] Proses filter tanpa reload/refresh halaman.
- [ ] Design card mengikuti style yang sudah ada pada project.

## Information Gathered

- [x] `templates/transaction/piutang/rekap.html`:
  - Sudah ada stats cards + DataTable rekap partner.
  - Belum ada filter bulan.
  - Belum ada section grouped per tanggal.
  - Masih statis ke endpoint summary/datatable tanpa parameter periode.
- [x] `internal/handler/piutang_handler.go`:
  - Sudah ada `Rekap`, `DatatableRekap`, `Summary`.
  - Belum ada endpoint grouped data per tanggal untuk rekap bulan.
- [x] `internal/repository/piutang_repo.go`:
  - Sudah ada `DatatableRekapRows` dan `Summary` (global).
  - Belum ada metode agregasi grouped-by-date untuk pasangan data piutang vs pembayaran.
- [x] `cmd/main.go` routes:
  - Sudah ada route `/transaction/piutang/rekap`, `/transaction/piutang/rekap/datatable`, `/transaction/piutang/summary`.
  - Belum ada route API khusus grouped rekap bulanan.

## Plan (File-level)

- [ ] `internal/repository/piutang_repo.go`
  - [ ] Tambah struct DTO row/group untuk grouped rekap bulanan per tanggal.
  - [ ] Tambah query repository untuk:
    - [ ] summary berdasarkan bulan (stats cards)
    - [ ] grouped data per tanggal (piutang + pembayaran)
    - [ ] grand total bulanan.
  - [ ] Update query `DatatableRekapRows` agar bisa filter bulan (default bulan berjalan).
- [ ] `internal/service/piutang_service.go`
  - [ ] Extend interface + implement method baru untuk summary bulan dan grouped rekap bulan.
- [ ] `internal/handler/piutang_handler.go`
  - [ ] Tambah endpoint JSON baru untuk grouped rekap bulanan.
  - [ ] Ubah `Summary` agar menerima parameter bulan (fallback bulan berjalan).
- [ ] `cmd/main.go`
  - [ ] Tambah route endpoint grouped rekap bulan.
- [ ] `templates/transaction/piutang/rekap.html`
  - [ ] Tambah filter periode bulan.
  - [ ] Ubah card rekap existing jadi collapsible (default close).
  - [ ] Tambah card baru berisi table grouped per tanggal sesuai format.
  - [ ] Tambah subtotal per kelompok + grand total bawah halaman.
  - [ ] Integrasi ajax fetch (tanpa reload) saat filter berubah / submit.

## Dependent Files to be Edited

- [ ] `internal/repository/piutang_repo.go`
- [ ] `internal/service/piutang_service.go`
- [ ] `internal/handler/piutang_handler.go`
- [ ] `cmd/main.go`
- [ ] `templates/transaction/piutang/rekap.html`

## Follow-up Steps

- [ ] Build validation: `go build ./...`
- [ ] Testing UI/API sesuai preferensi user (bisa skip jika diminta).
