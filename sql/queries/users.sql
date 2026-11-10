-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
  $1,
  $2,
  $3,
  $4
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE name = $1 LIMIT 1;

-- name: DropTable :exec
DELETE FROM users;

-- name: GetUsers :many
SELECT * FROM users;

-- name: FindUserByID :one
SELECT users.name FROM users INNER JOIN feeds ON users.id = feeds.user_id
WHERE users.id = $1 LIMIT 1;
