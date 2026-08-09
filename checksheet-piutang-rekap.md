# Checksheet Perbaikan Endpoint /transaction/piutang/rekap

Tanggal: 2026-06-14
Owner: Tim Development

## Tujuan

Memastikan 3 poin perbaikan halaman Rekap Piutang selesai dan terverifikasi.

## Daftar Kontrol

| No  | Task                                            | Kriteria Selesai                                                                                                                | Status     | Catatan Uji                                                                                                    |
| --- | ----------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- | ---------- | -------------------------------------------------------------------------------------------------------------- |
| 1   | Filter bulan tidak terbatas 2026                | Input periode bisa pindah lintas tahun (sebelum/sesudah 2026) dan request API memakai nilai bulan yang dipilih                  | ✅ Selesai | Uji manual: klik tombol « / » pada filter bulan, pastikan data reload per bulan terpilih                       |
| 2   | Tabel pengelompokan dipisah per tanggal         | Setiap tanggal memiliki blok tabel sendiri dengan header tabel masing-masing                                                    | ✅ Selesai | Uji manual: cek card "Pengelompokan Harian Piutang & Pembayaran", tiap tanggal tampil sebagai section terpisah |
| 3   | Grand total dipindah ke atas card pengelompokan | "Grand Total Piutang" dan "Grand Total Pembayaran Piutang" tampil di bagian paling atas area konten card (sebelum tabel harian) | ✅ Selesai | Uji manual: buka card pengelompokan, cek posisi kedua grand total                                              |

## Regression Checklist

| No  | Area                       | Kriteria                                                                                                   | Status    |
| --- | -------------------------- | ---------------------------------------------------------------------------------------------------------- | --------- |
| R1  | Rekap ringkasan DataTables | Tabel ringkasan partner tetap load, search/sort/pagination tetap berjalan                                  | ☐ Pending |
| R2  | Endpoint summary           | `/transaction/piutang/summary?bulan=YYYY-MM` tetap mengembalikan JSON sukses                               | ☐ Pending |
| R3  | Endpoint grouped           | `/transaction/piutang/rekap/grouped?bulan=YYYY-MM` tetap mengembalikan struktur `groups` dan `grand_total` | ☐ Pending |
| R4  | Empty state                | Saat data kosong, pesan "Tidak ada data pada bulan yang dipilih" tetap muncul                              | ☐ Pending |

## Sign-off

- Developer: ********\_\_\_\_********
- QA: ********\_\_\_\_********
- Tanggal verifikasi: ********\_\_\_\_********
