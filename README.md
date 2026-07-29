# Chirpy

boot.dev backend in go course

# How to setup connection to the databse

0. Prerequisites
   - Install postgresql (or your prefered database)
     - Take note of your "connection string" that follows this format: `postgres://username:password@host:port/database`
       - Note that note all users have the same access to the databse, so some of the commands in this guide may fail when using a specific user with less permissions.
     - Save the "connection string" as `DB_URL=<value>`.
     - The specific variable `DB_URL` can be changed as long as it's consistent within the application.
       - Access within the application code via `dbUrl := os.GetEnv("DB_URL")`
   - Install the following as system packages `go install ...`:
     - Goose - `github.com/pressly/goose/v3/cmd/goose@latest`
     - SQLC - `github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
   - Install the following as dependencies of the project `go get ...`:
     - dot env for go `github.com/joho/godotenv`
     - google's uuid package `github.com/google/uuid`
       - This mainly used for parsing uuid's and not the one responsible for generating the uuid's in the database
       - SQLC generated code depends on this.
     - postgresql driver for go `github.com/lib/pq`
       - Must be imported at the root of the project as `import _ "github.com/lib/pq"`
       - Create an SQL connection within the app code via `db, err := sql.Open("postgres", dbUrl)`
   - Setup `sqlc` config [link](https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html) at the root of the project as `sqlc.yaml`.
     - Generally appears like the following (from [boot.dev's example](https://www.boot.dev/lessons/e5bddf3d-d96b-487e-97e6-7a5aa06b1ee1)).
     - Sqlc can target other programming languages as well, not just go.
     ```yaml
     version: "2"
     sql:
       - schema: "sql/schema"
         queries: "sql/queries"
         engine: "postgresql"
         gen:
           go:
             out: "internal/database"
     ```

1. Create an SQL Query `.sql` migration file (optional)
   - This is only needed when you want to make modifications to the database setup (e.g. adding a table or modifying columns, etc.).
     - For this project, the file is saved in `/sql/schema/`, but this isn't scrict.
   - Make sure to following the naming format: `<migration#>_<table_name>.sql` (e.g. `001_users.sql`)
     - Prepending with 0's is optional; `001` is equivalent to both `1` and `0001`.
     - Add (required) comments to the `.sql` migration file to indicate up and down migrations:

     ```sql
     -- 001_users.sql --

     -- +goose Up
     CREATE TABLE users (
       id UUID PRIMARY KEY,
       created_at TIMESTAMP,
       updated_at TIMESTAMP,
       email TEXT NOT NULL
     );

     -- +goose Down
     DROP TABLE IF EXISTS users;
     ```

2. Running the migration (optional)
   - **Within the where the migration files are saved** (`/sql/schema/` for this project), run:
     - `goose postgres <connection_string> up` to apply or
     - `goose postgres <connection_string> down` to revert
3. Use sqlc to generate the application code equivalent of the queries.
   - Write the query as a `.sql` file.
     - For this project, we're saving to `/sql/queries`, but it could be anything as long as state it in your `sqlc.yaml`
     - Should generally be similar to the following syntax. Consult the [documentation](https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html) for more specicic implementations.
       - Note that the comments are required.
       - I believe you can have more than 1 query in 1 file.

     ```sql
     -- name: CreateUser :one
     INSERT INTO users (id, created_at, updated_at, email)
     VALUES (
       gen_random_uuid(),
       NOW(),
       NOW(),
       $1
     )
     RETURNING *;
     ```

   - At the root of the project (where `sqlc.yaml` should also exist) run:
     - `sqlc generate`

4. Use the generated go query functions as part of your code.
   - You may need to transform your variables to align with the types of the functions.
   - You also need to have setup the db connection prior.
   - Refer to example in [main.go](/main.go) as a guide
