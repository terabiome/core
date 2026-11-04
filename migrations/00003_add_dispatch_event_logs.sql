-- +goose Up
CREATE TABLE IF NOT EXISTS dispatch_event_logs (
    id BIGSERIAL PRIMARY KEY,
    dispatch_uuid UUID NOT NULL REFERENCES dispatches(uuid) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb, -- e.g. task pos/name, sequence pos/name, etc
    data JSONB NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    generated_at TIMESTAMP WITHOUT TIME ZONE,
    committed_at TIMESTAMP WITHOUT TIME ZONE
);

-- +goose Down
DROP TABLE IF EXISTS dispatch_event_logs;
