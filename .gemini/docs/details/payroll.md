# Modul: Payroll

Modul untuk mengkalkulasi dan mencetak slip gaji karyawan.

## 1. Sub-Menu: Data Payroll

- **Tujuan:** Daftar riwayat penggajian dan fitur _Generate_ Gaji bulanan.
- **Logika Bisnis (Backend):**
  - **Generate Proses:** Saat tombol "Generate Gaji" diklik, sistem akan menarik: Gaji Pokok (`kary_data`), Tunjangan aktif, Total Reward bulan tersebut, dan Potongan aktif (termasuk potongan Kasbon otomatis jika ada).
  - Simpan rekapitulasi ke `kary_gaji` (Header) dan detailnya ke `kary_gajidetail`.
- **UI/UX Behavior:**
  - Tampilkan riwayat payroll dalam bentuk tabel.
  - Sediakan tombol Aksi "Cetak Slip Gaji" (menghasilkan file PDF) untuk setiap baris karyawan.
