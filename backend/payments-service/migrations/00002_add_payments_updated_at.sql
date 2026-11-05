-- +goose Up
ALTER TABLE payments ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- +goose Down
ALTER TABLE payments DROP COLUMN updated_at;
