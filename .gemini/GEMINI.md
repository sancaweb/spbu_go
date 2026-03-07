# PROJECT CONTEXT: SPBU Management System (BBM Sales)
This project is a Point of Sales and Management system for Gas Stations (SPBU).
Focus: High performance, concurrency handling, and strict data consistency.

## TECH STACK
- **Language:** Go (Golang) version 1.22+
- **Framework:** Gin Web Framework (github.com/gin-gonic/gin)
- **Database:** PostgreSQL 14+
- **Driver:** pgx (github.com/jackc/pgx/v5) preferred over GORM for performance, UNLESS instructed otherwise.
- **Architecture:** Clean Architecture / Hexagonal (Handler -> Service -> Repository).

## KNOWLEDGE BASE & REFERENCES
Before writing code, ALWAYS check these context files to understand the SPBU system:
1. **App Menu:** Read `docs/struktur_menu.md` to know available features.
2. **Details:** When working on a specific feature, read the corresponding file in the `docs/details/` folder for Business Logic and UI/UX rules.
3. **Database:** Read `docs/database_schema.md` for table structure and relationships.
4. **Data Seeding:** Read `docs/seed_data.md` when asked to create seeders or populate tables with initial data.


## CODING RULES (GOLANG SPECIFIC)
1. **Error Handling:** - Never ignore errors (`_`). Always handle them explicitly.
   - Use custom error types for domain logic (e.g., `ErrInsufficientFuel`).
   - Wrap errors with context: `fmt.Errorf("failed to process transaction: %w", err)`.

2. **Concurrency:**
   - Use Go routines and Channels for heavy tasks (e.g., syncing pump data).
   - ALWAYS implement `context.Context` in every function passing through layers to handle timeouts/cancellation.

3. **Database:**
   - Use Database Transactions (`BEGIN`, `COMMIT`, `ROLLBACK`) for any operation involving money or fuel stock decrement. This is CRITICAL for an SPBU app.

4. **Code Style:**
   - Follow standard `gofmt`.
   - Variable names should be descriptive (e.g., `fuelTankCapacity`, not `cap`).
   - Use dependency injection for Services and Repositories.

## NUMBER FORMAT CONVENTION (WAJIB — Berlaku Global)
Semua angka numerik di seluruh project HARUS menggunakan format Indonesia:

| Jenis | Delimiter | Contoh |
|-------|-----------|--------|
| Ribuan | **titik** (`.`) | `1.234.567` |
| Desimal | **koma** (`,`) | `1.234,56` |

### Rules:
1. **Display (Tabel/Label):** Semua angka yang ditampilkan ke user (harga, stok, quantity, margin, total, dll) HARUS diformat dengan format Indonesia. Gunakan `toLocaleString('id-ID')` di JavaScript.
2. **Input (Form):** Semua input angka HARUS menggunakan `<input type="text">` (BUKAN `type="number"`), dengan live formatting Indonesia saat user mengetik. Contoh: user ketik `1234567` → otomatis tampil `1.234.567`.
3. **Parsing (Submit):** Sebelum mengirim data ke server, hapus titik (`.`) dan ganti koma (`,`) menjadi titik (`.`) untuk mendapat angka raw. Contoh: `1.234,56` → parse ke `1234.56`.
4. **Backend (Go):** Terima angka sebagai `float64` atau `int64`. Semua konversi format dilakukan di frontend, backend menyimpan angka raw tanpa formatting.
5. **Scope:** Berlaku untuk SEMUA field numerik: harga, margin, stok, quantity, total, gaji, potongan, tunjangan, saldo, dll.

### Helper Functions (JavaScript):
- `formatIDR(value)` — Format integer ke display ribuan: `1234567` → `1.234.567`
- `formatStock(rawValue)` — Format integer stok ke desimal berdasarkan `stock_decimal_places`
- `formatInputIDR(val)` — Live format input ribuan (tanpa desimal)
- `formatInputStock(val)` — Live format input stok (dengan desimal koma)
- `formatInputDecimal(val)` — Live format input desimal umum (e.g. persen)
- `parseIDR(val)` — Parse string Indonesia ke raw number: `1.234,56` → `1234.56`

## MODAL/POPUP SCROLL RULE (WAJIB)
Semua modal/popup yang kontennya bisa melebihi tinggi layar HARUS punya scroll. Pada inner content div (bukan backdrop), tambahkan: `max-h-[90vh] overflow-y-auto`. Jangan tambahkan scroll pada list internal secara terpisah — biarkan seluruh form yang scroll.

## BEHAVIORAL INSTRUCTIONS
- When asked to create a feature, first Plan step-by-step in pseudocode.
- If the Database Schema in docs/database_schema.md is missing a column required for a feature, propose the specific `ALTER TABLE` query first.
- Always include Unit Test guidelines for critical financial logic.