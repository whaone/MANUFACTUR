# PRD — ManufactPro

<aside>
📄

**Product Requirement Document** — Deskripsi naratif tentang siapa user, masalah apa yang dipecahkan, dan fitur apa yang akan dibangun.

</aside>

## 1. Latar Belakang

Pemilik usaha manufaktur kecil-menengah (konveksi, F&B, kerajinan) kesulitan mengelola:

- Bahan baku yang tersebar di banyak gudang/cabang
- Resep produksi (BOM) yang berbeda per varian produk (ukuran, warna, rasa)
- Perhitungan HPP yang akurat saat harga bahan berubah
- Sinkronisasi stok saat koneksi internet tidak stabil

**ManufactPro** hadir sebagai PWA ringan yang bisa jalan di HP cashier/staff produksi maupun di komputer admin, dengan offline-first sebagai prinsip utama.

## 2. User Person

### 👤 Owner / Admin

- Butuh visibilitas stok semua gudang real-time
- Butuh laporan HPP, profit, waste
- Setup master data (product, material, BOM)

### 👷 Staff Produksi

- Eksekusi production order: scan barcode material, input hasil produksi
- Catat reject/waste
- Mobile-first, satu tangan, layar HP

### 📦 Staff Gudang

- Terima barang dari supplier (Goods Receipt)
- Transfer barang antar gudang
- Stock opname

## 3. Fitur Utama (User Stories)

### 3.1 Manajemen Bahan Baku (Material)

- Sebagai admin, saya bisa menambah bahan baku (Kain, Rib, Label, Cat Sablon, Plastik, Beras, dll) dengan satuan (meter, pcs, gram, liter).
- Setiap bahan punya **stok minimum** untuk alert reorder.
- Bahan bisa dikategorikan (mis. "Tekstil", "Aksesoris", "Packaging").

### 3.2 Manajemen Produk + Varian

- Sebagai admin, saya bisa membuat produk (Kaos Polos, Nasi Box Ayam, dll).
- Produk bisa memiliki **varian** (S/M/L/XL, Warna Merah/Biru, Rasa Original/Pedas).
- Setiap varian = 1 SKU dengan barcode sendiri.

### 3.3 Bill of Materials (BOM)

- Sebagai admin, saya bisa menentukan resep produksi:
    - 1 Kaos Polos ukuran **M** butuh: 1.2 m Kain + 0.3 m Rib + 1 pcs Label + 8 gram Cat + 1 pcs Plastik
    - 1 Kaos Polos ukuran **XL** butuh: 1.5 m Kain + 0.4 m Rib + 1 pcs Label + 10 gram Cat + 1 pcs Plastik
- BOM bisa berbeda per varian.
- BOM bisa multi-level (sub-assembly): mis. "Adonan Sablon" sebagai material setengah jadi.

### 3.4 Production Order

- Sebagai staff produksi, saya membuat Production Order: "Produksi 100 Kaos M warna Merah".
- Sistem otomatis menghitung kebutuhan material berdasarkan BOM.
- Sistem cek apakah stok material cukup di gudang yang dipilih.
- Setelah produksi selesai, input: jumlah jadi, jumlah reject, alasan reject.
- Sistem otomatis kurangi stok material & tambah stok produk jadi.

### 3.5 Multi-Gudang

- Setiap stok dicatat **per gudang**.
- Transfer antar gudang sebagai transaksi terpisah (out dari gudang A, in ke gudang B).
- Production Order ter-attach ke 1 gudang sumber.

### 3.6 Supplier & Purchase Order

- Master data supplier (nama, kontak, term pembayaran).
- Buat PO ke supplier → terima barang via Goods Receipt → stok material bertambah dengan harga beli batch baru.

### 3.7 Perhitungan HPP

- Material disimpan dengan **harga per batch** (FIFO atau Weighted Average — configurable).
- HPP produk = ∑ (qty material × harga material saat ini) + overhead opsional.
- Laporan margin: harga jual − HPP.

### 3.8 Barcode & QR

- Setiap material & SKU produk punya barcode.
- Scan via kamera HP (PWA) untuk: input GR, stock opname, eksekusi production, dan pencarian cepat.

### 3.9 Offline-First

- Semua transaksi bisa dilakukan offline.
- Data masuk ke queue lokal (IndexedDB).
- Otomatis sync ke server saat online.
- Konflik diselesaikan dengan **last-write-wins** + log audit.

### 3.10 Laporan & Export

- Dashboard: stok rendah, produksi hari ini, top material consumption.
- Laporan stok per gudang, history movement, HPP per produk.
- Export Excel & PDF.

## 4. Non-Goals (Tidak Termasuk di MVP)

- Akuntansi & jurnal keuangan lengkap
- Integrasi dengan marketplace / POS eksternal
- Forecasting demand berbasis AI
- Multi-currency
- Modul SDM / payroll

## 5. Kriteria Sukses

- Time-to-first-transaction < 5 menit setelah install
- Offline transaction berhasil di-sync 100% saat reconnect
- Akurasi HPP ±1% vs perhitungan manual
- Bisa dipakai di HP Android entry-level (RAM 2GB) tanpa lag