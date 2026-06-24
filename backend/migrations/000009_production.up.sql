CREATE TYPE production_status AS ENUM ('draft', 'in_progress', 'completed', 'cancelled');

CREATE TABLE production_orders (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id        UUID NOT NULL REFERENCES workspaces(id),
    warehouse_id        UUID NOT NULL REFERENCES warehouses(id),
    product_variant_id  UUID NOT NULL REFERENCES product_variants(id),
    qty_planned         NUMERIC NOT NULL,
    status              production_status NOT NULL DEFAULT 'draft',
    planned_at          TIMESTAMPTZ,
    started_at          TIMESTAMPTZ,
    completed_at        TIMESTAMPTZ,
    total_cost          NUMERIC NOT NULL DEFAULT 0,
    created_by          UUID REFERENCES users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE production_outputs (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    production_order_id UUID NOT NULL REFERENCES production_orders(id),
    qty_good            NUMERIC NOT NULL DEFAULT 0,
    qty_reject          NUMERIC NOT NULL DEFAULT 0,
    qty_waste           NUMERIC NOT NULL DEFAULT 0,
    reject_reason       TEXT,
    waste_reason        TEXT,
    recorded_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
