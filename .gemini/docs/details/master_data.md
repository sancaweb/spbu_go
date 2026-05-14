# Detail Master Data

Detail ini menjelaskan tentang penjelasan dari setiap menu yang ada.

## 1. Menu: Employee Management

- **Tujuan:** Halaman untuk melakukan management data Employee.
- **Alur bisnis (Logic):**
  1. Saat akses halaman ini, halaman akan memunculkan data table list employee yang hanya berstatus aktif. Disertakan juga tombol untuk melihat employee yang sudah tidak aktif.
  2. Didalam list data employee yang tidak aktif, disetiap row datanya disertakan tombol detail dan restore. Saat tombol detail di klik, munculkan popup detail dari employee
  3. Di halaman ini terdapat tombol Create new employee
  4. Disetiap row data yang muncul terdapat tombol, edit dan detail untuk melakukan perubahan data dan juga view detail data employee.

- **UI/UX Behavior:**
  - Penambahan data maupun editing data, menggunakan popup form.
  - Saat proses submit data (penambahan maupun editing data), memunculkan loading proses. loading tidak hilang sebelum proses submit data selesai.
  - Saat proses selesai, munculkan info atau pemberitahuan yang estetik dan menarik.
  - Jangan gunakan _page reload_. Gunakan AJAX/Fetch API agar proses data bisa lebih cepat.

## 2. Menu: Partner

- **Tujuan:** Halaman untuk melakukan management data Partner.
- **Alur bisnis (Logic):**
  1. Saat akses halaman ini, halaman akan memunculkan data table list partner yang hanya berstatus aktif. Disertakan juga tombol untuk melihat partner yang sudah tidak aktif.
  2. Didalam list data partner yang tidak aktif, disetiap row datanya disertakan tombol detail dan restore. Saat tombol detail di klik, munculkan popup detail dari partner
  3. Di halaman ini terdapat tombol Create new partner
  4. Disetiap row data yang muncul terdapat tombol, edit dan detail untuk melakukan perubahan data dan juga view detail data partner.

- **UI/UX Behavior:**
  - Penambahan data maupun editing data, menggunakan popup form.
  - Saat proses submit data (penambahan maupun editing data), memunculkan loading proses. loading tidak hilang sebelum proses submit data selesai.
  - Saat proses selesai, munculkan info atau pemberitahuan yang estetik dan menarik.
  - Jangan gunakan _page reload_. Gunakan AJAX/Fetch API agar proses data bisa lebih cepat.

## 3. Menu: Data BBM

- **Tujuan:** Halaman untuk melakukan management data BBM.
- **Alur bisnis (Logic):**
  1. Saat akses halaman ini, halaman akan memunculkan data table list BBM yang hanya berstatus aktif. Disertakan juga tombol untuk melihat BBM yang sudah tidak aktif.
  2. Didalam list data BBM yang tidak aktif, disetiap row datanya disertakan tombol detail dan restore. Saat tombol detail di klik, munculkan popup detail dari BBM
  3. Di halaman ini terdapat tombol Create new BBM
  4. Disetiap row data yang muncul terdapat tombol, edit dan detail untuk melakukan perubahan data dan juga view detail data BBM.

- **UI/UX Behavior:**
  - Penambahan data maupun editing data, menggunakan popup form.
  - Saat proses submit data (penambahan maupun editing data), memunculkan loading proses. loading tidak hilang sebelum proses submit data selesai.
  - Saat proses selesai, munculkan info atau pemberitahuan yang estetik dan menarik.
  - Jangan gunakan _page reload_. Gunakan AJAX/Fetch API agar proses data bisa lebih cepat.

## 3.2 Menu: Tiang & Nozle

- **Tujuan:** Halaman untuk melakukan management data Tiang dan Nozle juga diperlukan untuk menentukan bbm apa yang ada di dalam nozle tertentu.
- **Alur bisnis (Logic):**
  1. Saat akses halaman ini, halaman akan memunculkan data table list Tiang & Nozle yang hanya berstatus aktif. Disertakan juga tombol untuk melihat Tiang & Nozle yang sudah tidak aktif.
  2. Didalam list data Tiang & Nozle yang tidak aktif, disetiap row datanya disertakan tombol detail dan restore. Saat tombol detail di klik, munculkan popup detail dari Tiang & Nozle
  3. Di halaman ini terdapat tombol Create new Tiang & Nozle
  4. Disetiap row data yang muncul terdapat tombol, edit dan detail untuk melakukan perubahan data dan juga view detail data Tiang & Nozle.

- **UI/UX Behavior:**
  - Penambahan data maupun editing data, menggunakan popup form.
  - Saat proses submit data (penambahan maupun editing data), memunculkan loading proses. loading tidak hilang sebelum proses submit data selesai.
  - Saat proses selesai, munculkan info atau pemberitahuan yang estetik dan menarik.
  - Jangan gunakan _page reload_. Gunakan AJAX/Fetch API agar proses data bisa lebih cepat.

## 4 Menu: Finance Accounting (COA Management)

- **Tujuan:** Halaman untuk melakukan management data COA (Chart of Account).
- **Alur bisnis (Logic):**
  1. Saat akses halaman ini, halaman akan memunculkan data table list COA yang hanya berstatus aktif. Disertakan juga tombol untuk melihat COA yang sudah tidak aktif.
  2. Didalam list data COA yang tidak aktif, disetiap row datanya disertakan tombol detail dan restore. Saat tombol detail di klik, munculkan popup detail dari COA
  3. Di halaman ini terdapat tombol Create new COA
  4. Disetiap row data yang muncul terdapat tombol, edit dan detail untuk melakukan perubahan data dan juga view detail data COA.

- **UI/UX Behavior:**
  - Penambahan data maupun editing data, menggunakan popup form.
  - Saat proses submit data (penambahan maupun editing data), memunculkan loading proses. loading tidak hilang sebelum proses submit data selesai.
  - Saat proses selesai, munculkan info atau pemberitahuan yang estetik dan menarik.
  - Jangan gunakan _page reload_. Gunakan AJAX/Fetch API agar proses data bisa lebih cepat.

## 4.2 Menu: Finance Accounting (COA Setting)

- **Tujuan:** Halaman untuk melakukan pemetaan Account COA ke setiap transaksi yang memiliki efek ke accounting.
- **Alur bisnis (Logic):**
  1. Saat akses halaman ini, halaman akan memunculkan data table list transaksi yang memiliki efek ke accounting dilengkapi dengan Account COA yang berelasi disetiap list transaksinya.
  2. Di halaman ini terdapat tombol Create new Transaction Mapping untuk melakukan mapping transaksi dan dihubungkan ke account COA.
  3. Disetiap row data yang muncul terdapat tombol, edit dan detail untuk melakukan perubahan data dan juga view detail data.

- **UI/UX Behavior:**
  - Penambahan data maupun editing data, menggunakan popup form.
  - Saat proses submit data (penambahan maupun editing data), memunculkan loading proses. loading tidak hilang sebelum proses submit data selesai.
  - Saat proses selesai, munculkan info atau pemberitahuan yang estetik dan menarik.
  - Jangan gunakan _page reload_. Gunakan AJAX/Fetch API agar proses data bisa lebih cepat.

## 4.3 Menu: Wallet

- **Tujuan:** Halaman untuk melakukan management data Wallet.
- **Alur bisnis (Logic):**
  1. Saat akses halaman ini, halaman akan memunculkan data table list Wallet yang hanya berstatus aktif. Disertakan juga tombol untuk melihat Wallet yang sudah tidak aktif.
  2. Didalam list data Wallet yang tidak aktif, disetiap row datanya disertakan tombol detail dan restore. Saat tombol detail di klik, munculkan popup detail dari Wallet
  3. Di halaman ini terdapat tombol Create new Wallet
  4. Disetiap row data yang muncul terdapat tombol, edit dan detail untuk melakukan perubahan data dan juga view detail data Wallet.

- **UI/UX Behavior:**
  - Penambahan data maupun editing data, menggunakan popup form.
  - Saat proses submit data (penambahan maupun editing data), memunculkan loading proses. loading tidak hilang sebelum proses submit data selesai.
  - Saat proses selesai, munculkan info atau pemberitahuan yang estetik dan menarik.
  - Jangan gunakan _page reload_. Gunakan AJAX/Fetch API agar proses data bisa lebih cepat.
