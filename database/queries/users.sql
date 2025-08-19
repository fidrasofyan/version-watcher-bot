-- name: GetUser :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: IsUserExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 LIMIT 1);

-- name: CreateUser :one
INSERT INTO users (id, username, first_name, last_name, created_at) 
VALUES ($1, $2, $3, $4, $5) 
RETURNING *;