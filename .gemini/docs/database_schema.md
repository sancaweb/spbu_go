# SPBU Database Schema Reference

> **Constraint Codes:** PK = Primary Key, AI = Auto Increment, NN = Not Null, NULL = Nullable, UQ = Unique, IDX = Indexed, FK = Foreign Key
>
> **Status Conventions:** `0` = inactive/pending/belum lunas, `1` = active/completed/lunas
>
> **Audit Columns** (on most transactional tables): `created_date`, `created_by`, `modified_date`, `modified_by`

---

## 1. Authentication & Authorization

### 1.1 `users` — User accounts (Ion Auth)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | int(10) unsigned | PK, AI, NN | |
| ip_address | varchar(45) | NN | |
| username | varchar(100) | NULL | |
| password | varchar(255) | NN | |
| email | varchar(254) | NN, UQ | |
| activation_selector | varchar(255) | NULL, UQ | |
| activation_code | varchar(255) | NULL | |
| forgotten_password_selector | varchar(255) | NULL, UQ | |
| forgotten_password_code | varchar(255) | NULL | |
| forgotten_password_time | int(10) unsigned | NULL | |
| remember_selector | varchar(255) | NULL, UQ | |
| remember_code | varchar(255) | NULL | |
| created_on | int(10) unsigned | NN | |
| last_login | int(10) unsigned | NULL | |
| active | tinyint(3) unsigned | NULL | |
| first_name | varchar(50) | NULL | |
| last_name | varchar(50) | NULL | |
| company | varchar(100) | NULL | |
| phone | varchar(20) | NULL | |
| image | varchar(255) | NN | |

### 1.2 `roles` — Master roles

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | bigint(20) unsigned | PK, AI, NN | |
| name | varchar(125) | NN | e.g. 'admin', 'operator', 'kasir' |
| guard_name | varchar(125) | NN | e.g. 'web' |
| created_at | timestamp | NULL | |
| updated_at | timestamp | NULL | |

**Unique:** `(name, guard_name)`

### 1.3 `permissions` — Master permissions

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | bigint(20) unsigned | PK, AI, NN | |
| name | varchar(125) | NN | e.g. 'penjualan.create', 'kasbon.approve' |
| guard_name | varchar(125) | NN | e.g. 'web' |
| created_at | timestamp | NULL | |
| updated_at | timestamp | NULL | |

**Unique:** `(name, guard_name)`

### 1.4 `model_has_roles` — Pivot: model ↔ roles (polymorphic)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| role_id | bigint(20) unsigned | NN, FK→`roles.id` CASCADE | |
| model_type | varchar(125) | NN | e.g. 'users' |
| model_id | bigint(20) unsigned | NN | e.g. users.id |

**PK:** `(role_id, model_id, model_type)` · **Index:** `(model_id, model_type)`

### 1.5 `model_has_permissions` — Pivot: model ↔ permissions (polymorphic, direct)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| permission_id | bigint(20) unsigned | NN, FK→`permissions.id` CASCADE | |
| model_type | varchar(125) | NN | e.g. 'users' |
| model_id | bigint(20) unsigned | NN | e.g. users.id |

**PK:** `(permission_id, model_id, model_type)` · **Index:** `(model_id, model_type)`

### 1.6 `role_has_permissions` — Pivot: roles ↔ permissions

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| permission_id | bigint(20) unsigned | NN, FK→`permissions.id` CASCADE | |
| role_id | bigint(20) unsigned | NN, FK→`roles.id` CASCADE | |

**PK:** `(permission_id, role_id)`

### 1.7 `login_attempts` — Failed login tracking (Ion Auth)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | int(10) unsigned | PK, AI, NN | |
| ip_address | varchar(45) | NN | |
| login | varchar(100) | NN | |
| time | int(10) unsigned | NULL | |

---

## 2. Karyawan & HR

### 2.1 `kary_data` — Master data karyawan

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_karyawan | int(11) | PK, AI, NN | |
| nik | varchar(20) | NN | |
| nama_karyawan | varchar(100) | NN | |
| id_jabatan | int(11) | NN, FK→`kary_jabatan.id_jabatan` | |
| gaji_pokok | int(11) | NN | Gaji pokok Rp |
| status_nikah | varchar(15) | NN | |
| jumlah_anak | int(11) | NN | |
| tempat_lahir | varchar(50) | NN | |
| tanggal_lahir | date | NN | |
| jenis_kelamin | varchar(1) | NN | L/P |
| agama | varchar(20) | NN | |
| alamat | text | NN | |
| no_tlp | varchar(20) | NN | |
| pendidikan | varchar(20) | NN | |
| tgl_penerimaan | date | NN | |
| tgl_resign | date | NN | |
| image | varchar(255) | NN | |
| status | varchar(1) | NN, default '1' | 1=aktif, 0=resign |

### 2.2 `kary_jabatan` — Master jabatan

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_jabatan | int(11) | PK, AI, NN | |
| kode_jabatan | varchar(10) | NN | |
| nama_jabatan | varchar(50) | NN | |
| reward_persen | varchar(5) | NN | % reward |

### 2.3 `kary_gaji` — Header penggajian bulanan

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_gaji | int(11) | PK, AI, NN | |
| no_slip | varchar(10) | NN | |
| gapok | int(11) | NN | Gaji pokok total |
| tunjangan | int(11) | NN | Total tunjangan |
| potongan | int(11) | NN | Total potongan |
| take_home | int(11) | NN | Take home pay |
| bulan | date | NN | Periode YYYY-MM-01 |
| periode_mulai | date | NN | |
| periode_sampai | date | NN | |
| created_date | datetime | NN, default NOW | |
| created_by | int(11) | NN, FK→`users.id` | |
| modified_date | datetime | NN, default NOW | |
| modified_by | int(11) | NN, FK→`users.id` | |

### 2.4 `kary_gajidetail` — Detail gaji per karyawan

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_rincian_gaji | int(11) | PK, AI, NN | |
| id_gaji | int(11) | NN, FK→`kary_gaji.id_gaji` | |
| no_slip | varchar(10) | NN | |
| id_karyawan | int(11) | NN, FK→`kary_data.id_karyawan` | |
| gapok | int(11) | NN | |
| tunjangan | int(11) | NN | |
| potongan | int(11) | NN | |
| take_home | int(11) | NN | |
| created_date | datetime | NN, default NOW | |
| created_by | int(11) | NN | |
| modified_date | datetime | NN, default NOW | |
| modified_by | int(11) | NN | |

### 2.5 `kary_tunjangan` — Master jenis tunjangan

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_tunjangan | int(11) | PK, AI, NN | |
| kode_tunjangan | varchar(11) | NN | |
| nama_tunjangan | varchar(100) | NN | |

### 2.6 `kary_junc_tunjangan` — Junction: gaji ↔ tunjangan

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_junc_tunjangan | int(11) | PK, AI, NN | |
| id_karyawan | int(11) | NN, FK→`kary_data.id_karyawan` | |
| id_gaji | int(11) | NN, FK→`kary_gaji.id_gaji` | |
| id_tunjangan | int(11) | NN, FK→`kary_tunjangan.id_tunjangan` | |
| value | int(11) | NN | Nominal Rp |

### 2.7 `kary_potongan` — Master jenis potongan

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_potongan | int(11) | PK, AI, NN | |
| kode_potongan | varchar(11) | NN | |
| nama_potongan | varchar(100) | NN | |

### 2.8 `kary_junc_potongan` — Junction: gaji ↔ potongan

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_junc_potongan | int(11) | PK, AI, NN | |
| id_karyawan | int(11) | NN, FK→`kary_data.id_karyawan` | |
| id_gaji | int(11) | NN, FK→`kary_gaji.id_gaji` | |
| id_potongan | int(11) | NN, FK→`kary_potongan.id_potongan` | |
| value | int(11) | NN | Nominal Rp |

### 2.9 `kary_kasbon` — Kasbon/pinjaman karyawan

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_kasbon | bigint(20) | PK, AI, NN | |
| id_karyawan | int(11) | NN, FK→`kary_data.id_karyawan` | |
| rp_kasbon | int(11) | NN | Nominal pinjaman |
| tgl_kasbon | datetime | NN | |
| keterangan | text | NN | |
| sisa_kasbon | int(11) | NN | Sisa belum dibayar |
| status | varchar(2) | NN, default '0' | 0=belum lunas, 1=lunas |
| created_date | datetime | NN, default NOW | |
| created_by | int(11) | NN | |
| modified_date | datetime | NN, default NOW | |
| modified_by | int(11) | NN | |

### 2.10 `backup_kary_kasbon` — Backup tabel kasbon

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_kasbon | bigint(20) | PK, AI, NN | |
| id_karyawan | int(11) | NN | |
| sisa_kasbon | int(11) | NN | |
| status | varchar(1) | NN, default '0' | |
| created_date | datetime | NN, default NOW | |
| created_by | int(11) | NN | |
| modified_date | datetime | NN, default NOW | |
| modified_by | int(11) | NN | |

### 2.11 `kary_pembayaran_kasbon` — Cicilan kasbon

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_pembayaran_kasbon | bigint(20) | PK, AI, NN | |
| kasbon_id | bigint(20) | NN, FK→`kary_kasbon.id_kasbon` | |
| id_karyawan | int(11) | NN, FK→`kary_data.id_karyawan` | |
| rp_pembayaran | int(11) | NN | Nominal cicilan |
| tgl_pembayaran | datetime | NN | |
| created_date | datetime | NN, default NOW | |
| created_by | int(11) | NN | |
| modified_date | datetime | NN, default NOW | |
| modified_by | int(11) | NN | |

### 2.12 `kary_reward_bbm` — Reward per jenis BBM per bulan

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_reward_bbm | int(11) | PK, AI, NN | |
| id_bbm | int(11) | NN, FK→`tb_bbm.id_bbm` | |
| penjualan | varchar(20) | NN | Total liter |
| reward_rp | int(11) | NN | Nominal Rp |
| bulan | date | NN | |
| created_date | datetime | NN, default NOW | |
| created_by | int(11) | NN | |
| modified_date | datetime | NN, default NOW | |
| modified_by | int(11) | NN | |

### 2.13 `kary_reward_jabatan` — Distribusi reward per jabatan

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_reward_jabatan | int(11) | PK, AI, NN | |
| id_jabatan | int(11) | NN, FK→`kary_jabatan.id_jabatan` | |
| jml_karyawan | int(11) | NN | |
| reward_rp_jabatan | int(11) | NN | Total reward jabatan |
| reward_rp_karyawan | int(11) | NN | Per orang |
| bulan | date | NN | |
| created_date | datetime | NN, default NOW | |
| created_by | int(11) | NN | |
| modified_date | datetime | NN, default NOW | |
| modified_by | int(11) | NN | |

---

## 3. Keuangan

### 3.1 `keu_type_coa` — Kategori COA

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_type_coa | int(11) | PK, AI, NN | |
| code | int(11) | NN | |
| nama | varchar(200) | NN | |
| keterangan | varchar(200) | NN | |

### 3.2 `keu_coa` — Chart of Account

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_coa | int(11) | PK, AI, NN | |
| type_coa | int(11) | NN, FK→`keu_type_coa.id_type_coa` | |
| code | int(11) | NN | |
| nama | varchar(200) | NN | |
| keterangan | varchar(200) | NN | |

### 3.3 `keu_dompet` — Dompet: Kas / Bank

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_dompet | int(11) | PK, AI, NN | |
| nama_wallet | varchar(50) | NN | |
| saldo | bigint(20) | NN | Saldo terkini |
| default_wallet | varchar(2) | NN | 1=default |
| keterangan | varchar(250) | NN | |

### 3.4 `keu_transaction` — Jurnal transaksi utama

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_transaction | bigint(20) | PK, AI, NN | |
| id_dompet | int(11) | NULL, FK→`keu_dompet.id_dompet` | |
| id_coa | int(11) | NN, FK→`keu_coa.id_coa` | |
| masuk | bigint(20) | NN | Uang masuk Rp |
| keluar | bigint(20) | NN | Uang keluar Rp |
| keterangan | varchar(250) | NN | |
| waktu | datetime | NN | Waktu transaksi |
| created_date | datetime | NN | |
| created_by | int(11) | NN | |
| modified_date | datetime | NN | |
| modified_by | int(11) | NN | |

### 3.5 `keu_saldo_kas` — Snapshot saldo akhir bulan

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | bigint(20) | PK, AI, NN | |
| bulan | varchar(10) | NN | Format YYYY-MM |
| saldo | bigint(20) | NN | |
| created_date | datetime | NN, default NOW | |
| created_by | int(11) | NN | |
| modified_date | datetime | NN, default NOW | |
| modified_by | int(11) | NN | |

---

## 4. BBM & Stok

### 4.1 `tb_bbm` — Master jenis BBM

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_bbm | int(11) | PK, AI, NN | |
| nama_bbm | varchar(20) | NN | |
| margin | int(11) | NN | Margin per liter Rp |
| harga_jual | int(11) | NN | Harga jual per liter |
| stok_liter | bigint(20) | NN | Stok dalam integer, desimal diatur via `settings.stock_decimal_places` |
| reward_persen | varchar(5) | NN | % reward |
| status | varchar(2) | NN, default '1' | 1=aktif |

> **Catatan:** Stok disimpan sebagai integer. Tampilan desimal diatur oleh setting `stock_decimal_places` di tabel `settings`. Contoh: jika `stock_decimal_places = 2`, maka nilai `123456` ditampilkan sebagai `1234.56` (dibagi 10²).

### 4.2 `tb_tiang` — Master tiang/island dispenser

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_tiang | int(11) | PK, AI, NN | |
| nama_tiang | varchar(15) | NN | |
| slug | varchar(10) | NN | |

### 4.3 `tb_nozle` — Master nozzle per tiang

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_nozle | int(11) | PK, AI, NN | |
| id_tiang | int(11) | NN, FK→`tb_tiang.id_tiang` | |
| ket | varchar(20) | NN | |
| jenis_bbm | int(11) | NN, FK→`tb_bbm.id_bbm` | |
| status | varchar(1) | NN, default '1' | 1=aktif |

### 4.4 `tb_penebusan` — Header penebusan BBM ke Pertamina

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_penebusan | bigint(20) | PK, AI, NN | |
| no_so | varchar(12) | NN | No Sales Order |
| tgl_penebusan | datetime | NN | |
| adm_bank | int(11) | NN | Biaya admin bank |
| harga_netto | int(11) | NN | |
| created_date | datetime | NN | |
| created_by | int(11) | NN | |
| modified_date | datetime | NN | |
| modified_by | int(11) | NN | |
| id_transaction | bigint(20) | NN, FK→`keu_transaction.id_transaction` | |

### 4.5 `tb_detail_penebusan` — Detail penebusan per jenis BBM

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_detail_penebusan | bigint(20) | PK, AI, NN | |
| id_penebusan | bigint(20) | NN, FK→`tb_penebusan.id_penebusan` | |
| jenis_bbm | bigint(20) | NN, FK→`tb_bbm.id_bbm` | |
| jml_liter | int(11) | NN | |
| margin | int(11) | NN | |
| harga_jual | int(11) | NN | |
| harga_beli | int(11) | NN | |
| harga_bruto | int(11) | NN | |
| rp_pajak | int(11) | NN | PPN/PPh |
| total | int(11) | NN | |
| id_trans_penebusan | bigint(20) | NN, FK→`keu_transaction.id_transaction` | |
| id_trans_pajak | bigint(20) | NN, FK→`keu_transaction.id_transaction` | |

### 4.6 `tb_stok_do` — Stok Delivery Order

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_stok | bigint(20) | PK, AI, NN | |
| no_so | varchar(12) | NN | |
| id_penebusan | bigint(20) | NN, FK→`tb_penebusan.id_penebusan` | |
| jenis_bbm | int(11) | NN, FK→`tb_bbm.id_bbm` | |
| jml_liter | int(11) | NN | |
| sisa_liter | varchar(11) | NN | |
| status_penerimaan | varchar(2) | NN, default '0' | 0=pending, 1=diterima |

### 4.7 `tb_kedatangan_bbm` — Log kedatangan fisik BBM

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_kedatangan | bigint(20) | PK, AI, NN | |
| no_so | varchar(12) | NN | |
| no_lo | varchar(20) | NN | No Loading Order |
| tgl_kedatangan | datetime | NN | |
| shift | varchar(2) | NN | |
| jenis_bbm | int(11) | NN, FK→`tb_bbm.id_bbm` | |
| jml_liter | varchar(11) | NN | |
| nama_driver | varchar(50) | NN | |
| no_pol | varchar(10) | NN | |
| created_date | datetime | NN | |
| created_by | int(11) | NN | |
| modified_date | datetime | NN | |
| modified_by | int(11) | NN | |

### 4.8 `tb_penyusutan` — Penyusutan/susut BBM harian

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_penyusutan | bigint(20) | PK, AI, NN | |
| id_penjualan | bigint(20) | NN, FK→`tb_penjualan.id_penjualan` | |
| no_form | varchar(25) | NN | |
| shift | varchar(2) | NN | |
| waktu | datetime | NN | |
| jenis_bbm | int(11) | NN, FK→`tb_bbm.id_bbm` | |
| stok_awal | varchar(11) | NN | |
| stok_akhir_aktual | varchar(11) | NN | Pembacaan tangki |
| stok_akhir_catatan | varchar(11) | NN | Stok hitung |
| created_date | datetime | NN | |
| created_by | int(11) | NN | |
| modified_date | datetime | NN | |
| modified_by | int(11) | NN | |
| edited | int(11) | NN, default 0 | Flag edit |
| id_transaction | bigint(20) | NN, FK→`keu_transaction.id_transaction` | |

---

## 5. Penjualan

### 5.1 `tb_penjualan` — Header penjualan per shift

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_penjualan | bigint(20) | PK, AI, NN | |
| no_form | varchar(25) | NN | No form unik |
| shift | varchar(2) | NN | 1/2/3 |
| waktu_mulai | datetime | NN | |
| waktu_akhir | datetime | NN | |
| total_rp_penjualan_tot | int(11) | NN | Total dari totalisator |
| total_penerimaan | int(11) | NN | Penjualan - Piutang |
| aktual_uang | int(11) | NN | Uang fisik |
| selisih | int(11) | NN | aktual - penerimaan |
| created_date | datetime | NN | |
| created_by | int(11) | NN | |
| modified_date | datetime | NN | |
| modified_by | int(11) | NN | |
| id_transaction | bigint(20) | NN, FK→`keu_transaction.id_transaction` | |

### 5.2 `tb_detail_penjualan` — Detail per nozzle per shift

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_detail_penjualan | bigint(20) | PK, AI, NN | |
| id_penjualan | bigint(20) | NN, FK→`tb_penjualan.id_penjualan` | |
| no_form | varchar(25) | NN | |
| shift | varchar(2) | NN | |
| waktu_mulai | datetime | NN | |
| waktu_akhir | datetime | NN | |
| id_tiang | int(11) | NN, FK→`tb_tiang.id_tiang` | |
| id_nozle | int(11) | NN, FK→`tb_nozle.id_nozle` | |
| jenis_bbm | int(11) | NN, FK→`tb_bbm.id_bbm` | |
| harga_bbm | int(11) | NN | Harga saat transaksi |
| margin | int(11) | NN | |
| totalisator_awal | varchar(30) | NN | Meter awal |
| totalisator_akhir | varchar(30) | NN | Meter akhir |
| jml_liter | varchar(11) | NN | akhir - awal |
| jml_rupiah | int(11) | NN | liter × harga |
| id_transaction | bigint(20) | NN, FK→`keu_transaction.id_transaction` | |

### 5.3 `tb_jenis_tes` — Master jenis tes BBM

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_jenis_tes | int(11) | PK, AI, NN | |
| nama_tes | varchar(50) | NN | |

### 5.4 `tb_pengeluaran_tes` — Pengeluaran BBM untuk tes

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_pengeluaran | bigint(20) | PK, AI, NN | |
| id_penjualan | bigint(20) | NN, FK→`tb_penjualan.id_penjualan` | |
| no_form | varchar(25) | NN | |
| shift | varchar(2) | NN | |
| waktu | datetime | NN | |
| jenis_tes | int(11) | NN, FK→`tb_jenis_tes.id_jenis_tes` | |
| jenis_bbm | int(11) | NN, FK→`tb_bbm.id_bbm` | |
| harga_bbm | int(11) | NN | |
| liter | varchar(11) | NN | |
| jml_rupiah | int(11) | NN | |
| id_transaction | bigint(20) | NN, FK→`keu_transaction.id_transaction` | |

---

## 6. Piutang & Pembayaran

### 6.1 `tb_pelanggan` — Master pelanggan korporat

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_pelanggan | int(11) | PK, AI, NN | |
| nama_pelanggan | varchar(50) | NN | |
| alamat | text | NN | |
| start_date | datetime | NN | Mulai kerjasama |
| end_date | datetime | NN | Akhir kerjasama |
| status | varchar(1) | NN, default '1' | 1=aktif |

### 6.2 `tb_piutang` — Header piutang

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_piutang | bigint(20) | PK, AI, NN | |
| id_penjualan | bigint(20) | NN, FK→`tb_penjualan.id_penjualan` | |
| no_form | varchar(25) | NN | |
| shift | varchar(2) | NN | |
| waktu | datetime | NN | |
| id_pelanggan | int(11) | NN, FK→`tb_pelanggan.id_pelanggan` | |
| tagihan_pelanggan | int(11) | NN | Nominal tagihan Rp |
| status | varchar(2) | NN, default '0' | 0=belum lunas, 1=lunas |
| add_invoice | varchar(2) | NN, default '0' | 0=belum, 1=di-invoice |
| created_date | datetime | NN | |
| created_by | int(11) | NN | |
| modified_date | datetime | NN | |
| modified_by | int(11) | NN | |
| id_transaction | bigint(20) | NN, FK→`keu_transaction.id_transaction` | |

### 6.3 `tb_detail_piutang` — Rincian piutang (nopol, driver, liter)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_detail_piutang | bigint(20) | PK, AI, NN | |
| id_penjualan | bigint(20) | NN, FK→`tb_penjualan.id_penjualan` | |
| no_form | varchar(25) | NN | |
| id_piutang | bigint(20) | NN, FK→`tb_piutang.id_piutang` | |
| id_pelanggan | int(11) | NN, FK→`tb_pelanggan.id_pelanggan` | |
| shift | varchar(2) | NN | |
| waktu | datetime | NN | |
| no_voucher | varchar(50) | NN | |
| no_pol | varchar(10) | NN | |
| nama_driver | varchar(50) | NN | |
| jenis_bbm | int(11) | NN, FK→`tb_bbm.id_bbm` | |
| harga_bbm | int(11) | NN | |
| jml_liter | varchar(11) | NN | |
| jml_tagihan | int(11) | NN | |
| created_date | datetime | NN | |
| created_by | int(11) | NN | |
| modified_date | datetime | NN | |
| modified_by | int(11) | NN | |

### 6.4 `tb_detail_piutang_tmp` — Staging (identik tb_detail_piutang)

Struktur identik dengan `tb_detail_piutang`, digunakan sebagai staging area sebelum data di-commit.

### 6.5 `tb_invoice` — Invoice penagihan

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_invoice | bigint(20) | PK, AI, NN | |
| no_invoice | varchar(50) | NN | |
| waktu_cetak | datetime | NN | |
| id_pelanggan | int(11) | NN, FK→`tb_pelanggan.id_pelanggan` | |
| total_tagihan | int(11) | NN | |
| periode_mulai | date | NN | |
| periode_sampai | date | NN | |
| status_bayar | varchar(2) | NN, default '0' | 0=belum, 1=lunas |
| created_date | datetime | NN | |
| created_by | int(11) | NN | |
| modified_date | datetime | NN | |
| modified_by | int(11) | NN | |

### 6.6 `tb_detail_invoice` — Junction: invoice ↔ piutang

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_detail_invoice | bigint(20) | PK, AI, NN | |
| id_invoice | bigint(20) | NN, FK→`tb_invoice.id_invoice` | |
| id_piutang | bigint(20) | NN, FK→`tb_piutang.id_piutang` | |

### 6.7 `tb_pembayaran` — Header pembayaran piutang

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_pembayaran | bigint(20) | PK, AI, NN | |
| no_invoice | varchar(30) | NN | |
| tgl_pembayaran | datetime | NN | |
| id_pelanggan | int(11) | NN, FK→`tb_pelanggan.id_pelanggan` | |
| rp_pembayaran | int(11) | NN | Nominal dibayar |
| periode_mulai | date | NN | |
| periode_sampai | date | NN | |
| created_date | datetime | NN | |
| created_by | int(11) | NN | |
| modified_date | datetime | NN | |
| modified_by | int(11) | NN | |
| id_transaction | bigint(20) | NN, FK→`keu_transaction.id_transaction` | |

### 6.8 `tb_detail_pembayaran` — Junction: pembayaran ↔ piutang

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_detail_pembayaran | bigint(20) | PK, AI, NN | |
| id_pembayaran | bigint(20) | NN, FK→`tb_pembayaran.id_pembayaran` | |
| id_pelanggan | int(11) | NN, FK→`tb_pelanggan.id_pelanggan` | |
| id_piutang | bigint(20) | NN, FK→`tb_piutang.id_piutang` | |

---

## 7. Sistem

### 7.1 `settings` — Pengaturan aplikasi (key-value)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | int(11) | PK, AI, NN | |
| setting_name | varchar(100) | NN, UQ | Nama setting |
| setting_value | varchar(255) | NN, default '' | Nilai setting |
| created_at | timestamp | default NOW | |
| updated_at | timestamp | default NOW | |
| deleted_at | timestamp | NULL | Soft delete |

### 7.2 `logs` — Audit log aktivitas

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id_log | bigint(20) | PK, AI, NN | |
| id_user | int(11) | NN, FK→`users.id` | |
| date | datetime | NN | |
| ip_address | varchar(15) | NN | |
| device | varchar(100) | NN | |
| platform | varchar(100) | NN | |
| rincian | text | NN | Deskripsi aktivitas |
