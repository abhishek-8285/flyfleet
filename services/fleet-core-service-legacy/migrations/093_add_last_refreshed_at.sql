-- +goose Up
-- +goose StatementBegin
ALTER TABLE vehicle_daily_stats ADD COLUMN last_refreshed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();
ALTER TABLE driver_daily_stats ADD COLUMN last_refreshed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();
ALTER TABLE fleet_hourly_stats ADD COLUMN last_refreshed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE fleet_hourly_stats DROP COLUMN last_refreshed_at;
ALTER TABLE driver_daily_stats DROP COLUMN last_refreshed_at;
ALTER TABLE vehicle_daily_stats DROP COLUMN last_refreshed_at;
-- +goose StatementEnd
