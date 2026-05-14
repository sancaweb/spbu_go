# Template Cepat — Transaksi (Header + Detail + List)

Gunakan template ini untuk modul transaksi seperti Penebusan dan modul transaksi baru lain.

## A) File yang biasanya dibuat/diubah

- `internal/entity/<trx>_entity.go` (header + detail)
- `internal/repository/<trx>_repo.go`
- `internal/service/<trx>_service.go`
- `internal/handler/<trx>_handler.go`
- `templates/transaction/<trx>/index.html`
- `cmd/main.go` (manual migration + AutoMigrate + DI + routes)
- `database/schema.dbml`
- `seeders/seeder.go` (opsional)

## B) Route pattern

- `GET /transaction/<trx>` -> halaman list + form
- `POST /transaction/<trx>/datatable` -> DataTables server-side
- `GET /transaction/<trx>/:id/detail` -> detail AJAX
- Jika ada mutasi:
  - `POST /transaction/<trx>` -> create
  - `POST /transaction/<trx>/:id` -> update
  - `POST /transaction/<trx>/:id/delete` -> delete/void

## C) Entity & data rules

- Header/detail dipisah tabel; detail `ON DELETE CASCADE` dari header.
- Simpan nilai finansial sebagai integer (`bigint`) sesuai pola existing.
- Status transaksi gunakan enum string konstan di entity (default konvensi: `DR`/`CO`).
- Bila relevan, gunakan audit `UpdatedBy/Updater` + soft delete di header.

## D) Repository rules

- Datatable: hitung `recordsTotal` dan `recordsFiltered` terpisah.
- Search pakai `ILIKE`.
- Preload relasi untuk detail view (`Wallet`, `Updater`, `Details`, `Details.<Relasi>`).
- `Create`/`Update` pakai `Omit(...)` untuk relasi.

## E) Handler/UI rules

- Embed master data ke frontend via `json.Marshal` + `<script type="application/json">`.
- Komponen Alpine terpisah untuk form vs detail panel.
- Datatable row click untuk load detail via AJAX.
- Jika ada form mutasi di halaman transaksi, gunakan popup modal/drawer dan submit via AJAX/fetch tanpa reload halaman.
- Untuk form create dokumen transaksi, gunakan 3 tombol aksi:
  - `Batal` = jika belum save: reset + close modal.
    Jika sudah save: ubah label menjadi `Close`, aksi clear form + close modal.
  - `Save` = simpan status `DR`.
  - `Save & Complete` = simpan status `CO`.
- Hindari dropdown status manual di form create (status ditentukan oleh tombol aksi).
- Render HTML wajib kirim `User`, `Favicon`, `Title`, `ActiveMenu`.

## F) Integrasi accounting (jika menyentuh COA)

- Gunakan service mapping COA (lihat pola COA Mapping / Penebusan).
- Jangan hardcode akun jika sudah tersedia mapping.
- Simpan referensi transaksi (`ref_type`, `ref_id`) saat posting jurnal.

## G) Done checklist

- [ ] Build `go build ./...`
- [ ] Test `go test ./...`
- [ ] Datatable + detail AJAX berjalan
- [ ] Kalkulasi subtotal/ppn/total konsisten backend-frontend
- [ ] Preload detail tidak N+1 berlebihan
- [ ] `database/schema.dbml` ikut diperbarui
