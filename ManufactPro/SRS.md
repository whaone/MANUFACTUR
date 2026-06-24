# SRS — ManufactPro

<aside>
📐

**System Requirement Specification** — Functional & non-functional requirements yang lebih formal, termasuk constraint teknis.

</aside>

## 1. Functional Requirements (FR)

### FR-1 Autentikasi & Otorisasi

- FR-1.1 User login via email + password (Better Auth).
- FR-1.2 Role-based access: `owner`, `admin`, `production`, `warehouse`, `viewer`.
- FR-1.3 Setiap user terikat pada 1 atau lebih cabang/gudang.

### FR-2 Master Data

- FR-2.1 CRUD Material dengan field: `name`, `sku`, `unit`, `category`, `min_stock`, `barcode`, `image`.
- FR-2.2 CRUD Product dengan dukungan Variant (atribut: size, color, flavor, dll — dinamis).
- FR-2.3 CRUD BOM per variant: list `(material_id, quantity, unit, is_optional)`.
- FR-2.4 CRUD Supplier, Warehouse, Branch.

### FR-3 Inventory

- FR-3.1 Stok dicatat per `(item_id, warehouse_id)`.
- FR-3.2 Setiap perubahan stok = 1 row di `stock_movements` (immutable log).
- FR-3.3 Tipe movement: `IN_PURCHASE`, `IN_PRODUCTION`, `OUT_PRODUCTION`, `OUT_SALES`, `TRANSFER_OUT`, `TRANSFER_IN`, `ADJUSTMENT`, `WASTE`, `REJECT`.
- FR-3.4 Transfer antar gudang = 2 row (OUT + IN) dengan `transfer_id` yang sama.

### FR-4 Production

- FR-4.1 Production Order memiliki status: `draft`, `in_progress`, `completed`, `cancelled`.
- FR-4.2 Saat status → `in_progress`, sistem melakukan reservasi material (soft hold).
- FR-4.3 Saat status → `completed`, sistem commit: kurangi material + tambah produk jadi.
- FR-4.4 Hasil produksi dipecah: `quantity_good`, `quantity_reject`, `quantity_waste` dengan alasan.

### FR-5 Procurement

- FR-5.1 Purchase Order: `draft`, `sent`, `partial_received`, `received`, `cancelled`.
- FR-5.2 Goods Receipt bisa partial (terima sebagian).
- FR-5.3 Setiap GR menciptakan `material_batch` dengan harga beli & tanggal.

### FR-6 HPP

- FR-6.1 Costing method configurable per workspace: `FIFO` atau `Weighted Average`.
- FR-6.2 HPP produk dihitung saat production completed dan disimpan sebagai snapshot.
- FR-6.3 Tampilan HPP breakdown per material di production order.

### FR-7 Barcode

- FR-7.1 Generate barcode (Code128 / QR) otomatis untuk setiap SKU & material.
- FR-7.2 Scan via kamera HP browser (@zxing/browser).
- FR-7.3 Print barcode label (PDF, ukuran sticker standard 32×25mm, 50×25mm).

### FR-8 Offline & Sync

- FR-8.1 Semua master data di-cache di IndexedDB saat login.
- FR-8.2 Mutasi offline masuk ke `sync_queue` lokal dengan timestamp.
- FR-8.3 Auto-retry sync setiap reconnect.
- FR-8.4 Conflict resolution: `last_write_wins` berdasarkan `updated_at` server.
- FR-8.5 Audit log untuk semua mutasi yang di-sync.

### FR-9 Reporting

- FR-9.1 Dashboard real-time: low stock, today's production, top materials.
- FR-9.2 Stock report per gudang, per kategori, per tanggal.
- FR-9.3 Production report dengan filter periode.
- FR-9.4 HPP & margin report per produk.
- FR-9.5 Export Excel (.xlsx) dan PDF.

## 2. Non-Functional Requirements (NFR)

### NFR-1 Performance

- NFR-1.1 First Contentful Paint < 1.5s pada 4G.
- NFR-1.2 Time to Interactive < 3s pada HP entry-level (Android Go).
- NFR-1.3 Barcode scan recognition < 300ms.
- NFR-1.4 Submit production order < 500ms (online), instan (offline).

### NFR-2 Resource Footprint

- NFR-2.1 Bundle JS awal ≤ 200KB gzipped.
- NFR-2.2 RAM usage idle ≤ 80MB di mobile browser.
- NFR-2.3 IndexedDB storage ≤ 50MB untuk 10.000 item + 30 hari movement.

### NFR-3 Usability

- NFR-3.1 Touch target minimum 44×44px.
- NFR-3.2 Responsive breakpoint: 320px, 768px, 1024px, 1440px.
- NFR-3.3 Dark mode support.
- NFR-3.4 Bahasa: Indonesia (default), English (opsional).

### NFR-4 Reliability

- NFR-4.1 Zero data loss saat offline → online.
- NFR-4.2 Soft delete untuk semua entity utama (recovery dalam 30 hari).
- NFR-4.3 Database backup harian

### NFR-5 Security

- NFR-5.1 Row Level Security (RLS) per workspace + role.
- NFR-5.2 HTTPS only.
- NFR-5.3 Audit log untuk perubahan harga & stok.
- NFR-5.4 Password policy: min 8 char, mixed case + number.

### NFR-6 Compatibility

- NFR-6.1 Browser: Chrome ≥ 100, Safari ≥ 15, Edge ≥ 100, Firefox ≥ 100.
- NFR-6.2 OS: Android 8+, iOS 14+, Windows 10+, macOS 11+.
- NFR-6.3 PWA installable di Android & iOS.

## 3. Hardware Integration (Opsional Phase 2)

| Hardware | Cara Integrasi |
| --- | --- |
| Barcode Scanner USB (HID mode) | Bekerja sebagai keyboard input — no extra code needed |
| Kamera HP | @zxing/browser via getUserMedia |
| Thermal Printer (Bluetooth) | Web Bluetooth API (ESC/POS commands) |
| Label Printer | Print PDF via browser print dialog |

## 4. Constraint Teknis

- Semua mutasi via Drizzle + Server Actions (type-safe end-to-end).
- State client di-cache via TanStack Query, dipersist ke IndexedDB via Dexie.