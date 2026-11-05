SELECT COUNT(*)
FROM users
WHERE ($1::timestamptz IS NULL OR created_at >= $1::timestamptz)
