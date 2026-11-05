-- Таблица подписок для ручки GET /subscriptions.
-- users лежат в таблице users (uuid как текст), plans заведены в internal/mock
-- (не выносим в БД), payments не моделируем - их кейсы (failed/retries/refunds)
-- покрывает seeder-service на уровне генерации данных.

-- +goose Up
CREATE TABLE subscriptions (
    id           UUID        PRIMARY KEY,
    user_id      TEXT        NOT NULL,
    plan_id      SMALLINT    NOT NULL,
    period_start TIMESTAMPTZ NOT NULL CHECK (period_start <= period_end),
    period_end   TIMESTAMPTZ NOT NULL,
    status       TEXT        NOT NULL DEFAULT 'active',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_subscriptions_created_at ON subscriptions (created_at DESC);
CREATE INDEX idx_subscriptions_user_id    ON subscriptions (user_id);

-- +goose Down
DROP TABLE IF EXISTS subscriptions;
