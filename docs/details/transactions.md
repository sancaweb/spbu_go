# Modul: Transactions

Modul ini memiliki logika paling kompleks karena melibatkan perubahan uang dan stok. WAJIB menggunakan Database Transaction (`DB.Begin()`, `Commit()`, `Rollback()`) di Golang.

## 1. Generate Reward

- **Tujuan:** Menghitung otomatis bonus karyawan berdasarkan penjualan BBM.
- **Logika Bisnis:** Ambil data penjualan, kalikan dengan `rewardPersen` di tabel BBM/Jabatan, simpan ke `kary_reward_bbm` dan `kary_reward_jabatan`.

## 2. Penebusan & Stok Masuk

- **Tujuan:** Mencatat pembelian BBM dari Pertamina hingga fisik tiba.
- **Logika Bisnis & UI/UX:**
  - **Penebusan BBM:** Mencatat PO/Pembayaran ke Pertamina (`tb_detail_penebusan`).
  - **Stock DO:** Pantau status Delivery Order.
  - **Kedatangan BBM:** Saat data diinput (`tb_kedatangan_bbm`), sistem **WAJIB MENAMBAH** `stokLiter` di `tb_bbm`.

## 3. Penjualan

- **Tujuan:** Mencatat omzet harian SPBU.
- **Logika Bisnis & UI/UX:**
  - **Data Penjualan:** Menginput `totalisatorAkhir` nozzle. Sistem menghitung liter terjual dan **WAJIB MENGURANGI** `stokLiter` di `tb_bbm`.
  - **Penyusutan:** Mencatat selisih takaran/penguapan BBM (mengurangi stok tanpa menghasilkan rupiah).

## 4. Piutang (Account Receivables)

- **Tujuan:** Mengelola hutang pelanggan B2B.
- **Logika Bisnis & UI/UX:**
  - **Data Piutang (Header):** Menampilkan rekapitulasi total piutang per _Partner_.
  - **Rincian Piutang:** Form untuk menginput bon piutang supir (`tb_detail_piutang`).
  - **Data Invoice:** Fitur untuk meng-generate tagihan (Invoice) berdasarkan kumpulan rincian piutang yang belum ditagih.
  - **Pembayaran:** Form pelunasan invoice (`tb_detail_pembayaran`), otomatis mengubah status tagihan menjadi lunas.

## 5. Kasbon Karyawan

- **Tujuan:** Manajemen pinjaman uang karyawan.
- **Logika Bisnis:** Fitur input pengajuan Kasbon (`kary_kasbon`), Pembayaran cicilan (`kary_pembayaran_kasbon`), dan Cetak Laporan historis.

## 6. Cash Management

- **Tujuan:** Mencatat arus kas manual di luar operasional BBM.
- **Logika Bisnis:** **Cash In** & **Cash Out**. Semua input di sini harus masuk ke tabel jurnal `keu_transaction` berdasarkan `idDompet` dan `idCoa` yang dipilih kasir.
