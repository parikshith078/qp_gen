// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: users.sql

package sqlc

import (
	"context"

	"github.com/google/uuid"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (
  name, email, username, password_hash, created_at, updated_at, last_activity
) VALUES (
  $1, $2, $3, $4, NOW(), NOW(), NOW()
)
RETURNING id, name, email, username, password_hash, created_at, updated_at, last_activity
`

type CreateUserParams struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.Name,
		arg.Email,
		arg.Username,
		arg.PasswordHash,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.Username,
		&i.PasswordHash,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.LastActivity,
	)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, name, email, username, password_hash, created_at, updated_at, last_activity FROM users
WHERE email = $1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.Username,
		&i.PasswordHash,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.LastActivity,
	)
	return i, err
}

const getUserByID = `-- name: GetUserByID :one
SELECT id, name, email, username, password_hash, created_at, updated_at, last_activity FROM users
WHERE id = $1
`

func (q *Queries) GetUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	row := q.db.QueryRow(ctx, getUserByID, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.Username,
		&i.PasswordHash,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.LastActivity,
	)
	return i, err
}

const getUserByUsername = `-- name: GetUserByUsername :one
SELECT id, name, email, username, password_hash, created_at, updated_at, last_activity FROM users
WHERE username = $1
`

func (q *Queries) GetUserByUsername(ctx context.Context, username string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByUsername, username)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.Username,
		&i.PasswordHash,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.LastActivity,
	)
	return i, err
}

const updateUserByID = `-- name: UpdateUserByID :one
UPDATE users
SET 
    name = COALESCE($2, name),
    username = COALESCE($3, username),
    email = COALESCE($4, email),
    password_hash = COALESCE($5, password_hash),
    last_activity = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, name, email, username, password_hash, created_at, updated_at, last_activity
`

type UpdateUserByIDParams struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
}

func (q *Queries) UpdateUserByID(ctx context.Context, arg UpdateUserByIDParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUserByID,
		arg.ID,
		arg.Name,
		arg.Username,
		arg.Email,
		arg.PasswordHash,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.Username,
		&i.PasswordHash,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.LastActivity,
	)
	return i, err
}

const updateUserLastActivity = `-- name: UpdateUserLastActivity :exec
UPDATE users
  SET   last_activity = CURRENT_TIMESTAMP
WHERE id = $1
`

func (q *Queries) UpdateUserLastActivity(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, updateUserLastActivity, id)
	return err
}
