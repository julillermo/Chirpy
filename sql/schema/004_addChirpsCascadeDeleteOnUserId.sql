-- +goose Up
ALTER TABLE chirps
DROP CONSTRAINT IF EXISTS chirps_user_id_fkey;

ALTER TABLE chirps
ADD
  CONSTRAINT chirps_user_id_fkey
  FOREIGN KEY (user_id)
  REFERENCES users(id)
  ON DELETE CASCADE;

-- +goose Down
ALTER TABLE chirps
DROP CONSTRAINT IF EXISTS chirps_user_id_fkey;

ALTER TABLE chirps
ADD
  CONSTRAINT chirps_user_id_fkey
  FOREIGN KEY (user_id)
  REFERENCES users(id);