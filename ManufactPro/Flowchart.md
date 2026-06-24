# Flowchart — ManufactPro

<aside>
🔀

**Flowchart** — Alur proses bisnis utama ManufactPro, termasuk pola **Event Sourcing** untuk pengamanan data stok.

</aside>

## 1. Big Picture: Alur Aplikasi

```mermaid
flowchart TD
	Start([User login]) --> Role{Role?}
	Role -->|Admin| Master[Setup Master Data:<br>Material, Product, BOM, Supplier]
	Role -->|Gudang| Procure[Procurement Flow]
	Role -->|Produksi| Production[Production Flow]
	Role -->|Owner| Report[Dashboard & Report]

	Master --> Ready[Master data siap]
	Ready --> Procure
	Ready --> Production

	Procure --> Stock[(Stock Movements<br>Event Log)]
	Production --> Stock
	Stock --> Aggregate[v_current_stock VIEW]
	Aggregate --> Report
	Aggregate --> Alert{Stok < min?}
	Alert -->|Ya| Notify[Notifikasi Reorder]
	Alert -->|Tidak| Idle((idle))
```

## 2. Flowchart: Production Order (Mengilustrasikan Event Sourcing)

```mermaid
flowchart TD
	A([Staff buka Production Order]) --> B[Pilih Product Variant + qty]
	B --> C[Load BOM dari DB]
	C --> D[Hitung kebutuhan tiap material]
	D --> E[Query v_current_stock per material]
	E --> F{Stok cukup?}
	F -->|Tidak| G[Tampilkan kekurangan]
	G --> End1([END: User adjust])
	F -->|Ya| H[User konfirmasi: EXECUTE]

	H --> I[/Generate EVENTS:/]
	I --> I1[Event: OUT_PRODUCTION material A -2.4m]
	I --> I2[Event: OUT_PRODUCTION material B -0.6m]
	I --> I3[Event: OUT_PRODUCTION material C -2 pcs]
	I --> I4[Event: IN_PRODUCTION variant X +2 pcs]

	I1 & I2 & I3 & I4 --> J[(APPEND ke stock_movements<br>dalam 1 transaction)]
	J --> K[Trigger: UPDATE material_batches.qty_remaining<br>FIFO consumption]
	K --> L[Hitung HPP = sum batch_cost]
	L --> M[Snapshot HPP ke production_order.total_cost]
	M --> N([SUCCESS])

	style I fill:#fef3c7
	style I1 fill:#fef3c7
	style I2 fill:#fef3c7
	style I3 fill:#fef3c7
	style I4 fill:#fef3c7
	style J fill:#dcfce7
```

<aside>
💡

**Inti Event Sourcing di sini:** Sistem tidak menulis `material.stock = 100`. Sistem **append** event `{material_id, qty: -2.4, type: OUT_PRODUCTION, ref: production_order_id}`. Stok saat ini = hasil SUM dari semua event. Ini yang mengamankan data.

</aside>

## 3. Flowchart: Goods Receipt (Penerimaan Barang)

```mermaid
flowchart TD
	A([Barang datang]) --> B[Buka PO terkait]
	B --> C[Scan barcode / pilih item]
	C --> D[Input qty diterima per item]
	D --> E{Qty = PO?}
	E -->|Ya| F[Status PO: received]
	E -->|Tidak| G[Status PO: partial_received]
	F & G --> H[/Generate EVENTS:/]
	H --> H1[Create material_batch<br>qty_remaining = qty_diterima<br>unit_cost = harga PO]
	H --> H2[Event: IN_PURCHASE +qty ke gudang]
	H1 & H2 --> I[(APPEND ke log)]
	I --> J[Print barcode label?]
	J -->|Ya| K[Generate PDF label]
	J -->|Tidak| L([END])
	K --> L

	style H fill:#fef3c7
	style H1 fill:#fef3c7
	style H2 fill:#fef3c7
	style I fill:#dcfce7
```

## 4. Flowchart: Offline-First Operation

```mermaid
flowchart TD
	A([User aksi: scan / input]) --> B{Online?}
	B -->|Ya| C[Direct ke Server Action]
	C --> D[(Postgres: append event)]
	D --> Z([Done])

	B -->|Tidak| E[Simpan event ke Dexie/IndexedDB<br>status: pending]
	E --> F[Push ke sync_queue]
	F --> G[UI tampilkan optimistic update]
	G --> H((Tunggu reconnect))

	H -.online detected.-> I[Service Worker:<br>baca sync_queue ASC by timestamp]
	I --> J[POST event ke server]
	J --> K{Response?}
	K -->|200 OK| L[Mark synced + update local id]
	K -->|409 Conflict| M[Resolve: last-write-wins<br>simpan ke audit_log]
	K -->|5xx / network| N[retries++, exponential backoff]
	N --> H
	L --> O{Queue empty?}
	M --> O
	O -->|Tidak| I
	O -->|Ya| Z

	style E fill:#fef3c7
	style F fill:#fef3c7
	style I fill:#dbeafe
```

<aside>
🔒

**Kenapa Event Sourcing aman untuk offline:** Yang dikirim ke server adalah **event** (aksi), bukan **state** (hasil akhir). Walaupun 5 device offline melakukan operasi, semua event masuk ke log dalam urutan timestamp. Tidak ada "overwrite" stok — hanya append.

</aside>

## 5. Flowchart: Transfer Antar Gudang

```mermaid
flowchart TD
	A([Staff buat transfer]) --> B[Pilih gudang asal & tujuan]
	B --> C[Scan items + qty]
	C --> D{Stok asal cukup?}
	D -->|Tidak| E[Error: stok kurang]
	D -->|Ya| F[Generate transfer_id]
	F --> G[/2 Events dalam 1 transaction:/]
	G --> G1[Event TRANSFER_OUT<br>warehouse=asal, qty=-N]
	G --> G2[Event TRANSFER_IN<br>warehouse=tujuan, qty=+N]
	G1 & G2 --> H[(APPEND keduanya)]
	H --> I[Broadcast via Realtime<br>ke device gudang tujuan]
	I --> J([END])

	style G fill:#fef3c7
	style G1 fill:#fef3c7
	style G2 fill:#fef3c7
	style H fill:#dcfce7
```

## 6. Flowchart: Stok Real-time (Read Model)

```mermaid
flowchart LR
	Events[(stock_movements<br>append-only log)] -->|aggregate| View[v_current_stock VIEW]
	Batches[(material_batches<br>qty_remaining)] -->|FIFO lookup| HPP[HPP Calculator]

	View --> UI1[Dashboard Stok]
	View --> UI2[Low Stock Alert]
	View --> UI3[Production Validation]
	HPP --> UI4[Cost Report]

	note1[/Setiap perubahan stok = 1 event→ di-append, tidak pernah di-update/delete\] -.- Events

	style Events fill:#dcfce7
	style View fill:#dbeafe
```

## 7. Ringkasan Pola Event Sourcing yang Dipakai

| Aspek | Implementasi |
| --- | --- |
| **Event Store** | Tabel `stock_movements` (immutable, append-only) |
| **Event Schema** | `{id, item_id, warehouse_id, qty, movement_type, ref, unit_cost, reason, ts, user}` |
| **Aggregate** | `v_current_stock` VIEW (SUM per item per warehouse) |
| **Snapshot** | `production_orders.total_cost` (HPP saat completed) |
| **Compensation** | Movement reverse (mis. `ADJUSTMENT` dengan qty berlawanan) — tidak pernah DELETE event |
| **Replay** | Bisa hitung ulang stok di tanggal X: `SUM(qty) WHERE created_at <= X` |
| **Offline event log** | `sync_queue` di IndexedDB — event yang menunggu sync |
| **Concurrency** | Optimistic + transaction Postgres + `last-write-wins` untuk offline conflict |

## 8. Yang TIDAK Pakai Event Sourcing

- Master data (Material, Product, BOM, Supplier) → CRUD biasa dengan `updated_at`.
- User profile, settings, dll → CRUD biasa.

Alasan: event sourcing untuk data yang **tidak sering berubah** dan **tidak butuh audit ketat** adalah overkill. Disiplin diterapkan hanya di tempat yang penting: **stok & uang**.