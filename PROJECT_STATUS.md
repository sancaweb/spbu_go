# Status Proyek SPBU Go

> Dokumen ini adalah snapshot historis audit 17 September 2026. Untuk melanjutkan pekerjaan, baca [AGENTS.md](AGENTS.md) lalu [DEVELOPMENT_HANDOFF.md](DEVELOPMENT_HANDOFF.md). Status test, fitur, dan working tree di bawah hanya berlaku pada saat audit tersebut.

Snapshot audit: 17 September 2026  
Branch: `main`  
Commit terakhir: `b27ec93` (`origin/main`)

## Ringkasan

Proyek sudah melewati tahap starter dan memiliki monolith Go yang berjalan dengan Gin, GORM, PostgreSQL, server-side HTML template, session authentication, repository, service, handler, middleware, entity, seeder, serta beberapa modul bisnis.

Validasi lokal saat audit:

- `go test ./...` berhasil untuk seluruh package.
- `go build ./...` berhasil.
- `go vet ./...` berhasil setelah diarahkan ke cache lokal proyek; cache global Windows tidak dapat ditulis oleh environment ini.
- Belum ada file `*_test.go`; hasil tersebut hanya memvalidasi kompilasi package.
- Banyak file Go lama belum mengikuti output `gofmt`; jangan melakukan format massal bersamaan dengan penyelesaian fitur penyusutan.

## Jejak development

| Commit | Tahap |
| --- | --- |
| `f6a1828` | Fondasi SPBU, auth/master awal, seeder, template dasar |
| `e302f38` | Modul penebusan BBM dan sebagian employee |
| `4331466` | Progress kedatangan BBM |
| `b9fdae0` | Penebusan ditutup, penjualan mulai |
| `a11d1dc` | Progress transaksi penjualan |
| `0068e12` | Penjualan dan piutang; rekap piutang masih menjadi pekerjaan lanjutan |
| `96773b1` | Progress penyusutan, perbaikan `is_cross_day` pada shift |
| `b27ec93` | Folder legacy `spbu` di-ignore dan tidak lagi dilacak |

## Status fitur

### Sudah tersedia di route dan source

- Login, logout, session, user, role, permission, middleware auth.
- Master BBM, tiang, nozzle, partner, employee, jabatan, pendapatan, potongan, shift, jenis test.
- Keuangan: wallet, COA, mapping COA, journal/accounting service.
- Transaksi: penebusan BBM, kedatangan BBM, penjualan, piutang, stok DO.
- Halaman dashboard, settings, template layout, seeder data contoh.

### Sedang berjalan / perlu verifikasi

- **Penyusutan:** perubahan lokal menambahkan report shift, token edit, halaman edit, dan penyimpanan stok awal/stok akhir aktual. Perubahan ini belum menjadi commit dan belum diuji terhadap database/UI nyata.
- **Rekap piutang:** halaman dan endpoint grouped sudah ada di source, tetapi `TODO.md` masih menggambarkan pekerjaan sebagai belum selesai. `checksheet-piutang-rekap.md` menyatakan tiga perubahan UI selesai, sedangkan empat regression check masih pending. Dokumentasi perlu dianggap belum final sampai regression test dijalankan.

## Kondisi Git saat audit

Perubahan fitur lokal yang harus dipertahankan:

- `cmd/main.go`
- `internal/handler/penyusutan_handler.go`
- `internal/repository/penyusutan_repo.go`
- `internal/service/penyusutan_service.go`
- `templates/master/shift/index.html`
- `templates/transaction/penyusutan/index.html`
- `internal/helper/`
- `templates/transaction/penyusutan/edit.html`

Housekeeping yang dipisahkan dari kode fitur:

- `.gocache/` sebelumnya ikut ter-track sebanyak 1.150 file.
- `config/.env` ikut ter-track; file lokal ini tidak boleh masuk version control.
- `main.exe`, `fix_template.exe`, log, dan file `.tmp_*` adalah artefak lokal.
- Folder `spbu/` sudah di-ignore oleh commit terakhir dan dibiarkan sebagai folder legacy lokal.

## Risiko teknis prioritas

1. Secret token penyusutan masih hard-coded di `internal/helper`; pindahkan ke konfigurasi/secret environment sebelum dipakai di production.
2. `config/.env` pernah masuk histori Git. Meng-ignore file sekarang tidak menghapus histori; credential yang pernah aktif perlu dirotasi dan histori hanya dibersihkan bila memang dibutuhkan.
3. Migrasi SQL belum merepresentasikan seluruh entity. Startup masih melakukan banyak SQL manual dan `AutoMigrate`, sehingga deployment belum memiliki satu sumber kebenaran schema.
4. Tidak ada automated test, CI workflow, README setup, atau prosedur migration/seed yang terdokumentasi sebagai jalur resmi.
5. Query dan perubahan lokal penyusutan perlu regression test untuk tanggal lintas hari, penerimaan per shift, token invalid, dan konsistensi stok.

## Urutan kerja yang direkomendasikan

1. Commit/pisahkan perubahan lokal penyusutan setelah review dan uji UI/API.
2. Jalankan regression checklist rekap piutang dan perbarui `TODO.md`/checksheet agar tidak kontradiktif.
3. Tambahkan test repository/service/handler minimal untuk modul transaksi utama.
4. Pindahkan schema startup ke migration versioned dan hilangkan ketergantungan pada `AutoMigrate` untuk production.
5. Externalize secret, rotasi credential yang pernah tersimpan, lalu rapikan format Go secara bertahap per modul.
