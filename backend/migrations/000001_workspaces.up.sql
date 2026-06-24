CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE costing_method AS ENUM ('fifo', 'weighted_average');

CREATE TABLE workspaces (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            TEXT NOT NULL,
    costing_method  costing_method NOT NULL DEFAULT 'fifo',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
