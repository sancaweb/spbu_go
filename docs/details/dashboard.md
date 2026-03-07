# Modul: Dashboard & Pelaporan

Modul ini berfokus pada agregasi data (Read-Only) untuk memberikan _insight_ kepada pemilik/manajemen SPBU.

## 1. Sub-Menu: Dashboard

- **Tujuan:** Memberikan gambaran sekilas tentang performa SPBU hari ini dan status inventaris.
- **Logika Bisnis (Backend):**
  - **Grafik Penjualan:** Lakukan agregasi (SUM) pendapatan dari tabel `tb_detail_penjualan` berdasarkan rentang waktu (harian/mingguan).
  - **Stok BBM:** Ambil data `stokLiter` terkini dari tabel `tb_bbm`.
- **UI/UX Behavior:**
  - Gunakan _Chart.js_ atau _Recharts_ untuk visualisasi grafik.
  - Tampilkan _Card_ indikator warna merah jika `stokLiter` suatu BBM berada di bawah batas minimum (kritis).

## 2. Sub-Menu: Rekap Piutang

- **Tujuan:** Menampilkan laporan piutang pelanggan dengan format yang mudah dibaca.
- **Logika Bisnis (Backend):**
  - **Tab Summary:** Lakukan `GROUP BY` bulan pada tabel `tb_piutang` untuk mendapatkan total nilai piutang bulanan.
  - **Tab Rincian:** Lakukan query detail ke `tb_detail_piutang` berdasarkan parameter bulan yang dikirim dari frontend.
- **UI/UX Behavior:**
  - Gunakan antarmuka **Tabs** (Tab 1: Summary Bulanan, Tab 2: Rincian Harian).
  - Sediakan _Dropdown_ Filter Bulan dan Tahun yang secara asinkron (AJAX) memperbarui data di tabel tanpa _reload_ halaman.
