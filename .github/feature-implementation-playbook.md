# Feature Implementation Playbook — spbu_go

Tujuan: menjaga konsistensi gaya implementasi fitur baru (arsitektur, query, routing, template, dan UX) sesuai pattern project saat ini.

## 0) Pre-flight (wajib)

- Tarik konteks dari [cmd/main.go](../cmd/main.go) untuk melihat modul aktif + pola route.
- Cek pola modul serupa yang paling dekat (copy gaya, jangan invent pola baru).
- Jalankan validasi minimal:
  - `go build ./...`
  - `go test ./...`

## 1) Pilih tipe fitur

- **Master data CRUD**: ikuti pola Partner/Karyawan.
- **Transaksi + detail + list**: ikuti pola Penebusan.
- **Konfigurasi/key-value**: ikuti pola Settings.

## 2) Scaffold backend berurutan (jangan lompat)

1. Entity di [internal/entity](../internal/entity)
2. Repository di [internal/repository](../internal/repository)
3. Service di [internal/service](../internal/service)
4. Handler di [internal/handler](../internal/handler)
5. Wiring + route di [cmd/main.go](../cmd/main.go)

Aturan:

- Setiap entity wajib punya `TableName()`.
- Untuk tabel bisnis, gunakan `gorm.DeletedAt` + audit `UpdatedBy`/`Updater` jika relevan.
- Di repository `Create()`/`Update()`, gunakan `Omit(...)` untuk relasi agar tidak terjadi nested insert tak sengaja.

## 3) Perubahan schema (aturan project ini)

- Tambahkan SQL migration manual (idempotent) di startup [cmd/main.go](../cmd/main.go):
  - `CREATE TABLE IF NOT EXISTS ...`
  - blok `DO $$ ... $$` untuk rename kolom/constraint/backward compatibility.
- Daftarkan model di `AutoMigrate(...)` dalam file yang sama.
- Sinkronkan dokumentasi schema di [database/schema.dbml](../database/schema.dbml).
- Catatan: SQL dalam [migrations](../migrations) saat ini tidak otomatis dieksekusi runtime.

## 4) Contract repository/service

- Interface + implementasi repository diletakkan per file modul (contoh: [internal/repository/partner_repo.go](../internal/repository/partner_repo.go)).
- Shared service interface tetap di [internal/service/interfaces.go](../internal/service/interfaces.go).
- Service interface khusus modul boleh di file modulnya (contoh: [internal/service/partner_service.go](../internal/service/partner_service.go)).

## 5) Handler HTTP pattern

- Render HTML selalu sertakan `User`, `Favicon`, `Title`, `ActiveMenu`.
- Mutasi data gunakan `POST` (create/update/delete/restore).
- Untuk list besar, pakai endpoint DataTables server-side:
  - bind request pakai `dto.DatatableRequest`
  - response pakai `dto.DatatableResponse`
  - search gunakan PostgreSQL `ILIKE`.
- Untuk archive/restore, pakai soft delete + `.Unscoped()` di repository.

## 6) Routing & dependency injection

- Tambahkan constructor wiring di [cmd/main.go](../cmd/main.go) urutan: repo → service → handler.
- Tambahkan route pada group yang tepat:
  - master: `/master/...`
  - transaksi: `/transaction/...`
- Ikuti naming `ActiveMenu` yang konsisten dengan sidebar template.

## 7) Frontend template pattern

- Lokasi template sesuai domain, path relatif `templates/`.
- Register Alpine di `document.addEventListener("alpine:init", ...)`.
- Standar CRUD UI: create/update/delete dilakukan via AJAX/fetch tanpa reload halaman.
- Untuk modul CRUD, form input utama ditampilkan dalam popup modal (bukan pindah page), kecuali ada alasan UX kuat.
- Jika perlu data server ke JS:
  - encode di handler (`json.Marshal`)
  - embed lewat `<script type="application/json">...`.
- Untuk tabel list, pakai DataTables server-side (lihat pola partner/employee/penebusan).
- Gunakan helper angka Indonesia yang sudah ada di [templates/includes/footer.html](../templates/includes/footer.html): `formatIDR`, `formatStock`, `parseIDR`, `formatInputIDR`, `formatInputStock`.

### Konvensi dokumen transaksi (wajib, unless domain khusus)

- Status dokumen disederhanakan ke 2 kode:
  - `DR` = Draft
  - `CO` = Complete
- Field dropdown status di form create **tidak ditampilkan**.
- Footer form create transaksi minimal memiliki 3 aksi:
  1. `Batal` → jika data belum tersimpan: batalkan input, reset form + tutup modal.
     Setelah data sudah tersimpan, tombol ini berubah menjadi `Close` dengan aksi clear form + tutup modal.
  2. `Save` → simpan status `DR` (belum memberi efek ke flow bisnis lanjutan/accounting effect).
  3. `Save & Complete` → simpan status `CO` (siap diproses ke flow bisnis berikutnya).

## 8) Validasi dan formatting input

- Validasi minimum di handler (required, format tanggal, enum, numeric).
- Normalisasi data yang jadi kebiasaan domain (contoh: title-case nama, sanitasi nomor HP).
- Gunakan pesan error yang ringkas dan konsisten (ID tidak valid, data tidak ditemukan, dll).

## 9) Seeder & data awal

- Jika modul butuh data awal, tambahkan idempotent seeder di [seeders/seeder.go](../seeders/seeder.go) dengan `FirstOrCreate`.
- Hindari `Create` langsung untuk seed agar aman saat startup berulang.

## 10) Checklist Definition of Done

- [ ] Build lulus: `go build ./...`
- [ ] Test sweep jalan: `go test ./...`
- [ ] Route bekerja (index, datatable, create, update, delete, restore bila ada)
- [ ] Query list pakai `ILIKE` + pagination/sorting sesuai DataTables
- [ ] Relasi aman (`Omit(...)` dipakai saat save)
- [ ] Soft delete + archive flow konsisten
- [ ] Template render tanpa error, `ActiveMenu` tepat
- [ ] [database/schema.dbml](../database/schema.dbml) ikut diperbarui
- [ ] Seeder tetap idempotent

## 11) Anti-pattern yang harus dihindari

- Menaruh business logic SQL langsung di handler.
- Menambah pola response baru yang berbeda tanpa alasan kuat.
- Menggunakan hard delete untuk data yang seharusnya bisa diarsipkan.
- Menambahkan endpoint `PUT/PATCH/DELETE` untuk modul yang UI-nya berbasis form POST.
- Menyimpan angka terformat Indonesia ke DB tanpa parsing ke nilai mentah.

## 12) Template cepat (siap pakai)

- Master CRUD: [.github/feature-templates/master-crud-template.md](feature-templates/master-crud-template.md)
- Transaksi (header + detail + list): [.github/feature-templates/transaction-template.md](feature-templates/transaction-template.md)
- Settings / key-value: [.github/feature-templates/settings-kv-template.md](feature-templates/settings-kv-template.md)

Cara pakai cepat:

1. Pilih template sesuai tipe fitur.
2. Ikuti daftar file + route pattern dari template.
3. Cocokkan naming `ActiveMenu`, DTO datatable, dan pola `Omit(...)` dengan modul pembanding.
