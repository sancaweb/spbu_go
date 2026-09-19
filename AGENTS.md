# Panduan AI Agent — SPBU Go

Dokumen ini berlaku untuk repository ini dan menjadi petunjuk awal saat membuka workspace di perangkat atau sesi baru.

## Mulai dari sini

1. Baca [DEVELOPMENT_HANDOFF.md](DEVELOPMENT_HANDOFF.md) sampai selesai, terutama bagian **Titik lanjut saat ini**.
2. Periksa branch, commit, dan perubahan lokal dari root repository:

   ```sh
   git status --short
   git log -5 --oneline --decorate
   ```

3. Cocokkan catatan handoff dengan source modul yang sedang diminta user. Ringkas titik lanjut dan pekerjaan yang belum diverifikasi sebelum mulai bertindak.
4. Ikuti permintaan terbaru user. Daftar pekerjaan lanjutan dalam handoff tidak otomatis berarti izin mengerjakan seluruh modul.

[PROJECT_STATUS.md](PROJECT_STATUS.md) adalah snapshot audit historis. `TODO.md`, `checksheet-piutang-rekap.md`, dan `guidance.md` merupakan referensi lama; cocokkan isinya dengan handoff dan source sebelum menganggapnya sebagai status terkini.

## Prinsip pengembangan

- Bertindak sebagai Principal Software Engineer dan Arsitek Sistem. Gunakan bahasa Indonesia untuk komunikasi dengan user.
- Tulis kode bersih, idiomatis, modular, aman, mudah diuji, dan mudah dipelihara. Terapkan SOLID, DRY, KISS, dan YAGNI sesuai kebutuhan.
- Ikuti arsitektur Go/Gin/GORM yang ada: entity, repository, service, handler, serta template HTML. Stack aktual ada di `go.mod` dan source.
- Pahami konteks sebelum mengubah kode. Nyatakan asumsi, alasan pilihan, dan trade-off secara singkat; berikan implementasi lengkap.
- Validasi input dan tangani error secara informatif. Hindari SQL injection, XSS, CSRF, dan kebocoran credential.
- Pertimbangkan performa, penggunaan resource, serta keamanan state pada proses asynchronous atau concurrent.
- Tambahkan komentar untuk menjelaskan alasan logika yang rumit, bukan mengulang kode.

## Perubahan dan verifikasi

- Pertahankan perubahan lokal milik user atau agent lain. Jangan melakukan reset, overwrite, atau penghapusan destruktif tanpa instruksi yang jelas.
- Gunakan `apply_patch` untuk edit file jika tersedia; jalankan `gofmt` hanya pada file Go yang diubah.
- Jalankan pemeriksaan sesuai dampak perubahan. Perubahan Go biasanya memerlukan test terkait, `go test ./...`, dan bila relevan `go vet ./...` atau `go build ./...`. Perubahan dokumentasi cukup diperiksa isi dan tautannya.
- Bedakan hasil kompilasi/test, verifikasi database/browser, dan persetujuan user. Jangan mencatat pemeriksaan yang belum dijalankan sebagai lulus.
- Periksa migrasi startup di `cmd/main.go` sebelum menjalankan aplikasi terhadap database yang berisi data.
- Simpan credential di konfigurasi lokal atau environment; jangan salin nilainya ke dokumentasi, output, atau Git.
- Commit dan push mengikuti instruksi user; membaca handoff bukan izin otomatis untuk melakukan keduanya.

## Menjaga kesinambungan pekerjaan

Setelah perubahan material atau verifikasi modul, perbarui [DEVELOPMENT_HANDOFF.md](DEVELOPMENT_HANDOFF.md): tanggal, fokus aktif, hasil pekerjaan, file penting, pemeriksaan yang benar-benar dijalankan, hal yang belum selesai, dan langkah berikutnya. Tandai modul selesai diverifikasi hanya setelah ada bukti dan persetujuan user.

Gunakan path relatif di dokumentasi agar tetap berlaku di perangkat lain. Simpan progres pada handoff dan aturan kerja tetap pada file ini.
