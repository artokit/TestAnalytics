SELECT id,
       user_id,
       plan_id,
       period_start,
       period_end,
       status,
       created_at,
       updated_at
FROM subscriptions
WHERE ($3::timestamptz IS NULL OR updated_at >= $3::timestamptz)
ORDER BY updated_at DESC, id
LIMIT $1::bigint OFFSET $2::bigint;
