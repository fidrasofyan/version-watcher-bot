-- name: IsChatExists :one
SELECT EXISTS(SELECT 1 FROM chats WHERE id = $1 LIMIT 1);

-- name: GetChat :one
SELECT * FROM chats WHERE id = $1 LIMIT 1;

-- name: CreateChat :one
INSERT INTO chats (id, command, step, data, created_at) 
VALUES ($1, $2, $3, $4, $5) 
RETURNING *;

-- name: UpdateChat :one
UPDATE chats
SET command = $1, step = $2, data = $3, updated_at = $4
WHERE id = $5
RETURNING *;

-- name: DeleteChat :exec
DELETE FROM chats WHERE id = $1;