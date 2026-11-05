SELECT id,
       telegram_id,
       created_at,
       utm_link
FROM users
WHERE ($3::timestamptz IS NULL OR created_at >= $3::timestamptz)
ORDER BY id
LIMIT $1::bigint OFFSET $2::bigint;
