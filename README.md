# ManufactPro

[![CI](https://github.com/whaone/MANUFACTUR/actions/workflows/ci.yml/badge.svg)](https://github.com/whaone/MANUFACTUR/actions/workflows/ci.yml)

Aplikasi manajemen manufaktur (PWA) untuk UKM: kelola material, produk + varian, resep (BOM), stok multi-gudang, pembelian (procurement), produksi dengan HPP metode FIFO, dan laporan. Mendukung mode offline (antrian sync).

- **Backend:** Go 1.25 + chi v5 + pgx/v5 + PostgreSQL, JWT auth.
- **Frontend:** Svelte 5 + Vite + TailwindCSS.

---

## 1. Prasyarat

- Go 1.25+
- Node.js 22+ & npm
- PostgreSQL (default dipakai port **5434**)

---

## 2. Setup Database

Pastikan Postgres jalan di port 5434, lalu buat database + jalankan migrasi (folder `backend/migrations`). Konfigurasi via `backend/.env`:

```
DATABASE_URL=postgres://user:pass@localhost:5434/manufactpro?sslmode=disable
JWT_SECRET=ganti-dengan-rahasia-kuat
PORT=8080
```

> `JWT_SECRET` wajib stabil. Kalau berubah, semua token lama langsung invalid (user harus login ulang).

---

## 3. Jalankan Backend

```bash
cd backend
go run ./cmd/api/main.go
# → ManufactPro API listening on :8080
```

Cek sehat: `curl http://localhost:8080/health` → `{"status":"ok"}`

---

## 4. Jalankan Frontend

```bash
cd ManufactPro/Frontend/ManufactPro-UI
npm install      # sekali saja
npm run dev
# → http://localhost:5173
```

`VITE_API_URL` default `http://localhost:8080`. Override lewat env bila perlu.

---

## 5. Isi Data Demo (simulasi alur penuh)

Script ini menjalankan seluruh alur dari nol: register workspace → gudang → material → supplier → produk + varian S/M/L → BOM → PO → terima barang (stok masuk) → produksi → output.

```bash
cd backend
bash scripts/seed_demo.sh
```

Di akhir akan tercetak kredensial login demo, contoh:

```
Login demo:  demo_1782222348@manufactpro.test / demo1234
Dashboard:   {"completed_orders":1,"total_qty_produced":19,"stock_value":2210000,...}
```

Login pakai kredensial itu di `http://localhost:5173` untuk melihat datanya. (Tiap run membuat workspace baru.)

---

## 6. Alur Pakai Aplikasi (manual)

Urutan penting — produksi butuh stok, stok butuh pembelian, pembelian butuh material & resep.

1. **Login / Register** — buat workspace.
2. **Gudang** (menu Gudang) — buat minimal 1 cabang + 1 gudang.
3. **Materials** — daftarkan bahan baku. Unit harus dari pilihan (meter, pcs, gram, liter, kg, lusin).
4. **Products** — buat produk, lalu tambah **Varian** (mis. ukuran S/M/L). Bisa pakai **Generate** (matrix) untuk banyak varian sekaligus.
5. **BOM** — untuk tiap varian, isi resep: material apa + qty per unit produk. **Wajib** sebelum produksi.
6. **Supplier** — daftarkan pemasok.
7. **Procurement** — buat Purchase Order → **Kirim** → **Terima Barang**. Terima barang = stok masuk + batch + harga (dasar HPP).
8. **Production**:
   - **Buat Order** (pilih varian + gudang + qty + tanggal).
   - **Mulai Produksi** → sistem potong material FIFO, hitung HPP. Kalau stok/resep kurang, muncul peringatan jelas (lengkapi dulu).
   - **Catat Output** → barang jadi masuk stok, order selesai.
9. **Inventory** — pantau stok per gudang, nilai stok, transfer antar gudang.
10. **Reports** — dashboard KPI, HPP vs harga jual, margin, tren produksi.

---

## 7. Catatan Penting

- **Produksi butuh stok**: kalau material kosong atau BOM belum diisi, tombol "Mulai Produksi" akan menolak dengan peringatan. Tambah stok lewat Procurement → Terima Barang dulu.
- **Hapus material = soft delete**: material disembunyikan dari daftar tapi riwayatnya tetap utuh (untuk integritas batch/movement lama).
- **Hapus produk** menghapus varian + BOM-nya sekaligus (cascade).
- **Stock movement append-only**: stok tidak pernah di-edit langsung; semua perubahan lewat movement (purchase / produksi / transfer / adjustment).
- **Offline / Sync**: perubahan saat offline diantrikan, otomatis dikirim ke server saat koneksi kembali.
- **Sesi**: access token 24 jam, auto-refresh pakai refresh token (7 hari). Kalau sesi benar-benar habis, otomatis diarahkan ke halaman login.

---

## 8. Test

Smoke test end-to-end seluruh fase (butuh backend jalan):

```bash
cd backend
bash scripts/test_phases.sh all
```

Type-check frontend:

```bash
cd ManufactPro/Frontend/ManufactPro-UI
npm run check
```
