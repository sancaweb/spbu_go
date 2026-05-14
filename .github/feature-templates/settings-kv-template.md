# Template Cepat — Konfigurasi / Key-Value Settings

Gunakan template ini untuk fitur konfigurasi berbasis tabel `settings`.

## A) Kapan pakai

- Konfigurasi aplikasi runtime (favicon, decimal places, parameter UI/logic kecil).
- Tidak cocok untuk data relasional kompleks.

## B) File yang biasanya diubah

- `internal/repository/setting_repo.go`
- `internal/service/setting_service.go`
- `internal/handler/setting_handler.go`
- `internal/middleware/settings_middleware.go` (jika perlu inject global)
- `templates/settings/index.html`

## C) Pola implementasi

- Read: `FindByKey(key)`
- Write: `Upsert(key, value)`
- Conversion helper di service (contoh `GetInt(key, defaultVal)`)
- Endpoint mutasi tetap `POST`.

## D) Handler rules

- Validasi minimal: `setting_name` wajib.
- Response JSON ringkas (`status`, `message`).
- Mutasi settings dari UI dilakukan via AJAX/fetch tanpa reload halaman.
- Untuk file setting (contoh favicon): validasi ekstensi, simpan di `static/uploads`, lalu simpan URL ke settings.

## E) Middleware/global context

- Jika setting dipakai global di template, inject melalui middleware (contoh `favicon`).
- Pastikan fallback default saat key belum ada.

## F) Done checklist

- [ ] Build `go build ./...`
- [ ] Test `go test ./...`
- [ ] Upsert setting berhasil
- [ ] Nilai setting terbaca di halaman yang membutuhkan
- [ ] Fallback default aman jika setting kosong/tidak ada
