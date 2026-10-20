-- +goose Up
CREATE TABLE IF NOT EXISTS dispatched_sequences (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispatch_uuid UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT 
        fk_dispatched_sequences_unit_dispatch
        FOREIGN KEY (dispatch_uuid)
        REFERENCES unit_dispatches(uuid)
        ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_dispatched_sequences_dispatch_uuid
ON dispatched_sequences(dispatch_uuid);

-- +goose Down
DROP INDEX IF EXISTS idx_dispatched_sequences_dispatch_uuid;
DROP TABLE IF EXISTS dispatched_sequences;
