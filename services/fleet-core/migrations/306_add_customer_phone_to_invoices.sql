-- +goose Up
-- +goose StatementBegin
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS customer_phone VARCHAR(50);
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS customer_email VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE invoices DROP COLUMN IF EXISTS customer_phone;
ALTER TABLE invoices DROP COLUMN IF EXISTS customer_email;
-- +goose StatementEnd
