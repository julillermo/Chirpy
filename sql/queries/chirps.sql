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
SELECT *
FROM chirps
WHERE sqlc.narg('user_id')::uuid IS NULL
  OR user_id = sqlc.narg('user_id')::uuid
ORDER BY
  CASE
    WHEN sqlc.narg('sort')::text = 'desc'
    THEN created_at
  END DESC,
  CASE
    WHEN sqlc.narg('sort')::text = 'asc'
    THEN created_at
  END ASC;

-- name: GetChirp :one
SELECT * FROM chirps WHERE id=$1;

-- -- name: GetAllChirpsByUserId :many
-- SELECT * FROM chirps WHERE user_id=$1;

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