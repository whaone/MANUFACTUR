# PLAN.md — ManufactPro Backend Implementation

## Database Schema Ground Truth
Lihat: `ManufactPro/Data_Model.md`

## Go Project Structure
```
backend/
├── cmd/api/main.go
├── internal/
│   ├── auth/handler.go + service.go + middleware.go
│   ├── material/handler.go + service.go
│   ├── product/handler.go + service.go
│   ├── bom/handler.go + service.go
│   ├── stock/handler.go + service.go
│   ├── production/handler.go + service.go
│   ├── procurement/handler.go + service.go
│   ├── reports/handler.go + service.go
│   ├── sync/handler.go + service.go
│   ├── warehouse/handler.go + service.go
│   └── db/
│       ├── db.go
│       ├── queries/*.sql
│       └── generated/
├── migrations/
│   ├── 000001_workspaces.{up,down}.sql
│   ├── 000002_branches_warehouses.{up,down}.sql
│   ├── 000003_users.{up,down}.sql
│   ├── 000004_materials.{up,down}.sql
│   ├── 000005_material_batches.{up,down}.sql
│   ├── 000006_products_variants.{up,down}.sql
│   ├── 000007_bom_items.{up,down}.sql
│   ├── 000008_stock_movements.{up,down}.sql
│   ├── 000009_production.{up,down}.sql
│   ├── 000010_procurement.{up,down}.sql
│   ├── 000011_indexes.{up,down}.sql
│   └── 000012_views.{up,down}.sql
├── pkg/response/json.go
├── sqlc.yaml
├── go.mod
├── .env.example
└── Makefile
```

## API Routes

### Auth (no JWT)
```
POST /api/auth/register
POST /api/auth/login
POST /api/auth/refresh
GET  /api/auth/me
```

### Master Data (JWT required)
```
GET/POST       /api/materials
GET/PUT/DELETE /api/materials/:id
GET/POST       /api/suppliers
GET/PUT/DELETE /api/suppliers/:id
GET/POST       /api/products
GET/PUT/DELETE /api/products/:id
GET/POST       /api/products/:productId/variants
PUT/DELETE     /api/products/:productId/variants/:id
GET            /api/variants
GET/POST       /api/warehouses
GET/POST       /api/branches
```

### BOM
```
GET            /api/variants/:id/bom
POST           /api/variants/:id/bom
PUT            /api/bom/:bomId
DELETE         /api/bom/:bomId
```

### Stock
```
GET            /api/stock
GET            /api/stock/movements
POST           /api/stock/transfer
POST           /api/stock/adjustment
```

### Production
```
GET/POST       /api/production/orders
GET            /api/production/orders/:id
PATCH          /api/production/orders/:id/status
POST           /api/production/orders/:id/execute
POST           /api/production/orders/:id/output
```

### Procurement
```
GET/POST       /api/procurement/purchase-orders
PATCH          /api/procurement/purchase-orders/:id/status
POST           /api/procurement/purchase-orders/:id/goods-receipts
```

### Reports
```
GET            /api/reports/dashboard
GET            /api/reports/stock
GET            /api/reports/production
GET            /api/reports/hpp-margin
```

### Sync
```
POST           /api/sync
```

## JWT Claims Struct
```go
type Claims struct {
    UserID      uuid.UUID `json:"user_id"`
    WorkspaceID uuid.UUID `json:"workspace_id"`
    Role        string    `json:"role"`
    jwt.RegisteredClaims
}
```

## FIFO Execute Production (1 DB transaction)
1. SELECT production_order WHERE id=$1 AND status='in_progress' FOR UPDATE
2. SELECT bom_items WHERE product_variant_id = order.product_variant_id
3. Per bom_item:
   - qty_needed = bom_item.qty × order.qty_planned
   - SELECT material_batches WHERE material_id=? AND warehouse_id=? AND qty_remaining>0
     ORDER BY received_at ASC FOR UPDATE SKIP LOCKED
   - Loop batch: take = MIN(batch.qty_remaining, remaining), UPDATE batch, accumulate cost
   - Jika stok kurang → ROLLBACK dengan error "insufficient stock: <material_name>"
   - INSERT stock_movements (OUT_PRODUCTION, qty=-qty_needed, unit_cost=weighted_avg)
4. INSERT production_outputs (qty_good, qty_reject, qty_waste)
5. INSERT stock_movements (IN_PRODUCTION, qty=+qty_good, item_type='product_variant', unit_cost=HPP/unit)
6. UPDATE production_orders SET status='completed', total_cost=grand_total, completed_at=NOW()

## Goods Receipt (1 DB transaction)
1. INSERT goods_receipts
2. Per item:
   - INSERT material_batches (qty_remaining = qty_received, received_at = NOW())
   - INSERT stock_movements (IN_PURCHASE, qty=+qty_received, unit_cost=unit_cost)
   - UPDATE po_items SET qty_received += qty_received
3. Cek SUM(qty_received) vs SUM(qty_ordered) → update po status

## Test Protocol per Phase

**Phase 1:**
```bash
curl http://localhost:8080/health
# → {"status":"ok"}
```

**Phase 2:**
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@test.com","password":"rahasia123","workspace_name":"Test WS"}'

curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"rahasia123"}'
# → {"access_token":"...","user":{...}}

export TOKEN="<token>"
curl http://localhost:8080/api/auth/me -H "Authorization: Bearer $TOKEN"
```

**Phase 3:**
```bash
curl http://localhost:8080/api/materials -H "Authorization: Bearer $TOKEN"
# → []
curl -X POST http://localhost:8080/api/materials \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"sku":"FAB-001","name":"Kain Cotton","unit":"meter","min_stock":50,"is_active":true}'
# → {"id":"...","sku":"FAB-001",...}
```

**Phase 4:**
```bash
VARIANT_ID="<id>"
curl http://localhost:8080/api/variants/$VARIANT_ID/bom -H "Authorization: Bearer $TOKEN"
curl -X POST http://localhost:8080/api/variants/$VARIANT_ID/bom \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"material_id":"...","qty":1.2,"unit":"meter","is_optional":false}'
```

**Phase 5:**
```bash
curl http://localhost:8080/api/stock -H "Authorization: Bearer $TOKEN"
curl -X POST http://localhost:8080/api/stock/transfer \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"from_warehouse_id":"...","to_warehouse_id":"...","item_type":"material","item_id":"...","qty":10}'
# Verify: 2 movements TRANSFER_OUT + TRANSFER_IN
```

**Phase 6:**
```bash
# Buat order, execute, verify HPP + stock movements
curl -X POST http://localhost:8080/api/production/orders/:id/execute \
  -H "Authorization: Bearer $TOKEN"
```

**Phase 7:**
```bash
curl -X POST http://localhost:8080/api/procurement/purchase-orders/:id/goods-receipts \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"items":[{"po_item_id":"...","qty_received":100,"unit_cost":5000}]}'
# Verify: material_batch + IN_PURCHASE movement
```

**Phase 8:**
```bash
curl http://localhost:8080/api/reports/dashboard -H "Authorization: Bearer $TOKEN"
```

**Phase 9:**
```bash
curl -X POST http://localhost:8080/api/sync \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"items":[{"id":"uuid","operation":"create","entity":"material","payload":{...},"client_timestamp":1234567890}]}'
```

## Aturan 300 Lines
Split file jika > 300 baris:
- handler.go: HTTP decode/encode saja
- service.go: business logic
- queries/*.sql: SQL saja, di-generate oleh sqlc
