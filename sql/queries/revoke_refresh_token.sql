-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW(),
    updated_at = NOW()
WHERE user_id = $1;