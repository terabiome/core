-- +goose Up
CREATE INDEX IF NOT EXISTS idx_uow_status ON units_of_work(status);
CREATE INDEX IF NOT EXISTS idx_dispatches_uow_uuid ON dispatches(uow_uuid);
CREATE INDEX IF NOT EXISTS idx_event_logs_dispatch_uuid ON dispatch_event_logs(dispatch_uuid);

-- +goose Down
DROP INDEX IF EXISTS idx_event_logs_dispatch_uuid;
DROP INDEX IF EXISTS idx_dispatches_uow_uuid;
DROP INDEX IF EXISTS idx_uow_status;
