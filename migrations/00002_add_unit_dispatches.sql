-- +goose Up
CREATE TABLE unit_dispatches IF NOT EXISTS (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    unit_of_work_uuid UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT 
        fk_unit_dispatches_unit_of_work 
        FOREIGN KEY (unit_of_work_uuid)
        REFERENCES units_of_work(uuid)
        ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_unit_dispatches_units_of_work_uuid
ON unit_dispatches(unit_of_work_uuid);

-- +goose Down
DROP INDEX IF EXISTS idx_unit_dispatches_units_of_work_uuid;
DROP TABLE IF EXISTS unit_dispatches;
