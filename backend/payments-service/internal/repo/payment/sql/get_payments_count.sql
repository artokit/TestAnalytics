SELECT COUNT(*)
FROM payments
WHERE ($1::timestamptz IS NULL OR updated_at >= $1::timestamptz)
