-- Idempotency log for offline sync replay.
-- Each client queue operation has a unique op_id; once processed we cache its
-- result so replays (e.g. lost responses) return the same outcome without re-applying.
CREATE TABLE IF NOT EXISTS sync_processed (
    workspace_id  uuid        NOT NULL,
    op_id         text        NOT NULL,
    status        text        NOT NULL,
    server_id     text,
    error         text,
    processed_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (workspace_id, op_id)
);
