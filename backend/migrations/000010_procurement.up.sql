CREATE TYPE po_status AS ENUM ('draft', 'sent', 'partial_received', 'received', 'cancelled');

CREATE TABLE suppliers (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    contact         TEXT,
    email           TEXT,
    phone           TEXT,
    payment_term    TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE purchase_orders (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id    UUID NOT NULL REFERENCES workspaces(id),
    supplier_id     UUID NOT NULL REFERENCES suppliers(id),
    warehouse_id    UUID NOT NULL REFERENCES warehouses(id),
    po_number       TEXT NOT NULL,
    status          po_status NOT NULL DEFAULT 'draft',
    total_amount    NUMERIC NOT NULL DEFAULT 0,
    ordered_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expected_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(workspace_id, po_number)
);

CREATE TABLE po_items (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    po_id           UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    material_id     UUID NOT NULL REFERENCES materials(id),
    qty_ordered     NUMERIC NOT NULL,
    qty_received    NUMERIC NOT NULL DEFAULT 0,
    unit_price      NUMERIC NOT NULL DEFAULT 0
);

CREATE TABLE goods_receipts (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    po_id           UUID NOT NULL REFERENCES purchase_orders(id),
    received_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    received_by     UUID REFERENCES users(id),
    note            TEXT
);
