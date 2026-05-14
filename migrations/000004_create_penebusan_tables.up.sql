-- ============================================================
-- Migration: 000004 — Penebusan BBM Tables
-- ============================================================
-- Tabel ini mencatat transaksi pembelian/penebusan BBM ke Pertamina.
--
-- ALUR BISNIS:
--   1. [draft]             → Buat order penebusan (pilih jenis BBM + volume)
--   2. [paid]              → SPBU transfer uang ke Pertamina via bank
--                            JURNAL: Dr Uang Muka Pertamina / Cr Bank
--                            JURNAL: Dr Biaya Admin Bank / Cr Bank  (jika adm_bank > 0)
--   3. [partial_delivered] → Sebagian DO sudah diterima (via trx_stok_do)
--   4. [delivered]         → Semua DO sudah diterima
--   5. [cancelled]         → Penebusan dibatalkan (sebelum paid)
--
-- JURNAL KEDATANGAN BBM (di modul terpisah trx_kedatangan_bbm):
--   Dr Persediaan BBM xxx / Cr Uang Muka Pertamina
-- ============================================================

-- ------------------------------------------------------------
-- 1. trx_penebusan — Header penebusan
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS trx_penebusan (
    id              BIGSERIAL       PRIMARY KEY,

    -- Nomor Dokumen
    no_penebusan    VARCHAR(25)     UNIQUE NOT NULL,   -- Internal: PNB/2026/04/0001
    no_so           VARCHAR(25)     NULL,              -- Nomor SO dari Pertamina (diisi setelah konfirmasi)

    -- Tanggal
    tgl_penebusan   DATE            NOT NULL,          -- Tanggal order dibuat
    tgl_bayar       DATE            NULL,              -- Tanggal transfer ke Pertamina (saat status → paid)

    -- Pembayaran
    wallet_id       INT             NULL
                        REFERENCES wallets(id)
                        ON DELETE SET NULL,            -- Rekening/dompet yang digunakan transfer
    adm_bank        BIGINT          NOT NULL DEFAULT 0, -- Biaya admin bank (Rp × 1; unit utuh)

    -- Status alur
    -- draft | paid | partial_delivered | delivered | cancelled
    status          VARCHAR(20)     NOT NULL DEFAULT 'draft',

    -- Catatan bebas
    catatan         TEXT            NULL,

    -- Kalkulasi total (semua dalam Rp, disimpan sebagai integer)
    subtotal        BIGINT          NOT NULL DEFAULT 0, -- Σ (harga_dasar × jml_liter) sebelum PPN
    total_ppn       BIGINT          NOT NULL DEFAULT 0, -- Σ PPN semua item
    total_bayar     BIGINT          NOT NULL DEFAULT 0, -- subtotal + total_ppn + adm_bank

    -- Audit
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by      INT             NULL
                        REFERENCES users(id)
                        ON DELETE SET NULL,
    deleted_at      TIMESTAMP       NULL
);

COMMENT ON TABLE  trx_penebusan                IS 'Header penebusan BBM ke Pertamina';
COMMENT ON COLUMN trx_penebusan.no_penebusan   IS 'Nomor dokumen internal, format: PNB/YYYY/MM/NNNN';
COMMENT ON COLUMN trx_penebusan.no_so          IS 'Nomor Sales Order dari Pertamina, nullable (bisa diisi setelah konfirmasi Pertamina)';
COMMENT ON COLUMN trx_penebusan.adm_bank       IS 'Biaya administrasi bank saat transfer, dibebankan ke akun Biaya Admin Bank';
COMMENT ON COLUMN trx_penebusan.subtotal       IS 'Total harga dasar tanpa PPN: Σ(harga_dasar × jml_liter)';
COMMENT ON COLUMN trx_penebusan.total_ppn      IS 'Total PPN dari semua item detail';
COMMENT ON COLUMN trx_penebusan.total_bayar    IS 'Grand total yang ditransfer ke Pertamina: subtotal + total_ppn + adm_bank';

CREATE INDEX IF NOT EXISTS idx_trx_penebusan_status      ON trx_penebusan(status);
CREATE INDEX IF NOT EXISTS idx_trx_penebusan_tgl         ON trx_penebusan(tgl_penebusan);
CREATE INDEX IF NOT EXISTS idx_trx_penebusan_deleted_at  ON trx_penebusan(deleted_at);

-- ------------------------------------------------------------
-- 2. trx_penebusan_detail — Item per jenis BBM
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS trx_penebusan_detail (
    id              BIGSERIAL       PRIMARY KEY,
    penebusan_id    BIGINT          NOT NULL
                        REFERENCES trx_penebusan(id)
                        ON DELETE CASCADE,

    -- Jenis BBM
    bbm_id          INT             NOT NULL
                        REFERENCES bbm(id)
                        ON DELETE RESTRICT,   -- tidak boleh hapus BBM yang masih ada di penebusan

    -- Volume
    -- Disimpan sebagai integer × 10^stock_decimal_places (konsisten dengan stok BBM master)
    jml_liter       BIGINT          NOT NULL,

    -- Harga — snapshot saat transaksi (tidak berubah meski harga master berubah)
    harga_dasar     BIGINT          NOT NULL,           -- Harga beli dari Pertamina per liter
    harga_jual      BIGINT          NOT NULL,           -- Harga jual SPBU per liter (dari master BBM saat ini)
    margin          BIGINT          NOT NULL DEFAULT 0, -- Harga jual - harga dasar per liter

    -- PPN
    ppn_persen      DECIMAL(5,2)    NOT NULL DEFAULT 0, -- % PPN (contoh: 11.00 untuk 11%)
    ppn_rp          BIGINT          NOT NULL DEFAULT 0, -- Nominal PPN = subtotal × (ppn_persen/100)

    -- Kalkulasi baris
    subtotal        BIGINT          NOT NULL DEFAULT 0, -- harga_dasar × jml_liter (sebelum PPN)
    total           BIGINT          NOT NULL DEFAULT 0, -- subtotal + ppn_rp

    -- Audit (no soft delete — ikut cascade dari header)
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE  trx_penebusan_detail              IS 'Detail penebusan per jenis BBM';
COMMENT ON COLUMN trx_penebusan_detail.jml_liter    IS 'Jumlah liter × 10^stock_decimal_places (integer, konsisten dengan stok BBM)';
COMMENT ON COLUMN trx_penebusan_detail.harga_dasar  IS 'Harga beli dari Pertamina per liter, snapshot saat transaksi (Rp integer)';
COMMENT ON COLUMN trx_penebusan_detail.harga_jual   IS 'Harga jual SPBU per liter, snapshot dari master BBM saat penebusan dibuat';
COMMENT ON COLUMN trx_penebusan_detail.margin       IS 'Selisih harga_jual - harga_dasar per liter, snapshot';
COMMENT ON COLUMN trx_penebusan_detail.ppn_persen   IS 'Persentase PPN yang berlaku, misal 11.00 = 11%';
COMMENT ON COLUMN trx_penebusan_detail.ppn_rp       IS 'Nominal PPN = subtotal × ppn_persen / 100';
COMMENT ON COLUMN trx_penebusan_detail.subtotal     IS 'harga_dasar × jml_liter, sebelum PPN';
COMMENT ON COLUMN trx_penebusan_detail.total        IS 'subtotal + ppn_rp, nilai yang masuk ke jurnal per item BBM';

CREATE INDEX IF NOT EXISTS idx_trx_pen_detail_penebusan ON trx_penebusan_detail(penebusan_id);
CREATE INDEX IF NOT EXISTS idx_trx_pen_detail_bbm       ON trx_penebusan_detail(bbm_id);

-- ============================================================
-- CATATAN JURNAL (ditangani di aplikasi via journal_entries):
--
-- Saat status → paid:
--   Dr  Uang Muka Pertamina (1131)     = subtotal + total_ppn
--   Cr  Bank/Kas (wallet_id)           = subtotal + total_ppn
--   --- jika adm_bank > 0:
--   Dr  Biaya Admin Bank (5110)        = adm_bank
--   Cr  Bank/Kas (wallet_id)           = adm_bank
--
-- Saat kedatangan BBM (di modul trx_kedatangan_bbm):
--   Dr  Persediaan BBM 112X (per BBM)  = subtotal detail masing-masing BBM
--   Cr  Uang Muka Pertamina (1131)     = total yang diterima
-- ============================================================
