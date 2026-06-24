# Sequence Diagram — ManufactPro

<aside>
🔁

**Sequence Diagrams** untuk skenario kritikal: Production Order, Offline Sync, Purchase Order & Goods Receipt.

</aside>

## 1. Production Order (Online)

```mermaid
sequenceDiagram
	participant U as Staff Produksi
	participant UI as PWA (Next.js)
	participant SA as Server Action
	participant DB as Supabase Postgres

	U->>UI: Pilih Product Variant + qty
	UI->>SA: createProductionOrder()
	SA->>DB: SELECT BOM untuk variant
	SA->>DB: SELECT v_current_stock per material
	SA-->>UI: validasi stok cukup?
	alt stok cukup
		UI->>U: Tampilkan rincian material
		U->>UI: Konfirmasi & jalankan
		UI->>SA: executeProduction()
		SA->>DB: BEGIN TRANSACTION
		SA->>DB: INSERT stock_movements (OUT_PRODUCTION) per material<br>(pakai harga FIFO/WAC dari material_batches)
		SA->>DB: UPDATE material_batches.qty_remaining
		SA->>DB: INSERT production_outputs (good/reject/waste)
		SA->>DB: INSERT stock_movements (IN_PRODUCTION) untuk product variant
		SA->>DB: UPDATE production_orders status='completed', total_cost=HPP
		SA->>DB: COMMIT
		SA-->>UI: success + ringkasan HPP
	else stok kurang
		SA-->>UI: error + daftar material yang kurang
	end
```

## 2. Offline-First: Operasi Saat Tidak Ada Koneksi

```mermaid
sequenceDiagram
	participant U as Staff
	participant UI as PWA
	participant DX as Dexie (IndexedDB)
	participant SW as Service Worker
	participant API as Supabase

	Note over UI,API: Koneksi terputus 📴

	U->>UI: Scan barcode material + input qty masuk
	UI->>DX: SELECT material by barcode (dari cache)
	UI->>DX: INSERT stock_movement (status=pending)
	UI->>DX: INSERT sync_queue { entity, payload, ts }
	UI-->>U: ✅ Tersimpan offline

	Note over UI,API: Koneksi pulih 📶

	SW->>DX: Ambil sync_queue WHERE status=pending
	loop tiap item dalam queue
		SW->>API: POST mutasi via Server Action
		alt sukses
			API-->>SW: 200 + server_id
			SW->>DX: UPDATE sync_queue status=synced
			SW->>DX: UPDATE local row dengan server_id + updated_at
		else conflict
			API-->>SW: 409 + server_version
			SW->>DX: Resolve last-write-wins<br>(simpan ke audit_log)
		else network error
			SW->>DX: UPDATE retries++, exponential backoff
		end
	end
```

## 3. Purchase Order → Goods Receipt → Material Batch

```mermaid
sequenceDiagram
	participant A as Admin
	participant W as Staff Gudang
	participant UI as PWA
	participant SA as Server Action
	participant DB as Supabase

	A->>UI: Buat PO (pilih supplier + materials + qty)
	UI->>SA: createPurchaseOrder()
	SA->>DB: INSERT purchase_orders + po_items (status='draft')
	A->>UI: Kirim PO ke supplier
	UI->>SA: updatePO(status='sent')

	Note over W,DB: Barang datang dari supplier 📦

	W->>UI: Scan barcode PO / cari PO
	W->>UI: Input qty diterima per item (bisa partial)
	UI->>SA: createGoodsReceipt()
	SA->>DB: BEGIN TRANSACTION
	SA->>DB: INSERT goods_receipts
	SA->>DB: INSERT material_batches<br>(qty_received, qty_remaining, unit_cost, received_at)
	SA->>DB: INSERT stock_movements (IN_PURCHASE) per item
	SA->>DB: UPDATE po_items.qty_received
	SA->>DB: UPDATE purchase_orders.status<br>(partial_received | received)
	SA->>DB: COMMIT
	SA-->>UI: success + label print suggestion
	UI->>W: Tawarkan print label barcode material
```

## 4. Transfer Antar Gudang

```mermaid
sequenceDiagram
	participant U as Staff
	participant UI as PWA
	participant SA as Server Action
	participant DB as Supabase

	U->>UI: Pilih gudang asal & tujuan, scan items + qty
	UI->>SA: createTransfer()
	SA->>DB: BEGIN TRANSACTION
	SA->>DB: INSERT stock_movements (TRANSFER_OUT) di gudang asal
	SA->>DB: INSERT stock_movements (TRANSFER_IN) di gudang tujuan<br>(transfer_id sama)
	SA->>DB: COMMIT
	SA-->>UI: success
	Note over UI: Realtime channel mem-broadcast ke<br>device di gudang tujuan untuk konfirmasi terima
```

## 5. Perhitungan HPP (FIFO)

```mermaid
sequenceDiagram
	participant SA as Server Action (executeProduction)
	participant DB as Postgres

	loop tiap material di BOM
		SA->>DB: SELECT material_batches<br>WHERE material_id=? AND warehouse_id=? AND qty_remaining > 0<br>ORDER BY received_at ASC
		Note over SA: Konsumsi qty dari batch tertua dulu
		loop sampai qty terpenuhi
			SA->>SA: take = min(batch.qty_remaining, sisa_butuh)
			SA->>DB: UPDATE batch SET qty_remaining -= take
			SA->>SA: cost += take × batch.unit_cost
		end
	end
	SA->>DB: production_orders.total_cost = sum(cost)
	SA->>DB: stock_movements (IN_PRODUCTION) unit_cost = total_cost / qty_good
```