-- name: Add :exec
INSERT INTO users (username, pass)
VALUES (?, ?);

-- name: GetByTelegramName :one
SELECT username
FROM users
WHERE username = ?
LIMIT 1;

-- name: GetPassByUsername :one
SELECT pass
FROM users
WHERE username = ?
LIMIT 1;