#!/usr/bin/env bash
# ManufactPro — smoke test per phase
# Usage: ./scripts/test_phases.sh [phase]  (default: all)
# Requires: server running at BASE_URL, jq installed

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
PASS=0; FAIL=0

green(){ echo -e "\033[32m✓ $1\033[0m"; }
red(){ echo -e "\033[31m✗ $1\033[0m"; }

check() {
  local name="$1" expected="$2" actual="$3"
  if echo "$actual" | grep -q "$expected"; then
    green "$name"
    PASS=$((PASS+1))
  else
    red "$name — expected '$expected', got: $actual"
    FAIL=$((FAIL+1))
  fi
}

check_status() {
  local name="$1" expected="$2" actual="$3"
  if [[ "$actual" == "$expected" ]]; then
    green "$name"
    PASS=$((PASS+1))
  else
    red "$name — expected HTTP $expected, got HTTP $actual"
    FAIL=$((FAIL+1))
  fi
}

# ─── Phase 1: Health ─────────────────────────────────────────────────────────
phase1() {
  echo -e "\n=== Phase 1: Health ==="
  resp=$(curl -s "$BASE_URL/health")
  check "GET /health returns ok" "ok" "$resp"
}

# ─── Phase 2: Auth ───────────────────────────────────────────────────────────
phase2() {
  echo -e "\n=== Phase 2: Auth ==="
  EMAIL="test_$(date +%s)@smoke.test"

  # Register
  resp=$(curl -s -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"Smoke Tester\",\"email\":\"$EMAIL\",\"password\":\"rahasia123\",\"workspace_name\":\"Smoke WS\"}")
  check "POST /api/auth/register → access_token" "access_token" "$resp"
  TOKEN=$(echo "$resp" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

  # Login
  resp=$(curl -s -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"rahasia123\"}")
  check "POST /api/auth/login → access_token" "access_token" "$resp"
  TOKEN=$(echo "$resp" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

  # Wrong password
  status=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"salah\"}")
  check_status "POST /api/auth/login wrong password → 401" "401" "$status"

  # Me
  resp=$(curl -s "$BASE_URL/api/auth/me" -H "Authorization: Bearer $TOKEN")
  check "GET /api/auth/me → email" "$EMAIL" "$resp"

  # Me without token
  status=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/auth/me")
  check_status "GET /api/auth/me no token → 401" "401" "$status"

  # Refresh: login response carries refresh_token → exchange for new access_token
  resp=$(curl -s -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"rahasia123\"}")
  REFRESH=$(echo "$resp" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)
  resp=$(curl -s -X POST "$BASE_URL/api/auth/refresh" \
    -H "Content-Type: application/json" \
    -d "{\"refresh_token\":\"$REFRESH\"}")
  check "POST /api/auth/refresh → access_token" "access_token" "$resp"

  # Refresh with garbage token → 401
  status=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/auth/refresh" \
    -H "Content-Type: application/json" -d '{"refresh_token":"not-a-jwt"}')
  check_status "POST /api/auth/refresh bad token → 401" "401" "$status"

  # Refresh without token → 400
  status=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/auth/refresh" \
    -H "Content-Type: application/json" -d '{}')
  check_status "POST /api/auth/refresh missing → 400" "400" "$status"

  export SMOKE_TOKEN="$TOKEN"
}

# ─── Phase 3: Master Data ────────────────────────────────────────────────────
phase3() {
  echo -e "\n=== Phase 3: Master Data ==="
  TOKEN="${SMOKE_TOKEN:-}"
  if [[ -z "$TOKEN" ]]; then
    red "Phase 3 requires TOKEN — run phase2 first"
    return
  fi

  # Materials
  resp=$(curl -s "$BASE_URL/api/materials" -H "Authorization: Bearer $TOKEN")
  check "GET /api/materials → array" "\[" "$resp"

  resp=$(curl -s -X POST "$BASE_URL/api/materials" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d '{"sku":"SMK-001","name":"Smoke Material","unit":"kg","min_stock":5,"is_active":true}')
  check "POST /api/materials → id" "id" "$resp"
  MAT_ID=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

  resp=$(curl -s "$BASE_URL/api/materials/$MAT_ID" -H "Authorization: Bearer $TOKEN")
  check "GET /api/materials/:id → SMK-001" "SMK-001" "$resp"

  resp=$(curl -s -X PUT "$BASE_URL/api/materials/$MAT_ID" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d '{"name":"Smoke Material Updated"}')
  check "PUT /api/materials/:id → Updated" "Updated" "$resp"

  status=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE \
    "$BASE_URL/api/materials/$MAT_ID" -H "Authorization: Bearer $TOKEN")
  check_status "DELETE /api/materials/:id → 204" "204" "$status"

  # Deleted material must NOT appear in list (soft-delete hidden)
  resp=$(curl -s "$BASE_URL/api/materials" -H "Authorization: Bearer $TOKEN")
  if echo "$resp" | grep -q "SMK-001"; then
    red "Deleted material still in list — expected hidden"
    FAIL=$((FAIL+1))
  else
    green "Deleted material hidden from list"
    PASS=$((PASS+1))
  fi

  # Suppliers
  resp=$(curl -s -X POST "$BASE_URL/api/suppliers" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d '{"name":"Smoke Supplier","contact":"PIC","email":"sup@smoke.test","phone":"021","payment_term":"30 hari"}')
  check "POST /api/suppliers → id" "id" "$resp"

  resp=$(curl -s "$BASE_URL/api/suppliers" -H "Authorization: Bearer $TOKEN")
  check "GET /api/suppliers → array" "\[" "$resp"

  # Products
  resp=$(curl -s -X POST "$BASE_URL/api/products" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d '{"sku":"PRD-SMK","name":"Smoke Product","category":"test","unit":"pcs","is_active":true}')
  check "POST /api/products → id" "id" "$resp"
  PROD_ID=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

  # Variants
  resp=$(curl -s -X POST "$BASE_URL/api/products/$PROD_ID/variants" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d '{"sku":"VAR-SMK","name":"Smoke Variant","sell_price":10000,"attributes":{}}')
  check "POST /api/products/:id/variants → id" "id" "$resp"

  # Delete product with variants → cascade (no FK error), gone from list
  status=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE \
    "$BASE_URL/api/products/$PROD_ID" -H "Authorization: Bearer $TOKEN")
  check_status "DELETE /api/products/:id (with variant) → 204" "204" "$status"
  resp=$(curl -s "$BASE_URL/api/products" -H "Authorization: Bearer $TOKEN")
  if echo "$resp" | grep -q "PRD-SMK"; then
    red "Deleted product still in list"
    FAIL=$((FAIL+1))
  else
    green "Product (with variant) deleted via cascade"
    PASS=$((PASS+1))
  fi
}

# ─── Phase 5: Stock ──────────────────────────────────────────────────────────
phase5() {
  echo -e "\n=== Phase 5: Stock ==="
  TOKEN="${SMOKE_TOKEN:-}"
  if [[ -z "$TOKEN" ]]; then
    red "Phase 5 requires TOKEN — run phase2 first"
    return
  fi

  # Create branch + 2 warehouses
  resp=$(curl -s -X POST "$BASE_URL/api/branches" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d '{"name":"Smoke Branch","address":"Jl. Test"}')
  BRANCH_ID=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  check "POST /api/branches → id" "id" "$resp"

  resp=$(curl -s -X POST "$BASE_URL/api/warehouses" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "{\"branch_id\":\"$BRANCH_ID\",\"name\":\"Gudang A\",\"code\":\"GDA\",\"is_default\":true}")
  WH1=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  check "POST /api/warehouses GudangA → id" "id" "$resp"

  resp=$(curl -s -X POST "$BASE_URL/api/warehouses" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "{\"branch_id\":\"$BRANCH_ID\",\"name\":\"Gudang B\",\"code\":\"GDB\",\"is_default\":false}")
  WH2=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  check "POST /api/warehouses GudangB → id" "id" "$resp"

  # Create material for stock ops
  resp=$(curl -s -X POST "$BASE_URL/api/materials" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d '{"sku":"STK-001","name":"Stock Test Material","unit":"kg","min_stock":1,"is_active":true}')
  MAT_ID=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  check "POST /api/materials for stock → id" "id" "$resp"

  # GET /api/stock (empty ok)
  resp=$(curl -s "$BASE_URL/api/stock" -H "Authorization: Bearer $TOKEN")
  check "GET /api/stock → array" "\[" "$resp"

  # GET /api/stock/movements (empty ok)
  resp=$(curl -s "$BASE_URL/api/stock/movements" -H "Authorization: Bearer $TOKEN")
  check "GET /api/stock/movements → array" "\[" "$resp"

  # Adjustment: add 100kg to Gudang A
  resp=$(curl -s -X POST "$BASE_URL/api/stock/adjustment" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "{\"warehouse_id\":\"$WH1\",\"item_type\":\"material\",\"item_id\":\"$MAT_ID\",\"qty\":100,\"reason\":\"Initial stock\"}")
  check "POST /api/stock/adjustment +100 → movement_type" "ADJUSTMENT" "$resp"

  # GET /api/stock → material should appear
  resp=$(curl -s "$BASE_URL/api/stock" -H "Authorization: Bearer $TOKEN")
  check "GET /api/stock after adjustment → STK-001" "STK-001" "$resp"

  # Transfer 40kg from Gudang A to Gudang B
  resp=$(curl -s -X POST "$BASE_URL/api/stock/transfer" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "{\"from_warehouse_id\":\"$WH1\",\"to_warehouse_id\":\"$WH2\",\"item_type\":\"material\",\"item_id\":\"$MAT_ID\",\"qty\":40,\"reason\":\"Transfer test\"}")
  check "POST /api/stock/transfer → ok" "ok" "$resp"

  # Transfer more than available → should fail 400
  status=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/stock/transfer" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "{\"from_warehouse_id\":\"$WH1\",\"to_warehouse_id\":\"$WH2\",\"item_type\":\"material\",\"item_id\":\"$MAT_ID\",\"qty\":9999,\"reason\":\"Should fail\"}")
  check_status "POST /api/stock/transfer insufficient → 400" "400" "$status"

  # Movements should have ADJUSTMENT + TRANSFER_OUT + TRANSFER_IN
  resp=$(curl -s "$BASE_URL/api/stock/movements" -H "Authorization: Bearer $TOKEN")
  check "GET /api/stock/movements → ADJUSTMENT" "ADJUSTMENT" "$resp"
  check "GET /api/stock/movements → TRANSFER_OUT" "TRANSFER_OUT" "$resp"
  check "GET /api/stock/movements → TRANSFER_IN" "TRANSFER_IN" "$resp"
}

# ─── Phase 6: Production ─────────────────────────────────────────────────────
phase6() {
  echo -e "\n=== Phase 6: Production ==="
  TOKEN="${SMOKE_TOKEN:-}"
  if [[ -z "$TOKEN" ]]; then
    red "Phase 6 requires TOKEN — run phase2 first"
    return
  fi

  # Need warehouse + variant for order creation
  resp=$(curl -s -X POST "$BASE_URL/api/branches" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d '{"name":"Prod Branch"}')
  BRANCH_ID=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

  resp=$(curl -s -X POST "$BASE_URL/api/warehouses" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "{\"branch_id\":\"$BRANCH_ID\",\"name\":\"Gudang Prod\",\"code\":\"GPR\",\"is_default\":true}")
  WH_ID=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

  resp=$(curl -s -X POST "$BASE_URL/api/products" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d '{"sku":"PROD-P6","name":"Prod Phase6","category":"test","unit":"pcs","is_active":true}')
  PROD_ID=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

  resp=$(curl -s -X POST "$BASE_URL/api/products/$PROD_ID/variants" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d '{"sku":"VAR-P6","sell_price":50000,"attributes":{},"is_active":true}')
  VAR_ID=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

  # GET /api/production/orders
  resp=$(curl -s "$BASE_URL/api/production/orders" -H "Authorization: Bearer $TOKEN")
  check "GET /api/production/orders → array" "\[" "$resp"

  # Create draft order
  resp=$(curl -s -X POST "$BASE_URL/api/production/orders" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "{\"warehouse_id\":\"$WH_ID\",\"product_variant_id\":\"$VAR_ID\",\"qty_planned\":10}")
  check "POST /api/production/orders → id" "id" "$resp"
  check "POST /api/production/orders → draft" "draft" "$resp"
  ORDER_ID=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

  # Start order (no BOM → FIFO skips material deduction, still transitions to in_progress)
  resp=$(curl -s -X POST "$BASE_URL/api/production/orders/$ORDER_ID/start" \
    -H "Authorization: Bearer $TOKEN")
  check "POST /start → in_progress" "in_progress" "$resp"

  # Record output
  resp=$(curl -s -X POST "$BASE_URL/api/production/orders/$ORDER_ID/output" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d '{"qty_good":9,"qty_reject":1,"qty_waste":0}')
  check "POST /output → completed" "completed" "$resp"

  # Cancel a second order
  resp=$(curl -s -X POST "$BASE_URL/api/production/orders" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "{\"warehouse_id\":\"$WH_ID\",\"product_variant_id\":\"$VAR_ID\",\"qty_planned\":5}")
  ORDER2=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

  resp=$(curl -s -X PATCH "$BASE_URL/api/production/orders/$ORDER2/cancel" \
    -H "Authorization: Bearer $TOKEN")
  check "PATCH /cancel → cancelled" "cancelled" "$resp"

  # Stock movements should contain IN_PRODUCTION for qty_good
  resp=$(curl -s "$BASE_URL/api/stock/movements" -H "Authorization: Bearer $TOKEN")
  check "GET /stock/movements → IN_PRODUCTION" "IN_PRODUCTION" "$resp"

  # Guard: BOM material with NO stock → Start must FAIL (insufficient stock), not silently proceed
  resp=$(curl -s -X POST "$BASE_URL/api/materials" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d '{"sku":"MAT-NOSTOCK","name":"No Stock Mat","unit":"kg","min_stock":0,"is_active":true}')
  NOSTOCK_MAT=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

  curl -s -X POST "$BASE_URL/api/variants/$VAR_ID/bom" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "{\"material_id\":\"$NOSTOCK_MAT\",\"qty\":5,\"unit\":\"kg\",\"is_optional\":false}" >/dev/null

  resp=$(curl -s -X POST "$BASE_URL/api/production/orders" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "{\"warehouse_id\":\"$WH_ID\",\"product_variant_id\":\"$VAR_ID\",\"qty_planned\":3}")
  ORDER3=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  resp=$(curl -s -X POST "$BASE_URL/api/production/orders/$ORDER3/start" \
    -H "Authorization: Bearer $TOKEN")
  check "POST /start with no stock → insufficient stock error" "insufficient stock" "$resp"

  # Order must remain draft after failed start (rollback)
  resp=$(curl -s "$BASE_URL/api/production/orders" -H "Authorization: Bearer $TOKEN")
  o3=$(echo "$resp" | tr '}' '\n' | grep "$ORDER3")
  check "Order stays draft after failed start" "draft" "$o3"
}

# ─── Phase 7: Procurement ────────────────────────────────────────────────────
phase7() {
  echo -e "\n=== Phase 7: Procurement ==="
  TOKEN="${SMOKE_TOKEN:-}"
  if [[ -z "$TOKEN" ]]; then red "Phase 7 requires TOKEN"; return; fi

  # Setup: branch + warehouse + supplier + material
  resp=$(curl -s -X POST "$BASE_URL/api/branches" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d '{"name":"Branch P7"}')
  BRANCH_ID=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

  resp=$(curl -s -X POST "$BASE_URL/api/warehouses" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "{\"branch_id\":\"$BRANCH_ID\",\"name\":\"Gudang P7\",\"code\":\"GP7\",\"is_default\":false}")
  WH_ID=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

  resp=$(curl -s -X POST "$BASE_URL/api/suppliers" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d '{"name":"Supplier P7","contact":"PIC","email":"p7@test.com","phone":"021","payment_term":"30"}')
  SUP_ID=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

  resp=$(curl -s -X POST "$BASE_URL/api/materials" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d '{"sku":"MAT-P7","name":"Material P7","unit":"kg","min_stock":1,"is_active":true}')
  MAT_ID=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

  # GET /api/procurement/purchase-orders
  resp=$(curl -s "$BASE_URL/api/procurement/purchase-orders" -H "Authorization: Bearer $TOKEN")
  check "GET /api/procurement/purchase-orders → array" "\[" "$resp"

  # Create PO
  resp=$(curl -s -X POST "$BASE_URL/api/procurement/purchase-orders" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "{\"supplier_id\":\"$SUP_ID\",\"warehouse_id\":\"$WH_ID\",\"items\":[{\"material_id\":\"$MAT_ID\",\"qty_ordered\":50,\"unit_price\":10000}]}")
  check "POST /api/procurement/purchase-orders → id" "id" "$resp"
  check "POST /api/procurement/purchase-orders → draft" "draft" "$resp"
  PO_ID=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  ITEM_ID=$(echo "$resp" | grep -o '"id":"[^"]*"' | sed -n '2p' | cut -d'"' -f4)

  # GET detail
  resp=$(curl -s "$BASE_URL/api/procurement/purchase-orders/$PO_ID" -H "Authorization: Bearer $TOKEN")
  check "GET /api/procurement/purchase-orders/:id → items" "items" "$resp"

  # Send
  resp=$(curl -s -X PATCH "$BASE_URL/api/procurement/purchase-orders/$PO_ID/send" \
    -H "Authorization: Bearer $TOKEN")
  check "PATCH /send → sent" "sent" "$resp"

  # Receive (full)
  resp=$(curl -s -X POST "$BASE_URL/api/procurement/purchase-orders/$PO_ID/receive" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "{\"note\":\"Goods arrived\",\"items\":[{\"po_item_id\":\"$ITEM_ID\",\"qty_received\":50,\"batch_no\":\"BATCH-P7-001\"}]}")
  check "POST /receive → received" "received" "$resp"

  # Stock movement IN_PURCHASE should exist
  resp=$(curl -s "$BASE_URL/api/stock/movements" -H "Authorization: Bearer $TOKEN")
  check "GET /stock/movements → IN_PURCHASE" "IN_PURCHASE" "$resp"

  # Stock should show material
  resp=$(curl -s "$BASE_URL/api/stock" -H "Authorization: Bearer $TOKEN")
  check "GET /api/stock → MAT-P7" "MAT-P7" "$resp"

  # Cancel a new draft PO
  resp=$(curl -s -X POST "$BASE_URL/api/procurement/purchase-orders" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "{\"supplier_id\":\"$SUP_ID\",\"warehouse_id\":\"$WH_ID\",\"items\":[{\"material_id\":\"$MAT_ID\",\"qty_ordered\":10,\"unit_price\":5000}]}")
  PO2=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  resp=$(curl -s -X PATCH "$BASE_URL/api/procurement/purchase-orders/$PO2/cancel" \
    -H "Authorization: Bearer $TOKEN")
  check "PATCH /cancel → cancelled" "cancelled" "$resp"

  # Guard: over-receive must be rejected (order 20, receive 25)
  resp=$(curl -s -X POST "$BASE_URL/api/procurement/purchase-orders" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "{\"supplier_id\":\"$SUP_ID\",\"warehouse_id\":\"$WH_ID\",\"items\":[{\"material_id\":\"$MAT_ID\",\"qty_ordered\":20,\"unit_price\":10000}]}")
  PO3=$(echo "$resp" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  ITEM3=$(echo "$resp" | grep -o '"id":"[^"]*"' | sed -n '2p' | cut -d'"' -f4)
  curl -s -X PATCH "$BASE_URL/api/procurement/purchase-orders/$PO3/send" -H "Authorization: Bearer $TOKEN" >/dev/null
  resp=$(curl -s -X POST "$BASE_URL/api/procurement/purchase-orders/$PO3/receive" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "{\"items\":[{\"po_item_id\":\"$ITEM3\",\"qty_received\":25}]}")
  check "POST /receive over-receive → rejected" "over-receive" "$resp"

  # PO3 must stay 'sent' (rollback), nothing received
  resp=$(curl -s "$BASE_URL/api/procurement/purchase-orders/$PO3" -H "Authorization: Bearer $TOKEN")
  check "PO stays sent after rejected over-receive" "sent" "$resp"
}

# ─── Phase 8: Reports ────────────────────────────────────────────────────────
phase8() {
  echo -e "\n=== Phase 8: Reports ==="
  EMAIL="p8_$(date +%s)@smoke.test"
  resp=$(curl -s -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"P8 User\",\"email\":\"$EMAIL\",\"password\":\"pass1234\",\"workspace_name\":\"WS-P8\"}")
  TOKEN=$(echo "$resp" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

  # Dashboard — even with no data, should return struct
  resp=$(curl -s "$BASE_URL/api/reports/dashboard" -H "Authorization: Bearer $TOKEN")
  check "GET /api/reports/dashboard → completed_orders" "completed_orders" "$resp"
  check "GET /api/reports/dashboard → stock_value" "stock_value" "$resp"
  check "GET /api/reports/dashboard → material_cost_month" "material_cost_month" "$resp"

  # HPP margin — empty array OK
  resp=$(curl -s "$BASE_URL/api/reports/hpp-margin" -H "Authorization: Bearer $TOKEN")
  check "GET /api/reports/hpp-margin → array" "\[" "$resp"

  # Production trend — empty array OK
  resp=$(curl -s "$BASE_URL/api/reports/production-trend" -H "Authorization: Bearer $TOKEN")
  check "GET /api/reports/production-trend → array" "\[" "$resp"

  # Unauthenticated → 401
  status=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/reports/dashboard")
  check_status "GET /api/reports/dashboard unauthenticated → 401" "401" "$status"
}

# ─── Phase 9: Sync (offline queue replay) ────────────────────────────────────
phase9() {
  echo -e "\n=== Phase 9: Sync ==="
  EMAIL="p9_$(date +%s)@smoke.test"
  resp=$(curl -s -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"P9 User\",\"email\":\"$EMAIL\",\"password\":\"pass1234\",\"workspace_name\":\"WS-P9\"}")
  TOKEN=$(echo "$resp" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

  # Empty items → 400
  status=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/sync" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d '{"items":[]}')
  check_status "POST /api/sync empty → 400" "400" "$status"

  # Batch: create material (ok) + create supplier (ok) + bad entity (fail)
  resp=$(curl -s -X POST "$BASE_URL/api/sync" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d '{"items":[
      {"id":"q1","operation":"create","entity":"material","client_timestamp":1,"payload":{"sku":"SYN-001","name":"Sync Material","unit":"kg","min_stock":5,"is_active":true}},
      {"id":"q2","operation":"create","entity":"supplier","client_timestamp":2,"payload":{"name":"Sync Supplier","contact":"PIC","email":"s@s.com","phone":"021","payment_term":"30"}},
      {"id":"q3","operation":"create","entity":"unknown","client_timestamp":3,"payload":{}}
    ]}')
  check "POST /api/sync → q1 synced" "\"id\":\"q1\",\"status\":\"synced\"" "$resp"
  check "POST /api/sync → q2 synced" "\"id\":\"q2\",\"status\":\"synced\"" "$resp"
  check "POST /api/sync → q3 failed" "\"id\":\"q3\",\"status\":\"failed\"" "$resp"
  check "POST /api/sync → synced count 2" "\"synced\":2" "$resp"
  check "POST /api/sync → failed count 1" "\"failed\":1" "$resp"

  # Created material persisted
  resp2=$(curl -s "$BASE_URL/api/materials" -H "Authorization: Bearer $TOKEN")
  check "Synced material persisted → SYN-001" "SYN-001" "$resp2"

  # Update via sync using server_id from q1
  MAT_SID=$(echo "$resp" | grep -o '"id":"q1","status":"synced","server_id":"[^"]*"' | grep -o 'server_id":"[^"]*"' | cut -d'"' -f3)
  resp=$(curl -s -X POST "$BASE_URL/api/sync" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "{\"items\":[{\"id\":\"q4\",\"operation\":\"update\",\"entity\":\"material\",\"client_timestamp\":4,\"payload\":{\"id\":\"$MAT_SID\",\"name\":\"Sync Material Updated\"}}]}")
  check "POST /api/sync update → q4 synced" "\"id\":\"q4\",\"status\":\"synced\"" "$resp"
  resp2=$(curl -s "$BASE_URL/api/materials" -H "Authorization: Bearer $TOKEN")
  check "Updated name persisted" "Sync Material Updated" "$resp2"

  # Delete via sync
  resp=$(curl -s -X POST "$BASE_URL/api/sync" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "{\"items\":[{\"id\":\"q5\",\"operation\":\"delete\",\"entity\":\"material\",\"client_timestamp\":5,\"payload\":{\"id\":\"$MAT_SID\"}}]}")
  check "POST /api/sync delete → q5 synced" "\"id\":\"q5\",\"status\":\"synced\"" "$resp"

  # Product entity via sync
  resp=$(curl -s -X POST "$BASE_URL/api/sync" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d '{"items":[{"id":"qp1","operation":"create","entity":"product","client_timestamp":6,"payload":{"name":"Sync Product","category":"test"}}]}')
  check "POST /api/sync product create → synced" "\"id\":\"qp1\",\"status\":\"synced\"" "$resp"
  resp2=$(curl -s "$BASE_URL/api/products" -H "Authorization: Bearer $TOKEN")
  check "Synced product persisted" "Sync Product" "$resp2"

  # Idempotency: replay a brand-new create op TWICE → second is deduped, no duplicate row
  resp=$(curl -s -X POST "$BASE_URL/api/sync" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d '{"items":[{"id":"qdup","operation":"create","entity":"material","client_timestamp":7,"payload":{"sku":"SYN-DUP","name":"Dup Material","unit":"kg","min_stock":1,"is_active":true}}]}')
  check "POST /api/sync qdup first → synced" "\"id\":\"qdup\",\"status\":\"synced\"" "$resp"
  # replay identical batch (same op id) — should still be synced (cached), not error
  curl -s -X POST "$BASE_URL/api/sync" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d '{"items":[{"id":"qdup","operation":"create","entity":"material","client_timestamp":7,"payload":{"sku":"SYN-DUP","name":"Dup Material","unit":"kg","min_stock":1,"is_active":true}}]}' >/dev/null
  # count SYN-DUP occurrences in materials list → must be exactly 1
  dupcount=$(curl -s "$BASE_URL/api/materials" -H "Authorization: Bearer $TOKEN" | grep -o "SYN-DUP" | wc -l | tr -d ' ')
  if [[ "$dupcount" == "1" ]]; then
    green "Idempotent replay → no duplicate (SYN-DUP count=1)"
    PASS=$((PASS+1))
  else
    red "Idempotent replay failed — SYN-DUP count=$dupcount (expected 1)"
    FAIL=$((FAIL+1))
  fi

  # Unauthenticated → 401
  status=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/sync" \
    -H "Content-Type: application/json" -d '{"items":[]}')
  check_status "POST /api/sync unauthenticated → 401" "401" "$status"
}

# ─── Run ──────────────────────────────────────────────────────────────────────
PHASE="${1:-all}"
case "$PHASE" in
  1) phase1 ;;
  2) phase2 ;;
  3) phase1; phase2; phase3 ;;
  5) phase1; phase2; phase5 ;;
  6) phase1; phase2; phase6 ;;
  7) phase1; phase2; phase7 ;;
  8) phase1; phase2; phase8 ;;
  9) phase1; phase2; phase9 ;;
  all) phase1; phase2; phase3; phase5; phase6; phase7; phase8; phase9 ;;
  *) echo "Usage: $0 [1|2|3|5|6|7|8|9|all]"; exit 1 ;;
esac

echo -e "\n─────────────────────────────"
echo "Results: ${PASS} passed, ${FAIL} failed"
[[ $FAIL -eq 0 ]] && exit 0 || exit 1
