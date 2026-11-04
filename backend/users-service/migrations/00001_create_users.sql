-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id           UUID PRIMARY KEY,
    telegram_id  BIGINT      NOT NULL UNIQUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    utm_link     TEXT        NOT NULL DEFAULT ''
);

-- +goose Down
DROP TABLE IF EXISTS users;
