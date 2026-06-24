CREATE TYPE material_unit AS ENUM ('meter', 'pcs', 'gram', 'liter', 'kg', 'lusin');

CREATE TABLE materials (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    sku             TEXT NOT NULL,
    name            TEXT NOT NULL,
    unit            material_unit NOT NULL,
    category        TEXT,
    min_stock       NUMERIC NOT NULL DEFAULT 0,
    barcode         TEXT,
    image_url       TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(workspace_id, sku)
);
