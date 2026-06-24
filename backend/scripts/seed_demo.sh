#!/usr/bin/env bash
# ManufactPro — seed data dummy + simulasi alur penuh (dari nol sampai produksi jadi).
# Pakai: bash scripts/seed_demo.sh
# Syarat: backend jalan di BASE_URL, akun demo dibuat otomatis.

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
TS=$(date +%s)
# Fresh workspace tiap run → tidak bentrok unique constraint. Kredensial dicetak di akhir.
EMAIL="${DEMO_EMAIL:-demo_$TS@manufactpro.test}"
PASS="${DEMO_PASS:-demo1234}"

green(){ echo -e "\033[32m✓ $1\033[0m"; }
info(){ echo -e "\033[36m→ $1\033[0m"; }
die(){ echo -e "\033[31m✗ $1\033[0m"; exit 1; }

# Ambil value JSON pertama untuk key tertentu (string). Tolerant (no match → kosong).
jval(){ echo "$1" | grep -o "\"$2\":\"[^\"]*\"" | head -1 | cut -d'"' -f4 || true; }
# Ambil "id" ke-N (1-based). Tolerant.
jid(){ echo "$1" | grep -o '"id":"[^"]*"' | sed -n "${2}p" | cut -d'"' -f4 || true; }

POST(){ curl -s -X POST "$BASE_URL$1" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d "$2"; }
PATCHR(){ curl -s -X PATCH "$BASE_URL$1" -H "Authorization: Bearer $TOKEN"; }

echo "=== ManufactPro — Seed Demo ==="

# 1) Register atau login
info "Auth ($EMAIL)"
reg=$(curl -s -X POST "$BASE_URL/api/auth/register" -H "Content-Type: application/json" \
  -d "{\"name\":\"Demo Owner\",\"email\":\"$EMAIL\",\"password\":\"$PASS\",\"workspace_name\":\"Demo Workshop\"}")
TOKEN=$(jval "$reg" access_token)
[[ -n "$TOKEN" ]] || die "gagal register: $reg"
green "Register workspace demo baru"

# 2) Branch + Warehouse
info "Branch + Gudang"
br=$(POST /api/branches '{"name":"Cabang Pusat","address":"Jl. Mawar 1"}')
BRANCH=$(jid "$br" 1)
wh=$(POST /api/warehouses "{\"branch_id\":\"$BRANCH\",\"name\":\"Gudang Utama\",\"code\":\"GU-$TS\",\"is_default\":true}")
WH=$(jid "$wh" 1)
[[ -n "$WH" ]] || die "gagal buat gudang: $wh"
green "Gudang Utama"

# 3) Materials
info "Materials"
m1=$(POST /api/materials "{\"sku\":\"KAIN-$TS\",\"name\":\"Kain Katun\",\"unit\":\"meter\",\"category\":\"Tekstil\",\"min_stock\":20,\"is_active\":true}")
MAT_KAIN=$(jid "$m1" 1)
m2=$(POST /api/materials "{\"sku\":\"BNG-$TS\",\"name\":\"Benang Jahit\",\"unit\":\"pcs\",\"category\":\"Tekstil\",\"min_stock\":10,\"is_active\":true}")
MAT_BENANG=$(jid "$m2" 1)
[[ -n "$MAT_BENANG" ]] || die "gagal buat material benang: $m2"
m3=$(POST /api/materials "{\"sku\":\"KNC-$TS\",\"name\":\"Kancing\",\"unit\":\"pcs\",\"category\":\"Aksesoris\",\"min_stock\":100,\"is_active\":true}")
MAT_KANCING=$(jid "$m3" 1)
green "3 material: Kain Katun, Benang, Kancing"

# 4) Supplier
sup=$(POST /api/suppliers '{"name":"CV Tekstil Jaya","contact":"Budi","email":"budi@tekstiljaya.id","phone":"021555","payment_term":"30 hari"}')
SUP=$(jid "$sup" 1)
green "Supplier: CV Tekstil Jaya"

# 5) Product + Variants (S, M, L)
info "Product + Variants"
prod=$(POST /api/products "{\"sku\":\"KAOS-$TS\",\"name\":\"Kaos Polos\",\"category\":\"Apparel\",\"unit\":\"pcs\",\"is_active\":true}")
PROD=$(jid "$prod" 1)
declare -A VAR
for size in S M L; do
  v=$(POST "/api/products/$PROD/variants" "{\"sku\":\"KAOS-$TS-$size\",\"name\":\"Kaos $size\",\"sell_price\":75000,\"attributes\":{\"Ukuran\":\"$size\"}}")
  VAR[$size]=$(jid "$v" 1)
done
green "Produk Kaos Polos + varian S/M/L"

# 6) BOM per variant (1.5m kain, 1 roll benang, 3 kancing)
info "BOM (resep) tiap varian"
for size in S M L; do
  vid=${VAR[$size]}
  POST "/api/variants/$vid/bom" "{\"material_id\":\"$MAT_KAIN\",\"qty\":1.5,\"unit\":\"meter\",\"is_optional\":false}" >/dev/null
  POST "/api/variants/$vid/bom" "{\"material_id\":\"$MAT_BENANG\",\"qty\":1,\"unit\":\"pcs\",\"is_optional\":false}" >/dev/null
  POST "/api/variants/$vid/bom" "{\"material_id\":\"$MAT_KANCING\",\"qty\":3,\"unit\":\"pcs\",\"is_optional\":false}" >/dev/null
done
green "BOM terisi (kain + benang + kancing)"

# 7) Purchase Order → Send → Receive (stok masuk + HPP)
info "Procurement: PO → kirim → terima barang"
[[ -n "$SUP" && -n "$WH" && -n "$MAT_KAIN" ]] || die "id kosong sebelum PO (SUP=$SUP WH=$WH MAT=$MAT_KAIN)"
po=$(POST /api/procurement/purchase-orders "{\"supplier_id\":\"$SUP\",\"warehouse_id\":\"$WH\",\"items\":[{\"material_id\":\"$MAT_KAIN\",\"qty_ordered\":100,\"unit_price\":25000},{\"material_id\":\"$MAT_BENANG\",\"qty_ordered\":50,\"unit_price\":8000},{\"material_id\":\"$MAT_KANCING\",\"qty_ordered\":500,\"unit_price\":500}]}")
PO=$(jid "$po" 1)
ITEM_KAIN=$(jid "$po" 2)
ITEM_BENANG=$(jid "$po" 3)
ITEM_KANCING=$(jid "$po" 4)
[[ -n "$PO" && -n "$ITEM_KAIN" && -n "$ITEM_BENANG" && -n "$ITEM_KANCING" ]] || die "PO/item id kosong: $po"
snd=$(PATCHR "/api/procurement/purchase-orders/$PO/send")
echo "$snd" | grep -q '"status":"sent"' || die "send gagal: $snd"
rcv=$(POST "/api/procurement/purchase-orders/$PO/receive" "{\"note\":\"Kedatangan awal\",\"items\":[{\"po_item_id\":\"$ITEM_KAIN\",\"qty_received\":100,\"batch_no\":\"B-KAIN-1\"},{\"po_item_id\":\"$ITEM_BENANG\",\"qty_received\":50,\"batch_no\":\"B-BNG-1\"},{\"po_item_id\":\"$ITEM_KANCING\",\"qty_received\":500,\"batch_no\":\"B-KNC-1\"}]}")
echo "$rcv" | grep -q '"status":"received"' || die "receive gagal: $rcv"
green "Stok masuk: kain 100m, benang 50pcs, kancing 500pcs"

# 8) Production: order → start (FIFO HPP) → output
info "Production: order → mulai → catat output"
ord=$(POST /api/production/orders "{\"warehouse_id\":\"$WH\",\"product_variant_id\":\"${VAR[M]}\",\"qty_planned\":20}")
ORD=$(jid "$ord" 1)
st=$(POST "/api/production/orders/$ORD/start" '')
echo "$st" | grep -q "in_progress" || die "start gagal: $st"
POST "/api/production/orders/$ORD/output" '{"qty_good":19,"qty_reject":1,"qty_waste":0,"reject_reason":"jahitan miring"}' >/dev/null
green "Produksi Kaos M: 20 direncanakan → 19 baik, 1 reject"

# 9) Ringkasan dari Reports
info "Reports dashboard"
dash=$(curl -s "$BASE_URL/api/reports/dashboard" -H "Authorization: Bearer $TOKEN")
echo ""
echo "=== SELESAI ==="
echo "Login demo:  $EMAIL / $PASS"
echo "Dashboard:   $dash"
echo ""
echo "Buka frontend, login pakai kredensial di atas untuk lihat datanya."
