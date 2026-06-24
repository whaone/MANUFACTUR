CREATE TABLE material_batches (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    material_id     UUID NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
    warehouse_id    UUID NOT NULL REFERENCES warehouses(id),
    batch_no        TEXT NOT NULL,
    qty_received    NUMERIC NOT NULL,
    qty_remaining   NUMERIC NOT NULL,
    unit_cost       NUMERIC NOT NULL DEFAULT 0,
    supplier_id     UUID,
    received_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expiry_at       TIMESTAMPTZ
);
