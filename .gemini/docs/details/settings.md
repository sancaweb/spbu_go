# Modul: Settings

Modul untuk manajemen keamanan aplikasi.

## 1. Sub-Menu: User Management

- **Tujuan:** Mengatur siapa saja yang bisa masuk ke aplikasi dan apa yang bisa mereka lakukan.
- **Logika Bisnis (Backend):**
  - CRUD tabel `users` untuk akun login (Username, Password Hashing menggunakan bcrypt).
  - Penugasan Role (Admin, Operator, Pengawas) menggunakan tabel `groups` dan `users_groups`.
- **UI/UX Behavior:**
  - Form pembuatan user baru.
  - _Checklist/Checkbox_ untuk mengatur _Permissions_ (Hak Akses) tiap-tiap Role. Jangan tampilkan password di tabel UI demi keamanan.
