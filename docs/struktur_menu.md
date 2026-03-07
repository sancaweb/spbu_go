# Struktur Menu & Navigasi Aplikasi SPBU

Dokumen ini menjelaskan hierarki menu aplikasi dan gambaran rute API yang akan digunakan.

## 1. Dashboard

Kelompok menu untuk mengakses report-report.
_(Untuk detail logika dan UI/UX modul ini, silakan baca `docs/details/dashboard.md`)_

- **Dashboard** (Halaman grafik-grafik report penjualan BBM dan juga kondisi stock BBM)
- **Rekap Piutang** (Halaman rekap piutang yang memiliki dua Tab. Tab untuk summary dalam satu bulan dan ada fitur filter bulan. Juga tab untuk rincian piutang setiap harinya dalam satu bulan yang dipilih.)

## 2. Master Data

Kelompok menu untuk mengatur data induk aplikasi
_(Untuk detail logika dan UI/UX modul ini, silakan baca `docs/details/master_data.md`)_

- **Employee Management**
  - Data Karyawan (Management Data Karyawan)
  - Jabatan (Management Data Jabatan yang ada diperusahaan)
  - Tunjangan (Management Data tunjangan yang memungkinkan untuk dijadikan sebagai pendapatan tambahan setiap karyawan)
  - Potongan (Management data potongan yang memungkinkan untuk dijadikan sebagai pengurangan pendapatan setiap karyawan)

- **Partner** (Management data partner)

- **Data BBM**
  - Data BBM
  - Tiang & Nozle (management tiang, nozle dan juga tempat assign jenis BBM)

- **Finance Accounting**
  - COA Management
  - COA Settings
  - Wallet

## 3. Transactions

_(Untuk detail logika dan UI/UX modul ini, silakan baca `docs/details/transactions.md`)_

- **Generate Reward**
- **Penebusan**
  - Penebusan BBM
  - Stock DO
  - Kedatangan BBM
- **Penjualan**
  - Data Penjualan
  - Penyusutan
- **Piutang**
  - Data Piutang (Data Header piutang, berisi list data piutang dari beberapa partner yang nilai ammount nya merupakan akumulasi dari rincian piutang)
  - Rincian Piutang (list data detail atau rincian piutang)
  - Data Invoice (List data invoice yang sudah digenerate)
  - Pembayaran (List History pembayaran invoice. Yang juga bisa menambahkan data data pembayaran)

- **Kasbon**
  - Data Kasbon (List data karyawan yang hanya memiliki kasbon, dilengkapi dengan data total pinjaman yang belum selesai)
  - Pembayaran (List data history pembayaran kasbon)
  - Laporan (List data karyawan yang hanya memiliki kasbon dilengkapi dengan history pembayaran yang sudah dilakukan hingga pada bulan yang ditentukan dalam filter.)

- **Cash Management** Management data masuk dan keluar uang yang diinput secara manual
  - Cash In (List history penginputan uang masuk secara manual)
  - Cash Out (List history penginputan uang keluar secara manual)

## 4. Payroll

_(Untuk detail logika dan UI/UX modul ini, silakan baca `docs/details/payroll.md`)_

- **Data Payroll** (List data gaji yang sudah di generate)

## 5 Settings

_(Untuk detail logika dan UI/UX modul ini, silakan baca `docs/details/settings.md`)_

- **User Management** (Management user, role and permissions)
