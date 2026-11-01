-- +goose Up
CREATE TABLE IF NOT EXISTS dispatches (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    uow_uuid UUID NOT NULL REFERENCES units_of_work(uuid) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'NEW',
    target JSONB DEFAULT '{}'::jsonb, -- target routing, topic, etc, matching attributes, etc
    metadata JSONB DEFAULT '{}'::jsonb, -- NATS topic, status transition, etc
    payload JSONB NOT NULL, -- transformed sequences, ordered by group + other stuffs
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS dispatches;
