-- +goose Up
CREATE TABLE IF NOT EXISTS units_of_work (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'NEW',
    config_json JSONB NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS units_of_work;
