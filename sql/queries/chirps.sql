-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
  gen_random_uuid(),
  NOW(),
  NOW(),
  $1,
  $2
)
RETURNING *;

-- name: GetAllChirps :many
SELECT * FROM chirps;

-- name: GetChirp :one
SELECT * FROM chirps WHERE id=$1;

-- name: DeleteAllChirps :exec
DELETE FROM chirps;

-- name: DeleteUserChirpById :execresult
DELETE FROM chirps
WHERE 
  id=$1
  AND user_id=$2;

-- name: DeleteChirpById :execresult
DELETE FROM chirps
WHERE 
  id=$1;