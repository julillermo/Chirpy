-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (id, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES(
  $1,
  NOW(),
  NOW(),
  $2,
  $3,
  NULL
)
RETURNING *;

-- name: UpdateRefreshTokenById :one
UPDATE refresh_tokens
SET
  updated_at = NOW(),
  expires_at = COALESCE(sqlc.narg('expires_at'), expires_at),
  revoked_at = COALESCE(sqlc.narg('revoked_at'), revoked_at) 
WHERE 
  id=sqlc.arg('id')
  AND revoked_at IS NULL
RETURNING *;

-- name: ValidateRefreshToken :one
SELECT EXISTS (
    SELECT TRUE
    FROM refresh_tokens
    WHERE id = $1
      AND expires_at >= NOW()
      AND revoked_at IS NULL
) AS is_valid;

-- name: GetRefreshTokenById :one
SELECT * FROM refresh_tokens WHERE id=$1;