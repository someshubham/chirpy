-- name: GetUserIdFromRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1 LIMIT 1;