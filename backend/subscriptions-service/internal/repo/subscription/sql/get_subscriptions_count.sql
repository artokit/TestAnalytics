-- Общее количество записей для пагинации (/subscriptions total, total_pages).
SELECT COUNT(*)
FROM subscriptions
WHERE ($1::timestamptz IS NULL OR updated_at >= $1::timestamptz)
