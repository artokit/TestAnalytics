-- +goose Up
CREATE TABLE payments (
    id              UUID        PRIMARY KEY,
    user_id         TEXT        NOT NULL,
    subscription_id TEXT        NOT NULL,
    amount          BIGINT      NOT NULL,
    currency        TEXT        NOT NULL DEFAULT 'RUB',
    status          TEXT        NOT NULL CHECK (status IN ('succeeded', 'failed', 'refunded')),
    period_start    TIMESTAMPTZ NOT NULL,
    period_end      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payments_created_at   ON payments (created_at DESC);
CREATE INDEX idx_payments_subscription ON payments (subscription_id);
CREATE INDEX idx_payments_user_id      ON payments (user_id);

-- +goose Down
DROP TABLE IF EXISTS payments;
