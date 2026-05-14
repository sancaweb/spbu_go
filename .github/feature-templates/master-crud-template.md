# Template Cepat — Master Data CRUD

Gunakan template ini untuk modul master seperti Partner, Jabatan, Pendapatan, Potongan, Wallet.

## A) File yang biasanya dibuat/diubah

- `internal/entity/<modul>_entity.go`
- `internal/repository/<modul>_repo.go`
- `internal/service/<modul>_service.go`
- `internal/handler/<modul>_handler.go`
- `templates/master/<modul>/index.html` (atau `templates/<modul>/index.html` mengikuti pola modul)
- `cmd/main.go` (manual migration + AutoMigrate + DI + routes)
- `database/schema.dbml`
- `seeders/seeder.go` (opsional, jika perlu data awal)

## B) Route pattern (POST untuk mutasi)

- `GET /master/<modul>` -> index
- `GET /master/<modul>/archive` -> arsip (jika pakai soft delete)
- `POST /master/<modul>/datatable` -> DataTables server-side
- `POST /master/<modul>` -> create
- `POST /master/<modul>/:id` -> update
- `POST /master/<modul>/:id/delete` -> delete (soft delete)
- `POST /master/<modul>/:id/restore` -> restore (opsional)

## C) Query & repository rules

- List aktif: `WHERE is_active = true`
- List arsip: `Unscoped()` + `is_active = false OR deleted_at IS NOT NULL`
- Search: `ILIKE` untuk kolom utama
- Create/Update: selalu `Omit(...)` relasi (contoh `Omit("Updater")`)

## D) Handler rules

- HTML render kirim: `User`, `Favicon`, `Title`, `ActiveMenu`.
- Datatable pakai `dto.DatatableRequest` / `dto.DatatableResponse`.
- Validasi minimal: required + format + pesan ringkas.
- Isi `UpdatedBy` dari context `user` jika field tersedia di entity.

## E) Template/UI rules

- Daftar data: DataTables server-side.
- Form create/edit pakai modal popup.
- Submit create/update/delete via AJAX/fetch tanpa reload halaman.
- Gunakan helper angka Indonesia dari `templates/includes/footer.html`.
- Register komponen Alpine di `document.addEventListener("alpine:init", ...)`.

## F) Done checklist

- [ ] Build `go build ./...`
- [ ] Test `go test ./...`
- [ ] Route index/datatable/create/update/delete berjalan
- [ ] Search/sort/paging DataTables sesuai
- [ ] Soft delete + restore konsisten (jika diperlukan)
- [ ] `database/schema.dbml` ikut diperbarui
