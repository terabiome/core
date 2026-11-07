-- +goose Up
CREATE TABLE IF NOT EXISTS units_of_work (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    config_json JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    staged_at TIMESTAMP WITH TIME ZONE
);

-- +goose Down
DROP TABLE IF EXISTS units_of_work;
