# AGENTS.md — ManufactPro Backend

Panduan untuk agent/LLM agar tidak halusinasi dan tetap konsisten.

## Aturan Eksekusi
- MAX 300 lines per file per eksekusi/commit
- Satu phase selesai + ditest DULU sebelum lanjut ke phase berikutnya
- Selalu baca file ini + PLAN.md sebelum mulai kerja
- Jangan buat file baru kecuali ada di PLAN.md
- Jangan ubah database schema tanpa tambah migration baru

## Paths Penting
- Backend: `/home/whaone17/Documents/PROJECT/GudangPawon/backend/`
- Frontend API: `ManufactPro/Frontend/ManufactPro-UI/src/lib/api/index.ts`
- Frontend Auth: `ManufactPro/Frontend/ManufactPro-UI/src/lib/stores/auth.ts`
- Data Model: `ManufactPro/Data_Model.md` (ground truth schema)
- Plan: `/home/whaone17/Documents/PROJECT/GudangPawon/PLAN.md`

## Tech Stack Backend
- Language: Go 1.23
- Router: chi v5
- DB Driver: pgx/v5
- Query tool: sqlc (generate dari .sql → Go structs)
- Auth: golang-jwt/jwt v5 + bcrypt
- Migrations: raw SQL files (golang-migrate)
- Config: godotenv + OS env

## Prinsip Kritis
- `stock_movements` = APPEND ONLY. Tidak ada UPDATE/DELETE movements.
- Semua query pakai `WHERE workspace_id = $1` (workspace isolation dari JWT)
- FIFO: `SELECT material_batches ORDER BY received_at ASC FOR UPDATE SKIP LOCKED`
- Transfer stock = 2 movements (TRANSFER_OUT + TRANSFER_IN) dalam 1 transaksi
- Goods Receipt = INSERT material_batch + INSERT stock_movement (IN_PURCHASE) dalam 1 transaksi
- Jangan pernah trust `workspace_id` dari request body — selalu ambil dari JWT claims

## API Base URL
- Dev: `http://localhost:8080`
- Frontend env: `VITE_API_URL=http://localhost:8080`

## Status Phase (update ini setiap phase selesai)
- [x] Phase 0: Anchor files + env setup
- [x] Phase 1: Go project init + migrations + DB connection
- [x] Phase 2: Auth (JWT login/register/me)
- [x] Phase 3: Master data CRUD (materials, products, variants, suppliers, warehouses)
- [x] Phase 4: BOM (bill of materials per variant)
- [x] Phase 5: Stock (movements, current stock view, transfer, adjustment)
- [x] Phase 6: Production (orders, execute FIFO HPP, record output)
- [x] Phase 7: Procurement (PO, goods receipt → batch + movement)
- [x] Phase 8: Reports (dashboard KPIs, stock, HPP/margin)
- [x] Phase 9: Sync endpoint (offline queue) — material + supplier create/update/delete
- [x] Phase 10: Frontend integration (semua route ke HTTP; sync flush on online)

## Go Package Layout (tiap package = internal/<name>/)
```
handler.go  — HTTP handlers, decode request, call service, encode response
service.go  — business logic, panggil DB queries, return domain types
```

Tidak ada layer repository terpisah — service langsung pakai pgxpool/sqlc.

## Cara Update Status Phase
Setelah phase selesai dan test pass, edit baris `- [ ]` jadi `- [x]` di file ini.

## Known Issues
- [ ] **Offline sync write-path unwired** — `enqueue()` di `ManufactPro/Frontend/ManufactPro-UI/src/lib/sync/flush.ts` tidak pernah dipanggil. `flushSyncQueue()` menguras antrian, tapi tidak ada mutation yang mengisi `db.syncQueue`. Akibat: edit saat offline hilang (tidak ter-queue). Fix: panggil `enqueue(op, entity, payload)` di mutation API (`src/lib/api/index.ts`) ketika offline, lalu flush saat `online`.
