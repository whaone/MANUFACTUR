CREATE TABLE bom_items (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_variant_id  UUID NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    material_id         UUID NOT NULL REFERENCES materials(id),
    qty                 NUMERIC NOT NULL,
    unit                TEXT NOT NULL,
    is_optional         BOOLEAN NOT NULL DEFAULT FALSE,
    note                TEXT,
    UNIQUE(product_variant_id, material_id)
);
