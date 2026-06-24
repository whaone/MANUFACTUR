CREATE TYPE item_type AS ENUM ('material', 'product_variant');

CREATE TYPE movement_type AS ENUM (
    'IN_PURCHASE',
    'IN_PRODUCTION',
    'OUT_PRODUCTION',
    'OUT_SALES',
    'TRANSFER_OUT',
    'TRANSFER_IN',
    'ADJUSTMENT',
    'WASTE',
    'REJECT'
);

CREATE TABLE stock_movements (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id    UUID NOT NULL REFERENCES workspaces(id),
    warehouse_id    UUID NOT NULL REFERENCES warehouses(id),
    item_type       item_type NOT NULL,
    item_id         UUID NOT NULL,
    qty             NUMERIC NOT NULL,
    movement_type   movement_type NOT NULL,
    reference_type  TEXT,
    reference_id    UUID,
    unit_cost       NUMERIC NOT NULL DEFAULT 0,
    reason          TEXT,
    created_by      UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
    -- APPEND ONLY: no UPDATE or DELETE allowed
);
