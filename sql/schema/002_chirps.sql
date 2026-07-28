-- +goose Up
CREATE TABLE chirps (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  body TEXT NOT NULL,
  user_id UUID REFERENCES users(id)
  
  /* The above is equivalent to this
  - Note that the sql program would automatically decide
    the constraint name
  - Note that the type still has to match what it's referencing
  */
  -- user_id UUID,
  -- CONSTRAINT fk_users
  --   FOREIGN KEY (user_id)
  --   REFERENCES users(id)
);

-- +goose Down
DROP TABLE IF EXISTS chirps;
