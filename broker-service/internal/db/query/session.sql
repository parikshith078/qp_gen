-- name: CreateSessionToken :one
INSERT INTO session_tokens (
  user_id, token, expires_at, created_at
) VALUES (
  $1, $2, $3, NOW() 
)
RETURNING *;

-- name: CreateCsrfToken :one
INSERT INTO csrf_tokens (
  session_id, token, expires_at, created_at
) VALUES (
  $1, $2, $3, NOW() 
)
RETURNING *;

-- name: GetSessionTokenByToken :one
SELECT * FROM session_tokens
WHERE token = $1;

-- name: GetCsrfTokenBySessionID :one
SELECT * FROM csrf_tokens
WHERE session_id = $1;