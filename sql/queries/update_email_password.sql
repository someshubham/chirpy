-- name: UpdateUserEmailPassword :one
UPDATE users 
SET hashed_password = $3, email = $2
WHERE id = $1
RETURNING *;
