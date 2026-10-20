-- +goose Up
CREATE TABLE IF NOT EXISTS dispatched_tasks (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sequence_uuid UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT 
        fk_dispatched_tasks_dispatched_sequence
        FOREIGN KEY (sequence_uuid)
        REFERENCES dispatched_sequences(uuid)
        ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_dispatched_tasks_sequence_uuid
ON dispatched_tasks(sequence_uuid);

-- +goose Down
DROP INDEX IF EXISTS idx_dispatched_tasks_sequence_uuid;
DROP TABLE IF EXISTS dispatched_tasks;
