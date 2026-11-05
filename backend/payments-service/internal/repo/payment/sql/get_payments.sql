SELECT id,
       user_id,
       subscription_id,
       amount,
       currency,
       status,
       period_start,
       period_end,
       created_at,
       updated_at
FROM payments
WHERE ($1::timestamptz IS NULL OR updated_at >= $1::timestamptz)
ORDER BY updated_at DESC, id
LIMIT $2::bigint OFFSET $3::bigint;
