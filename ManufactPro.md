# ManufactPro

<aside>
🎯

**ManufactPro** — Aplikasi manajemen produksi & bahan baku berbasis Web/PWA untuk multi-cabang, multi-gudang, dengan dukungan varian produk, BOM, HPP, supplier, dan offline-first.

</aside>

## 📌 Ringkasan Project

| Aspek | Detail |
| --- | --- |
| **Tipe** | Web App + PWA (installable, offline-first) |
| **Target user** | Pemilik UMKM / staff produksi / staff gudang |
| **Skala** | Multi-cabang & multi-gudang |
| **Platform** | Desktop browser + Mobile (responsive, touch-friendly) |

## 🧱 Tech Stack

| Layer | Pilihan |
| --- | --- |
| Frontend | Next.js 15 (App Router) + TypeScript |
| UI | Tailwind CSS + shadcn/ui |
| PWA | @serwist/next |
| Data fetching | TanStack Query |
| Backend | Go (Postgres + Auth + Realtime + Storage) |
| ORM | Drizzle ORM |
| Validasi | Zod |
| Offline DB | Dexie.js (IndexedDB) + sync queue |
| Barcode/QR | @zxing/browser |
| Export | SheetJS (xlsx) + pdfmake |
| Form | React Hook Form + Zod |
| Chart | Recharts / Tremor |

## 📚 Dokumen

Dokumen-dokumen detail akan otomatis menjadi sub-page di bawah halaman ini setelah dibuat.

## 🗺️ Roadmap Pengembangan

- [ ]  **Phase 1 — Foundation** (1-2 minggu): Setup, Auth, Warehouse, Material CRUD
- [ ]  **Phase 2 — Product & BOM** (1-2 minggu): Product + Variant + BOM editor
- [ ]  **Phase 3 — Stock & Movement** (1 minggu): Stock in/out, transfer antar gudang, barcode scan
- [ ]  **Phase 4 — Production** (2 minggu): Production Order, konsumsi material, HPP
- [ ]  **Phase 5 — Procurement** (1-2 minggu): Supplier, PO, Goods Receipt
- [ ]  **Phase 6 — Offline + Sync** (2 minggu): Dexie cache, sync queue, conflict handling
- [ ]  **Phase 7 — Report & Export** (1 minggu): Dashboard, laporan, export Excel/PDF

[Sequence Diagram — ManufactPro](ManufactPro/Sequence_Diagram.md)

[Data Model — ManufactPro](ManufactPro/Data_Model.md)

[SRS — ManufactPro](ManufactPro/SRS.md)

[PRD — ManufactPro](ManufactPro/PRD.md)

[Flowchart — ManufactPro](ManufactPro/Flowchart.md)