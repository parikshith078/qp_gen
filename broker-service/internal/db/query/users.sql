-- name: CreateUser :one
INSERT INTO users (
  name, email, username, password_hash, created_at, updated_at, last_activity
) VALUES (
  $1, $2, $3, $4, NOW(), NOW(), NOW()
)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1;

-- name: UpdateUserLastActivity :exec
UPDATE users
  SET   last_activity = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: UpdateUserByID :one
UPDATE users
SET 
    name = COALESCE($2, name),
    username = COALESCE($3, username),
    email = COALESCE($4, email),
    password_hash = COALESCE($5, password_hash),
    last_activity = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

