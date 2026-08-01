-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
  gen_random_uuid(),
  NOW(),
  NOW(),
  $1,
  $2
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email=$1;

-- name: GetUserById :one
SELECT * FROM users WHERE id=$1;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: UpdateUserById :one
UPDATE users
SET
  updated_at = NOW(),
  email = COALESCE(sqlc.narg('email'), email),
  hashed_password = COALESCE(sqlc.narg('hashed_password'), hashed_password)
WHERE
  id=sqlc.arg('id')
RETURNING *;

-- name: UpgradeUserToRedById :one
UPDATE users
SET
  updated_at = NOW(),
  is_chirpy_red = TRUE
WHERE
  id=$1
RETURNING *;
