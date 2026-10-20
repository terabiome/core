-- +goose Up
CREATE TABLE IF NOT EXISTS units_of_work (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    config_json JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_units_of_work_status
ON units_of_work(status);

-- +goose Down
DROP INDEX IF EXISTS idx_units_of_work_status;
DROP TABLE IF EXISTS units_of_work;
