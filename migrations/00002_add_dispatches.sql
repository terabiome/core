-- +goose Up
-- Each target matching pattern is correspondent to 1 dispatch, which can match multiple targets.
CREATE TABLE IF NOT EXISTS dispatches (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    uow_uuid UUID NOT NULL REFERENCES units_of_work(uuid) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'NEW', -- NEW, DISPATCHED, FAILED (mostly due to dispatch failure, not downstream error)
    target JSONB DEFAULT '{}'::jsonb, -- target routing, topic, etc, matching attributes, etc
    metadata JSONB DEFAULT '{}'::jsonb, -- NATS topic, status transition, etc
    payload JSONB NOT NULL, -- transformed sequences, ordered by group + other stuffs
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS dispatches;
