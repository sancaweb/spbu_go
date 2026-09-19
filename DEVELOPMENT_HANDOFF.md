# Development Handoff — SPBU Go

Dokumen ini adalah titik awal bagi developer atau AI agent berikutnya yang melanjutkan workspace ini.

Last updated: 20 September 2026  
Branch: `main`  
Last committed revision: `b27ec93` (`Untrack and ignore spbu directory`)

## Titik lanjut saat ini

- **Fokus aktif:** `/settings/import-data`. User meminta modul import bertahap untuk migrasi data sistem lama, dimulai dari master BBM.
- **Tahap terakhir:** menu Settings → Import Data, tab modul, konfigurasi koneksi MySQL/MariaDB, download data BBM lama dalam Excel, download template Excel, dan upload/import master BBM sudah tersedia.
- **Verifikasi terbaru:** parser Excel BBM, round-trip template, validasi header/angka/status, render template Gin, `go test ./...`, `go vet ./...`, dan `go build ./...` lulus. Warning dependency MySQL pada `go.mod` telah diperbaiki dengan memindahkannya sebagai dependency langsung; `go mod tidy` dan pemeriksaan lanjutan lulus. Tidak ada koneksi atau data sistem lama yang diubah dalam pengujian.
- **Langkah berikutnya:** user menguji pengisian koneksi sistem lama, Tes Koneksi, Get Data, Get Import Template, dan Import Data BBM dengan salinan Excel. Setelah BBM disetujui, implementasikan tab modul berikutnya sesuai arahan user.
- **Batas bukti:** belum dilakukan pengujian browser langsung maupun koneksi ke database lama yang sebenarnya pada sesi ini. Fitur import tab selain BBM masih berupa placeholder.

## Cara membaca status workspace

Mulai dari [AGENTS.md](AGENTS.md), lalu baca dokumen ini, periksa `git status`, dan cocokkan source dengan catatan sebelum melanjutkan pekerjaan. Agent baru mungkin tidak memiliki percakapan sebelumnya; dokumentasi ini menyimpan konteks yang diperlukan di dalam repository.

Dokumen ini adalah acuan progres terbaru. [PROJECT_STATUS.md](PROJECT_STATUS.md) menyimpan audit 17 September 2026; beberapa pernyataannya, termasuk belum adanya test, sudah tidak menggambarkan workspace sekarang. `TODO.md` dan `guidance.md` juga memuat rencana lama yang perlu dicocokkan dengan implementasi.

Workspace saat ini memiliki banyak perubahan lokal dari beberapa tahap development. Jangan menjalankan `git reset --hard`, `git checkout --`, atau menghapus perubahan tanpa instruksi eksplisit dari user.

## Melanjutkan di perangkat lain

1. Bawa dokumen dan source versi yang sama. Jika memakai Git, review perubahan lokal, commit file yang diperlukan, lalu push dan clone/pull di perangkat tujuan. Commit dokumentasi saja tidak membawa implementasi yang masih belum di-commit. Jika menyalin folder, sertakan working tree terbaru, file baru, serta `.git` untuk membawa histori.
2. Buka folder root `spbu_go` di editor atau aplikasi agent, tempat `AGENTS.md` dan dokumen ini berada.
3. Kirim instruksi awal berikut agar agent membaca konteks, termasuk bila alatnya tidak memuat `AGENTS.md` secara otomatis:

   > Baca AGENTS.md dan DEVELOPMENT_HANDOFF.md di root workspace. Periksa git status dan source yang relevan, lalu jelaskan titik lanjut serta bagian yang belum diverifikasi. Fokus terakhir adalah /settings/import-data, khususnya import master BBM. Lanjutkan sesuai catatan dan arahan saya; jangan mengimplementasikan tab modul lain tanpa instruksi.

4. Untuk menjalankan aplikasi, siapkan Go sesuai `go.mod`, PostgreSQL, serta `config/.env` atau environment lokal yang sesuai. Database berjalan terpisah dari folder Git: data uji yang diperlukan harus tersedia di perangkat tujuan. Transfer konfigurasi berisi credential melalui cara privat, bukan commit Git.

Saat panduan ini diperbarui pada 20 September 2026, `AGENTS.md` dan `DEVELOPMENT_HANDOFF.md` belum di-commit/push. Perubahan source terbaru juga masih berada pada working tree. Tidak ada operasi commit/push yang dilakukan dalam sesi implementasi import data ini.

## Kondisi umum proyek

- Backend: Go 1.24, Gin, GORM, PostgreSQL.
- Frontend: server-side HTML template, Alpine.js, Tailwind CDN, jQuery DataTables, SweetAlert.
- Struktur backend utama: `internal/entity`, `internal/repository`, `internal/service`, `internal/handler`.
- Inisialisasi schema saat startup masih menggunakan SQL manual di `cmd/main.go` dan GORM `AutoMigrate`.
- Database utama dikonfigurasi melalui `config/.env`. Jangan menyalin credential ke dokumentasi, source code, log, atau commit.
- User sedang melakukan verifikasi modul satu per satu. Fokus aktif terakhir adalah `/settings/import-data`, khususnya master BBM; jangan beralih ke implementasi tab modul lain tanpa arahan user.

## Pekerjaan yang sudah dilakukan

### Master BBM

- Datatable BBM sudah diarahkan untuk mengurutkan data berdasarkan abjad pada kolom `name`.
- Data BBM yang dipakai pada proses penjualan dibatasi pada data aktif.

### Transaksi penjualan

- Halaman `/transaction/penjualan/create` memiliki detail penjualan per nozzle/tiang.
- Tab `Piutang B2B` dipertahankan pada form penjualan.
- Piutang B2B menggunakan multiple row dengan partner/konsumen dan nominal tagihan.
- Total piutang ditampilkan pada footer tabel.
- Pengeluaran test tetap diinput manual.
- Saat form penjualan disimpan, alur bisnis mencakup:
  1. header dan detail penjualan;
  2. header piutang B2B;
  3. pengeluaran test.
- Partner dan BBM yang muncul pada form transaksi diambil dari data aktif.

### Upload laporan penjualan

- Upload laporan penjualan tersedia pada `/transaction/penjualan/create`.
- Format utama upload dan template download adalah `.xlsx`.
- Upload hanya mengisi field totalisator pada form di browser.
- Upload tidak langsung menyimpan transaksi ke database; penyimpanan baru terjadi ketika user menekan tombol `Simpan`.
- Piutang B2B dan Pengeluaran Test tidak diambil dari file upload dan tetap diinput manual.
- Parser dan generator XLSX dibuat tanpa dependency spreadsheet eksternal, menggunakan library standard Go.
- Template Excel dibuat berdasarkan nozzle aktif dan memiliki kolom:
  - `nozzle_id`
  - `nozzle`
  - `totalisator_awal`
  - `totalisator_akhir`
- Validasi upload mencakup ukuran file maksimal 5 MB dan kewajiban memuat seluruh nozzle aktif.
- Loading upload sudah memiliki spinner tombol dan overlay pada panel upload.

File terkait:

- `internal/handler/penjualan_upload.go`
- `internal/handler/penjualan_upload_test.go`
- `internal/handler/testdata/penjualan_6846_upload.xlsx`
- `templates/transaction/penjualan/form.html`
- Route berada di `cmd/main.go`:
  - `GET /transaction/penjualan/upload/template`
  - `POST /transaction/penjualan/upload`

### Data testing penjualan

- Data transaksi sample dengan ID penjualan `6846` digunakan sebagai referensi/testing.
- Sample upload tersedia dalam `internal/handler/testdata/`.
- Jangan memasukkan credential database sample ke source atau dokumentasi.

### ERD dan dokumentasi database

- ERD aktual database tersedia di `database/erd_actual.dbml`.
- Migrations awal tersedia di folder `migrations/`.
- Perlu diingat bahwa schema runtime belum sepenuhnya dipindahkan ke migration versioned; sebagian masih dibuat/diperbarui dari `cmd/main.go`.

### Warning editor Go

- Warning `writestring` pada `internal/handler/penjualan_upload.go` sudah diperbaiki.
- Konkatenasi string di dalam `WriteString` diganti dengan `fmt.Fprintf` atau penulisan bertahap.

## Modul partner — perubahan terbaru

Modul yang sedang diverifikasi user adalah partner; halaman aktif terakhir `/master/partner/archive`.

Field yang sudah ditambahkan:

- `start_date` — tipe database `DATE`, nullable.
- `end_date` — tipe database `DATE`, nullable.
- `isactive` — tipe database `BOOLEAN NOT NULL DEFAULT TRUE`.
- `auto_off` — tipe database `BOOLEAN NOT NULL DEFAULT FALSE`.

> Catatan penting: kolom status partner diminta bernama `isactive`, bukan `is_active`.

Implementasi berada pada:

- `internal/entity/partner_entity.go`
- `internal/handler/partner_handler.go`
- `internal/repository/partner_repo.go`
- `templates/master/partner/index.html`
- `cmd/main.go`

Perilaku yang sudah diterapkan:

- Form create/edit memiliki Start Date, End Date, Status, dan Auto Off.
- Datatable dan detail partner menampilkan field baru.
- End Date tidak boleh lebih awal dari Start Date.
- End Date wajib diisi jika Auto Off diaktifkan.
- Saat partner dibaca melalui repository, partner dengan `auto_off = true` dan `end_date < CURRENT_DATE` akan diperbarui menjadi `isactive = false`.
- Partner yang statusnya tidak aktif tidak muncul pada daftar partner aktif untuk transaksi.
- Tanggal End Date masih dianggap aktif sampai tanggal tersebut selesai; penonaktifan terjadi pada hari berikutnya.
- Start Date dicatat sebagai informasi periode kerja sama; belum ada penjadwalan aktivasi berdasarkan tanggal mulai.

> Auto Off saat ini dievaluasi ketika data partner diakses, bukan melalui background scheduler. Jadi database akan berubah saat halaman master/transaksi atau endpoint partner berikutnya melakukan pembacaan. Pembanding tanggal adalah `CURRENT_DATE` PostgreSQL; zona waktu database perlu diperiksa saat verifikasi di perangkat baru.

Migrasi kompatibilitas di `cmd/main.go`:

- Tabel baru memakai kolom `isactive`.
- Database lama yang masih memiliki `is_active` akan mengganti nama kolom tersebut menjadi `isactive` saat startup.
- Kolom `start_date`, `end_date`, dan `auto_off` ditambahkan jika belum ada.

### Hapus permanen pada arsip partner

Implementasi tanggal 20 September 2026:

- Cakupan arsip mengikuti query yang sudah ada: `isactive = false` **atau** `deleted_at IS NOT NULL`.
- Datatable mengembalikan flag boolean `can_delete_permanent`, dihitung dengan `NOT EXISTS` pada `trx_piutang.pelanggan_id = partners.id`. Flag berada di DTO, bukan kolom baru pada tabel partner; tidak ada migrasi tambahan.
- Tombol **Hapus Permanen** hanya dirender di halaman arsip ketika flag bernilai `true`. Tombol aktifkan kembali tetap tersedia.
- Konfirmasi menjelaskan bahwa penghapusan tidak dapat dikembalikan, dengan tombol batal dan indikator proses. Klik berulang dicegah selama dialog/proses berlangsung.
- Endpoint baru: `POST /master/partner/:id/delete-permanent`, berada dalam kelompok route terautentikasi yang sama dengan partner lainnya.
- Repository mengunci baris partner (`FOR UPDATE`), memeriksa ulang status arsip serta keberadaan piutang, lalu melakukan hard delete di dalam satu transaksi database. Pengecekan tidak dibatasi status atau nominal piutang; header piutang saja sudah cukup untuk menolak.
- Partner aktif ditolak; partner yang tidak ditemukan menghasilkan 404; konflik status/relasi menghasilkan 409. Kegagalan pemeriksaan database tidak melanjutkan penghapusan. Pelanggaran foreign key membatalkan transaksi.
- Jika status berubah sejak datatable dimuat, penolakan 409/404 menyegarkan tabel tanpa menampilkan pesan sukses palsu.
- Pemeriksaan metadata database lokal secara read-only memastikan foreign key `trx_piutang.pelanggan_id -> partners.id` dengan `ON DELETE RESTRICT` dan index pada `pelanggan_id` tersedia.

### Modul Import Data

Implementasi tanggal 20 September 2026, dimulai dari master BBM:

- Menu baru berada di `/settings/import-data` dan ditautkan di bagian Settings pada sidebar.
- Halaman memakai tab per modul. Tab Master BBM aktif; tab Master Tiang, Nozzle, Partner, Karyawan, Jabatan, Pendapatan, Potongan, Shift, Jenis Test, Wallet, Kedatangan BBM, Penebusan, Penjualan, Piutang, dan Penyusutan sudah disiapkan sebagai placeholder untuk tahap berikutnya.
- Form koneksi menerima driver MySQL/MariaDB, host, port, database, username, password, dan mode SSL. Password disimpan terenkripsi AES-GCM menggunakan `SESSION_SECRET` dan tidak dikirim kembali ke browser; field password kosong berarti memakai password tersimpan.
- Tombol **Tes Koneksi** menguji koneksi MySQL/MariaDB dengan timeout. Mode SSL dipetakan ke konfigurasi driver (`false`, `preferred`, `true`, atau `skip-verify`).
- Tombol **Get Data** menjalankan query read-only ke tabel lama `tb_bbm`, membaca kolom `idBbm`, `namaBbm`, `margin`, `hargaJual`, `stokLiter`, `rewardPersen`, dan `status`, kemudian mengunduh `data_bbm_sistem_lama.xlsx`.
- Tombol **Get Import Template** mengunduh `template_import_bbm.xlsx` dengan header `source_id`, `name`, `margin`, `price`, `stock`, `reward_percent`, dan `is_active`.
- **Import Data** hanya menerima `.xlsx` maksimal 5 MB. Data divalidasi, di-upsert berdasarkan nama BBM (case-insensitive), dan disimpan dalam satu transaksi PostgreSQL. `source_id` hanya referensi dan tidak dipakai sebagai ID target.
- Import dilakukan dari file Excel yang dipilih; proses upload tidak otomatis membaca atau mengubah database lama. Pengambilan data lama hanya terjadi melalui Get Data.
- Parser Excel bersama pada `internal/handler/penjualan_upload.go` dipakai oleh import BBM; implementasi generator/parser BBM berada di `internal/handler/import_data_xlsx.go`.
- Endpoint utama: `GET /settings/import-data`, `POST /settings/import-data/connection`, `POST /settings/import-data/connection/test`, `GET /settings/import-data/bbm/data`, `GET /settings/import-data/bbm/template`, `POST /settings/import-data/bbm/import`. Semua berada di route terautentikasi.

File penting:

- `internal/entity/legacy_db_connection_entity.go`
- `internal/dto/import_data.go`
- `internal/repository/legacy_db_connection_repo.go`
- `internal/service/import_data_service.go`
- `internal/handler/import_data_handler.go`
- `internal/handler/import_data_xlsx.go`
- `internal/handler/import_data_xlsx_test.go`
- `templates/settings/import_data.html`
- `templates/includes/header.html`
- `cmd/main.go`

File utama tambahan/perubahan:

- `internal/dto/partner.go`
- `internal/entity/partner_entity.go`
- `internal/repository/partner_repo.go`
- `internal/service/partner_service.go`
- `internal/handler/partner_handler.go`
- `templates/master/partner/index.html`
- `cmd/main.go`
- `internal/repository/partner_repo_test.go`
- `internal/handler/partner_handler_test.go`
- `tests/partner_archive.test.cjs`

## Validasi terakhir

Pada sesi implementasi import data tanggal 20 September 2026, pemeriksaan berikut berhasil:

- `go test ./... -count=1` dengan `SPBU_PARTNER_DB_TEST=1`.
- `go vet ./...`
- `go build ./...`
- `node --test tests/partner_archive.test.cjs` — tujuh test logika UI lulus.
- `go test ./internal/handler ./internal/server -count=1` — test parser Excel dan render template import lulus.
- `git diff --check` — tidak ada error whitespace; Git menampilkan pemberitahuan normalisasi LF/CRLF.

Test repository partner mencakup arsip tanpa piutang, piutang belum lunas/lunas/header saja, partner aktif, ID tidak ditemukan, flag datatable serta field partner/updater yang sudah ada, perubahan kelayakan setelah daftar dibaca, kegagalan pemeriksaan piutang, dan rollback jika ada referensi foreign key lain. Test handler mencakup respons sukses/error serta validasi ID. Test import BBM mencakup round-trip XLSX, header alternatif, escaping XML, status aktif, dan penolakan file template tanpa data. Test server menjalankan render template import dari working tree.

Test integrasi PostgreSQL bersifat opt-in. Konfigurasi berasal dari environment `DB_*` atau `config/.env`; jangan menyalin credential ke kode. Koneksi khusus memakai `search_path=pg_temp` dan tabel sementara, tanpa menjalankan startup/migrasi aplikasi atau mengubah tabel bisnis asli. Contoh PowerShell dari root repository:

```powershell
$env:SPBU_PARTNER_DB_TEST = '1'
go test ./internal/repository ./internal/handler -run TestPartner -count=1
node --test tests/partner_archive.test.cjs
```

Tanpa flag tersebut, test repository PostgreSQL dilewati; test handler tetap berjalan. Test sample penjualan/upload/penyusutan yang sebelumnya ada juga lulus saat suite Go dijalankan penuh. Browser end-to-end, uji race paralel, verifikasi semua field partner, serta migrasi field partner pada database lama belum dinyatakan selesai.

## Jejak Git dan perubahan lokal

Histori commit utama yang tercatat:

| Commit | Tahap |
| --- | --- |
| `f6a1828` | Fondasi SPBU, auth/master awal, seeder, template dasar |
| `e302f38` | Modul penebusan BBM dan sebagian employee |
| `4331466` | Progress kedatangan BBM |
| `b9fdae0` | Penebusan selesai, penjualan mulai |
| `a11d1dc` | Progress transaksi penjualan |
| `0068e12` | Penjualan dan piutang |
| `96773b1` | Progress penyusutan dan perbaikan `is_cross_day` shift |
| `b27ec93` | Folder legacy `spbu` di-ignore |

Perubahan terbaru dari sesi-sesi development berada di working tree dan belum seluruhnya menjadi commit. File yang relevan antara lain:

- `cmd/main.go`
- `internal/entity/partner_entity.go`
- `internal/dto/partner.go`
- `internal/handler/partner_handler.go`
- `internal/handler/partner_handler_test.go`
- `internal/repository/partner_repo.go`
- `internal/repository/partner_repo_test.go`
- `internal/service/partner_service.go`
- `templates/master/partner/index.html`
- `tests/partner_archive.test.cjs`
- `internal/handler/penjualan_upload.go`
- `internal/handler/penjualan_upload_test.go`
- `templates/transaction/penjualan/form.html`
- file testing/sample pada `internal/handler/testdata/`
- perubahan pada modul penyusutan, piutang, penjualan, dan shift dari pekerjaan sebelumnya

Status aktual harus selalu dikonfirmasi dengan:

```powershell
git status --short
git log -5 --oneline --decorate
```

Jangan menganggap semua perubahan lokal berasal dari sesi terakhir. Workspace ini memang pernah berpindah antar-agent, sehingga perubahan lama dan perubahan baru dapat bercampur.

## Pekerjaan yang masih perlu dilakukan

Prioritas terdekat:

1. Uji UI `/settings/import-data` secara manual dengan koneksi/database lama yang aman:
   - simpan konfigurasi koneksi, Tes Koneksi berhasil/gagal dengan pesan yang sesuai, dan pastikan password tidak tampil kembali;
   - Get Import Template menghasilkan `.xlsx` yang dapat dibuka dan memiliki header yang terdokumentasi;
   - Get Data menghasilkan `.xlsx` dari `tb_bbm` tanpa menulis database lama;
   - edit salinan Excel, upload melalui Import Data, dan pastikan jumlah created/updated benar;
   - file invalid, nama kosong, angka invalid, status invalid, duplikat nama, dan file >5 MB ditolak tanpa perubahan parsial;
   - jangan memakai password database produksi untuk uji coba tanpa persetujuan dan backup.
2. Verifikasi ulang `/master/partner/archive` yang masih tercatat pending:
   - tombol Hapus Permanen hanya muncul untuk arsip tanpa piutang;
   - piutang lunas maupun belum lunas tetap menyembunyikan tombol;
   - batal tidak menghapus, konfirmasi menghapus hanya partner yang dipilih;
   - daftar disegarkan dan partner hilang setelah penghapusan berhasil;
   - permintaan langsung atau halaman yang sudah kedaluwarsa tetap ditolak jika partner memiliki piutang atau sudah aktif kembali.
3. Catatan verifikasi `/master/partner` dari tahap sebelumnya yang belum selesai; kerjakan sesuai arahan user:
   - tambah partner dengan semua field;
   - edit partner;
   - simpan status `false` saat create maupun update, lalu baca kembali dari database;
   - kosongkan tanggal yang sebelumnya terisi, serta uji rentang tanggal tidak valid;
   - Auto Off dengan End Date lampau, hari ini, dan mendatang; pastikan tanpa End Date ditolak;
   - daftar aktif dan archive, termasuk cara memperbarui masa kerja sama partner yang telah kedaluwarsa;
   - penggunaan partner aktif pada transaksi.
4. Pastikan migrasi `is_active` ke `isactive` berhasil pada database yang sudah berisi data, bila user melanjutkan verifikasi tahap tersebut.
5. Catat hasil verifikasi modul dan konfirmasi hasil dengan user sebelum mengimplementasikan tab import berikutnya. Commit/push mengikuti arahan user.

Backlog di luar fokus aktif; kerjakan setelah user mengarahkan perpindahan modul:

- Verifikasi alur upload XLSX pada `/transaction/penjualan/create` menggunakan sample Excel.
- Rekonsiliasi `TODO.md` dengan `checksheet-piutang-rekap.md`; keduanya masih memuat status yang belum sepenuhnya sinkron.

## Aturan kerja untuk agent berikutnya

- Baca dokumen ini sebelum melakukan perubahan.
- Ikuti [AGENTS.md](AGENTS.md) dan instruksi terbaru user.
- Pertahankan perubahan lokal yang sudah ada.
- Gunakan `apply_patch` untuk edit file.
- Setelah perubahan Go, jalankan minimal `gofmt`, `go test ./...`, dan bila relevan `go vet ./...` serta `go build ./...`.
- Jangan mencetak credential dari `config/.env` ke output atau dokumentasi.
- Jangan melakukan destructive Git operation tanpa persetujuan user.
- Jika user memindahkan workspace tanpa commit/push, pastikan seluruh working tree ikut disalin; clone repository saja hanya membawa commit yang sudah dipush.
- Setelah pekerjaan material, perbarui bagian **Titik lanjut saat ini** dan hasil verifikasi agar agent berikutnya memperoleh konteks terbaru.
